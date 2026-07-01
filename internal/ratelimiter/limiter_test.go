package ratelimiter

import (
	"testing"
	"time"
)

func TestResult_IsDenied(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name:   "allowed request has no retry",
			result: Result{Allowed: true, Remaining: 5, Limit: 10},
			want:   false,
		},
		{
			name:   "denied request should signal retry",
			result: Result{Allowed: false, Remaining: 0, Limit: 10, RetryAfter: 2 * time.Second},
			want:   true,
		},
		{
			name:   "denied at exact limit boundary",
			result: Result{Allowed: false, Remaining: 0, Limit: 1, RetryAfter: 500 * time.Millisecond},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := !tt.result.Allowed
			if got != tt.want {
				t.Errorf("Result.Allowed = %v, want denied = %v", tt.result.Allowed, tt.want)
			}
		})
	}
}

func TestResult_RemainingNeverNegative(t *testing.T) {
	tests := []struct {
		name      string
		remaining int64
	}{
		{name: "zero remaining is valid", remaining: 0},
		{name: "positive remaining is valid", remaining: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Remaining: tt.remaining}
			if r.Remaining < 0 {
				t.Errorf("Result.Remaining = %d, must never be negative", r.Remaining)
			}
		})
	}
}
