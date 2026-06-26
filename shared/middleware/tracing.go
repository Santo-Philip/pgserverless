package middleware

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/gofiber/fiber/v2"
)

var tracer trace.Tracer

func InitTracing(ctx context.Context, serviceName, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	if otlpEndpoint == "" {
		tracer = trace.NewNoopTracerProvider().Tracer(serviceName)
		return nil, nil
	}

	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(otlpEndpoint),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp.Tracer(serviceName)
	return tp, nil
}

func TracingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		headers := make(map[string][]string)
		for k, vals := range c.GetReqHeaders() {
			headers[k] = vals
		}
		ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(headers))
		ctx, span := tracer.Start(ctx, c.Method()+" "+c.Path(),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.Path()),
				attribute.String("http.host", c.Hostname()),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)
		err := c.Next()

		status := c.Response().StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", status))
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

func Tracer() trace.Tracer {
	if tracer == nil {
		return trace.NewNoopTracerProvider().Tracer("default")
	}
	return tracer
}
