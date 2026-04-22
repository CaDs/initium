// check-parity asserts every /api/ path in the OpenAPI spec has at least
// one consumer in the web or mobile codebase. Catches the round-2 failure
// mode: spec defines /api/notes, but no client calls it — a ghost endpoint.
//
// Run: go run ./cmd/check-parity
// Exits non-zero on orphan paths. Wired into `make check:parity`.
//
// Rules:
//   - Every path under /api/ must appear as a literal string in either
//     web/src/** or mobile/lib/**.
//   - Paths with URL params like /api/notes/{id} are checked by their
//     static prefix (/api/notes) since params vary by request.
//   - Exclusions in `excluded` below (operational routes).
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

// excluded are spec paths that intentionally have no client consumer
// (operational endpoints scraped by infra, not called by the apps).
var excluded = map[string]bool{
	"/healthz":       true,
	"/readyz":        true,
	"/metrics":       true,
	"/_debug/routes": true,
}

func main() {
	specPath := flag.String("spec", "api/openapi.yaml", "path to OpenAPI spec (relative to backend/)")
	webRoot := flag.String("web", "../web/src", "path to web source tree")
	mobileRoot := flag.String("mobile", "../mobile/lib", "path to mobile source tree")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	orphans, err := check(*specPath, *webRoot, *mobileRoot)
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
		fmt.Fprintf(os.Stderr, "  %s (not referenced in web/src or mobile/lib)\n", p)
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Either add a consumer in web or mobile, or remove the path from openapi.yaml.")
	os.Exit(1)
}

func check(specPath, webRoot, mobileRoot string) ([]string, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("loading spec %s: %w", specPath, err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid spec: %w", err)
	}

	webText, err := concatSources(webRoot, ".ts", ".tsx")
	if err != nil {
		return nil, fmt.Errorf("reading web tree %s: %w", webRoot, err)
	}
	mobileText, err := concatSources(mobileRoot, ".dart")
	if err != nil {
		return nil, fmt.Errorf("reading mobile tree %s: %w", mobileRoot, err)
	}

	var orphans []string
	for path := range doc.Paths.Map() {
		if excluded[path] || !strings.HasPrefix(path, "/api/") {
			continue
		}
		// Strip {param} suffixes for the grep target. /api/notes/{id} → /api/notes.
		staticPrefix := staticPathPrefix(path)
		if !strings.Contains(webText, staticPrefix) && !strings.Contains(mobileText, staticPrefix) {
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
// Skips common build output dirs.
func concatSources(root string, suffixes ...string) (string, error) {
	skip := regexp.MustCompile(`(^|/)(node_modules|\.next|build|\.dart_tool|dist|coverage)(/|$)`)
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
		b, err := os.ReadFile(p)
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
