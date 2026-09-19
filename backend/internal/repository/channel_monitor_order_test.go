package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/channelmonitor"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestChannelMonitorListOrdersBySortOrderThenID(t *testing.T) {
	var capturedSQL string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureEntQueryMatcher{actual: &capturedSQL}))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })

	mock.ExpectQuery("count monitors").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("list monitors").WillReturnRows(sqlmock.NewRows(channelmonitor.Columns))

	repo := NewChannelMonitorRepository(client, db)
	items, total, err := repo.List(context.Background(), service.ChannelMonitorListParams{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Empty(t, items)
	require.Zero(t, total)
	require.NoError(t, mock.ExpectationsWereMet())
	assertChannelMonitorSortSQL(t, capturedSQL)
}

func TestChannelMonitorListEnabledOrdersBySortOrderThenID(t *testing.T) {
	var capturedSQL string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureEntQueryMatcher{actual: &capturedSQL}))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })

	mock.ExpectQuery("list enabled monitors").WillReturnRows(sqlmock.NewRows(channelmonitor.Columns))

	repo := NewChannelMonitorRepository(client, db)
	items, err := repo.ListEnabled(context.Background())

	require.NoError(t, err)
	require.Empty(t, items)
	require.NoError(t, mock.ExpectationsWereMet())
	assertChannelMonitorSortSQL(t, capturedSQL)
}

func assertChannelMonitorSortSQL(t *testing.T, query string) {
	t.Helper()
	normalized := normalizeSQLWhitespace(query)
	orderIndex := strings.Index(normalized, ` ORDER BY `)
	require.NotEqual(t, -1, orderIndex, "query must contain ORDER BY: %s", normalized)
	orderClause := normalized[orderIndex:]
	require.Regexp(t, `ORDER BY .*"sort_order" ASC, .*"id" ASC`, orderClause)
}
