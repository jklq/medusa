package medusa

import (
	"fmt"
	"os"
	"path/filepath"

	"log/slog"
)

type Builder struct {
	store Store

	workingDir      string
	source          string
	destination     string
	transformers    []Transformer
	files           []File
	log             slog.Logger
	skipFrontmatter bool

	autoConfirm bool
}

// A function that change files and a
// global store as a part of a larger chain.
type Transformer func(files *[]File, store *Store) error

// Used by transformers to hold arbitrary data.
type Store map[string]any

// Returns empty transformer with error
func ErrTransformer(err error) Transformer {
	return func(files *[]File, store *Store) error {
		return err
	}
}

// Creates a new builder struct.
// The config parameter is optional.
// If more than one configs are passed, it uses the first one.
func NewBuilder(optionalConfig ...Config) *Builder {
	config := configDefault(optionalConfig...)

	return &Builder{
		workingDir:      config.WorkingDir,
		log:             *config.Logger,
		skipFrontmatter: config.SkipFrontmatterParsing,
		autoConfirm:     config.AutoConfirm,
	}
}

// Defines the source directory, relative to WorkingDir
// as specified in [Config].
func (b *Builder) Source(source string) {
	source = filepath.Join(b.workingDir, source)
	b.source = source
}

// Defines the destination directory, relative to WorkingDir
// as specified in [Config].
func (b *Builder) Destination(destination string) {
	destination = filepath.Join(b.workingDir, destination)
	b.destination = destination
}

// Adds a transformer function to the stack.
func (b *Builder) Use(transformer Transformer) {
	b.transformers = append(b.transformers, transformer)
}

// Build applies the transformers in the stack to the contents of
// every file in the source directory, and writes them to the
// destination.
func (b *Builder) Build() error {
	b.store = make(Store)

	err := checkSourceAndDestination(b.source, b.destination)
	if err != nil {
		return err
	}

	err = prepareDestination(b.destination, b.autoConfirm)
	if err != nil {
		return err
	}

	// Access file tree and run file transformers
	err = filepath.WalkDir(b.source, b.srcWalker)
	if err != nil {
		return err
	}

	for _, transformer := range b.transformers {
		err := transformer(&b.files, &b.store)
		if err != nil {
			return err
		}
	}

	// Build to destination
	err = writeFiles(b.destination, b.files)
	if err != nil {
		return err
	}

	return nil
}

func writeFiles(destination string, files []File) error {
	for _, file := range files {
		writePath := filepath.Join(destination, file.Path)

		err := os.MkdirAll(filepath.Dir(writePath), 0755)
		if err != nil {
			return err
		}

		err = os.WriteFile(writePath, file.content, file.FileInfo.Mode().Perm())
		if err != nil {
			return err
		}

	}
	return nil
}

func checkSourceAndDestination(source string, destination string) error {
	if destination == "" {
		return fmt.Errorf("destination directory not defined")
	}
	if source == "" {
		return fmt.Errorf("source directory not defined")
	}
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist")
	}
	return nil
}

func prepareDestination(destination string, autoconfirm bool) error {
	if !autoconfirm {
		if _, err := os.Stat(destination); err == nil {
			var answer string
			fmt.Fprintf(os.Stderr,
				"Destination directory \"%v\" exists. Delete and proceed? (y/n)",
				destination)

			fmt.Scanf("%v", &answer)
			if answer != "y" {
				os.Exit(0)
			}
		}
	}
	err := os.RemoveAll(destination)
	if err != nil {
		return err
	}
	err = os.MkdirAll(destination, 0755)
	if err != nil {
		return err
	}
	return nil
}
