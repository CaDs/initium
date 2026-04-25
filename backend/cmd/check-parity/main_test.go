package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// check-parity is the gate that prevents the spec from drifting away
// from clients. The orphan check is the load-bearing logic — these
// tests cover the path through `check()` end-to-end with synthetic
// trees so the test doesn't depend on the real codebase's state.

const minimalSpec = `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
paths:
  /api/landing:
    get:
      responses:
        '200':
          description: ok
  /api/notes:
    get:
      responses:
        '200':
          description: ok
  /api/notes/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
  /healthz:
    get:
      responses:
        '200':
          description: ok
`

// writeSpec writes the fixture spec to a temp file and returns its path.
func writeSpec(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// writeFile creates a file at the given path inside dir, including any
// parent directories.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	full := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o600))
}

func TestCheck_AllPathsReferenced_ReturnsNoOrphans(t *testing.T) {
	t.Parallel()

	specPath := writeSpec(t, minimalSpec)

	webDir := t.TempDir()
	writeFile(t, webDir, "src/api.ts",
		`fetch("/api/landing"); fetch("/api/notes");`)

	orphans, err := check(specPath, []treeSpec{
		{label: "web", root: webDir, suffixes: []string{".ts"}},
	})

	require.NoError(t, err)
	assert.Empty(t, orphans,
		"all spec paths are referenced in a client tree (or excluded), so no orphans expected")
}

func TestCheck_OrphanPath_ReportsIt(t *testing.T) {
	t.Parallel()

	specPath := writeSpec(t, minimalSpec)

	webDir := t.TempDir()
	// Only references /api/landing — /api/notes is orphaned.
	writeFile(t, webDir, "src/api.ts", `fetch("/api/landing");`)

	orphans, err := check(specPath, []treeSpec{
		{label: "web", root: webDir, suffixes: []string{".ts"}},
	})

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"/api/notes", "/api/notes/{id}"}, orphans,
		"the unreferenced API paths must be reported as orphans")
}

func TestCheck_ExcludedPaths_NotReported(t *testing.T) {
	t.Parallel()

	specPath := writeSpec(t, minimalSpec)

	webDir := t.TempDir()
	writeFile(t, webDir, "src/api.ts",
		`fetch("/api/landing"); fetch("/api/notes");`)

	// /healthz is in `excluded` — it must not be flagged even though
	// no client references it.
	orphans, err := check(specPath, []treeSpec{
		{label: "web", root: webDir, suffixes: []string{".ts"}},
	})

	require.NoError(t, err)
	for _, o := range orphans {
		assert.NotEqual(t, "/healthz", o,
			"/healthz is in `excluded` and must not be reported")
	}
}

func TestCheck_ParametrizedPath_MatchedByStaticPrefix(t *testing.T) {
	t.Parallel()

	specPath := writeSpec(t, minimalSpec)

	webDir := t.TempDir()
	// /api/notes/{id} is checked by its static prefix /api/notes,
	// which appears in the source — so it must NOT be orphaned.
	writeFile(t, webDir, "src/api.ts",
		`fetch("/api/landing"); fetch("/api/notes/" + id);`)

	orphans, err := check(specPath, []treeSpec{
		{label: "web", root: webDir, suffixes: []string{".ts"}},
	})

	require.NoError(t, err)
	assert.NotContains(t, orphans, "/api/notes/{id}",
		"path with {id} suffix should match by its static prefix /api/notes")
}

func TestCheck_MissingClientTree_TolerantOfMissingDirs(t *testing.T) {
	t.Parallel()

	specPath := writeSpec(t, minimalSpec)

	webDir := t.TempDir()
	writeFile(t, webDir, "src/api.ts",
		`fetch("/api/landing"); fetch("/api/notes");`)

	// One tree exists, one doesn't — check() should still succeed.
	orphans, err := check(specPath, []treeSpec{
		{label: "web", root: webDir, suffixes: []string{".ts"}},
		{label: "android", root: "/this/does/not/exist", suffixes: []string{".kt"}},
	})

	require.NoError(t, err)
	assert.Empty(t, orphans)
}

func TestCheck_MissingSpec_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := check("/this/spec/does/not/exist.yaml", nil)
	assert.Error(t, err, "missing spec must surface a clear error")
}

func TestStaticPathPrefix(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"/api/notes":            "/api/notes",
		"/api/notes/{id}":       "/api/notes",
		"/api/users/{id}/notes": "/api/users",
		"/healthz":              "/healthz",
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, staticPathPrefix(in))
		})
	}
}
