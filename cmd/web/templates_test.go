package main

import (
	"testing"
)

func TestIntRange(t *testing.T) {
	tests := []struct {
		name  string
		start int
		stop  int
		want  []int
	}{
		{
			name:  "one to five",
			start: 1,
			stop:  5,
			want:  []int{1, 2, 3, 4, 5},
		},
		{
			name:  "start equals stop",
			start: 5,
			stop:  5,
			want:  []int{5},
		},
		{
			name:  "ten to five",
			start: 10,
			stop:  5,
			want:  []int{10, 9, 8, 7, 6, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var values []int
			for value := range intRange(tt.start, tt.stop) {
				values = append(values, value)
			}

			if len(values) == len(tt.want) {
				for i, value := range values {
					if value != tt.want[i] {
						t.Errorf("at index %d, got %d, want %d", i, value, tt.want[i])
					}
				}
			} else {
				t.Errorf("got %v; want %v", values, tt.want)
			}
		})
	}
}
