package strategyengine

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestA047ShipsNoRuntimeOrderOrExitWiring(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	modulePath := "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		if strings.HasPrefix(filepath.ToSlash(relative), "internal/strategyengine/") {
			return nil
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, spec := range file.Imports {
			if strings.Trim(spec.Path.Value, "\"") == modulePath {
				t.Errorf("a047 dormant engine is runtime-wired by %s", filepath.ToSlash(relative))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
