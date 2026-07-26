package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeCapturesFunctionStructure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := `package sample
func Worker(input []int) int {
	total := 0
	for _, value := range input {
		if value < 0 { return total }
		total += helper(value)
	}
	go helper(total)
	defer helper(0)
	return total
}
func helper(v int) int { return v }
`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := analyze(path, "Worker")
	if err != nil {
		t.Fatal(err)
	}
	if got.Function != "Worker" || got.Package != "sample" {
		t.Fatalf("unexpected identity: %#v", got)
	}
	if len(got.SourceSHA256) != 64 {
		t.Fatalf("missing source hash: %#v", got)
	}
	if got.Signature != "Worker(params=1, results=1)" {
		t.Fatalf("signature = %q", got.Signature)
	}
	if got.Start.Line != 2 || got.End.Line <= got.Start.Line {
		t.Fatalf("invalid source range: %#v..%#v", got.Start, got.End)
	}
	if len(got.Branches) < 2 {
		t.Fatalf("expected range and if branches, got %#v", got.Branches)
	}
	for index, branch := range got.Branches {
		want := fmt.Sprintf("B%d", index+1)
		if branch.ID != want {
			t.Fatalf("branch id = %q, want %q", branch.ID, want)
		}
	}
	if len(got.Returns) != 2 {
		t.Fatalf("expected two returns, got %#v", got.Returns)
	}
	if len(got.Calls) < 3 || len(got.GoStatements) != 1 || len(got.Defers) != 1 {
		t.Fatalf("missing call/go/defer evidence: %#v", got)
	}
	if len(got.Assignments) < 2 {
		t.Fatalf("missing assignment evidence: %#v", got.Assignments)
	}
}

func TestListFunctionsReturnsQualifiedSourceRanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := "package sample\nfunc One() {}\ntype Runner struct{}\nfunc (Runner) Run() {}\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := listFunctions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Receiver != "Runner" || len(got[0].SourceSHA256) != 64 {
		t.Fatalf("unexpected list evidence: %#v", got)
	}
}

func TestAnalyzeMethodUsesReceiverQualifiedName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := "package sample\ntype Runner struct{}\nfunc (Runner) Run() {}\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := analyze(path, "Runner.Run")
	if err != nil {
		t.Fatal(err)
	}
	if got.Receiver != "Runner" {
		t.Fatalf("receiver = %q", got.Receiver)
	}
}

func TestAnalyzeRejectsAmbiguousAndMissingFunctions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := "package sample\ntype A struct{}\ntype B struct{}\nfunc (A) Run() {}\nfunc (B) Run() {}\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := analyze(path, "Run"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
	if _, err := analyze(path, "Missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestAnalyzeRejectsInvalidGo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.go")
	if err := os.WriteFile(path, []byte("package broken\nfunc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := analyze(path, "broken"); err == nil {
		t.Fatal("expected parse error")
	}
}
