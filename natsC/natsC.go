package natsC

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	"jeager1/natsC/submoduleC"

	"github.com/kokizzu/gotro/L"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type NatsC struct{}

const otelTraceSpanIdPair = `otelTraceSpanIdPair`

func (n *NatsC) StartServer(environment, serviceName, version string) {
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

	// will propagate trace-ID to next request properly, not used since we send directly parent context as header
	//otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	opts := &server.Options{}

	// Initialize new server with options
	ns, err := server.NewServer(opts)

	if err != nil {
		panic(err)
	}

	// Start the server via goroutine
	go ns.Start()

	// Wait for server to be ready for connections
	if !ns.ReadyForConnections(4 * time.Second) {
		panic("not ready for connection")
	}

	// Connect to server
	nc, err := nats.Connect(ns.ClientURL())
	L.PanicIf(err, `nats.Connect`)

	// check this example for tracing (outdated, no longer works)
	// https://github.com/nats-io/not.go

	// check this for nats embedding
	// https://dev.to/karanpratapsingh/embedding-nats-in-go-19o

	const topic1 = "my-topic1"
	const topic2 = "my-topic2"

	// Subscribe to the topic1
	_, err = nc.QueueSubscribe(topic1, "my-queue", func(msg *nats.Msg) {
		ctx, span := otel.Tracer(`natsC`).Start(createRemoteCtxFromHeader(topic1, msg), topic1,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		submoduleC.SomeFuncC(ctx)

		data := string(msg.Data)
		fmt.Println(topic1, data)

		reply := sendMessage(nc, topic2, `something`, ctx)

		err = msg.Respond([]byte(data + reply))
		L.IsError(err, `msg.Respond`) // ignore error
	})

	_, err = nc.QueueSubscribe(topic2, "my-queue", func(msg *nats.Msg) {
		ctx, span := otel.Tracer(`natsC`).Start(createRemoteCtxFromHeader(topic2, msg), topic2,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		submoduleC.SomeFuncC(ctx)

		// Print message data
		data := string(msg.Data)
		err = msg.Respond([]byte(data))
		L.IsError(err, `msg.Respond`) // ignore error
	})

	// Publish data to the topic1
	go sendMessage(nc, topic1, `whatever`, context.Background())

	// Wait for server shutdown
	ns.WaitForShutdown()
}

func sendMessage(nc *nats.Conn, topic string, payload string, ctx context.Context) string {
	_, span := otel.Tracer("natsC").Start(ctx, "publish")
	defer span.End()
	spanCtx := span.SpanContext()
	traceId := spanCtx.TraceID().String()
	spanId := spanCtx.SpanID().String()
	log.Println(traceId, spanId)
	msg, err := nc.RequestMsg(&nats.Msg{
		Subject: topic, Data: []byte(payload), Header: nats.Header{
			otelTraceSpanIdPair: []string{
				traceId,
				spanId,
			},
		},
	}, 2*time.Second)
	if L.IsError(err, `nc.Publish`) {
		return ""
	}
	log.Println(`reply:`, topic, string(msg.Data))
	return string(msg.Data)
}

func createRemoteCtxFromHeader(topic string, msg *nats.Msg) (spanContext context.Context) {
	traceSpanIdPair := msg.Header.Values(otelTraceSpanIdPair)
	log.Println(topic, traceSpanIdPair)
	if len(traceSpanIdPair) != 2 {
		spanContext = context.Background()
		return
	}

	traceID, err := trace.TraceIDFromHex(traceSpanIdPair[0])
	L.IsError(err, `TraceIDFromHex`)
	var spanID trace.SpanID
	spanID, err = trace.SpanIDFromHex(traceSpanIdPair[1])
	L.IsError(err, `SpanIDFromHex`)

	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = true
	parentSpanCtx := trace.NewSpanContext(spanContextConfig)
	spanContext = trace.ContextWithRemoteSpanContext(context.Background(), parentSpanCtx)
	return
}
