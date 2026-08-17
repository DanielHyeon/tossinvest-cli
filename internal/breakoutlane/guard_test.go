package breakoutlane

import (
	"crypto/sha256"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAcceptedL0GoldensMatchManifest(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens")
	manifest := string(mustRead(t, filepath.Join(dir, "manifest.sha256")))
	for _, name := range []string{"breakout-evidence-and-sizing-v1.json", "four-family-runtime-v1.json"} {
		want := ""
		for _, line := range strings.Split(manifest, "\n") {
			f := strings.Fields(line)
			if len(f) == 2 && f[1] == name {
				want = f[0]
			}
		}
		sum := sha256.Sum256(mustRead(t, filepath.Join(dir, name)))
		if want == "" || fmtHex(sum[:]) != want {
			t.Fatalf("golden %s mismatch", name)
		}
	}
}

func TestAcceptedL0GoldensBindConfigAndDormantDescriptors(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens")
	var breakout struct {
		SetupIdentity struct {
			Domain           string `json:"domain"`
			PreimageEncoding string `json:"preimage_encoding"`
			Separator        struct {
				Byte int `json:"byte"`
			} `json:"separator"`
		} `json:"setup_identity"`
		Thresholds struct {
			OpeningRange int `json:"opening_range_regular_session_minutes"`
			RVOL         int `json:"rvol_min_ppm"`
			Wick         int `json:"upper_wick_range_max_ppm"`
		} `json:"thresholds"`
	}
	if err := json.Unmarshal(mustRead(t, filepath.Join(dir, "breakout-evidence-and-sizing-v1.json")), &breakout); err != nil {
		t.Fatal(err)
	}
	if breakout.SetupIdentity.Domain != "tossos.breakout.setup.v1" || breakout.SetupIdentity.PreimageEncoding != "UTF-8" || breakout.SetupIdentity.Separator.Byte != 0 || breakout.Thresholds.OpeningRange != 15 || breakout.Thresholds.RVOL != 1_500_000 || breakout.Thresholds.Wick != 350_000 {
		t.Fatalf("breakout golden contract drift: %+v", breakout)
	}
	var runtime struct {
		Descriptors []struct {
			Market    string `json:"market"`
			Family    string `json:"family"`
			LaneID    string `json:"lane_id"`
			Desired   string `json:"desired"`
			Effective string `json:"effective"`
			Runtime   string `json:"runtime"`
		} `json:"descriptors"`
	}
	if err := json.Unmarshal(mustRead(t, filepath.Join(dir, "four-family-runtime-v1.json")), &runtime); err != nil {
		t.Fatal(err)
	}
	want := map[string]Descriptor{"KR": {MarketKR, KRLaneID, "v1", "SHORT", "OFF", "OFF", "UNOBSERVED"}, "US": {MarketUS, USLaneID, "v1", "SHORT", "OFF", "OFF", "UNOBSERVED"}}
	seen := map[string]bool{}
	for _, item := range runtime.Descriptors {
		if item.Family != "BREAKOUT_RETEST" {
			continue
		}
		got, ok := want[item.Market]
		if !ok || item.LaneID != got.LaneID || item.Desired != "OFF" || item.Effective != "OFF" || item.Runtime != "UNOBSERVED" {
			t.Fatalf("runtime golden breakout descriptor mismatch: %+v", item)
		}
		seen[item.Market] = true
	}
	if !seen["KR"] || !seen["US"] || Descriptors() != [2]Descriptor{want["KR"], want["US"]} {
		t.Fatal("breakout descriptors do not bind frozen runtime golden")
	}
}
func TestResolvedProductionClosureAllowlist(t *testing.T) {
	root := repoRoot(t)
	files, _ := filepath.Glob(filepath.Join(root, "internal/breakoutlane", "*.go"))
	allow := map[string]bool{"crypto/sha256": true, "encoding/hex": true, "errors": true, "math/bits": true, "strconv": true, "strings": true, "unicode/utf8": true}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			path := strings.Trim(im.Path.Value, "\"")
			if !allow[path] {
				t.Fatalf("forbidden import %s in %s", path, file)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if ok && (fn.Name.Name == "Apply" || fn.Name.Name == "NewMachine" || fn.Name.Name == "NewQuote" || fn.Name.Name == "NewFX") {
				t.Fatalf("raw or bypass API %s", fn.Name)
			}
			return true
		})
	}
}
func TestPublicSurfaceCannotAssertEventsOrForgeMachine(t *testing.T) {
	root := repoRoot(t)
	files, _ := filepath.Glob(filepath.Join(root, "internal/breakoutlane", "*.go"))
	forbidden := map[string]bool{"Event": true, "Machine": true, "State": true, "Apply": true, "Propose": true, "Consume": true}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && forbidden[d.Name.Name] {
					t.Fatalf("public bypass function %s", d.Name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if named, ok := spec.(*ast.TypeSpec); ok && forbidden[named.Name.Name] {
						t.Fatalf("public bypass type %s", named.Name)
					}
				}
			}
		}
	}
}
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	for d := filepath.Dir(file); ; d = filepath.Dir(d) {
		if _, e := os.Stat(filepath.Join(d, "go.mod")); e == nil {
			return d
		}
		if filepath.Dir(d) == d {
			t.Fatal("root")
		}
	}
}
func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func fmtHex(v []byte) string {
	const d = "0123456789abcdef"
	out := make([]byte, len(v)*2)
	for i, b := range v {
		out[2*i], out[2*i+1] = d[b>>4], d[b&15]
	}
	return string(out)
}
