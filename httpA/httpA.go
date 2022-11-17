package httpA

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kokizzu/gotro/L"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"jeager1/httpA/submoduleA"
)

type HttpA struct{}

func (h *HttpA) StartServer(environment, serviceName, version string) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	L.PanicIf(err, `jeager.New`)
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

	mux.Handle("/test", otelhttp.WithRouteTag("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := otel.Tracer("httpA").Start(ctx, "POST /test")
			defer span.End()

			time.Sleep(6 * time.Millisecond)
			submoduleA.SomeFuncA(ctx)
			time.Sleep(4 * time.Millisecond)

			// r.Header.Get(`traceparent`) // will get traceparent from previous request

			_, _ = w.Write([]byte(`post /test happened`))
			w.WriteHeader(http.StatusOK)
		})))

	log.Fatal(http.ListenAndServe(":3000", otelhttp.NewHandler(&mux, "server",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)))
}
