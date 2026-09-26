package openai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultModelsIncludeBareGPT56Alias(t *testing.T) {
	require.Contains(t, DefaultModelIDs(), "gpt-5.6")
}

func TestDefaultModelsIncludeLatestGPT6Models(t *testing.T) {
	require.NotEmpty(t, DefaultModels)
	require.Equal(t, "gpt-6-astra", DefaultModels[0].ID)
	require.Subset(t, DefaultModelIDs(), []string{"gpt-6-astra", "gpt-6-sol", "gpt-6-luna"})
}
