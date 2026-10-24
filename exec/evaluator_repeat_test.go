package exec

import "testing"

func TestStringRepeatSize(t *testing.T) {
	const limit = 100 * 1024 * 1024
	maxInt := int(^uint(0) >> 1)
	// Exercise the production guard without allocating strings at the limit.
	for _, tc := range []struct {
		name    string
		length  int
		count   int
		wantErr bool
	}{
		{name: "empty", length: 0, count: maxInt},
		{name: "zero", length: maxInt, count: 0},
		{name: "below_limit", length: 1, count: limit - 1},
		{name: "at_limit", length: 1, count: limit},
		{name: "above_limit", length: 1, count: limit + 1, wantErr: true},
		{name: "one_copy_at_limit", length: limit, count: 1},
		{name: "one_copy_above_limit", length: limit + 1, count: 1, wantErr: true},
		{name: "two_bytes_at_limit", length: 2, count: limit / 2},
		{name: "two_bytes_above_limit", length: 2, count: limit/2 + 1, wantErr: true},
		{name: "three_bytes_below_limit", length: 3, count: limit / 3},
		{name: "three_bytes_above_limit", length: 3, count: limit/3 + 1, wantErr: true},
		{name: "count_overflow", length: 4, count: maxInt/2 + 1, wantErr: true},
		{name: "length_overflow", length: maxInt, count: 2, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkStringRepeatSize(tc.length, tc.count)
			if (err != nil) != tc.wantErr {
				t.Fatalf("length=%d, count=%d: got %v, want error=%v", tc.length, tc.count, err, tc.wantErr)
			}
		})
	}
}
