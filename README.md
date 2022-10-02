
# jeager1

This is example how to use jeager to trace between services.
Since most of the examples on the internet already outdated (jeager client and opentracing deprecated).

```shell
docker-compose up -d
go run main.go
curl -v localhost:3000

```

open [localhost:16686](http://localhost:16686) to see the trace.

## TODO

- grpc
- nats
