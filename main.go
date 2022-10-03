package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/kokizzu/gotro/S"
	"jeager1/grpcB"
	"jeager1/httpA"
	"jeager1/natsC"
)

func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return S.TrimChars(info.Deps[len(info.Deps)-1].Version, `()`) // importing binary always on the last of Deps
	}
	return `unknown`
}

var DEPLOY_ENV = `development`

func main() {
	if len(os.Args) <= 1 {
		fmt.Println(`usage:
` + os.Args[0] + ` httpA|grpcB|natsC`)
		return
	}

	// https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/jaeger
	// environment variables used:
	/*
		OTEL_EXPORTER_JAEGER_AGENT_HOST 	WithAgentHost 	localhost
		OTEL_EXPORTER_JAEGER_AGENT_PORT 	WithAgentPort 	6831
		OTEL_EXPORTER_JAEGER_ENDPOINT 	WithEndpoint 	http://localhost:14268/api/traces
		OTEL_EXPORTER_JAEGER_USER 	WithUsername
		OTEL_EXPORTER_JAEGER_PASSWORD 	WithPassword
	*/

	mode := os.Args[1]
	switch mode {
	case `httpA`:
		server := httpA.HttpA{}
		server.StartServer(DEPLOY_ENV, mode, GetVersion())
	case `grpcB`:
		server := grpcB.GrpcB{}
		server.StartServer(DEPLOY_ENV, mode, GetVersion())
	case `natsC`:
		server := natsC.NatsC{}
		server.StartServer(DEPLOY_ENV, mode, GetVersion())
	}
}
