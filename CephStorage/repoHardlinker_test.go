package CephStorage

import (
	"fmt"
	"os"
	"testing"
)

type errorWithCode int

func (e errorWithCode) Error() string {
	return "coded error"
}

func (e errorWithCode) ErrorCode() int {
	return int(e)
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "standard already exists", err: os.ErrExist, want: true},
		{name: "Ceph EEXIST", err: errorWithCode(-17), want: true},
		{name: "wrapped Ceph EEXIST", err: fmt.Errorf("symlink: %w", errorWithCode(-17)), want: true},
		{name: "other Ceph error", err: errorWithCode(-2), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlreadyExists(tt.err); got != tt.want {
				t.Fatalf("isAlreadyExists(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}
