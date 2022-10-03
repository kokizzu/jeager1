package natsC

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kokizzu/gotro/L"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type natsC struct{}

func (n *natsC) StartServer(environment, serviceName, version string) {
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

	// TODO: check this for tracing
	// https://github.com/nats-io/not.go

	// TODO: check this for embedding
	// https://dev.to/karanpratapsingh/embedding-nats-in-go-19o

	subject := "my-subject"

	// Subscribe to the subject
	_, err = nc.QueueSubscribe(subject, "my-queue", func(msg *nats.Msg) {
		// Print message data
		data := string(msg.Data)
		fmt.Println(data)
		err := msg.Respond(msg.Data)
		L.IsError(err, `msg.Respond`) // ignore error
	})

	// Publish data to the subject
	msg, err := nc.Request(subject, []byte("Hello embedded NATS!"), 2*time.Second)
	if L.IsError(err, `nc.Publish`) {
		return
	}
	log.Println(`reply:`, msg)

	// Wait for server shutdown
	ns.WaitForShutdown()
}
