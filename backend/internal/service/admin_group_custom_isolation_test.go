//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func customIsolationTestGroup(id int64, platform string) *Group {
	return &Group{
		ID:               id,
		Name:             "group",
		Platform:         platform,
		RateMultiplier:   1,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}
}

func TestCanCopyAccountsFromGroupPlatformRejectsCustomSourceForComposite(t *testing.T) {
	require.False(t, canCopyAccountsFromGroupPlatform(PlatformComposite, PlatformCustom))
	require.True(t, canCopyAccountsFromGroupPlatform(PlatformComposite, PlatformOpenAI))
	require.True(t, canCopyAccountsFromGroupPlatform(PlatformCustom, PlatformCustom))
}

func TestCreateCustomGroupRejectsNonCustomAccountHiddenInCustomSource(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{
		getByIDsAccounts: []*Account{{ID: 101, Platform: PlatformOpenAI}},
	}
	bindCalled := false
	groupRepo := &groupRepoStubForAdmin{
		createID: 99,
		getByIDByID: map[int64]*Group{
			10: customIsolationTestGroup(10, PlatformCustom),
		},
		getAccountIDsByGroupIDsFn: func([]int64) ([]int64, error) {
			return []int64{101}, nil
		},
		bindAccountsToGroupFn: func(int64, []int64) error {
			bindCalled = true
			return nil
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	group, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
		Name:                     "custom",
		Platform:                 PlatformCustom,
		RateMultiplier:           1,
		CopyAccountsFromGroupIDs: []int64{10},
	})

	require.Nil(t, group)
	require.ErrorContains(t, err, "custom accounts and groups can only be assigned to each other")
	require.Nil(t, groupRepo.created, "validation must happen before the group is created")
	require.False(t, bindCalled)
	require.Equal(t, []int64{101}, accountRepo.getByIDsIDs)
}

func TestCreateCustomGroupBindsCustomAccountsAfterValidation(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{
		getByIDsAccounts: []*Account{{ID: 111, Platform: PlatformCustom, Type: AccountTypeAPIKey}},
	}
	var boundAccountIDs []int64
	groupRepo := &groupRepoStubForAdmin{
		createID: 19,
		getByIDByID: map[int64]*Group{
			11: customIsolationTestGroup(11, PlatformCustom),
		},
		getAccountIDsByGroupIDsFn: func([]int64) ([]int64, error) {
			return []int64{111}, nil
		},
		bindAccountsToGroupFn: func(groupID int64, accountIDs []int64) error {
			require.Equal(t, int64(19), groupID)
			boundAccountIDs = append([]int64(nil), accountIDs...)
			return nil
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	group, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
		Name:                     "custom",
		Platform:                 PlatformCustom,
		RateMultiplier:           1,
		CopyAccountsFromGroupIDs: []int64{11},
	})

	require.NoError(t, err)
	require.Equal(t, PlatformCustom, group.Platform)
	require.Equal(t, []int64{111}, accountRepo.getByIDsIDs)
	require.Equal(t, []int64{111}, boundAccountIDs)
}

func TestCreateCompositeGroupRejectsCustomAccountHiddenInConcreteSource(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{
		getByIDsAccounts: []*Account{{ID: 201, Platform: PlatformCustom}},
	}
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[int64]*Group{
			20: customIsolationTestGroup(20, PlatformOpenAI),
		},
		getAccountIDsByGroupIDsFn: func([]int64) ([]int64, error) {
			return []int64{201}, nil
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	group, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
		Name:                     "composite",
		Platform:                 PlatformComposite,
		RateMultiplier:           1,
		CopyAccountsFromGroupIDs: []int64{20},
	})

	require.Nil(t, group)
	require.ErrorContains(t, err, "custom accounts and groups can only be assigned to each other")
	require.Nil(t, groupRepo.created)
}

func TestCreateCompositeGroupRejectsCustomSourceBeforeReadingAccounts(t *testing.T) {
	getAccountsCalled := false
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[int64]*Group{
			30: customIsolationTestGroup(30, PlatformCustom),
		},
		getAccountIDsByGroupIDsFn: func([]int64) ([]int64, error) {
			getAccountsCalled = true
			return []int64{301}, nil
		},
	}
	svc := &adminServiceImpl{groupRepo: groupRepo}

	group, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
		Name:                     "composite",
		Platform:                 PlatformComposite,
		RateMultiplier:           1,
		CopyAccountsFromGroupIDs: []int64{30},
	})

	require.Nil(t, group)
	require.ErrorContains(t, err, "platform mismatch")
	require.False(t, getAccountsCalled)
	require.Nil(t, groupRepo.created)
}

func TestUpdateCompositeGroupRejectsCustomSourceBeforeUpdateOrClear(t *testing.T) {
	getAccountsCalled := false
	clearCalled := false
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[int64]*Group{
			70: customIsolationTestGroup(70, PlatformComposite),
			71: customIsolationTestGroup(71, PlatformCustom),
		},
		getAccountIDsByGroupIDsFn: func([]int64) ([]int64, error) {
			getAccountsCalled = true
			return []int64{701}, nil
		},
		deleteAccountGroupsByGroupIDFn: func(int64) (int64, error) {
			clearCalled = true
			return 0, nil
		},
	}
	svc := &adminServiceImpl{groupRepo: groupRepo}

	group, err := svc.UpdateGroup(context.Background(), 70, &UpdateGroupInput{
		CopyAccountsFromGroupIDs: []int64{71},
	})

	require.Nil(t, group)
	require.ErrorContains(t, err, "platform mismatch")
	require.Nil(t, groupRepo.updated)
	require.False(t, getAccountsCalled)
	require.False(t, clearCalled)
}

func TestUpdateGroupRejectsCustomAccountBeforeClearingBindings(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{
		getByIDsAccounts: []*Account{{ID: 401, Platform: PlatformCustom}},
	}
	clearCalled := false
	bindCalled := false
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[int64]*Group{
			40: customIsolationTestGroup(40, PlatformOpenAI),
			41: customIsolationTestGroup(41, PlatformOpenAI),
		},
		getAccountIDsByGroupIDsFn: func([]int64) ([]int64, error) {
			return []int64{401}, nil
		},
		deleteAccountGroupsByGroupIDFn: func(int64) (int64, error) {
			clearCalled = true
			return 0, nil
		},
		bindAccountsToGroupFn: func(int64, []int64) error {
			bindCalled = true
			return nil
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	group, err := svc.UpdateGroup(context.Background(), 40, &UpdateGroupInput{
		CopyAccountsFromGroupIDs: []int64{41},
	})

	require.Nil(t, group)
	require.ErrorContains(t, err, "custom accounts and groups can only be assigned to each other")
	require.Nil(t, groupRepo.updated, "validation must happen before the group row is updated")
	require.False(t, clearCalled, "validation must happen before existing bindings are cleared")
	require.False(t, bindCalled)
}

func TestUpdateGroupRejectsPlatformChangesAcrossCustomBoundaryWithExistingBindings(t *testing.T) {
	tests := []struct {
		name             string
		originalPlatform string
		targetPlatform   string
		accountPlatform  string
	}{
		{
			name:             "non-custom to custom",
			originalPlatform: PlatformOpenAI,
			targetPlatform:   PlatformCustom,
			accountPlatform:  PlatformOpenAI,
		},
		{
			name:             "custom to non-custom",
			originalPlatform: PlatformCustom,
			targetPlatform:   PlatformOpenAI,
			accountPlatform:  PlatformCustom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &accountRepoStubForBulkUpdate{
				getByIDsAccounts: []*Account{{ID: 501, Platform: tt.accountPlatform}},
			}
			groupRepo := &groupRepoStubForAdmin{
				getByID: customIsolationTestGroup(50, tt.originalPlatform),
				getAccountIDsByGroupIDsFn: func(groupIDs []int64) ([]int64, error) {
					require.Equal(t, []int64{50}, groupIDs)
					return []int64{501}, nil
				},
			}
			svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

			group, err := svc.UpdateGroup(context.Background(), 50, &UpdateGroupInput{Platform: tt.targetPlatform})

			require.Nil(t, group)
			require.ErrorContains(t, err, "custom accounts and groups can only be assigned to each other")
			require.Nil(t, groupRepo.updated)
			require.Equal(t, []int64{501}, accountRepo.getByIDsIDs)
		})
	}
}

func TestUpdateGroupAllowsCustomBoundaryChangeWithoutBindings(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{}
	groupRepo := &groupRepoStubForAdmin{
		getByID: customIsolationTestGroup(60, PlatformOpenAI),
		getAccountIDsByGroupIDsFn: func(groupIDs []int64) ([]int64, error) {
			require.Equal(t, []int64{60}, groupIDs)
			return nil, nil
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	group, err := svc.UpdateGroup(context.Background(), 60, &UpdateGroupInput{Platform: PlatformCustom})

	require.NoError(t, err)
	require.Equal(t, PlatformCustom, group.Platform)
	require.NotNil(t, groupRepo.updated)
	require.False(t, accountRepo.getByIDsCalled)
}
