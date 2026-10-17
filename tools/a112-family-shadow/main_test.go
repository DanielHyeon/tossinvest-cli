package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a112 0.5 리뷰 시험#9: 작성 도구 run()/document() 를 직접 부르는 시험.
// 커밋된 골든(strategyshadow/testdata)을 **이 도구가** 같은 값에서 바이트 그대로 만들어 내는지 대조함 —
// 골든은 소비자(적재기)만 재므로 생산자인 도구를 부르는 시험이 따로 있어야 함.

// goldenArgs 는 internal/strategyshadow/testdata/strategy-family-shadow-KR.json 의 값과 같은 인자임.
func goldenArgs(out string) []string {
	return []string{
		"-market", "KR", "-generation", "1",
		"-route-manifest-digest", "sha256:0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0",
		"-calibration-digest", "sha256:calibration-shadow-kr-v1",
		"-calendar-version", "kr-regular-2026.10",
		"-risk-policy-digest", "sha256:a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90",
		"-build-digest", "tossos-shadow-build-1",
		"-actor", "golden-operator",
		"-approved-at", "2026-10-05T00:00:00Z",
		"-issued-at", "2026-10-05T00:30:00Z",
		"-expires-at", "2026-10-05T23:00:00Z",
		"-shadow", "CONTINUATION,REVERSAL",
		"-out", out,
	}
}

// runWith 는 전역 flag 집합을 새로 깔고 인자를 넣어 run() 을 부름. stdout 은 잡아서 돌려줌.
func runWith(t *testing.T, args []string) (string, error) {
	t.Helper()
	savedArgs, savedFlags, savedStdout := os.Args, flag.CommandLine, os.Stdout
	t.Cleanup(func() { os.Args, flag.CommandLine, os.Stdout = savedArgs, savedFlags, savedStdout })
	os.Args = append([]string{"a112-family-shadow"}, args...)
	flag.CommandLine = flag.NewFlagSet("a112-family-shadow", flag.ContinueOnError)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	runErr := run()
	writer.Close()
	os.Stdout = savedStdout
	printed, _ := io.ReadAll(reader)
	reader.Close()
	return string(printed), runErr
}

func TestTheToolWritesTheCommittedGoldenBytesReadOnlyAndPrintsTheirPin(t *testing.T) {
	golden, err := os.ReadFile(filepath.Join("..", "..", "internal", "strategyshadow", "testdata", "strategy-family-shadow-KR.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "strategy-family-shadow-KR.json")
	printed, err := runWith(t, goldenArgs(out))
	if err != nil {
		t.Fatalf("run refused the golden values: %v", err)
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, bytes.TrimRight(golden, "\n")) {
		t.Fatalf("tool bytes differ from the committed golden\n got: %s\nwant: %s", written, golden)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o400 {
		t.Fatalf("manifest mode = %o, want 0400", mode)
	}
	// 골든 digest 는 strategyshadow/golden_test.go 가 고정한 값과 같아야 함 — 핀 줄 전체를 대조
	const pinLine = "TOSSOS_STRATEGY_FAMILY_SHADOW_KR_MANIFEST_SHA256=sha256:09e183e5cb709defb8e5053ad0cd8c6787ce7863b1f23e2044f2723e84327e39"
	if !strings.Contains(printed, pinLine+"\n") {
		t.Fatalf("printed pin line does not carry the golden digest:\n%s", printed)
	}
}

func TestTheToolNeverOverwritesAnExistingManifest(t *testing.T) {
	out := filepath.Join(t.TempDir(), "strategy-family-shadow-KR.json")
	// 쓰기 가능한(0600) 기존 파일로 둠 — 0400 이면 권한 거부가 대신 막아 O_EXCL 을 재지 못함(변이 T02 가 그렇게 살아남았음).
	if err := os.WriteFile(out, []byte("live"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := runWith(t, goldenArgs(out)); err == nil {
		t.Fatal("run overwrote an existing manifest")
	}
	kept, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(kept) != "live" {
		t.Fatalf("existing manifest changed: %q", kept)
	}
}

// 옮길 수 없는 인자 — 거절하고 파일을 만들지 않음(document() 의 갈래 · -out 누락)
func TestTheToolRefusesUntranslatableArgumentsWithoutWritingAFile(t *testing.T) {
	// 거절은 **그 인자를 옮기는 자리**에서 나야 함 — 오류가 그 플래그 이름을 말하는지 대조. 이름 없이 「거절됨」 만 보면 뒤의 인코더
	// 검사가 대신 막아도 통과함(변이 T01 — parseFamilies 의 Known 검사를 꺼도 인코더가 모르는 가족을 막아 살아남았음).
	cases := map[string]struct {
		mutate func([]string) []string
		names  string
	}{
		"market":         {func(a []string) []string { return replaceArg(a, "-market", "JP") }, "-market"},
		"approved-at":    {func(a []string) []string { return replaceArg(a, "-approved-at", "yesterday") }, "-approved-at"},
		"issued-at":      {func(a []string) []string { return replaceArg(a, "-issued-at", "2026-10-05") }, "-issued-at"},
		"expires-at":     {func(a []string) []string { return replaceArg(a, "-expires-at", "") }, "-expires-at"},
		"unknown-family": {func(a []string) []string { return replaceArg(a, "-shadow", "CONTINUATION,MOMENTUM") }, "-shadow"},
		"missing-out":    {func(a []string) []string { return replaceArg(a, "-out", "") }, "-out"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "strategy-family-shadow-KR.json")
			_, err := runWith(t, tc.mutate(goldenArgs(out)))
			if err == nil {
				t.Fatal("run accepted an untranslatable argument")
			}
			if !strings.Contains(err.Error(), tc.names) {
				t.Fatalf("refusal %q does not name %s — another guard stopped it", err, tc.names)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("refused run left files behind: %v", entries)
			}
		})
	}
}

func TestParseFamiliesKeepsAnEmptyListAllOff(t *testing.T) {
	families, err := parseFamilies("  ")
	if err != nil || families != nil {
		t.Fatalf("parseFamilies(blank) = %v, %v; want nil, nil (넷 다 OFF)", families, err)
	}
}

func replaceArg(args []string, name, value string) []string {
	out := append([]string(nil), args...)
	for i := 0; i+1 < len(out); i++ {
		if out[i] == name {
			out[i+1] = value
			return out
		}
	}
	panic("no such flag " + name)
}

func TestParseMarketKnowsBothMarketsAndNothingElse(t *testing.T) {
	for value, want := range map[string]string{"KR": "KR", " us ": "US"} {
		if got, err := parseMarket(value); err != nil || string(got) != want {
			t.Errorf("parseMarket(%q) = %q, %v; want %s", value, got, err, want)
		}
	}
	if _, err := parseMarket("JP"); err == nil {
		t.Error("parseMarket accepted JP")
	}
}
