package main_test

import (
	"testing"

	"jeager1/httpA"
)

func TestMain(m *testing.M) {
	//D.InitDockerTest(``)

	// TODO: continue this

	httpA := httpA.HttpA{}
	go httpA.StartServer("development", "httpA", "1.0.0")
	{

	}

	//  go grpcB.Run() {
	//
	//}
	//
	//go natsC.Run() {
	//
	//}
}
