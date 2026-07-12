package tuiweave

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPublicPackagesHaveREADMEs keeps package-level quickstarts complete as
// the public library grows. Examples and internal implementation packages have
// different documentation contracts and are intentionally excluded.
func TestPublicPackagesHaveREADMEs(t *testing.T) {
	t.Helper()

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if path != "." && skipDocumentationDir(entry.Name(), path) {
			return filepath.SkipDir
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasPackageSource := false
		for _, child := range entries {
			if !child.IsDir() && strings.HasSuffix(child.Name(), ".go") && !strings.HasSuffix(child.Name(), "_test.go") {
				hasPackageSource = true
				break
			}
		}
		if !hasPackageSource {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "README.md")); err != nil {
			if os.IsNotExist(err) {
				t.Errorf("public package %q is missing README.md", path)
				return nil
			}
			return err
		}
		readme, err := os.ReadFile(filepath.Join(path, "README.md"))
		if err != nil {
			return err
		}
		content := string(readme)
		if !strings.Contains(content, "## API highlights") {
			t.Errorf("public package %q README.md is missing an API highlights section", path)
		}
		if !strings.Contains(content, "## Related documentation") {
			t.Errorf("public package %q README.md is missing a related documentation section", path)
		}
		if !strings.Contains(content, "https://pkg.go.dev/github.com/ishan5ain/tuiweave") {
			t.Errorf("public package %q README.md is missing a pkg.go.dev API reference", path)
		}
		if !strings.Contains(content, "AGENTS.md") && !strings.Contains(content, "AGENT-CATALOG.md") && !strings.Contains(content, "examples/") {
			t.Errorf("public package %q README.md is missing a runnable example or central guide link", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func skipDocumentationDir(name, path string) bool {
	if strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor" {
		return true
	}
	return path == "examples" || path == "internal"
}
