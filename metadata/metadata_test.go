package metadata

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDocumentLeadingH1(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	filename = "testdata/header.md"

	markdown, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	actual := ExtractDocumentLeadingH1(markdown)

	assert.Equal(t, "a", actual)
}

func TestExtractDocumentLeadingH1OnlyMatchesRealHeadings(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{"shebang is not a heading", "#!/usr/bin/env node\nconsole.log(1)\n", ""},
		{"mid-line hash is not a heading", "fixes issue #12 in prod\n# Actual Title\n", "Actual Title"},
		{"no space after hash is not a heading", "#nope\n# Yes\n", "Yes"},
		{"h2 is not an h1", "## Subheading\nbody\n", ""},
		{"heading without trailing newline", "# Title", "Title"},
		{"trailing spaces are trimmed", "# Title  \n", "Title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ExtractDocumentLeadingH1([]byte(tt.markdown)))
		})
	}
}

func TestSetTitleFromFilename(t *testing.T) {
	t.Run("set title from filename", func(t *testing.T) {
		meta := &Meta{Title: ""}
		setTitleFromFilename(meta, "/path/to/test.md")
		assert.Equal(t, "Test", meta.Title)
	})

	t.Run("replace underscores with spaces", func(t *testing.T) {
		meta := &Meta{Title: ""}
		setTitleFromFilename(meta, "/path/to/test_with_underscores.md")
		assert.Equal(t, "Test With Underscores", meta.Title)
	})

	t.Run("replace dashes with spaces", func(t *testing.T) {
		meta := &Meta{Title: ""}
		setTitleFromFilename(meta, "/path/to/test-with-dashes.md")
		assert.Equal(t, "Test With Dashes", meta.Title)
	})

	t.Run("mixed underscores and dashes", func(t *testing.T) {
		meta := &Meta{Title: ""}
		setTitleFromFilename(meta, "/path/to/test_with-mixed_separators.md")
		assert.Equal(t, "Test With Mixed Separators", meta.Title)
	})

	t.Run("already title cased", func(t *testing.T) {
		meta := &Meta{Title: ""}
		setTitleFromFilename(meta, "/path/to/Already-Title-Cased.md")
		assert.Equal(t, "Already Title Cased", meta.Title)
	})
}

func TestExtractMetaOwner(t *testing.T) {
	t.Run("owner header is captured", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n<!-- Owner: 712020:35f9ab2f-1111-2222-3333-6da6d57a51e9 -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, "")
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, "712020:35f9ab2f-1111-2222-3333-6da6d57a51e9", meta.Owner)
	})

	t.Run("owner defaults to empty", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, "")
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, "", meta.Owner)
	})
}

func TestExtractMetaContentAppearance(t *testing.T) {
	t.Run("default fills missing content appearance", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, FixedContentAppearance)
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, FixedContentAppearance, meta.ContentAppearance)
	})

	t.Run("header takes precedence over default", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n<!-- Content-Appearance: full-width -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, FixedContentAppearance)
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, FullWidthContentAppearance, meta.ContentAppearance)
	})

	t.Run("falls back to full-width when default isn't set", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, "")
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, FullWidthContentAppearance, meta.ContentAppearance)
	})

	t.Run("default appearance via cli flag", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, DefaultContentAppearance)
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, DefaultContentAppearance, meta.ContentAppearance)
	})

	t.Run("default appearance via header", func(t *testing.T) {
		data := []byte("<!-- Space: DOC -->\n<!-- Title: Example -->\n<!-- Content-Appearance: default -->\n\nbody\n")

		meta, _, err := ExtractMeta(data, "", false, false, "", nil, false, "")
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, DefaultContentAppearance, meta.ContentAppearance)
	})
}
