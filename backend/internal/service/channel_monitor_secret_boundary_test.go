//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelMonitorManagementReadNeverReturnsCredential(t *testing.T) {
	repo := &groupMonitorRepoStub{existing: &ChannelMonitor{
		ID:     10,
		APIKey: "enc:server-only-secret",
	}}
	encryptor := &groupMonitorEncryptor{}
	svc := NewChannelMonitorService(repo, encryptor)

	monitor, err := svc.Get(context.Background(), 10)

	require.NoError(t, err)
	require.Empty(t, monitor.APIKey)
	require.False(t, monitor.APIKeyDecryptFailed)
	require.Equal(t, []string{"enc:server-only-secret"}, encryptor.decryptInputs)
}

func TestChannelMonitorExecutionReadIsOnlyPlaintextCredentialPath(t *testing.T) {
	repo := &groupMonitorRepoStub{existing: &ChannelMonitor{
		ID:     10,
		APIKey: "enc:server-only-secret",
	}}
	svc := NewChannelMonitorService(repo, &groupMonitorEncryptor{})

	monitor, err := svc.getForExecution(context.Background(), 10)

	require.NoError(t, err)
	require.Equal(t, "server-only-secret", monitor.APIKey)
	require.False(t, monitor.APIKeyDecryptFailed)
}

func TestChannelMonitorManagementReadReportsDecryptFailureWithoutReturningCiphertext(t *testing.T) {
	repo := &groupMonitorRepoStub{existing: &ChannelMonitor{
		ID:     10,
		APIKey: "enc:undecryptable-secret",
	}}
	svc := NewChannelMonitorService(repo, &groupMonitorEncryptor{decryptErr: assertSecretDecryptError{}})

	monitor, err := svc.Get(context.Background(), 10)

	require.NoError(t, err)
	require.Empty(t, monitor.APIKey)
	require.True(t, monitor.APIKeyDecryptFailed)
}

type assertSecretDecryptError struct{}

func (assertSecretDecryptError) Error() string { return "decrypt failed" }
