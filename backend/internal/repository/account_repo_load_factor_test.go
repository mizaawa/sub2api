package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBulkUpdatePersistsZeroLoadFactor(t *testing.T) {
	exec := &recordingSQLExecutor{result: rowsAffectedResult(1)}
	repo := newAccountRepositoryWithSQL(nil, exec, nil)
	loadFactor := 0

	rows, err := repo.BulkUpdate(context.Background(), []int64{27}, service.AccountBulkUpdate{
		LoadFactor: &loadFactor,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
	require.NotEmpty(t, exec.execQueries)
	require.Contains(t, normalizeSQLWhitespace(exec.execQueries[0]), "SET load_factor = $1")
	require.Len(t, exec.execArgs[0], 2)
	require.Equal(t, 0, exec.execArgs[0][0])
}
