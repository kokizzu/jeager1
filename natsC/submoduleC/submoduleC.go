package submoduleC

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func SomeFuncC(ctx context.Context) {
	// otel.Tracer will use tracer that set last from otel.SetTracerProvider
	ctx, span := otel.Tracer("submoduleC").Start(ctx, "SomeFuncC")

	// any attribute, eg. total records after query, response/code from 3rdParty
	span.SetAttributes(attribute.Key("key3").String("value3"))
	// TODO: do some process here
	time.Sleep(10 * time.Millisecond)

	span.AddEvent("event3") // any event, eg. branching/return

	defer span.End()
}
