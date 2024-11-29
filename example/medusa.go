package main

import (
	"log"
	"log/slog"

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
		AutoConfirm: true,
	}

	b := medusa.NewBuilder(config)
	b.Source("./src")
	b.Destination("./build")

	b.Use(metadata.New(
		map[string]any{
			"title":       "My Blog",
			"description": "The blog where I blog.",
		},
	))

	b.Use(collections.New(collections.CollectionConfig{
		Name: "blog",
		Store: map[string]any{
			"heading": "My awesome posts!",
		},
		Patterns: []string{"blog/*.md"},
	}))

	b.Use(markdown.New())

	b.Use(layouts.New(layouts.Config{
		LayoutPatterns:  []string{"template/*"},
		ContentPatterns: []string{"**/*.html"},
	}))

	err := b.Build()
	if err != nil {
		panic(err)
	}
}
