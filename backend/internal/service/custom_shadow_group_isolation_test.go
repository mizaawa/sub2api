//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateShadowRejectsCustomGroupBeforeCreatingAccount(t *testing.T) {
	repo := newSparkShadowRepoStub()
	parent := &Account{
		Name:        "parent",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Credentials: map[string]any{"access_token": "token"},
	}
	require.NoError(t, repo.Create(context.Background(), parent))

	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
		42: {ID: 42, Platform: PlatformCustom},
	}}
	svc := &adminServiceImpl{accountRepo: repo, groupRepo: groupRepo}

	_, err := svc.CreateShadow(context.Background(), parent.ID, ShadowOptions{GroupIDs: []int64{42}})

	require.ErrorContains(t, err, "custom accounts and groups can only be assigned to each other")
	require.Len(t, repo.accounts, 1, "validation must happen before the shadow is persisted")
}
