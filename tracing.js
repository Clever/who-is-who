// adapted from:
// https://github.com/open-telemetry/opentelemetry-js/blob/main/getting-started/ts-example/README.md
// https://github.com/open-telemetry/opentelemetry-js/tree/main/packages/opentelemetry-exporter-collector-grpc
const { NodeTracerProvider } = require("@opentelemetry/sdk-trace-node");
const { registerInstrumentations } = require("@opentelemetry/instrumentation");
const { ExpressInstrumentation } = require("@opentelemetry/instrumentation-express");
const { HttpInstrumentation } = require("@opentelemetry/instrumentation-http");
const { SimpleSpanProcessor } = require("@opentelemetry/sdk-trace-base");
const { OTLPTraceExporter } = require("@opentelemetry/exporter-trace-otlp-grpc");

const provider = new NodeTracerProvider({});
const exporter = new OTLPTraceExporter();
provider.addSpanProcessor(new SimpleSpanProcessor(exporter));

registerInstrumentations({
  tracerProvider: provider,
  instrumentations: [new ExpressInstrumentation(), new HttpInstrumentation()],
});

provider.register();
