/*
Package layouts is a medusa transformer that renders layouts.

It finds layout files and content files based on glob patterns,
and renderes the content files based on a layout as defined by
the "layout" key in the frontmatter. If the layout key is not
present, it will use a file that start with "default",
if that isn't found it will just pick a random layout to be
the default.

It assigns the whole file path (including extension) as layout name.
This is of course relative to the source directory.

Warning! This package trusts the content, and does not escape it.
Documentation (and package as a whole) is not very developed.
*/
package layouts
