from flask import Flask
from flask import request
from requests import get
from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.propagate import inject
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

trace.set_tracer_provider(
TracerProvider(
        resource=Resource.create({SERVICE_NAME: "my-helloworld-service"})
    )
)
tracer = trace.get_tracer(__name__)


# create a JaegerExporter
jaeger_exporter = JaegerExporter(
    # configure agent
    agent_host_name='localhost',
    agent_port=6831,
    # optional: configure also collector
    # collector_endpoint='http://localhost:14268/api/traces?format=jaeger.thrift',
    # username=xxxx, # optional
    # password=xxxx, # optional
    # max_tag_value_length=None # optional
)

# Create a BatchSpanProcessor and add the exporter to it
span_processor = BatchSpanProcessor(jaeger_exporter)


# add to the tracer
trace.get_tracer_provider().add_span_processor(span_processor)


app = Flask(__name__)

def get_carrier_traceparent(request):
    tp = request.headers.get_all('traceparent')
    if tp:
        return {"traceparent": tp[0]}
    return {}

# assuming this is another service
@app.route("/another_route")
def another_route():
    carrier = get_carrier_traceparent(request)
    ctx = TraceContextTextMapPropagator().extract(carrier)
    with tracer.start_as_current_span("another_route", context=ctx):
        current_span = trace.get_current_span()
        current_span.set_attribute("testing", 1)
        return "from another route"

# assuming this is the frontend service being hit
@app.route("/")
def root_route():
    with tracer.start_as_current_span("root_route"):
        # from: https://lightstep.com/blog/opentelemetry-for-python-the-hard-way
        carrier = {}
        TraceContextTextMapPropagator().inject(carrier)
        headers = {"traceparent": carrier["traceparent"]}
        inject(headers)
        requested = get(
            "http://localhost:5000/another_route",
            params={"param": 'foo'},
            headers=headers,
        )
        return "Hello World! " + requested.text


if __name__ == "__main__":
    app.run()
