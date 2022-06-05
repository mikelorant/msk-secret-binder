package app

import (
	"fmt"
	"testing"

	"github.com/MakeNowJust/heredoc"
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
			want: heredoc.Docf(`
				+apple
				+pear
				+orange
			`),
		}, {
			name: "remove",
			give: SecretChangeSet{
				remove: []string{"peach", "coconut"},
			},
			want: heredoc.Docf(`
				-peach
				-coconut
			`),
		}, {
			name: "add_remove",
			give: SecretChangeSet{
				add:    []string{"apple", "pear", "orange"},
				remove: []string{"peach", "coconut"},
			},
			want: heredoc.Docf(`
				+apple
				+pear
				+orange
				-peach
				-coconut
			`),
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
