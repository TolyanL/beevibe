package pgtest

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"

	pgmocks "github.com/TolyanL/beevibe/internal/adapter/postgres/mocks"
	dbtxmocks "github.com/TolyanL/beevibe/pkg/postgres/mocks"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

// DB — мок соединения, положенный в контекст: адаптеры берут его через transaction.TryExtractTX.
func DB(t *testing.T) (*dbtxmocks.MockDBTX, context.Context) {
	t.Helper()

	db := dbtxmocks.NewMockDBTX(t)

	return db, transaction.WithDBTX(context.Background(), db)
}

// Row — мок pgx.Row, чей Scan раскладывает значения по destination-указателям.
func Row(t *testing.T, values ...any) *pgmocks.MockRow {
	t.Helper()

	row := pgmocks.NewMockRow(t)
	scanRow(row, len(values)).Run(func(dest ...any) {
		fill(t, dest, values)
	}).Return(nil)

	return row
}

// RowError — мок pgx.Row, чей Scan завершается ошибкой.
func RowError(t *testing.T, err error, destinations int) *pgmocks.MockRow {
	t.Helper()

	row := pgmocks.NewMockRow(t)
	scanRow(row, destinations).Return(err)

	return row
}

// RowsError — мок pgx.Rows, который отдаёт ошибку итерации.
func RowsError(t *testing.T, err error) *pgmocks.MockRows {
	t.Helper()

	rows := pgmocks.NewMockRows(t)
	rows.EXPECT().Next().Return(false).Once()
	rows.EXPECT().Err().Return(err)
	rows.EXPECT().Close().Return()

	return rows
}

// Rows — мок pgx.Rows: Next выдаёт ровно len(sets) строк, затем false.
func Rows(t *testing.T, sets ...[]any) *pgmocks.MockRows {
	t.Helper()

	rows := pgmocks.NewMockRows(t)

	if len(sets) > 0 {
		width := len(sets[0])

		scanRows(rows, width).Run(func(dest ...any) {
			fill(t, dest, sets[0])
		}).Return(nil).Once()

		for _, set := range sets[1:] {
			scanRows(rows, width).Run(func(dest ...any) {
				fill(t, dest, set)
			}).Return(nil).Once()
		}
	}

	if len(sets) > 0 {
		rows.EXPECT().Next().Return(true).Times(len(sets))
	}

	rows.EXPECT().Next().Return(false).Once()
	rows.EXPECT().Err().Return(nil)
	rows.EXPECT().Close().Return()

	return rows
}

// Args — сопоставление вариативных аргументов запроса: mock.Anything в values
// работает как wildcard, остальные значения сравниваются как есть.
func Args(values ...any) any {
	return mock.MatchedBy(func(args []any) bool {
		if len(args) != len(values) {
			return false
		}

		for i := range values {
			if values[i] == mock.Anything {
				continue
			}

			if !reflect.DeepEqual(args[i], values[i]) {
				return false
			}
		}

		return true
	})
}

// Tag — тег команды pgx с указанным числом затронутых строк.
func Tag(rows int64) pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", rows))
}

// UniqueViolation — ошибка нарушения уникальности, как её отдаёт pgx.
func UniqueViolation(constraint string) error {
	return &pgconn.PgError{Code: "23505", ConstraintName: constraint}
}

func scanRow(row *pgmocks.MockRow, destinations int) *pgmocks.MockRow_Scan_Call {
	if destinations == 0 {
		return row.EXPECT().Scan()
	}

	return row.EXPECT().Scan(mock.Anything)
}

func scanRows(rows *pgmocks.MockRows, destinations int) *pgmocks.MockRows_Scan_Call {
	if destinations == 0 {
		return rows.EXPECT().Scan()
	}

	return rows.EXPECT().Scan(mock.Anything)
}

func fill(t *testing.T, dest []any, values []any) {
	t.Helper()

	if len(dest) != len(values) {
		t.Fatalf("scan destinations = %d, want %d", len(dest), len(values))

		return
	}

	for i, value := range values {
		target := reflect.ValueOf(dest[i])

		if target.Kind() != reflect.Pointer {
			t.Fatalf("scan destination %d is %T, want pointer", i, dest[i])

			return
		}

		if value == nil {
			target.Elem().SetZero()

			continue
		}

		source := reflect.ValueOf(value)
		if !source.Type().AssignableTo(target.Elem().Type()) {
			t.Fatalf("scan value %d is %T, destination wants %s", i, value, target.Elem().Type())

			return
		}

		target.Elem().Set(source)
	}
}
