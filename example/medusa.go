package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"git.sr.ht/~relay/medusa"
	"git.sr.ht/~relay/medusa/transformers/collections"
	"git.sr.ht/~relay/medusa/transformers/layouts"
	"git.sr.ht/~relay/medusa/transformers/markdown"
	"git.sr.ht/~relay/medusa/transformers/metadata"
)

func main() {
	config := medusa.Config{
		Logger: slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		AllowOverwrite: true,
	}

	b := medusa.NewBuilder(config)
	b.Source("./src")
	b.Destination("./build")

	b.Use(metadata.New(
		map[string]any{
			"Title":       "My Blog",
			"Description": "The blog where I blog.",
		},
	))

	b.Use(collections.New(collections.CollectionConfig{
		Name: "Blog",
		Store: map[string]any{
			"Heading": "My awesome posts!",
		},
		Patterns: []string{"blog/*.md"},
	}))

	b.Use(markdown.New())

	b.Use(layouts.New(layouts.Config{
		LayoutPatterns:  []string{"template/*"},
		ContentPatterns: []string{"*.html"},
	}))

	err := b.Build()
	if err != nil {
		if errors.Is(err, medusa.ErrDestinationExists) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please remove the destination directory or set AllowOverwrite to true.\n")
			os.Exit(1)
		} else {
			log.Fatalf("Build failed: %v", err)
		}
	}

}
