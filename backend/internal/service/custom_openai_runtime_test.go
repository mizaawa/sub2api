package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomOpenAIBaseURLFailsClosedWhenMissing(t *testing.T) {
	account := &Account{
		ID:       17,
		Platform: PlatformCustom,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "custom-secret",
		},
	}

	require.Empty(t, account.GetOpenAIBaseURL())
	_, err := requireOpenAIBaseURL(account)
	require.ErrorContains(t, err, "custom account missing base_url")

	svc := &OpenAIGatewayService{}
	_, err = svc.openAIChatCompletionsTargetURL(account)
	require.ErrorContains(t, err, "custom account missing base_url")
	_, err = svc.buildOpenAIResponsesWSURL(account)
	require.ErrorContains(t, err, "custom account missing base_url")
}

func TestCustomAccountRuntimeEligibilityRequiresAPIKeyAndBaseURL(t *testing.T) {
	valid := &Account{
		Platform:    PlatformCustom,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"api_key":  "custom-secret",
			"base_url": "https://custom.example/v1",
		},
	}
	require.True(t, isOpenAICompatibleAccountEligibleForRequest(context.Background(), valid, PlatformCustom, "", false, ""))

	missingBaseURL := *valid
	missingBaseURL.Credentials = map[string]any{"api_key": "custom-secret"}
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(context.Background(), &missingBaseURL, PlatformCustom, "", false, ""))

	oauth := *valid
	oauth.Type = AccountTypeOAuth
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(context.Background(), &oauth, PlatformCustom, "", false, ""))
}

func TestOfficialOpenAIBaseURLStillDefaultsToOpenAI(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	require.Equal(t, "https://api.openai.com", account.GetOpenAIBaseURL())
}
