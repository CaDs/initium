// check-dto-drift verifies hand-written Dart DTOs stay in sync with the OpenAPI spec.
//
// For each mapping declared in mobile/tool/dto_manifest.yaml, it loads the named
// schema from backend/api/openapi.yaml and asserts that every required field
// is referenced by its snake_case JSON name in the corresponding Dart file.
//
// Run: go run ./cmd/check-dto-drift
// Exits non-zero on drift. Used by `make check:openapi` (see Workstream C).
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

type manifest struct {
	Mappings []mapping `yaml:"mappings"`
}

type mapping struct {
	Schema    string `yaml:"schema"`
	DartFile  string `yaml:"dart_file"`
	DartClass string `yaml:"dart_class"`
}

func main() {
	specPath := flag.String("spec", "api/openapi.yaml", "path to OpenAPI spec (relative to backend/)")
	manifestPath := flag.String("manifest", "../mobile/tool/dto_manifest.yaml", "path to DTO manifest")
	mobileRoot := flag.String("mobile", "../mobile", "path to mobile project root")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	drifts, err := check(*specPath, *manifestPath, *mobileRoot)
	if err != nil {
		slog.Error("check failed", "error", err)
		os.Exit(2)
	}

	if len(drifts) == 0 {
		fmt.Println("dart DTO drift check: ok")
		return
	}

	fmt.Fprintln(os.Stderr, "dart DTO drift detected:")
	for _, d := range drifts {
		fmt.Fprintf(os.Stderr, "  %s [%s → %s]: %s\n", d.schema, d.dartClass, d.dartFile, d.msg)
	}
	os.Exit(1)
}

type drift struct {
	schema    string
	dartFile  string
	dartClass string
	msg       string
}

func check(specPath, manifestPath, mobileRoot string) ([]drift, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("loading spec %s: %w", specPath, err)
	}

	mf, err := loadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest %s: %w", manifestPath, err)
	}

	var drifts []drift
	for _, m := range mf.Mappings {
		schemaRef, ok := doc.Components.Schemas[m.Schema]
		if !ok {
			drifts = append(drifts, drift{m.Schema, m.DartFile, m.DartClass, "schema not found in spec"})
			continue
		}

		source, err := os.ReadFile(filepath.Join(mobileRoot, m.DartFile))
		if err != nil {
			drifts = append(drifts, drift{m.Schema, m.DartFile, m.DartClass, fmt.Sprintf("cannot read dart file: %v", err)})
			continue
		}
		body, ok := extractClassBody(string(source), m.DartClass)
		if !ok {
			drifts = append(drifts, drift{m.Schema, m.DartFile, m.DartClass, "dart class not found in file"})
			continue
		}

		required := map[string]bool{}
		for _, f := range schemaRef.Value.Required {
			required[f] = true
		}

		for name := range schemaRef.Value.Properties {
			present := jsonKeyReferenced(body, name)
			if required[name] && !present {
				drifts = append(drifts, drift{
					m.Schema, m.DartFile, m.DartClass,
					fmt.Sprintf("required field %q missing from dart class", name),
				})
			}
		}
	}

	return drifts, nil
}

func loadManifest(path string) (*manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mf manifest
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return nil, err
	}
	return &mf, nil
}

// extractClassBody returns the text between `class Name {` and its matching
// closing brace. Lightweight: assumes well-formed Dart source and no class
// bodies containing unbalanced braces in string literals (none in our DTOs).
func extractClassBody(source, className string) (string, bool) {
	re := regexp.MustCompile(`class\s+` + regexp.QuoteMeta(className) + `\b[^{]*\{`)
	loc := re.FindStringIndex(source)
	if loc == nil {
		return "", false
	}
	depth := 1
	i := loc[1]
	for i < len(source) && depth > 0 {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
		}
		i++
	}
	if depth != 0 {
		return "", false
	}
	return source[loc[1] : i-1], true
}

// jsonKeyReferenced returns true if the Dart class body references the JSON
// key either as:
//   - json['name'] / json["name"] — the fromJson pattern (response DTOs), or
//   - 'name': / "name":           — the map-literal pattern (toJson on request DTOs).
//
// Request-body DTOs don't need a fromJson; their toJson emits the key via
// a map literal. Accepting both patterns means the drift check covers
// request + response DTOs without forcing request classes to carry a
// decorative fromJson just to satisfy the checker.
func jsonKeyReferenced(body, name string) bool {
	return strings.Contains(body, `json['`+name+`']`) ||
		strings.Contains(body, `json["`+name+`"]`) ||
		strings.Contains(body, `'`+name+`':`) ||
		strings.Contains(body, `"`+name+`":`)
}
