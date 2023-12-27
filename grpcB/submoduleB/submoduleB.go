package submoduleB

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func SomeFuncA(ctx context.Context) {
	// otel.Tracer will use tracer that set last from otel.SetTracerProvider
	ctx, span := otel.Tracer("submoduleB").Start(ctx, "SomeFuncA")
	_ = ctx

	// any attribute, eg. total records after query, response/code from 3rdParty
	span.SetAttributes(attribute.Key("key2").String("value2"))
	// TODO: do some process here
	time.Sleep(10 * time.Millisecond)

	span.AddEvent("event2") // any event, eg. branching/return

	defer span.End()
}
