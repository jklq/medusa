package markdown

import (
	"path/filepath"
	"testing"

	"git.sr.ht/~relay/medusa"
)

func TestMarkdown(t *testing.T) {
	files := []medusa.File{
		{
			Path: "markdown.md",
		},
		{
			Path: "notmarkdown.txt",
		},
	}

	files[0].SetContent([]byte("# Hello"))
	files[1].SetContent([]byte("# Goodbye"))

	transformer := New()
	store := make(medusa.Store)

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<h1>Hello</h1>\n"
	actual := string(files[0].Content())
	if expected != actual {
		t.Fatalf("Unexpected heading render:\n expected %v, got %v\n",
			expected,
			actual)
	}
	expected = ".html"
	actual = filepath.Ext(files[0].Path)
	if expected != actual {
		t.Fatalf("Unexpected extension:\n expected %v, got %v\n",
			expected,
			actual)
	}

	expected = "# Goodbye"
	actual = string(files[1].Content())
	if expected != actual {
		t.Fatalf("Unexpected heading render:\n expected %v, got %v",
			expected,
			actual)
	}

}
