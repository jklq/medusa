package layouts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~relay/medusa"
)

// Helper type for testing FileInfo
type testFileInfo struct {
	name    string
	isDir   bool
	modTime time.Time
	mode    os.FileMode
	size    int64
}

func (fi *testFileInfo) Name() string       { return fi.name }
func (fi *testFileInfo) Size() int64        { return fi.size }
func (fi *testFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *testFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *testFileInfo) IsDir() bool        { return fi.isDir }
func (fi *testFileInfo) Sys() interface{}   { return nil }

// Helper to create simple files for testing
func makeFile(path string, content string, fm medusa.Store) medusa.File {
	f := medusa.File{
		Path:        path,
		FileInfo:    &testFileInfo{name: filepath.Base(path)},
		Frontmatter: fm,
	}
	f.SetContent([]byte(content))
	return f
}

// Helper to sort files by path for consistent comparison
func sortFiles(files []medusa.File) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func TestNew_ConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		files   []medusa.File
		store   medusa.Store
		wantErr error
	}{
		{
			name:    "empty config patterns",
			cfg:     Config{},
			files:   []medusa.File{makeFile("test.html", "content", nil)},
			wantErr: ErrNoLayoutPattern,
		},
		{
			name: "only layout patterns",
			cfg: Config{
				LayoutPatterns: []string{"*.layout"},
			},
			files:   []medusa.File{makeFile("test.html", "content", nil)},
			wantErr: ErrNoContentPattern,
		},
		{
			name: "no matching layouts but content needs one",
			cfg: Config{
				LayoutPatterns:  []string{"*.layout"},
				ContentPatterns: []string{"*.html"},
			},
			files:   []medusa.File{makeFile("test.html", "content", nil)}, // Needs default layout
			wantErr: ErrNoLayouts,
		},
		{
			name: "no matching layouts but content specifies one",
			cfg: Config{
				LayoutPatterns:  []string{"*.layout"},
				ContentPatterns: []string{"*.html"},
			},
			files: []medusa.File{
				makeFile("test.html", "content", medusa.Store{"layout": "missing.layout"}),
			},
			wantErr: ErrNoLayouts, // Because the specified layout wasn't found among potential layouts
		},
		{
			name: "no layouts and no content files",
			cfg: Config{
				LayoutPatterns:  []string{"*.layout"},
				ContentPatterns: []string{"*.html"},
			},
			files:   []medusa.File{makeFile("other.txt", "stuff", nil)}, // No content files match
			wantErr: nil,                                                // Should succeed without doing anything
		},
		{
			name: "no layouts but content files don't need them",
			cfg: Config{
				LayoutPatterns:  []string{"*.layout"},
				ContentPatterns: []string{"*.md"}, // No *.md files provided
			},
			files:   []medusa.File{makeFile("index.html", "content", nil)},
			wantErr: nil, // Should succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := New(tt.cfg)
			// Make copies to avoid modifying test case data between runs
			filesCopy := make([]medusa.File, len(tt.files))
			copy(filesCopy, tt.files)
			storeCopy := make(medusa.Store)
			for k, v := range tt.store {
				storeCopy[k] = v
			}

			err := transformer(&filesCopy, &storeCopy)

			// Use errors.Is for checking specific error types
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLayoutTransformation_Basic(t *testing.T) {
	layoutFile := makeFile("layouts/base.html", "<html><body>{{.Content}} | {{.File.Frontmatter.title}}</body></html>", nil)
	contentFile := makeFile("content.html", "<h1>Hello</h1>", medusa.Store{
		"title":  "Test Title",
		"layout": "layouts/base.html",
	})
	otherFile := makeFile("styles.css", "body { color: red; }", nil) // Should be ignored

	files := []medusa.File{layoutFile, contentFile, otherFile}
	store := medusa.Store{"globalVar": "Global Value"}

	cfg := Config{
		LayoutPatterns:  []string{"layouts/*.html"},
		ContentPatterns: []string{"*.html"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("Transformation failed: %v", err)
	}

	// Expecting content.html (processed) and styles.css (ignored)
	if len(files) != 2 {
		t.Fatalf("Expected 2 files after transformation, got %d", len(files))
	}

	sortFiles(files) // Ensure consistent order for checking

	// Check ignored file
	if files[1].Path != "styles.css" || string(files[1].Content()) != "body { color: red; }" {
		t.Errorf("Ignored file was modified or removed. Got: %+v", files[1])
	}

	// Check processed file
	processedFile := files[0]
	if processedFile.Path != "content.html" {
		t.Errorf("Expected processed file path 'content.html', got %s", processedFile.Path)
	}

	expectedContent := "<html><body><h1>Hello</h1> | Test Title</body></html>"
	if string(processedFile.Content()) != expectedContent {
		t.Errorf("Content transformation failed.\nGot:  %q\nWant: %q",
			string(processedFile.Content()), expectedContent)
	}
}

func TestLayoutTransformation_DefaultLayout(t *testing.T) {
	defaultLayout := makeFile("layouts/default.html", "DEFAULT: {{.Content}}", nil)
	otherLayout := makeFile("layouts/other.html", "OTHER: {{.Content}}", nil)
	contentFile := makeFile("page.md", "Markdown Content", nil) // No layout specified

	files := []medusa.File{otherLayout, defaultLayout, contentFile} // Order shouldn't matter
	store := medusa.Store{}

	cfg := Config{
		LayoutPatterns:  []string{"layouts/*.html"},
		ContentPatterns: []string{"*.md"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("Transformation failed: %v", err)
	}

	if len(files) != 1 {
		fmt.Println(files)
		t.Fatalf("Expected 1 file after transformation, got %d", len(files))
	}

	expectedContent := "DEFAULT: Markdown Content"
	if string(files[0].Content()) != expectedContent {
		t.Errorf("Default layout not applied correctly.\nGot:  %q\nWant: %q",
			string(files[0].Content()), expectedContent)
	}
}

func TestLayoutTransformation_FallbackDefaultLayout(t *testing.T) {
	layout1 := makeFile("layouts/first.html", "FIRST: {{.Content}}", nil)
	layout2 := makeFile("layouts/second.html", "SECOND: {{.Content}}", nil) // Should be picked as default
	contentFile := makeFile("page.md", "Markdown Content", nil)

	// Order matters here for the fallback logic (last parsed)
	files := []medusa.File{layout1, layout2, contentFile}
	store := medusa.Store{}

	cfg := Config{
		LayoutPatterns:  []string{"layouts/*.html"},
		ContentPatterns: []string{"*.md"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("Transformation failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file after transformation, got %d", len(files))
	}

	// Expecting the *last* parsed layout to be used as default
	expectedContent := "SECOND: Markdown Content"
	if string(files[0].Content()) != expectedContent {
		t.Errorf("Fallback default layout not applied correctly.\nGot:  %q\nWant: %q",
			string(files[0].Content()), expectedContent)
	}
}

func TestLayoutTransformation_WithPartials(t *testing.T) {
	headerPartial := makeFile("partials/header.html", "<head><title>{{ .File.Frontmatter.title }}</title></head>", nil)
	footerPartial := makeFile("partials/footer.html", "<footer>Copyright {{ .Global.year }}</footer>", nil)
	baseLayout := makeFile(
		"layouts/base.layout",
		`<!DOCTYPE html>
<html>
{{ template "partials/header.html" . }}
<body>
  <h1>Base Layout</h1>
  <main>{{ .Content }}</main>
  {{ template "partials/footer.html" . }}
</body>
</html>`,
		nil,
	)
	contentFile := makeFile("index.md", "## Page Content", medusa.Store{
		"title":  "My Site",
		"layout": "layouts/base.layout",
	})

	files := []medusa.File{headerPartial, footerPartial, baseLayout, contentFile}
	store := medusa.Store{"year": 2024}

	cfg := Config{
		// Include partials in layout patterns
		LayoutPatterns:  []string{"layouts/*", "partials/*"},
		ContentPatterns: []string{"*.md"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("Transformation with partials failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file after transformation, got %d", len(files))
	}

	expected := `<!DOCTYPE html>
<html>
<head><title>My Site</title></head>
<body>
  <h1>Base Layout</h1>
  <main>## Page Content</main>
  <footer>Copyright 2024</footer>
</body>
</html>`
	got := string(files[0].Content())

	// Normalize whitespace slightly for comparison if needed, though template output should be consistent
	expected = strings.Join(strings.Fields(expected), " ")
	got = strings.Join(strings.Fields(got), " ")

	if got != expected {
		t.Errorf("Transformation with partials failed.\nGot:  %q\nWant: %q", got, expected)
	}
}

func TestLayoutTransformation_LayoutNotFound(t *testing.T) {
	layoutFile := makeFile("layouts/exists.html", "Exists: {{.Content}}", nil)
	contentFile := makeFile("page.md", "Content", medusa.Store{"layout": "layouts/missing.html"})

	files := []medusa.File{layoutFile, contentFile}
	store := medusa.Store{}

	cfg := Config{
		LayoutPatterns:  []string{"layouts/*.html"},
		ContentPatterns: []string{"*.md"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)

	// Check for the specific error type
	var wantErr ErrLayoutNotFound
	if !errors.As(err, &wantErr) {
		t.Fatalf("Expected ErrLayoutNotFound, got: %v", err)
	}

	// Optionally check the details of the error
	if wantErr.layout != "layouts/missing.html" || wantErr.path != "page.md" {
		t.Errorf("ErrLayoutNotFound details incorrect. Got: %+v", wantErr)
	}
}

func TestLayoutTransformation_InvalidLayoutNameType(t *testing.T) {
	layoutFile := makeFile("layouts/default.html", "Default: {{.Content}}", nil)
	// Frontmatter layout value is not a string
	contentFile := makeFile("page.md", "Content", medusa.Store{"layout": 123})

	files := []medusa.File{layoutFile, contentFile}
	store := medusa.Store{}

	cfg := Config{
		LayoutPatterns:  []string{"layouts/*.html"},
		ContentPatterns: []string{"*.md"},
	}

	transformer := New(cfg)
	err := transformer(&files, &store)

	// Check for the specific error type
	var wantErr ErrInvalidLayoutName
	if !errors.As(err, &wantErr) {
		t.Fatalf("Expected ErrInvalidLayoutName, got: %v", err)
	}
	if wantErr.path != "page.md" {
		t.Errorf("ErrInvalidLayoutName path incorrect. Got: %s", wantErr.path)
	}
}

func TestLayoutTransformation_KeepsOtherFiles(t *testing.T) {
	layoutFile := makeFile("layouts/base.html", "Layout: {{.Content}}", nil)
	contentFile := makeFile("index.html", "Index Content", medusa.Store{"layout": "layouts/base.html"})
	scriptFile := makeFile("script.js", "console.log('hello');", nil)
	dataFile := makeFile("data.json", `{"key": "value"}`, nil)

	files := []medusa.File{layoutFile, contentFile, scriptFile, dataFile}
	store := medusa.Store{}

	cfg := Config{
		LayoutPatterns:  []string{"layouts/*.html"},
		ContentPatterns: []string{"*.html"}, // Only process .html files
	}

	transformer := New(cfg)
	err := transformer(&files, &store)
	if err != nil {
		t.Fatalf("Transformation failed: %v", err)
	}

	// Expected files: index.html (processed), script.js, data.json
	expectedPaths := []string{"index.html", "script.js", "data.json"}
	if len(files) != len(expectedPaths) {
		fmt.Println(files)
		t.Fatalf("Expected %d files, got %d", len(expectedPaths), len(files))
	}

	sortFiles(files) // Sort by path for consistent checking

	gotPaths := make([]string, len(files))
	for i, f := range files {
		gotPaths[i] = f.Path
	}
	sort.Strings(expectedPaths) // Ensure expected paths are sorted too

	if !reflect.DeepEqual(gotPaths, expectedPaths) {
		t.Errorf("Expected file paths %v, got %v", expectedPaths, gotPaths)
	}

	// Quick check on content of processed and one non-processed file
	if string(files[1].Content()) != "Layout: Index Content" { // index.html
		t.Errorf("Processed file content incorrect")
	}
	if string(files[2].Content()) != "console.log('hello');" { // script.js
		t.Errorf("Non-processed file content incorrect")
	}
}
