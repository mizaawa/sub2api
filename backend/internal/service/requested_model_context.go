package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

// WithRequestedPublicModel preserves the first non-empty model supplied by the
// downstream client. Channel and account mappings may rewrite request bodies
// later, but response-model masking must still use this public name.
func WithRequestedPublicModel(ctx context.Context, model string) context.Context {
	if ctx == nil {
		return ctx
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return ctx
	}
	if _, ok := RequestedPublicModelFromContext(ctx); ok {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.RequestedPublicModel, model)
}

func downstreamRequestedModel(ctx context.Context, fallback string) string {
	if model, ok := RequestedPublicModelFromContext(ctx); ok {
		return model
	}
	return fallback
}
