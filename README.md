
# jeager1

This is example how to use jeager to trace between services.
Since most of the examples on the internet already outdated (jeager client and opentracing deprecated).

```shell
docker-compose up -d
go run main.go httpA
curl -v localhost:3000
```

open [localhost:16686](http://localhost:16686) to see the trace.

![image](https://user-images.githubusercontent.com/1061610/193477550-a8e1b58e-1f5f-46c3-bcb9-0f866c05c15f.png)

## TODO

- grpc
- nats
