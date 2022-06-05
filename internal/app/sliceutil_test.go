package app

import (
	"reflect"
	"testing"
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
			name:    "removed",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"apple", "orange"},
			want:    []string{"pear"},
		}, {
			name:    "added",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"apple", "pear", "orange", "lemon"},
			want:    []string{},
		}, {
			name:    "reordered",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"orange", "apple", "pear"},
			want:    []string{},
		}, {
			name:    "empty",
			giveSrc: []string{},
			giveCmp: []string{"orange", "apple", "pear"},
			want:    []string{},
		}, {
			name:    "nothing",
			giveSrc: []string{"orange", "apple", "pear"},
			giveCmp: []string{},
			want:    []string{"orange", "apple", "pear"},
		}, {
			name:    "removed_duplicate",
			giveSrc: []string{"pear", "lemon", "pear"},
			giveCmp: []string{""},
			want:    []string{"pear", "lemon", "pear"},
		}, {
			name:    "added_duplicate",
			giveSrc: []string{"apple", "pear", "orange"},
			giveCmp: []string{"apple", "pear", "orange", "apple"},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diff(tt.giveSrc, tt.giveCmp)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

}
