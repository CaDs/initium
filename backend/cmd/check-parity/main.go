// check-parity asserts every /api/ path in the OpenAPI spec has at least
// one consumer in the web, iOS, or Android codebase. Catches the round-2
// failure mode: spec defines /api/notes, but no client calls it — a ghost
// endpoint.
//
// Run: go run ./cmd/check-parity
// Exits non-zero on orphan paths. Wired into `make check:parity`.
//
// Rules:
//   - Every path under /api/ must appear as a literal string in at least
//     one of: web/src/** (.ts, .tsx), mobile/ios/** (.swift), or
//     mobile/android/** (.kt).
//   - Paths with URL params like /api/notes/{id} are checked by their
//     static prefix (/api/notes) since params vary by request.
//   - Exclusions in `excluded` below (operational routes + paths
//     temporarily deferred during migrations).
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// excluded are spec paths that intentionally have no client consumer.
// Operational endpoints scraped by infra, not called by the apps.
var excluded = map[string]bool{
	"/healthz":       true,
	"/readyz":        true,
	"/metrics":       true,
	"/_debug/routes": true,
}

type treeSpec struct {
	label    string
	root     string
	suffixes []string
}

func main() {
	specPath := flag.String("spec", "api/openapi.yaml", "path to OpenAPI spec (relative to backend/)")
	webRoot := flag.String("web", "../web/src", "path to web source tree")
	iosRoot := flag.String("ios", "../mobile/ios", "path to iOS source tree")
	androidRoot := flag.String("android", "../mobile/android", "path to Android source tree")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	trees := []treeSpec{
		{label: "web/src", root: *webRoot, suffixes: []string{".ts", ".tsx"}},
		{label: "mobile/ios", root: *iosRoot, suffixes: []string{".swift"}},
		{label: "mobile/android", root: *androidRoot, suffixes: []string{".kt"}},
	}

	orphans, err := check(*specPath, trees)
	if err != nil {
		slog.Error("parity check failed", "error", err)
		os.Exit(2)
	}

	if len(orphans) == 0 {
		fmt.Println("spec ↔ client parity check: ok")
		return
	}

	fmt.Fprintln(os.Stderr, "spec paths with no client consumer:")
	for _, p := range orphans {
		fmt.Fprintf(os.Stderr, "  %s (not referenced in web/src, mobile/ios, or mobile/android)\n", p)
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Either add a consumer on one of the three surfaces, add the path")
	fmt.Fprintln(os.Stderr, "to `excluded` in check-parity/main.go, or remove the path from openapi.yaml.")
	os.Exit(1)
}

func check(specPath string, trees []treeSpec) ([]string, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("loading spec %s: %w", specPath, err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid spec: %w", err)
	}

	// Concatenate all tree sources into a single blob we can substring-match.
	// Missing trees (e.g. the android directory doesn't exist yet on a fork)
	// are tolerated — they contribute an empty string and the path must then
	// be satisfied by a different tree.
	var blob strings.Builder
	for _, t := range trees {
		if _, statErr := os.Stat(t.root); os.IsNotExist(statErr) {
			continue
		}
		text, err := concatSources(t.root, t.suffixes...)
		if err != nil {
			return nil, fmt.Errorf("reading %s tree %s: %w", t.label, t.root, err)
		}
		blob.WriteString(text)
		blob.WriteString("\n")
	}
	combined := blob.String()

	var orphans []string
	for path := range doc.Paths.Map() {
		if excluded[path] || !strings.HasPrefix(path, "/api/") {
			continue
		}
		// Strip {param} suffixes for the grep target. /api/notes/{id} → /api/notes.
		staticPrefix := staticPathPrefix(path)
		if !strings.Contains(combined, staticPrefix) {
			orphans = append(orphans, path)
		}
	}
	sort.Strings(orphans)
	return orphans, nil
}

// staticPathPrefix returns the portion of an OpenAPI path up to the first
// `{` placeholder — that's the stable substring clients will contain.
func staticPathPrefix(path string) string {
	if idx := strings.Index(path, "{"); idx > 0 {
		return strings.TrimSuffix(path[:idx], "/")
	}
	return path
}

// concatSources reads every file under root with one of the given suffixes
// and returns their concatenated contents as a single searchable string.
// Skips common build / IDE / generated dirs across all three stacks.
func concatSources(root string, suffixes ...string) (string, error) {
	skip := regexp.MustCompile(`(^|/)(node_modules|\.next|build|dist|coverage|\.gradle|\.idea|DerivedData|\.swiftpm|xcuserdata|generated)(/|$)`)
	var parts []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skip.MatchString(p) {
				return filepath.SkipDir
			}
			return nil
		}
		ok := false
		for _, s := range suffixes {
			if strings.HasSuffix(p, s) {
				ok = true
				break
			}
		}
		if !ok {
			return nil
		}
		b, err := os.ReadFile(p) //nolint:gosec // WalkDir only scans trusted repo source trees.
		if err != nil {
			return err
		}
		parts = append(parts, string(b))
		return nil
	})
	if err != nil {
		return "", err
	}
	return strings.Join(parts, "\n"), nil
}
