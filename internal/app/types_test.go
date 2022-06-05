package app

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name string
		give SecretChangeSet
		want string
	}{
		{
			name: "add",
			give: SecretChangeSet{
				add: []string{"apple", "pear", "orange"},
			},
			want: "+apple\n+pear\n+orange\n",
		}, {
			name: "remove",
			give: SecretChangeSet{
				remove: []string{"peach", "coconut"},
			},
			want: "-peach\n-coconut\n",
		}, {
			name: "add_remove",
			give: SecretChangeSet{
				add:    []string{"apple", "pear", "orange"},
				remove: []string{"peach", "coconut"},
			},
			want: "+apple\n+pear\n+orange\n-peach\n-coconut\n",
		}, {
			name: "empty",
			give: SecretChangeSet{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprint(tt.give)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
