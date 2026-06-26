package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepository_CreateUser(t *testing.T) {
	t.Parallel()

	type mockBehavior func(mock sqlmock.Sqlmock, email, password string)

	testTable := []struct {
		name     string
		email    string
		password string
		mockFn   mockBehavior
		want     string
		wantErr  bool
	}{
		{
			name:     "OK",
			email:    "test@mail.com",
			password: "hashed_password",
			mockFn: func(mock sqlmock.Sqlmock, email, password string) {
				mock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_USER)).
					WithArgs(email, password).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
			},
			want: "1",
		},
		{
			name:     "Unique Constraint Violation",
			email:    "test@mail.com",
			password: "hashed_password",
			mockFn: func(mock sqlmock.Sqlmock, email, password string) {
				mock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_USER)).
					WithArgs(email, password).WillReturnError(errors.New("duplicate key value violation"))
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "Internal DB Error",
			email:    "test@mail.com",
			password: "hashed_password",
			mockFn: func(mock sqlmock.Sqlmock, email, password string) {
				mock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_USER)).
					WithArgs(email, password).WillReturnError(errors.New("sql: database is closed"))
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "Scan Error",
			email:    "test@mail.com",
			password: "hashed_password",
			mockFn: func(mock sqlmock.Sqlmock, email, password string) {
				mock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_USER)).
					WithArgs(email, password).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(nil))
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sql mock: %v", err)
			}
			defer db.Close()

			repo := NewRepository(db)

			test.mockFn(mock, test.email, test.password)

			res, err := repo.CreateUser(context.Background(), test.email, test.password)
			if test.want != res {
				t.Errorf("invalid result, expected: %s, got: %s", test.want, res)
			}

			if test.wantErr {
				if err == nil {
					t.Errorf("invalid error, expected: error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("invalid error, expected: nil, got: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
