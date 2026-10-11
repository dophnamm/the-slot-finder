//go:build integration

package db_test

import (
	"context"
	"errors"
	"testing"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/isoweek"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type txProbe struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type weekProbe struct {
	ID        uint         `gorm:"primaryKey"`
	WeekStart isoweek.Week `gorm:"column:week_start"`
}

func newTestPostgres(t *testing.T) *db.PostgresDB {
	t.Helper()
	pdb := dbtest.NewPostgres(t)
	require.NoError(t, pdb.AutoMigrate(&txProbe{}, &weekProbe{}))
	return pdb
}

func countProbes(t *testing.T, pdb *db.PostgresDB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, pdb.Model(&txProbe{}).Count(&n).Error)
	return n
}

func TestTxManager(t *testing.T) {
	pdb := newTestPostgres(t)
	tx := db.ProvideTxManager(pdb)
	ctx := context.Background()
	insert := func(ctx context.Context, name string) error {
		return pdb.From(ctx).Create(&txProbe{Name: name}).Error
	}

	t.Run("commits", func(t *testing.T) {
		require.NoError(t, tx.Do(ctx, func(ctx context.Context) error { return insert(ctx, "a") }))
		assert.Equal(t, int64(1), countProbes(t, pdb))
	})

	t.Run("rolls back on error", func(t *testing.T) {
		err := tx.Do(ctx, func(ctx context.Context) error {
			require.NoError(t, insert(ctx, "b"))
			return errors.New("stop")
		})
		require.Error(t, err)
		assert.Equal(t, int64(1), countProbes(t, pdb))
	})

	t.Run("rolls back on panic", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = tx.Do(ctx, func(ctx context.Context) error {
				require.NoError(t, insert(ctx, "c"))
				panic("boom")
			})
		})
		assert.Equal(t, int64(1), countProbes(t, pdb))
	})

	t.Run("nested Do reuses the transaction", func(t *testing.T) {
		err := tx.Do(ctx, func(ctx context.Context) error {
			require.NoError(t, insert(ctx, "d"))
			require.NoError(t, tx.Do(ctx, func(ctx context.Context) error { return insert(ctx, "e") }))
			return errors.New("roll back both")
		})
		require.Error(t, err)
		assert.Equal(t, int64(1), countProbes(t, pdb))
	})
}

func TestWeekColumnRoundTrip(t *testing.T) {
	pdb := newTestPostgres(t)
	week, err := isoweek.Parse("2026-W53")
	require.NoError(t, err)
	require.NoError(t, pdb.Create(&weekProbe{WeekStart: week}).Error)

	var stored string
	require.NoError(t, pdb.Raw("SELECT week_start::text FROM week_probes").Scan(&stored).Error)
	assert.Equal(t, "2026-12-28", stored)

	var probe weekProbe
	require.NoError(t, pdb.First(&probe).Error)
	assert.Equal(t, "2026-W53", probe.WeekStart.String())
}
