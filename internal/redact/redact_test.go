package redact_test

// redact_test.go — a131 규칙 하나의 단위 시험. 센티널은 계좌 모양 둘(짧은 문자형 · 11자리 숫자형)을 씀.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

const (
	sentinelRef    = "987-65-SENTINEL"
	sentinelDigits = "98765432101"
	sentinelMasked = "*******2101"
	positionID     = "pos-0123456789abcdef01234567"
)

func init() { redact.Register(sentinelRef, sentinelDigits) }

// leaks 는 출력에 남은 계좌 흔적을 모두 셈 — 원문 · 대시 끼운 숫자열 · 마스크 · 포지션 id.
func leaks(out string) []string {
	var found []string
	for _, needle := range []string{sentinelRef, sentinelDigits, "987-654-32101", sentinelMasked, positionID} {
		if strings.Contains(out, needle) {
			found = append(found, needle)
		}
	}
	return found
}

func TestA131StringReplacesEveryAccountShape(t *testing.T) {
	in := fmt.Sprintf("a=%s b=%s c=987-654-32101 d=%s e=%s f=keep-123 g=12345", sentinelRef, sentinelDigits, sentinelMasked, positionID)
	out := redact.String(in)
	if l := leaks(out); len(l) != 0 {
		t.Fatalf("String left %v in %q", l, out)
	}
	// 계좌가 아닌 값은 그대로(가격 · 수량 · 짧은 숫자).
	for _, keep := range []string{"f=keep-123", "g=12345"} {
		if !strings.Contains(out, keep) {
			t.Fatalf("String damaged a non-account value %q: %q", keep, out)
		}
	}
	if !strings.Contains(out, "pos-[redacted]") {
		t.Fatalf("position id not replaced by its placeholder: %q", out)
	}
}

func prepare(t *testing.T, dir string) *redact.SealKey {
	t.Helper()
	key, err := redact.PrepareSealKey(filepath.Join(dir, "record.jsonl"))
	if err != nil {
		t.Fatalf("PrepareSealKey: %v", err)
	}
	return key
}

func TestA131StringCoversSeparatedMasksAndEveryDerivedID(t *testing.T) {
	in := "a ********5-12 b adj-0123456789abcdef01234567 c corr-0123456789abcdef0123456789abcdef d rec-0123456789abcdef01234567 e mode-0123456789abcdef01234567"
	out := redact.String(in)
	for _, leak := range []string{"5-12", "-12 ", "0123456789abcdef"} {
		if strings.Contains(out, leak) {
			t.Fatalf("String left %q in %q", leak, out)
		}
	}
	for _, want := range []string{"adj-[redacted]", "corr-[redacted]", "rec-[redacted]", "mode-[redacted]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("String = %q, want %q", out, want)
		}
	}
}

func TestA131PrepareSealKeyRefusesANewKeyForASealedRecord(t *testing.T) {
	dir := t.TempDir()
	rec := filepath.Join(dir, "record.jsonl")
	sealed, err := redact.SealRecordLine([]byte(`{"account_ref":"`+sentinelRef+`"}`+"\n"), prepare(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rec, sealed, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, redact.KeyFileName)); err != nil {
		t.Fatal(err)
	}
	if _, err := redact.PrepareSealKey(rec); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("PrepareSealKey on a sealed record without its key = %v; want a refusal", err)
	}
	if _, err := os.Stat(filepath.Join(dir, redact.KeyFileName)); !os.IsNotExist(err) {
		t.Fatalf("a new key was created for a sealed record: %v", err)
	}
}

func TestA131MalformedKeyIsRefused(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, redact.KeyFileName), []byte("abcd\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := redact.PrepareSealKey(filepath.Join(dir, "record.jsonl")); err == nil {
		t.Fatal("a 2-byte key was accepted")
	}
}

type panicky struct{}

func (panicky) MarshalJSON() ([]byte, error) { panic("boom") }

func TestA131HandlerSurvivesAPanickingValue(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(redact.Handler(slog.NewJSONHandler(&buf, nil)))
	log.Info("m", "bad", panicky{}, "symbol", "005930")
	if !strings.Contains(buf.String(), redact.Unloggable) || !strings.Contains(buf.String(), "005930") {
		t.Fatalf("a panicking value must become a placeholder and the line must survive: %q", buf.String())
	}
}

func TestA131StringLeavesShortNumbersAlone(t *testing.T) {
	// 등록 숫자열 규칙은 8자리 이상 참조에만 — 짧은 참조로 가격을 지우면 로그가 거짓이 됨.
	redact.Register("12-34")
	out := redact.String("price 1234 qty 12")
	if out != "price 1234 qty 12" {
		t.Fatalf("short registered ref damaged unrelated digits: %q", out)
	}
	if got := redact.String("acct 12-34"); strings.Contains(got, "12-34") {
		t.Fatalf("short registered ref itself not replaced: %q", got)
	}
}

func TestA131IsAccountKey(t *testing.T) {
	for _, k := range []string{"account", "account_ref", "Account_Ref", "accountNo", "account_seq", "account_refs", "masked_account"} {
		if !redact.IsAccountKey(k) {
			t.Errorf("IsAccountKey(%q) = false", k)
		}
	}
	for _, k := range []string{"symbol", "accounted", "position_id", "claimed_by"} {
		if redact.IsAccountKey(k) {
			t.Errorf("IsAccountKey(%q) = true", k)
		}
	}
}

type stringer struct{ s string }

func (s stringer) String() string { return s.s }

func TestA131HandlerDropsAccountKeysAndScrubsValues(t *testing.T) {
	var buf bytes.Buffer
	h := redact.Handler(slog.NewJSONHandler(&buf, nil))
	// 등록 전에 묶은 속성도 Handle 시점에 치환되어야 함 — 아래 값은 위 init 의 등록 값.
	log := slog.New(h).With("account", sentinelRef, "claimed_by", "engine:"+sentinelDigits+":1")
	wrapped := fmt.Errorf("journal: %w", fmt.Errorf("no position for account %s", sentinelDigits))
	log.Error("event "+sentinelRef, "error", wrapped.Error(), "err2", wrapped,
		"stringer", stringer{"x " + sentinelMasked}, "position_id", positionID,
		slog.Group("g", slog.String("account_ref", sentinelRef), slog.String("detail", sentinelRef)),
		"num", int64(98765432101), "list", []string{sentinelRef}, "symbol", "005930")
	out := buf.String()
	if l := leaks(out); len(l) != 0 {
		t.Fatalf("handler output leaked %v: %s", l, out)
	}
	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("handler output is not one JSON object: %v: %s", err, out)
	}
	if _, ok := line["account"]; ok {
		t.Fatalf("account key survived: %s", out)
	}
	if g, _ := line["g"].(map[string]any); g == nil || g["account_ref"] != nil || g["detail"] == nil {
		t.Fatalf("group: account_ref must be dropped and detail kept: %s", out)
	}
	if line["symbol"] != "005930" {
		t.Fatalf("non-account value changed: %s", out)
	}
}

func TestA131HandlerScrubsAttrsBoundBeforeRegistration(t *testing.T) {
	late := "555-44-LATEBOUND"
	var buf bytes.Buffer
	log := slog.New(redact.Handler(slog.NewJSONHandler(&buf, nil))).With("holder", "x:"+late)
	redact.Register(late) // 등록이 With 보다 늦다
	log.Info("m")
	if strings.Contains(buf.String(), late) {
		t.Fatalf("attribute bound before registration leaked: %s", buf.String())
	}
}

func TestA131WriterScrubsEveryWrite(t *testing.T) {
	var buf bytes.Buffer
	w := redact.Writer(&buf)
	msg := "account          " + sentinelDigits + "\n"
	n, err := w.Write([]byte(msg))
	if err != nil || n != len(msg) {
		t.Fatalf("Write = %d, %v; want %d, nil", n, err, len(msg))
	}
	fmt.Fprintf(w, "refused: %v\n", errors.New("attests "+sentinelMasked))
	if l := leaks(buf.String()); len(l) != 0 {
		t.Fatalf("writer leaked %v: %q", l, buf.String())
	}
}

func TestA131SealRecordLineRoundTrips(t *testing.T) {
	dir := t.TempDir()
	key := prepare(t, dir)
	line := []byte(`{"format_version":1,"account_ref":"` + sentinelRef + `","latency_ms":12345678901234567,"note":"<b>&` + sentinelDigits + `","nested":{"e":"x ` + sentinelRef + `","order_id":"` + positionID + `"},"request_digest":"` + positionID + `"}` + "\n")
	sealed, err := redact.SealRecordLine(line, key)
	if err != nil {
		t.Fatalf("SealRecordLine: %v", err)
	}
	if l := leaks(strings.ReplaceAll(string(sealed), positionID, "")); len(l) != 0 {
		t.Fatalf("sealed line leaked %v: %s", l, sealed)
	}
	// 식별자·다이제스트 키의 값은 치환하지 않음(대사·정리가 이 값으로 대상을 찾음).
	if strings.Count(string(sealed), positionID) != 2 {
		t.Fatalf("identity values were rewritten: %s", sealed)
	}
	if !strings.Contains(string(sealed), `"format_version":2`) {
		t.Fatalf("a sealed line must carry the sealed record format: %s", sealed)
	}
	if !bytes.HasSuffix(sealed, []byte("\n")) {
		t.Fatalf("trailing newline lost: %q", sealed)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(sealed, &obj); err != nil {
		t.Fatalf("sealed line is not JSON: %v", err)
	}
	if _, ok := obj["account_ref"]; ok {
		t.Fatalf("plaintext account_ref survived: %s", sealed)
	}
	if string(obj["latency_ms"]) != "12345678901234567" {
		t.Fatalf("number precision changed: %s", obj["latency_ms"])
	}
	if !strings.Contains(string(sealed), "<b>&") {
		t.Fatalf("HTML-escaping changed the text: %s", sealed)
	}
	var sealedValue string
	_ = json.Unmarshal(obj[redact.SealedField], &sealedValue)
	got, ok := redact.OpenAccount(sealedValue, dir)
	if !ok || got != sentinelRef {
		t.Fatalf("OpenAccount = %q, %v; want %q", got, ok, sentinelRef)
	}
	// 같은 값을 다시 봉인하면 다른 바이트 — 줄 사이 상관이 남지 않음.
	again, _ := redact.SealRecordLine(line, key)
	if bytes.Equal(again, sealed) {
		t.Fatalf("sealing is deterministic; two lines would correlate")
	}
	info, err := os.Stat(filepath.Join(dir, redact.KeyFileName))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("key file: %v, mode %v; want 0600", err, info)
	}
}

func TestA131OpenAccountFailsClosedWithoutTheKey(t *testing.T) {
	dir := t.TempDir()
	sealed, err := redact.SealRecordLine([]byte(`{"account_ref":"`+sentinelRef+`"}`), prepare(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]string
	_ = json.Unmarshal(sealed, &obj)
	other := t.TempDir()
	if got, ok := redact.OpenAccount(obj[redact.SealedField], other); ok || got != "" {
		t.Fatalf("opened without the key: %q, %v", got, ok)
	}
	if _, err := os.Stat(filepath.Join(other, redact.KeyFileName)); !os.IsNotExist(err) {
		t.Fatalf("a reader created a key file: %v", err)
	}
}

func TestA131PreviewRecordLineWritesNoKey(t *testing.T) {
	dir := t.TempDir()
	changed, err := redact.RecordLineNeedsRewrite([]byte(`{"account_ref":"` + sentinelRef + `"}`))
	if err != nil || !changed {
		t.Fatalf("RecordLineNeedsRewrite = %v, %v; want true, nil", changed, err)
	}
	unchanged, err := redact.RecordLineNeedsRewrite([]byte(`{"symbol":"005930"}`))
	if err != nil || unchanged {
		t.Fatalf("clean line reported as needing rewrite: %v, %v", unchanged, err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("preview touched the directory: %v", entries)
	}
}

func TestA131EmptyAccountRefIsLeftAlone(t *testing.T) {
	line := []byte(`{"account_ref":"","symbol":"005930"}`)
	changed, err := redact.RecordLineNeedsRewrite(line)
	if err != nil || changed {
		t.Fatalf("an empty account_ref is not account information: changed=%v err=%v", changed, err)
	}
}
