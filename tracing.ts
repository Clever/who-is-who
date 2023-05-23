// adapted from:
// https://github.com/open-telemetry/opentelemetry-js/blob/main/getting-started/ts-example/README.md
// https://github.com/open-telemetry/opentelemetry-js/tree/main/packages/opentelemetry-exporter-collector-grpc
import { NodeTracerProvider } from "@opentelemetry/sdk-trace-node";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { ExpressInstrumentation } from "@opentelemetry/instrumentation-express";
import { HttpInstrumentation } from "@opentelemetry/instrumentation-http";
import { SimpleSpanProcessor } from "@opentelemetry/sdk-trace-base";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-grpc";

const provider: NodeTracerProvider = new NodeTracerProvider({});
const exporter = new OTLPTraceExporter();
provider.addSpanProcessor(new SimpleSpanProcessor(exporter));

registerInstrumentations({
  tracerProvider: provider,
  instrumentations: [new ExpressInstrumentation(), new HttpInstrumentation()],
});

provider.register();
