/*
Pacakge medusa is a very simple static site generator that works by
chaining functions together. When [Builder.Build] is called,
It reads all the files in the source directory and represents
them all as a slice of [File]. A pointer to this slice, along with
a pointer to the global [Store], is passed to every [Transformer]
sequencially. At the end of the chain, it writes the state of the
[File] slice to the destination.

Each file also has a Frontmatter field where yaml/toml/json
frontmatter is parsed and stored. The builder returns [ErrFrontmatter]
if it failes to parse fronmatter of a file. Frontmatter parsing can
be skipped via [Config].

Configuration defaults generally rely on Go's zero values (e.g., false for
booleans, nil for pointers, "" for strings), with specific overrides in
[NewBuilder] for fields like WorkingDir (defaults to "./") and Logger
(defaults to a discard logger).

# Example Usage

	package main

	import (
		"log"
		"log/slog"
		"os"

		"git.sr.ht/~relay/medusa"
	)

	func main() {
		// Example with custom configuration
		config := medusa.Config{
			Logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
			AllowOverwrite: true, // Skip overwrite prompts
		}
		b := medusa.NewBuilder(config)

		// Example with default configuration
		// b := medusa.NewBuilder()

		b.Source("./src")
		b.Destination("./build")

		// Add transformers if needed
		// b.Use(...)

		err := b.Build()
		if err != nil {
			panic(err)
		}
	}
*/
package medusa
