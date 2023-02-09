
# Example Otel+Jaeger in Python


service to internal service via network example

```
docker-compose up

go run main.go httpA

cd python


OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces python3 simpleFlask.py

curl http://localhost:5000

```

open jaeger UI: http://localhost:16686