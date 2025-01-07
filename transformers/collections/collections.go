package collections

import (
	"errors"
	"path/filepath"
	"slices"

	"git.sr.ht/~relay/medusa"
)

var (
	ErrNoName    = errors.New("no collection name specified")
	ErrNoPattern = errors.New("no collection pattern specified")
)

// Pattern and Name is the only non-optional field
type CollectionConfig struct {
	// The name of the collection.
	Name string
	// Glob pattern for files to add in
	// collection.
	Patterns []string

	// Store defines the data to add
	// the files in the collections's stores
	Store medusa.Store

	// 0 if a == b,
	// -1 if a < b,
	// +1 if a > b.
	SortBy func(a medusa.File, b medusa.File) int

	FilterFunc func(file medusa.File) bool

	// Don't reverse.
	// By default it is false.
	DontReverse bool

	// zero is unlimited
	Limit int

	// Whether to include content in the
	// store of collections
	IncludeContent bool
}

// A file as represented in a [Collection]
type File struct {
	Frontmatter map[string]any

	// often nil
	Content []byte

	// Underlying medusa file
	medusaFile *medusa.File
}

type Collection struct {
	// Yet another store, this time
	// specific to the collection!
	Store medusa.Store

	Files []File
}

// Maps collection name to collection.
type Collections map[string]Collection

func defaultCollectionCfg(cfg *CollectionConfig) error {

	if cfg.Name == "" {
		return ErrNoName
	}
	if len(cfg.Patterns) == 0 {
		return ErrNoPattern
	}
	if cfg.Store == nil {
		cfg.Store = make(map[string]any)
	}
	if cfg.SortBy == nil {
		cfg.SortBy = func(a medusa.File, b medusa.File) int {
			return a.FileInfo.ModTime().Compare(b.FileInfo.ModTime())
		}
	}
	if cfg.FilterFunc == nil {
		cfg.FilterFunc = func(file medusa.File) bool { return true }
	}
	return nil
}

func fileMatchesPatterns(patterns []string, file medusa.File) (bool, error) {
	matchOne := false
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, file.Path)
		if err != nil {
			return false, err
		}
		if match {
			matchOne = true
		}
	}
	return matchOne, nil
}

// Adds a "Collections" key to global store.
// The value is of type [Collections].
func New(collectionCfgs ...CollectionConfig) medusa.Transformer {
	return func(files *[]medusa.File, store *medusa.Store) error {
		collections, ok := (*store)["Collections"].(Collections)

		if !ok {
			collections = Collections{}
		}

		for i := range collectionCfgs {
			var collectionFiles = []File{}
			var contentToInclude []byte

			cfg := collectionCfgs[i]

			err := defaultCollectionCfg(&cfg)
			if err != nil {
				return err
			}

			for _, file := range *files {
				match, err := fileMatchesPatterns(cfg.Patterns, file)
				if err != nil {
					return err
				}
				if !match || !cfg.FilterFunc(file) {
					continue
				}

				if cfg.IncludeContent {
					contentToInclude = file.Content()
				}

				collectionFiles = append(collectionFiles, File{
					Frontmatter: file.Frontmatter,
					Content:     contentToInclude,
					medusaFile:  &file,
				})

			}
			collectionStore := make(medusa.Store)
			for key, value := range cfg.Store {
				collectionStore[key] = value
			}
			slices.SortFunc(collectionFiles, func(a File, b File) int {
				if cfg.DontReverse {
					return cfg.SortBy(*a.medusaFile, *b.medusaFile)
				}
				return -cfg.SortBy(*a.medusaFile, *b.medusaFile)
			})

			collections[cfg.Name] = Collection{
				Store: collectionStore,
				Files: collectionFiles,
			}
		}

		(*store)["Collections"] = collections
		return nil
	}
}
