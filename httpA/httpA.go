package httpA

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kokizzu/gotro/L"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"jeager1/grpcB"
	"jeager1/httpA/submoduleA"
)

type HttpA struct{}

func (h *HttpA) StartServer(environment, serviceName, version string) {
	//exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	//L.PanicIf(err, `jeager.New`)

	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(`localhost:4318`))
	L.PanicIf(err, `otlptracehttp.New`)

	// only from go 1.18 -buildvcs
	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSampler(tracesdk.AlwaysSample()), // use ProbabilitySampler on production
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String(environment),
			semconv.ServiceVersionKey.String(version),
		)),
	)
	otel.SetTracerProvider(tracerProvider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			L.PanicIf(err, `tracerProvider.Shutdown`)
		}
	}(ctx)

	// will propagate trace-ID to next request properly
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// tracer := tracerProvider.Tracer("httpA") // equal to otel.Tracer("httpA")

	var mux http.ServeMux
	mux.Handle("/", otelhttp.WithRouteTag("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("httpA").Start(ctx, "GET /")
		defer span.End()

		time.Sleep(6 * time.Millisecond)
		submoduleA.SomeFuncA(ctx)
		time.Sleep(4 * time.Millisecond)

		// try post to lower handler (assuming this is another http service)
		client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
		req, err := http.NewRequestWithContext(ctx, "POST", `http://localhost:3000/test`, strings.NewReader(`{"key":"value"}`))
		if L.IsError(err, `http.NewRequest`) {
			span.SetAttributes(attribute.Key("error").String(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res, err := client.Do(req)
		if L.IsError(err, `client.Do`) {
			span.SetAttributes(attribute.Key("error").String(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(res.Body)
		if L.IsError(err, `io.ReadAll`) {
			span.SetAttributes(attribute.Key("error").String(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		span.SetAttributes(attribute.Key("customPost.response").String(string(body)))

		_, _ = w.Write([]byte(`test 123`))
		w.WriteHeader(http.StatusOK)
	})))

	mux.Handle("/test", otelhttp.WithRouteTag("/test", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := otel.Tracer("httpA").Start(ctx, "POST /test")
			defer span.End()

			time.Sleep(6 * time.Millisecond)
			submoduleA.SomeFuncA(ctx)
			time.Sleep(4 * time.Millisecond)

			// r.Header.Get(`traceparent`) // will get traceparent from previous request

			traceparent := r.Header.Get(`traceparent`) // will get traceparent from previous request
			tracestate := r.Header.Get(`tracestate`)
			ourTrace, _ := span.SpanContext().MarshalJSON()
			L.Describe(traceparent, tracestate, string(ourTrace))

			_, _ = w.Write([]byte(`post /test happened`))
			w.WriteHeader(http.StatusOK)
		})))

	mux.Handle(`/try-http2grpc`, otelhttp.WithRouteTag("/try-http2grpc", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := otel.Tracer("httpA").Start(ctx, "GET /try-http2grpc")
			defer span.End()

			conn, err := grpc.Dial(`127.0.0.1:3001`,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
			)
			if L.IsError(err, `grpc.Dial`) {
				return
			}

			grpcClient := grpcB.NewGrpcBClient(conn)
			res, err := grpcClient.PostAnything(ctx, &grpcB.PostAnythingRequest{
				Name: &[]string{`http2grpc`}[0],
			})
			if L.IsError(err, `grpcClient.PostAnything`) {
				return
			}

			// how to test this:
			/*
				go run main.go httpA
				go run main.go grpcB
				curl -v localhost:3000/try-http2grpc
			*/

			L.Print(`http2grpcB.PostAnything: `, res.GetValue())
			_, _ = w.Write([]byte(`http2grpcB.PostAnything: ` + res.GetValue()))
		})))

	mux.Handle("/traceparent-check", otelhttp.WithRouteTag("/traceparent-check", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := otel.Tracer("httpA").Start(ctx, "GET /traceparent-check")
			defer span.End()

			time.Sleep(5 * time.Millisecond)
			traceparent := r.Header.Get(`traceparent`) // will get traceparent from previous request

			ourTrace, _ := span.SpanContext().MarshalJSON()

			client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
			req, err := http.NewRequestWithContext(ctx, "POST", `http://localhost:3000/test`, strings.NewReader(`{"key":"value"}`))
			L.IsError(err, `http.NewRequest`)
			_, err = client.Do(req)
			L.IsError(err, `client.Do`)

			_, _ = w.Write([]byte(`post /test happened with ` + traceparent + ` and ` + string(ourTrace)))
			w.WriteHeader(http.StatusOK)

			// how to check:
			/*
				go run main.go httpA

				curl http://localhost:3000/traceparent-check \
				-H 'Traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01'

				// received properly
				post /test happened with 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01 and {"TraceID":"4db6e89b9d0f5cce699ab3bee9022e60","SpanID":"8aaba233dfa73840","TraceFlags":"01","TraceState":"","Remote":false}

				// log from /test
				"00-4db6e89b9d0f5cce699ab3bee9022e60-735403388675709d-01"
				"{"TraceID":"9a8561c47f92b1fdaa4977d4dc9291d3","SpanID":"fd3cdf685296b62b","TraceFlags":"01","TraceState":"","Remote":false}"
			*/
		})))

	log.Fatal(http.ListenAndServe(":3000", otelhttp.NewHandler(&mux, "server",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)))
}
