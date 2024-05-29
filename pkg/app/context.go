package app

import (
	"context"
)

const (
	CtxKeyLogAttrs = ctxKey(iota + 1)
	CtxKeyXRequestId
)

type ctxKey int

func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, CtxKeyXRequestId, reqID)
}

func RequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(CtxKeyXRequestId).(string); ok {
		return reqID
	}
	return ""
}

type Attrs map[string]any

func WithLogAttrs(ctx context.Context, attrs Attrs) context.Context {
	if ctx == nil {
		return ctx
	}
	var newAttrs Attrs
	if existingAttrs, ok := ctx.Value(CtxKeyLogAttrs).(Attrs); ok {
		newAttrs = make(Attrs, len(existingAttrs)+len(attrs))
		for k, v := range existingAttrs {
			newAttrs[k] = v
		}
	} else {
		newAttrs = make(Attrs, len(attrs))
	}
	for k, v := range attrs {
		newAttrs[k] = v
	}
	return context.WithValue(ctx, CtxKeyLogAttrs, newAttrs)
}

func LogAttrs(ctx context.Context) Attrs {
	if ctx == nil {
		return make(Attrs)
	}
	if attrs, ok := ctx.Value(CtxKeyLogAttrs).(Attrs); ok {
		return attrs
	}
	return make(Attrs)
}
