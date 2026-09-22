package service

import (
	"context"
	"strings"
)

type CompositeRouteResolver struct {
	repo CompositeModelRouteRepository
}

func NewCompositeRouteResolver(repo CompositeModelRouteRepository) *CompositeRouteResolver {
	return &CompositeRouteResolver{repo: repo}
}

func (r *CompositeRouteResolver) Resolve(ctx context.Context, groupID int64, model, endpoint string) (CompositeRouteDecision, error) {
	model = strings.TrimSpace(model)
	endpoint = normalizeCompositeRouteEndpoint(endpoint)
	decision := CompositeRouteDecision{
		GroupID:     groupID,
		PublicModel: model,
		Endpoint:    endpoint,
	}
	if model == "" {
		decision.Reason = "model is required"
		return decision, nil
	}
	return CompositeRouteDecision{
		Matched:        true,
		Source:         CompositeRouteSourceDetector,
		GroupID:        groupID,
		PublicModel:    model,
		TargetPlatform: PlatformCustom,
		UpstreamModel:  model,
		Endpoint:       endpoint,
		Reason:         "Custom groups always route to Custom accounts",
	}, nil
}
