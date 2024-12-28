package retries

import (
	"github.com/sshirox/isaac/internal/errors"
	"testing"
)

func TestRetry(t *testing.T) {
	type args struct {
		op func() error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				op: func() error { return nil },
			},
			wantErr: false,
		},
		{
			name: "Non retry error",
			args: args{
				op: func() error { return errors.ErrNonRetry },
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Retry(tt.args.op); (err != nil) != tt.wantErr {
				t.Errorf("Retry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
