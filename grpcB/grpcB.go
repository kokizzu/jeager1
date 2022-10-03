package grpcB

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/kokizzu/gotro/L"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"jeager1/grpcB/submoduleB"
)

type GrpcB struct {
	UnimplementedGrpcBServer
}

func (g *GrpcB) PostAnything(ctx context.Context, request *PostAnythingRequest) (*PostAnythingResponse, error) {
	ctx, span := otel.Tracer("grpcB").Start(ctx, "PostAnything")
	defer span.End()

	submoduleB.SomeFuncA(ctx)

	return &PostAnythingResponse{
		Value: request.Name,
	}, nil
}

func (g *GrpcB) GetSomething(ctx context.Context, request *GetSomethingRequest) (*GetSomethingResponse, error) {
	ctx, span := otel.Tracer("grpcB").Start(ctx, "GetSomething")
	defer span.End()

	submoduleB.SomeFuncA(ctx)

	conn, err := grpc.Dial(`127.0.0.1:3001`,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if L.IsError(err, `grpc.Dial`) {
		return nil, err
	}

	grpcClient := NewGrpcBClient(conn)
	res, err := grpcClient.PostAnything(ctx, &PostAnythingRequest{
		Name: request.Name,
	})
	if L.IsError(err, `grpcClient.PostAnything`) {
		return nil, err
	}

	ret := `returned from post: ` + res.GetValue()
	return &GetSomethingResponse{
		Value: &ret,
	}, nil
}

func (g *GrpcB) StartServer(environment, serviceName, version string) {
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

	server := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	exampleService := &GrpcB{}

	RegisterGrpcBServer(server, exampleService)
	reflection.Register(server)
	lis, err := net.Listen("tcp", `127.0.0.1:3001`)
	L.PanicIf(err, `net.Listen`)

	log.Fatal(server.Serve(lis))
}
