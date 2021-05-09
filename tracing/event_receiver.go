package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// EventReceiver provides an embeddable implementation of tyr.TracingEventReceiver
// powered by opentracing-go.
type EventReceiver struct{}

// SpanStart starts a new query span from ctx, then returns a new context with the new span.
func (EventReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	span, ctx := opentracing.StartSpanFromContext(ctx, eventName)
	ext.DBStatement.Set(span, query)
	ext.DBType.Set(span, "sql")
	return ctx
}

// SpanFinish finishes the span associated with ctx.
func (EventReceiver) SpanFinish(ctx context.Context) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.Finish()
	}
}

// SpanError adds an error to the span associated with ctx.
func (EventReceiver) SpanError(ctx context.Context, err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.String("event", "error"), log.Error(err))
	}
}
