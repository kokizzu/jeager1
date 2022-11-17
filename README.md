
# jeager1

This is example how to use jaeger (typo '__') to trace between services.
Since most of the examples on the internet already outdated (deprecated jaager client or deprecated opentracing library).

```shell
docker-compose up -d

# http example
go run main.go httpA
curl -v localhost:3000

# grpc example
go run main.go grpcB
grpcurl -plaintext 127.0.0.1:3001 list
grpcurl -plaintext -d '{"name":"BBB"}' 127.0.0.1:3001 GrpcB.GetSomething 

# nats example, will publish a message when server start
go run main.go natsC 
```

open [localhost:16686](http://localhost:16686) to see the trace.

![image](https://user-images.githubusercontent.com/1061610/193477550-a8e1b58e-1f5f-46c3-bcb9-0f866c05c15f.png)

![image](https://user-images.githubusercontent.com/1061610/193554547-f3f931e9-35ef-481f-8d80-0769175f289e.png)

![image](https://user-images.githubusercontent.com/1061610/202473625-522ced0e-ec91-4882-a42c-0649d82fd49d.png)

## TODO

- log all request and response payload


## Special Thanks

- [thetooth](https://github.com/thetooth) - fixing NATS traceID/spanID problem
