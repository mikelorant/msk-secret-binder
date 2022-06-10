package sliceutil

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name    string
		giveSrc []string
		giveCmp []string
		want    []string
	}{
		{
			name:    "equal",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"apple", "pear", "orange"},
			want:    []string{},
		}, {
			name:    "compare_has_less",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"apple", "orange"},
			want:    []string{"pear"},
		}, {
			name:    "compare_has_more",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"apple", "pear", "orange", "lemon"},
			want:    []string{},
		}, {
			name:    "equal_reordered",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"orange", "apple", "pear"},
			want:    []string{},
		}, {
			name:    "source_is_empty",
			giveSrc: []string{},
			giveCmp: []string{"orange", "apple", "pear"},
			want:    []string{},
		}, {
			name:    "compare_is_empty",
			giveSrc: []string{"orange", "apple", "pear"},
			giveCmp: []string{},
			want:    []string{"orange", "apple", "pear"},
		}, {
			name:    "source_has_duplicates",
			giveSrc: []string{"pear", "lemon", "pear"},
			giveCmp: []string{"lemon"},
			want:    []string{"pear", "pear"},
		}, {
			name:    "compare_has_duplicates",
			giveSrc: []string{"apple", "pear", "coconut", "orange"},
			giveCmp: []string{"apple", "pear", "orange", "apple"},
			want:    []string{"coconut"},
		}, {
			name:    "both_empty",
			giveSrc: []string{},
			giveCmp: []string{},
			want:    []string{},
		}, {
			name:    "source_is_nil",
			giveSrc: nil,
			giveCmp: []string{"apple", "pear", "orange"},
			want:    []string{},
		}, {
			name:    "compare_is_nil",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: nil,
			want:    []string{"apple", "pear", "orange"},
		}, {
			name:    "both_are_nil",
			giveSrc: nil,
			giveCmp: nil,
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Diff(tt.giveSrc, tt.giveCmp)
			td.Cmp(t, got, tt.want)
		})
	}
}
