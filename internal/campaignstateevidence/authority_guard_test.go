package campaignstateevidence

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const evidenceImport = "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"

func allowedImport(path string, source []byte) (bool, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		return false, err
	}
	laneAliases := map[string]bool{}
	for _, spec := range f.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return false, err
		}
		if strings.HasPrefix(p, "github.com/JungHoonGhae/tossinvest-cli/internal/") && (strings.HasSuffix(p, "/continuationlane") || strings.HasSuffix(p, "/reversallane") || strings.HasSuffix(p, "/weeklyvaluelane")) {
			alias := filepath.Base(p)
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			if alias == "." {
				return false, nil
			}
			laneAliases[alias] = true
		}
		if p == evidenceImport && path != "internal/continuationlane/offline_evidence.go" && path != "internal/reversallane/offline_evidence.go" && path != "internal/weeklyvaluelane/offline_evidence.go" {
			return false, nil
		}
	}
	allowed := true
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == "CaptureOfflineEvidence" {
			if id, ok := sel.X.(*ast.Ident); ok && laneAliases[id.Name] {
				allowed = false
			}
		}
		return true
	})
	return allowed, nil
}
func TestInspectionPackageHasNoRuntimeConsumers(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, dir := range []string{"internal", "cmd", "tools"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			allowed, err := allowedImport(filepath.ToSlash(rel), source)
			if err != nil {
				return err
			}
			if !allowed {
				t.Errorf("unexpected inspection import: %s", rel)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	for _, src := range []string{"package p; import `" + evidenceImport + "`", "package p; import \"github.com/JungHoonGhae/tossinvest-cli/internal/campaignstate\\u0065vidence\""} {
		ok, err := allowedImport("internal/app/engine/forged.go", []byte(src))
		if err != nil || ok {
			t.Fatal("guard missed raw or escaped import", err)
		}
	}
}

func TestGuardRejectsRuntimeCaptureUse(t *testing.T) {
	for _, imp := range []string{"`github.com/JungHoonGhae/tossinvest-cli/internal/reversallane`", "\"github.com/JungHoonGhae/tossinvest-cli/internal/reversall\\u0061ne\""} {
		src := "package p; import lane " + imp + "; var x = lane.CaptureOfflineEvidence"
		if ok, err := allowedImport("internal/app/engine/forged.go", []byte(src)); err != nil || ok {
			t.Fatalf("guard missed runtime reference: %v", err)
		}
	}
	src := "package p; import . `github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane`; var x = CaptureOfflineEvidence"
	if ok, err := allowedImport("internal/app/engine/forged.go", []byte(src)); err != nil || ok {
		t.Fatalf("guard missed dot import: %v", err)
	}
}
