
# Example Otel+Jaeger in Python+Flask


service to internal service via network example

```
docker-compose up

go run main.go httpA

cd python
OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces python3 simpleFlask.py

curl http://localhost:5000

```

open jaeger UI: http://localhost:16686

![image](https://user-images.githubusercontent.com/1061610/217779376-c9531e08-1d37-493e-b321-8f05e142f1e9.png)
