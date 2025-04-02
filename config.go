package medusa

import (
	"log/slog"
)

type Config struct {
	// Defines the working directory.
	// It is used to find the source and
	// destination directory.
	//
	// Optional. Defaults to the current directory ("./") if empty.
	WorkingDir string

	// Whether prompts like
	// "overwrite warnings" should be skipped.
	//
	// Optional. Defaults to false.
	AllowOverwrite bool

	// Defines which logger to use.
	//
	// Optional. Defaults to a discard logger if nil.
	Logger *slog.Logger

	// Whether frontmatter parsing should be skipped.
	//
	// Optional. Defaults to false.
	SkipFrontmatterParsing bool
}

// Helper function to set default values
// DEPRECATED: Defaults are now handled directly in NewBuilder using Go's
// zero values where appropriate.
// func configDefault(optionalConfig ...Config) Config {
// 	var config Config

// 	if len(optionalConfig) == 0 {
// 		config = Config{}
// 	} else {
// 		config = optionalConfig[0]
// 	}

// 	if config.Logger == nil {
// 		config.Logger = slog.New(discardHandler{})
// 	}

// 	// Set WorkingDir default value
// 	if config.WorkingDir == "" {
// 		config.WorkingDir = "./"
// 	}

// 	return config
// }
