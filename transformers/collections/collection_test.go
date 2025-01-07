package collections

import (
	"os"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~relay/medusa"
)

// Helper function to create test files
func createTestFile(t *testing.T, path string, content string, modTime time.Time) medusa.File {
	t.Helper()
	file := medusa.File{
		Path: path,
		FileInfo: &testFileInfo{
			modTime: modTime,
			name:    path,
		},
		Frontmatter: map[string]any{
			"title": path,
		},
		Store: make(medusa.Store),
	}

	file.SetContent([]byte(content))

	return file
}

// Test file info implementation
type testFileInfo struct {
	modTime time.Time
	name    string
}

func (fi *testFileInfo) Name() string       { return fi.name }
func (fi *testFileInfo) Size() int64        { return 0 }
func (fi *testFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *testFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *testFileInfo) IsDir() bool        { return false }
func (fi *testFileInfo) Sys() any           { return nil }

func TestCollectionBasicPatternMatching(t *testing.T) {
	files := []medusa.File{
		createTestFile(t, "posts/post1.md", "content1", time.Now()),
		createTestFile(t, "posts/post2.md", "content2", time.Now()),
		createTestFile(t, "pages/page1.md", "content3", time.Now()),
	}

	store := make(medusa.Store)
	transformer := New(CollectionConfig{
		Name:     "posts",
		Patterns: []string{"posts/*.md"},
	})

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collection := store["Collections"].(Collections)["posts"]
	if len(collection.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(collection.Files))
	}
}

func TestCollectionWithCustomSorting(t *testing.T) {
	now := time.Now()
	files := []medusa.File{
		createTestFile(t, "posts/post2.md", "content2", now.Add(-1*time.Hour)),
		createTestFile(t, "posts/post3.md", "content3", now.Add(-5*time.Hour)),
		createTestFile(t, "posts/post1.md", "content1", now.Add(-2*time.Hour)),
	}

	store := make(medusa.Store)
	transformer := New(CollectionConfig{
		Name:     "posts",
		Patterns: []string{"posts/*.md"},
		SortBy: func(a, b medusa.File) int {
			return strings.Compare(a.Path, b.Path)
		},
		DontReverse: true,
	})

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collection := store["Collections"].(Collections)["posts"]
	if len(collection.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(collection.Files))
	}

	// Verify files are sorted by path
	expectedPaths := []string{"posts/post1.md", "posts/post2.md", "posts/post3.md"}
	for i, expectedPath := range expectedPaths {
		if collection.Files[i].Frontmatter["title"] != expectedPath {
			t.Errorf("file at position %d: expected %s, got %s",
				i,
				expectedPath,
				collection.Files[i].Frontmatter["title"])
		}
	}
}

func TestCollectionWithCustomFilter(t *testing.T) {
	now := time.Now()
	files := []medusa.File{
		createTestFile(t, "posts/post2.md", "content2", now.Add(-1*time.Hour)),
		createTestFile(t, "posts/post3.md", "content3", now.Add(-5*time.Hour)),
		createTestFile(t, "posts/post1.md", "content1", now.Add(-2*time.Hour)),
		createTestFile(t, "posts/snoozepost1.md", "content1", now.Add(-2*time.Hour)),
		createTestFile(t, "posts/snoozepost2.md", "content1", now.Add(-2*time.Hour)),
		createTestFile(t, "posts/snoozepost3.md", "content1", now.Add(-2*time.Hour)),
		createTestFile(t, "posts/snoozepost4.md", "content1", now.Add(-2*time.Hour)),
		createTestFile(t, "posts/snooze.md", "content1", now.Add(-2*time.Hour)),
	}

	store := make(medusa.Store)
	transformer := New(CollectionConfig{
		Name:     "posts",
		Patterns: []string{"posts/*.md"},
		FilterFunc: func(file medusa.File) bool {
			return !strings.Contains(file.Path, "snooze")
		},
		DontReverse: true,
	})

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collection := store["Collections"].(Collections)["posts"]
	if len(collection.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(collection.Files))
	}
}

func TestCollectionWithStore(t *testing.T) {
	files := []medusa.File{
		createTestFile(t, "posts/post1.md", "content1", time.Now()),
	}

	store := make(medusa.Store)
	transformer := New(CollectionConfig{
		Name:     "posts",
		Patterns: []string{"posts/*.md"},
		Store: map[string]any{
			"description": "Blog posts",
			"category":    "blog",
		},
	})

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collection := store["Collections"].(Collections)["posts"]
	if collection.Store["description"] != "Blog posts" {
		t.Error("metadata not properly set in collection store")
	}
}

func TestCollectionWithContentInclusion(t *testing.T) {
	files := []medusa.File{
		createTestFile(t, "posts/post1.md", "content1", time.Now()),
	}

	store := make(medusa.Store)
	transformer := New(CollectionConfig{
		Name:           "posts",
		Patterns:       []string{"posts/*.md"},
		IncludeContent: true,
	})

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collection := store["Collections"].(Collections)["posts"]
	if collection.Files[0].Content == nil {
		t.Error("content not included despite IncludeContent being true")
	}
}

func TestCollectionWithMultiplePatterns(t *testing.T) {
	files := []medusa.File{
		createTestFile(t, "posts/post1.md", "content1", time.Now()),
		createTestFile(t, "articles/article1.md", "content2", time.Now()),
		createTestFile(t, "pages/page1.md", "content3", time.Now()),
	}

	store := make(medusa.Store)
	transformer := New(CollectionConfig{
		Name:     "writings",
		Patterns: []string{"posts/*.md", "articles/*.md"},
	})

	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collection := store["Collections"].(Collections)["writings"]
	if len(collection.Files) != 2 {
		t.Errorf("expected 2 files matching multiple patterns, got %d", len(collection.Files))
	}
}

func TestCollectionValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  CollectionConfig
		wantErr error
	}{
		{
			name: "missing name",
			config: CollectionConfig{
				Patterns: []string{"*.md"},
			},
			wantErr: ErrNoName,
		},
		{
			name: "missing patterns",
			config: CollectionConfig{
				Name: "posts",
			},
			wantErr: ErrNoPattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := New(tt.config)
			files := []medusa.File{}
			store := make(medusa.Store)

			err := transformer(&files, &store)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
