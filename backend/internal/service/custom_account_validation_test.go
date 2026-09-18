package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCustomAccountCredentials(t *testing.T) {
	tests := []struct {
		name        string
		platform    string
		accountType string
		credentials map[string]any
		wantErr     string
	}{
		{
			name:        "valid custom API key account",
			platform:    PlatformCustom,
			accountType: AccountTypeAPIKey,
			credentials: map[string]any{"api_key": "sk-test", "base_url": "https://api.example.com/v1"},
		},
		{
			name:        "custom rejects OAuth",
			platform:    PlatformCustom,
			accountType: AccountTypeOAuth,
			credentials: map[string]any{"api_key": "sk-test", "base_url": "https://api.example.com/v1"},
			wantErr:     "custom accounts only support apikey credentials",
		},
		{
			name:        "custom requires API key",
			platform:    PlatformCustom,
			accountType: AccountTypeAPIKey,
			credentials: map[string]any{"base_url": "https://api.example.com/v1"},
			wantErr:     "custom accounts require an api_key",
		},
		{
			name:        "custom requires base URL",
			platform:    PlatformCustom,
			accountType: AccountTypeAPIKey,
			credentials: map[string]any{"api_key": "sk-test"},
			wantErr:     "custom accounts require a base_url",
		},
		{
			name:        "other platforms keep existing credential rules",
			platform:    PlatformOpenAI,
			accountType: AccountTypeOAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomAccountCredentials(tt.platform, tt.accountType, tt.credentials)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestValidateCustomAccountGroupPlatform(t *testing.T) {
	require.NoError(t, validateCustomAccountGroupPlatform(PlatformCustom, PlatformCustom))
	require.NoError(t, validateCustomAccountGroupPlatform(PlatformOpenAI, PlatformComposite))
	require.ErrorContains(
		t,
		validateCustomAccountGroupPlatform(PlatformCustom, PlatformOpenAI),
		"custom accounts and groups can only be assigned to each other",
	)
	require.ErrorContains(
		t,
		validateCustomAccountGroupPlatform(PlatformCustom, PlatformComposite),
		"custom accounts and groups can only be assigned to each other",
	)
	require.ErrorContains(
		t,
		validateCustomAccountGroupPlatform(PlatformOpenAI, PlatformCustom),
		"custom accounts and groups can only be assigned to each other",
	)
}
