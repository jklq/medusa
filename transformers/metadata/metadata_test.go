package metadata

import (
	"reflect"
	"testing"

	"git.sr.ht/~relay/medusa"
)

func TestMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		initial  medusa.Store
		expected medusa.Store
	}{
		{
			name: "empty store",
			input: map[string]any{
				"title":  "My Site",
				"author": "John Doe",
			},
			initial: medusa.Store{},
			expected: medusa.Store{
				"title":  "My Site",
				"author": "John Doe",
			},
		},
		{
			name: "existing store",
			input: map[string]any{
				"description": "A blog",
			},
			initial: medusa.Store{
				"title": "My Site",
			},
			expected: medusa.Store{
				"title":       "My Site",
				"description": "A blog",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := New(tt.input)
			files := []medusa.File{}
			store := tt.initial

			err := transformer(&files, &store)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(store, tt.expected) {
				t.Errorf("got %v, want %v", store, tt.expected)
			}
		})
	}
}
