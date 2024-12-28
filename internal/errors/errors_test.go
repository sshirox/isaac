package errors

import (
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsRetryPGErr(t *testing.T) {
	type args struct {
		pgErr error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "AdminShutdown error",
			args: args{pgErr: &pgconn.PgError{Code: pgerrcode.AdminShutdown}},
			want: true,
		},
		{
			name: "ConnectionFailure error",
			args: args{pgErr: &pgconn.PgError{Code: pgerrcode.ConnectionFailure}},
			want: true,
		},
		{
			name: "ConnectionException error",
			args: args{pgErr: &pgconn.PgError{Code: pgerrcode.ConnectionException}},
			want: true,
		},
		{
			name: "DeadlockDetected error",
			args: args{pgErr: &pgconn.PgError{Code: pgerrcode.DeadlockDetected}},
			want: true,
		},
		{
			name: "SerializationFailure error",
			args: args{pgErr: &pgconn.PgError{Code: pgerrcode.SerializationFailure}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryPGErr(tt.args.pgErr)
			assert.Equal(t, got, tt.want)
		})
	}
}
