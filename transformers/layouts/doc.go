/*
Package layouts provides a Medusa transformer for applying Go html/template
layouts and partials to content files.

This transformer allows you to wrap your content (e.g., rendered Markdown)
within consistent site structures defined by layout templates (like base.html)
and reuse common UI elements (like headers, footers) via partial templates.

# Configuration

Use the `LayoutPatterns` option to specify glob patterns matching your layout
and partial template files (e.g., "layouts/*.html", "partials/**\/*.html").
Use the `ContentPatterns` option to specify glob patterns matching the content
files that should have layouts applied (e.g., "**\/*.md", "**\/*.html").

# Layout Selection

For each content file:
 1. If the file's frontmatter contains a `layout` key (e.g., `layout: "layouts/post.html"`),
    the specified template will be used. The path should be relative to the source root.
 2. Otherwise, the transformer looks for a template file whose name starts with
    "default." (e.g., "layouts/default.html") found via `LayoutPatterns`.
 3. If neither is found, a fallback layout is used (typically the last template parsed).

# Template Data Context

Within your layout and partial templates, you have access to the following data:
  - `{{ .File }}`: The medusa.File struct for the content file being processed.
    Access frontmatter via `{{ .File.Frontmatter.YourKey }}`.
  - `{{ .Global }}`: The global medusa.Store, containing site-wide data.
  - `{{ .Content }}`: The pre-rendered content of the input file, available as
    `template.HTML`. This is typically the output of a preceding transformer
    (like a Markdown renderer).

# Including Partials

Use the standard Go template action to include partials within layouts or other partials:

	{{ template "partials/header.html" . }}

The "." passes the current data context (`.File`, `.Global`, `.Content`) to the partial.

# File Handling

  - Files matching `LayoutPatterns` are consumed by this transformer and do not
    appear in the output file list.
  - Files matching `ContentPatterns` are processed, and their content is replaced
    with the result of rendering the chosen layout.
  - Files matching neither pattern set are passed through unmodified.

# Security Note

The content rendered via `{{ .Content }}` is treated as trusted HTML (`template.HTML`).
Ensure that any preceding transformers (e.g., Markdown rendering) produce safe HTML,
especially if processing user-generated or untrusted content. Sanitize content
beforehand if necessary.
*/
package layouts
