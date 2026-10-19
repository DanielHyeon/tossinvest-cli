package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplayWritesOnlyOwnedOutputAndPreservesInput(t *testing.T) {
	input := filepath.Join("testdata", "synthetic.jsonl")
	before, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out")
	var result bytes.Buffer
	if err = run([]string{"-input", input, "-out", dest}, &result); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.String(), "durable=1") {
		t.Fatal(result.String())
	}
	after, err := os.ReadFile(input)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("input modified", err)
	}
	result.Reset()
	if err = run([]string{"-input", input, "-out", dest}, &result); err != nil || !strings.Contains(result.String(), "duplicate=1") {
		t.Fatalf("replay: %v %s", err, result.String())
	}
}
func TestMalformedInputDoesNotCreateOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "bad")
	dest := filepath.Join(dir, "out")
	if err := os.WriteFile(input, []byte("{\"unexpected\":true}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-input", input, "-out", dest}, &bytes.Buffer{}); err == nil {
		t.Fatal("invalid accepted")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("destination created: %v", err)
	}
}
