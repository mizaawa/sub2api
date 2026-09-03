package service

import "testing"

func TestAccountSupportsPrivacySetting(t *testing.T) {
	tests := []struct {
		name     string
		account  *Account
		expected bool
	}{
		{
			name:     "OpenAI OAuth",
			account:  &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth},
			expected: true,
		},
		{
			name:     "Antigravity OAuth",
			account:  &Account{Platform: PlatformAntigravity, Type: AccountTypeOAuth},
			expected: true,
		},
		{
			name:     "OpenAI API key",
			account:  &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
			expected: false,
		},
		{
			name:     "OpenAI setup token",
			account:  &Account{Platform: PlatformOpenAI, Type: AccountTypeSetupToken},
			expected: false,
		},
		{
			name:     "Gemini OAuth",
			account:  &Account{Platform: PlatformGemini, Type: AccountTypeOAuth},
			expected: false,
		},
		{
			name:     "nil account",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.account.SupportsPrivacySetting(); got != test.expected {
				t.Fatalf("SupportsPrivacySetting() = %v, want %v", got, test.expected)
			}
		})
	}
}
