
# jeager1

This is example how to use jeager to trace between services.
Since most of the examples on the internet already outdated (jeager client and opentracing deprecated).

```shell
docker-compose up -d

# http example
go run main.go httpA
curl -v localhost:3000

# grpc example
go run main.go grpcB
grpcurl -plaintext 127.0.0.1:3001 list
grpcurl -plaintext -d '{"name":"BBB"}' 127.0.0.1:3001 GrpcB.GetSomething 
```

open [localhost:16686](http://localhost:16686) to see the trace.

![image](https://user-images.githubusercontent.com/1061610/193477550-a8e1b58e-1f5f-46c3-bcb9-0f866c05c15f.png)

![image](https://user-images.githubusercontent.com/1061610/193554547-f3f931e9-35ef-481f-8d80-0769175f289e.png)

## TODO

- grpc
- nats
