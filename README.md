> **WARNING**: This package is still in it's infancy.
> Breaking features/fixes will be introduced
> and bugs will be prevalent.

# medusa

> a static site generator based on transformer functions

This SSG works by chaining together transformer functions.
It is early in development, so expect: bugs, poor design
choices, and breaking changes.

Concurrency is not implemented at the moment.
I might explore it in the future.

```go
package medusa

// ...imports...

func main() {
	b := medusa.NewBuilder()
	b.Source("./src")
	b.Destination("./build")

	// add values to global store
	b.Use(metadata.New(
		map[string]any{
			"title":       "My Blog",
			"description": "The blog where I blog.",
		},
	))

	// adds "collections" to global store
	b.Use(collections.New(collections.CollectionConfig{
		Name: "blog",
		Store: map[string]any{
			"heading": "My awesome posts!",
		},
		Patterns: []string{"blog/*.md"},
	}))

	// transpile all md to html
	b.Use(markdown.New())

	// apply layouts to content
	b.Use(layouts.New(layouts.Config{
		LayoutPatterns:  []string{"template/*"},
		ContentPatterns: []string{"**/*.html"},
	}))

	// run all transformer functions
	err := b.Build()
	if err != nil {
		panic(err)
	}
}

```

## Metalsmith

This project is heaviliy inspired by and modelled after [Metalsmith](https://github.com/metalsmith/metalsmith).

## Support

Please feel free to send an email to ~relay/medusa@lists.sr.ht to submit patches or ask questions.
