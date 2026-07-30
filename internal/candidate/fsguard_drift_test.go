package candidate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// fsguard.go copies the ledger's filesystem allowlist rather than importing it,
// because importing internal/journal would put the order ledger's whole API
// inside this package's type universe and the isolation requirement forbids that.
//
// The price of a copy is drift, and this file is what stops the copy from being
// left to care. It reads internal/journal/fsguard.go as *text* — parsing a file
// is not importing it, so the isolation still holds — and fails when the two
// stop agreeing.
//
// The failure it prevents is quiet in the worst way: someone adds a filesystem to
// the ledger's rejected list because it was found to lose writes, and the
// discovery store keeps accepting it. Nothing errors; the store just silently
// stops being durable on that mount, and the row it loses may be a first_seen_at.

func journalFSGuardSource(t *testing.T) *ast.File {
	t.Helper()
	path := filepath.Join("..", "journal", "fsguard.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v — the ledger's allowlist is the original this package copies", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return file
}

// magicConstants reads every `MagicX int64 = 0x…` declaration out of a parsed
// file. The names are the comparison key, so a magic that is renamed on one side
// shows up as a difference rather than passing on its value.
func magicConstants(t *testing.T, file *ast.File) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range spec.Names {
			if len(spec.Values) <= i {
				continue
			}
			lit, ok := spec.Values[i].(*ast.BasicLit)
			if !ok || lit.Kind != token.INT {
				continue
			}
			v, err := strconv.ParseInt(lit.Value, 0, 64)
			if err != nil {
				continue
			}
			out[name.Name] = v
		}
		return true
	})
	return out
}

// TestTheAllowlistStillAgreesWithTheLedgers is the drift guard.
//
// It compares magic numbers by name in both directions. A magic present on one
// side and absent on the other is a difference, because both directions are
// dangerous: a filesystem the ledger newly rejects must be rejected here too, and
// a filesystem the ledger newly allows should not leave this store refusing to
// open on a host the engine runs on.
func TestTheAllowlistStillAgreesWithTheLedgers(t *testing.T) {
	theirs := magicConstants(t, journalFSGuardSource(t))

	mine := map[string]int64{
		"MagicExt": MagicExt, "MagicXFS": MagicXFS, "MagicBtrfs": MagicBtrfs,
		"MagicFUSE": MagicFUSE, "MagicTmpfs": MagicTmpfs, "MagicNFS": MagicNFS,
		"MagicSMB": MagicSMB, "MagicSMB2": MagicSMB2, "MagicOverlay": MagicOverlay,
		"MagicZFS": MagicZFS, "MagicExFAT": MagicExFAT, "MagicVFAT": MagicVFAT,
		"MagicNTFS3": MagicNTFS3, "MagicF2FS": MagicF2FS, "MagicCeph": MagicCeph,
	}

	if len(theirs) == 0 {
		t.Fatal("no magic constants were read from the ledger's fsguard.go; the guard is not looking at it")
	}
	for name, want := range theirs {
		got, ok := mine[name]
		if !ok {
			t.Errorf("internal/journal declares %s and this package does not; "+
				"a filesystem the ledger has an opinion about is unrecognised here", name)
			continue
		}
		if got != want {
			t.Errorf("%s = %#x here and %#x in internal/journal", name, got, want)
		}
	}
	for name := range mine {
		if _, ok := theirs[name]; !ok {
			t.Errorf("this package declares %s and internal/journal does not; "+
				"the copy has grown its own entry", name)
		}
	}
}

// mapKeys reads the keys of a `var name = map[...]...{ … }` composite literal out
// of a parsed file, as source text.
//
// Keys rather than values, because the allowlists are membership sets: what
// matters is which filesystems appear in allowedMagics, not what they map to.
func mapKeys(file *ast.File, name string) ([]string, bool) {
	var (
		keys  []string
		found bool
	)
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, ident := range spec.Names {
			if ident.Name != name || len(spec.Values) <= i {
				continue
			}
			lit, ok := spec.Values[i].(*ast.CompositeLit)
			if !ok {
				continue
			}
			found = true
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				// Only allowlisted (true) entries count as membership. A map that
				// gained an explicit `: false` entry has not widened.
				if truth, ok := kv.Value.(*ast.Ident); ok && truth.Name != "true" {
					continue
				}
				switch key := kv.Key.(type) {
				case *ast.Ident:
					keys = append(keys, key.Name)
				case *ast.BasicLit:
					keys = append(keys, strings.Trim(key.Value, `"`))
				}
			}
		}
		return true
	})
	sort.Strings(keys)
	return keys, found
}

// funcReturnStrings reads the string literals out of a `func Name() []string {
// return []string{…} }` body.
func funcReturnStrings(file *ast.File, name string) ([]string, bool) {
	var (
		out   []string
		found bool
	)
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != name || fn.Body == nil {
			return true
		}
		for _, stmt := range fn.Body.List {
			ret, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			for _, expr := range ret.Results {
				lit, ok := expr.(*ast.CompositeLit)
				if !ok {
					continue
				}
				found = true
				for _, elt := range lit.Elts {
					if s, ok := elt.(*ast.BasicLit); ok {
						out = append(out, strings.Trim(s.Value, `"`))
					}
				}
			}
		}
		return true
	})
	return out, found
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestTheAllowedSetIsTheLedgersAllowedSet compares the *decision*, not the
// constants.
//
// The constants agreeing is not the property that matters. The failure this file
// exists to catch — the one its header describes — is the ledger removing a
// filesystem from its allowed set because it was found to lose writes, while this
// store keeps accepting it. The magic numbers are unchanged in that scenario; only
// the membership moved. So the membership is what gets read out of the ledger's
// source and compared, in both directions.
func TestTheAllowedSetIsTheLedgersAllowedSet(t *testing.T) {
	theirs := journalFSGuardSource(t)

	theirMagics, ok := mapKeys(theirs, "allowedMagics")
	if !ok || len(theirMagics) == 0 {
		t.Fatal("no allowedMagics literal was read from the ledger's fsguard.go; " +
			"the guard is comparing nothing")
	}
	theirNames, ok := mapKeys(theirs, "allowedNames")
	if !ok || len(theirNames) == 0 {
		t.Fatal("no allowedNames literal was read from the ledger's fsguard.go")
	}
	theirAnswer, ok := funcReturnStrings(theirs, "AllowedFilesystems")
	if !ok || len(theirAnswer) == 0 {
		t.Fatal("the ledger's AllowedFilesystems() returned no literal list")
	}

	mineMagics := make([]string, 0, len(allowedMagics))
	for name, allowed := range map[string]bool{
		"MagicExt": allowedMagics[MagicExt], "MagicXFS": allowedMagics[MagicXFS],
		"MagicBtrfs": allowedMagics[MagicBtrfs], "MagicFUSE": allowedMagics[MagicFUSE],
		"MagicTmpfs": allowedMagics[MagicTmpfs], "MagicNFS": allowedMagics[MagicNFS],
		"MagicSMB": allowedMagics[MagicSMB], "MagicSMB2": allowedMagics[MagicSMB2],
		"MagicOverlay": allowedMagics[MagicOverlay], "MagicZFS": allowedMagics[MagicZFS],
		"MagicExFAT": allowedMagics[MagicExFAT], "MagicVFAT": allowedMagics[MagicVFAT],
		"MagicNTFS3": allowedMagics[MagicNTFS3], "MagicF2FS": allowedMagics[MagicF2FS],
		"MagicCeph": allowedMagics[MagicCeph],
	} {
		if allowed {
			mineMagics = append(mineMagics, name)
		}
	}
	sort.Strings(mineMagics)

	mineNames := make([]string, 0, len(allowedNames))
	for name, allowed := range allowedNames {
		if allowed {
			mineNames = append(mineNames, name)
		}
	}
	sort.Strings(mineNames)

	if !sameSet(mineMagics, theirMagics) {
		t.Errorf("allowedMagics membership: this package allows %v, internal/journal allows %v",
			mineMagics, theirMagics)
	}
	if !sameSet(mineNames, theirNames) {
		t.Errorf("allowedNames membership: this package allows %v, internal/journal allows %v",
			mineNames, theirNames)
	}
	if !sameSet(AllowedFilesystems(), theirAnswer) {
		t.Errorf("AllowedFilesystems() = %v here and %v in internal/journal",
			AllowedFilesystems(), theirAnswer)
	}

	// And the decision function agrees with the membership it was built from, so a
	// map that matches but an isAllowedFS that stopped consulting it still fails.
	for _, magic := range []int64{MagicExt, MagicXFS, MagicBtrfs} {
		if !isAllowedFS(FSInfo{Magic: magic}) {
			t.Errorf("%#x is in allowedMagics but isAllowedFS rejects it", magic)
		}
	}
	for _, magic := range []int64{
		MagicFUSE, MagicTmpfs, MagicNFS, MagicSMB, MagicSMB2, MagicOverlay,
		MagicZFS, MagicExFAT, MagicVFAT, MagicNTFS3, MagicF2FS, MagicCeph,
	} {
		if isAllowedFS(FSInfo{Magic: magic}) {
			t.Errorf("%#x is not in allowedMagics but isAllowedFS accepts it", magic)
		}
	}
}
