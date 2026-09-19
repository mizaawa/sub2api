//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestChannelMonitorUpdateSortOrdersIsAtomicAndLastDuplicateWins(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT COUNT\("channel_monitors"\."id"\) FROM "channel_monitors" WHERE "channel_monitors"\."id" IN \(\$1, \$2\)`).
		WithArgs(int64(41), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectExec(`(?s)UPDATE "channel_monitors" SET .*"sort_order" = \$2 WHERE "channel_monitors"\."id" = \$3`).
		WithArgs(sqlmock.AnyArg(), 30, int64(41)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE "channel_monitors" SET .*"sort_order" = \$2 WHERE "channel_monitors"\."id" = \$3`).
		WithArgs(sqlmock.AnyArg(), 15, int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewChannelMonitorRepository(client, db)
	err = repo.UpdateSortOrders(context.Background(), []service.ChannelMonitorSortOrderUpdate{
		{ID: 41, SortOrder: 30},
		{ID: 42, SortOrder: 10},
		{ID: 42, SortOrder: 15},
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelMonitorUpdateSortOrdersMissingIDRollsBackWithoutUpdates(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT COUNT\("channel_monitors"\."id"\) FROM "channel_monitors" WHERE "channel_monitors"\."id" IN \(\$1, \$2\)`).
		WithArgs(int64(41), int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	repo := NewChannelMonitorRepository(client, db)
	err = repo.UpdateSortOrders(context.Background(), []service.ChannelMonitorSortOrderUpdate{
		{ID: 41, SortOrder: 20},
		{ID: 999, SortOrder: 10},
	})

	require.ErrorIs(t, err, service.ErrChannelMonitorNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
