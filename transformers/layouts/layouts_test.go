package layouts

import (
	"os"
	"testing"
	"time"

	"git.sr.ht/~relay/medusa"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		files   []medusa.File
		store   medusa.Store
		wantErr error
	}{
		{
			name: "empty config patterns",
			cfg:  Config{},
			files: []medusa.File{
				{Path: "test.html"},
			},
			wantErr: ErrNoLayoutPattern,
		},
		{
			name: "only layout patterns",
			cfg: Config{
				LayoutPatterns: []string{"*.layout"},
			},
			files: []medusa.File{
				{Path: "test.html"},
			},
			wantErr: ErrNoContentPattern,
		},
		{
			name: "no matching layouts",
			cfg: Config{
				LayoutPatterns:  []string{"*.layout"},
				ContentPatterns: []string{"*.html"},
			},
			files: []medusa.File{
				{Path: "test.html"},
			},
			wantErr: ErrNoLayouts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := New(tt.cfg)
			err := transformer(&tt.files, &tt.store)
			if err != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileMatchesPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		file     medusa.File
		want     bool
		wantErr  bool
	}{
		{
			name:     "matching pattern",
			patterns: []string{"*.html"},
			file:     medusa.File{Path: "test.html"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "non-matching pattern",
			patterns: []string{"*.html"},
			file:     medusa.File{Path: "test.txt"},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "multiple patterns, one match",
			patterns: []string{"*.html", "*.txt"},
			file:     medusa.File{Path: "test.txt"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "invalid pattern",
			patterns: []string{"["},
			file:     medusa.File{Path: "test.txt"},
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fileMatchesPatterns(tt.patterns, tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("fileMatchesPatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fileMatchesPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLayoutTransformation(t *testing.T) {
	defaultLayout := medusa.File{
		Path:     "dodo.html",
		FileInfo: &testFileInfo{name: "dodo.html"},
	}
	defaultLayout.SetContent([]byte("<html>{{.Content}} {{.File.Frontmatter.title}}</html>"))

	contentFile := medusa.File{
		Path:     "content.html",
		FileInfo: &testFileInfo{name: "content.html"},
		Frontmatter: medusa.Store{
			"title":  "Test Title",
			"layout": "dodo.html",
		},
	}
	contentFile.SetContent([]byte(`<h1>Hello</h1>`))

	files := []medusa.File{defaultLayout, contentFile}
	store := medusa.Store{}

	cfg := Config{
		LayoutPatterns:  []string{"dodo.*"},
		ContentPatterns: []string{"*.html"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("Transformation failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file after transformation, got %d", len(files))
	}

	expectedContent := "<html><h1>Hello</h1> Test Title</html>"
	if string(files[0].Content()) != expectedContent {
		t.Errorf("Content transformation failed. Got %q, want %q",
			string(files[0].Content()), expectedContent)
	}
}

// Helper type for testing
type testFileInfo struct {
	name string
}

func (fi *testFileInfo) Name() string       { return fi.name }
func (fi *testFileInfo) Size() int64        { return 0 }
func (fi *testFileInfo) Mode() os.FileMode  { return 0 }
func (fi *testFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *testFileInfo) IsDir() bool        { return false }
func (fi *testFileInfo) Sys() interface{}   { return nil }
