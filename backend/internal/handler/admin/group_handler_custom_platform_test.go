package admin

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/require"
)

func TestGroupRequestsAcceptCustomPlatform(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		req := CreateGroupRequest{Name: "custom-group", Platform: "custom"}
		require.NoError(t, binding.Validator.ValidateStruct(req))
	})

	t.Run("update", func(t *testing.T) {
		req := UpdateGroupRequest{Platform: "custom"}
		require.NoError(t, binding.Validator.ValidateStruct(req))
	})

	t.Run("unknown remains invalid", func(t *testing.T) {
		req := CreateGroupRequest{Name: "unknown-group", Platform: "unknown"}
		require.Error(t, binding.Validator.ValidateStruct(req))
	})
}
