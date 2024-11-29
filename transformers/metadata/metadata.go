package metadata

import "git.sr.ht/~relay/medusa"

func New(metadata map[string]any) medusa.Transformer {
	return func(files *[]medusa.File, store *medusa.Store) error {
		if len(*store) == 0 {
			*store = metadata
		} else {
			for key, value := range metadata {
				(*store)[key] = value
			}
		}
		return nil
	}
}
