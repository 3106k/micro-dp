package featureflag

import (
	"context"

	"github.com/open-feature/go-sdk/openfeature"
)

// envProvider implements openfeature.FeatureProvider using in-memory flag values
// loaded from environment variables at startup.
type envProvider struct {
	flags map[string]bool
}

func newEnvProvider(flags map[string]bool) *envProvider {
	return &envProvider{flags: flags}
}

func (p *envProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{Name: "env"}
}

func (p *envProvider) BooleanEvaluation(_ context.Context, flag string, defaultValue bool, _ openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	v, ok := p.flags[flag]
	if !ok {
		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.DefaultReason,
			},
		}
	}
	return openfeature.BoolResolutionDetail{
		Value: v,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.StaticReason,
		},
	}
}

func (p *envProvider) StringEvaluation(_ context.Context, _ string, defaultValue string, _ openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	return openfeature.StringResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.DefaultReason,
		},
	}
}

func (p *envProvider) FloatEvaluation(_ context.Context, _ string, defaultValue float64, _ openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	return openfeature.FloatResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.DefaultReason,
		},
	}
}

func (p *envProvider) IntEvaluation(_ context.Context, _ string, defaultValue int64, _ openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	return openfeature.IntResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.DefaultReason,
		},
	}
}

func (p *envProvider) ObjectEvaluation(_ context.Context, _ string, defaultValue any, _ openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	return openfeature.InterfaceResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.DefaultReason,
		},
	}
}

func (p *envProvider) Hooks() []openfeature.Hook {
	return nil
}
