// check-parity asserts every /api/ path in the OpenAPI spec has at least
// one consumer in the web codebase. Catches the round-2 failure mode:
// spec defines /api/notes, but no client calls it — a ghost endpoint.
//
// Run: go run ./cmd/check-parity
// Exits non-zero on orphan paths. Wired into `make check:parity`.
//
// Rules:
//   - Every path under /api/ must appear as a literal string in web/src/**.
//   - Paths with URL params like /api/notes/{id} are checked by their
//     static prefix (/api/notes) since params vary by request.
//   - Exclusions in `excluded` below (operational routes + paths explicitly
//     deferred while the native iOS / Android apps catch up).
//
// Mobile parity note: the Flutter client was removed on branch
// feat/dropping_flutter in favor of two native apps (mobile/ios/,
// mobile/android/). While those apps are feature-light, /api/auth/mobile/*
// endpoints have no client consumer — they stay in the spec because the
// backend wiring is still correct, just temporarily orphaned. Add them to
// `excluded` below if they surface as orphans. Once the native apps call
// those endpoints, remove the exclusions and (optionally) teach this tool
// to scan .swift / .kt source under mobile/.
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
// - Operational endpoints scraped by infra, not called by the apps.
// - Mobile-only auth endpoints, deferred until native apps catch up.
var excluded = map[string]bool{
	"/healthz":                 true,
	"/readyz":                  true,
	"/metrics":                 true,
	"/_debug/routes":           true,
	"/api/auth/mobile/google":  true,
	"/api/auth/mobile/verify":  true,
}

func main() {
	specPath := flag.String("spec", "api/openapi.yaml", "path to OpenAPI spec (relative to backend/)")
	webRoot := flag.String("web", "../web/src", "path to web source tree")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	orphans, err := check(*specPath, *webRoot)
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
		fmt.Fprintf(os.Stderr, "  %s (not referenced in web/src — mobile parity is paused, see mobile/AGENTS.md)\n", p)
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Either add a consumer in web, add the path to `excluded` in check-parity/main.go,")
	fmt.Fprintln(os.Stderr, "or remove the path from openapi.yaml.")
	os.Exit(1)
}

func check(specPath, webRoot string) ([]string, error) {
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

	var orphans []string
	for path := range doc.Paths.Map() {
		if excluded[path] || !strings.HasPrefix(path, "/api/") {
			continue
		}
		// Strip {param} suffixes for the grep target. /api/notes/{id} → /api/notes.
		staticPrefix := staticPathPrefix(path)
		if !strings.Contains(webText, staticPrefix) {
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
	skip := regexp.MustCompile(`(^|/)(node_modules|\.next|build|dist|coverage)(/|$)`)
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
