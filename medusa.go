package medusa

import (
	"errors" // Added for errors.Is and defining new error types
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time" // Import time for duration calculation
)

// ErrDestinationExists indicates that the destination directory exists
// and overwriting was not explicitly allowed via Config.AllowOverwrite.
var ErrDestinationExists = errors.New("destination directory exists and overwrite not permitted")

type Builder struct {
	store Store

	workingDir      string
	source          string
	destination     string
	transformers    []Transformer
	files           []File
	log             *slog.Logger
	skipFrontmatter bool

	autoConfirm bool // Represents if overwriting the destination is allowed
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
// If more than one config is passed, it uses the first one.
// Defaults are applied based on Go's zero values where applicable:
// - WorkingDir: Defaults to "./" if empty.
// - AllowOverwrite: Defaults to false. If true, allows overwriting the destination.
// - Logger: Defaults to a discard logger if nil.
// - SkipFrontmatterParsing: Defaults to false.
func NewBuilder(optionalConfig ...Config) *Builder {
	var config Config
	if len(optionalConfig) > 0 {
		config = optionalConfig[0]
	}

	if config.WorkingDir == "" {
		config.WorkingDir = "./"
	}
	if config.Logger == nil {
		config.Logger = slog.New(discardHandler{})
	}

	logger := config.Logger.With("component", "medusa_builder")

	logger.Debug("Initializing new builder")

	return &Builder{
		workingDir:      config.WorkingDir,
		log:             logger,
		skipFrontmatter: config.SkipFrontmatterParsing,
		autoConfirm:     config.AllowOverwrite,
		store:           make(Store),
	}
}

// Defines the source directory, relative to WorkingDir
// as specified in [Config].
func (b *Builder) Source(source string) {
	absSource := filepath.Join(b.workingDir, source)
	b.log.Debug("Setting source directory", "raw", source, "absolute", absSource)
	b.source = absSource
}

// Defines the destination directory, relative to WorkingDir
// as specified in [Config].
func (b *Builder) Destination(destination string) {
	absDest := filepath.Join(b.workingDir, destination)
	b.log.Debug("Setting destination directory", "raw", destination, "absolute", absDest)
	b.destination = absDest
}

// Adds a transformer function to the stack.
func (b *Builder) Use(transformer Transformer) {
	b.log.Debug("Adding transformer", "current_count", len(b.transformers))
	b.transformers = append(b.transformers, transformer)
}

// Build applies the transformers in the stack to the contents of
// every file in the source directory, and writes them to the
// destination. It returns ErrDestinationExists if the destination
// directory exists and Config.AllowOverwrite is false.
func (b *Builder) Build() error {
	startTime := time.Now()
	b.log.Info("Build process started")
	b.log.Debug("Effective configuration",
		"source", b.source,
		"destination", b.destination,
		"allow_overwrite", b.autoConfirm,
		"skip_frontmatter", b.skipFrontmatter,
		"working_dir", b.workingDir,
	)

	err := b.checkSourceAndDestination()
	if err != nil {
		b.log.Error("Source/Destination check failed", "error", err)
		return err
	}

	err = b.prepareDestination()
	if err != nil {
		return err
	}

	b.log.Info("Walking source directory", "source", b.source)
	walkStart := time.Now()
	err = filepath.WalkDir(b.source, b.srcWalker)
	walkDuration := time.Since(walkStart)
	if err != nil {
		b.log.Error("Failed during source walk", "source", b.source, "error", err)
		return fmt.Errorf("error walking source directory %s: %w", b.source, err)
	}
	b.log.Info("Finished walking source directory", "files_found", len(b.files), "duration", walkDuration)

	b.log.Info("Applying transformers", "count", len(b.transformers))
	transformStart := time.Now()
	for i, transformer := range b.transformers {
		tfStartTime := time.Now()
		b.log.Debug("Executing transformer", "index", i)
		err := transformer(&b.files, &b.store)
		tfDuration := time.Since(tfStartTime)
		if err != nil {
			b.log.Error("Transformer failed", "index", i, "duration", tfDuration, "error", err)
			return fmt.Errorf("transformer at index %d failed: %w", i, err)
		}
		b.log.Debug("Finished transformer", "index", i, "duration", tfDuration)
	}
	transformDuration := time.Since(transformStart)
	b.log.Info("Finished applying all transformers", "count", len(b.transformers), "duration", transformDuration)

	b.log.Info("Writing files to destination", "destination", b.destination, "count", len(b.files))
	writeStart := time.Now()
	err = b.writeFiles()
	writeDuration := time.Since(writeStart)
	if err != nil {
		b.log.Error("Failed writing files to destination", "destination", b.destination, "duration", writeDuration, "error", err)
		return err
	}
	buildDuration := time.Since(startTime)
	b.log.Info("Build successful",
		"destination", b.destination,
		"files_written", len(b.files),
		"total_duration", buildDuration,
	)

	return nil
}

func (b *Builder) writeFiles() error {
	b.log.Debug("Starting file writing process", "count", len(b.files))
	for i, file := range b.files {
		writePath := filepath.Join(b.destination, file.Path)
		writeDir := filepath.Dir(writePath)

		b.log.Debug("Preparing to write file", "index", i, "target_path", writePath)

		err := os.MkdirAll(writeDir, 0755)
		if err != nil {
			b.log.Error("Failed to create directory for file", "directory", writeDir, "file_path", writePath, "error", err)
			return fmt.Errorf("failed to create directory %s: %w", writeDir, err)
		}

		err = os.WriteFile(writePath, file.content, file.FileInfo.Mode().Perm())
		if err != nil {
			b.log.Error("Failed to write file content", "file_path", writePath, "error", err)
			return fmt.Errorf("failed to write file %s: %w", writePath, err)
		}
		b.log.Debug("Successfully wrote file", "index", i, "path", writePath, "size", len(file.content))
	}
	b.log.Debug("Finished file writing process")
	return nil
}

func (b *Builder) checkSourceAndDestination() error {
	b.log.Debug("Checking source and destination paths")
	if b.destination == "" {
		err := fmt.Errorf("destination directory not defined")
		b.log.Error("Validation failed", "reason", err)
		return err
	}
	if b.source == "" {
		err := fmt.Errorf("source directory not defined")
		b.log.Error("Validation failed", "reason", err)
		return err
	}

	b.log.Debug("Checking source path existence and type", "source", b.source)
	sourceInfo, err := os.Stat(b.source)
	if os.IsNotExist(err) {
		err = fmt.Errorf("source directory '%s' does not exist", b.source)
		b.log.Error("Validation failed", "reason", err, "source", b.source)
		return err
	} else if err != nil {
		err = fmt.Errorf("failed to stat source directory '%s': %w", b.source, err)
		b.log.Error("Validation failed", "reason", "stat error", "source", b.source, "error", err)
		return err
	}
	if !sourceInfo.IsDir() {
		err = fmt.Errorf("source path '%s' is not a directory", b.source)
		b.log.Error("Validation failed", "reason", "source not a directory", "source", b.source)
		return err
	}
	b.log.Debug("Source and destination paths validated successfully")
	return nil
}

func (b *Builder) prepareDestination() error {
	b.log.Debug("Preparing destination directory", "destination", b.destination)
	_, err := os.Stat(b.destination)

	if err == nil {
		b.log.Debug("Destination directory exists", "destination", b.destination)
		if !b.autoConfirm {
			b.log.Warn("Destination exists but overwrite not permitted", "destination", b.destination)
			return fmt.Errorf("%w: %s", ErrDestinationExists, b.destination)
		}

		b.log.Info("Removing existing destination directory", "destination", b.destination)
		removeStart := time.Now()
		err = os.RemoveAll(b.destination)
		removeDuration := time.Since(removeStart)
		if err != nil {
			b.log.Error("Failed to remove existing destination directory", "destination", b.destination, "duration", removeDuration, "error", err)
			return fmt.Errorf("failed to remove existing destination directory %s: %w", b.destination, err)
		}
		b.log.Debug("Existing destination directory removed", "destination", b.destination, "duration", removeDuration)

	} else if !os.IsNotExist(err) {
		b.log.Error("Failed to check destination directory status", "destination", b.destination, "error", err)
		return fmt.Errorf("failed to check destination directory %s: %w", b.destination, err)
	} else {
		b.log.Debug("Destination directory does not exist, will create.", "destination", b.destination)
	}

	b.log.Debug("Creating destination directory", "destination", b.destination)
	createStart := time.Now()
	err = os.MkdirAll(b.destination, 0755)
	createDuration := time.Since(createStart)
	if err != nil {
		b.log.Error("Failed to create destination directory", "destination", b.destination, "duration", createDuration, "error", err)
		return fmt.Errorf("failed to create destination directory %s: %w", b.destination, err)
	}
	b.log.Info("Destination directory prepared", "destination", b.destination, "duration", createDuration)
	return nil
}
