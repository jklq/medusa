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
if it failes to parse fronmatter of a file.

# Example Usage

	package main

	package main

	import (
		"context"
		"git.sr.ht/~relay/medusa/pkg/medusa"
	)

	func main() {
		b := medusa.NewBuilder()
		b.Source("./src")
		b.Destination("./build")

		err := b.Build(context.Background())
		if err != nil {
			panic(err)
		}
	}
*/
package medusa
