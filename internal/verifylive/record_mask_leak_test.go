package verifylive

import (
	"bufio"
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// hashDigestShape 는 기록이 싣는 해시 digest 값의 모양임(`sha256:` + 16진). 해시 출력이라 계좌 글자를 품을 수 없음.
var hashDigestShape = regexp.MustCompile(`^sha256:[0-9a-f]+$`)

// recordStringsContaining 은 기록 줄마다 JSON 을 풀어 **키와 문자열 값** 중 needle 을 품은 것을 돌려줌.
// 원 바이트 부분 문자열 검사는 실행마다 바뀌는 16진 digest 와 우연히 일치해 0.17% 로 실패했음(a164 종결 실측 5/3000).
// 그래서 값 전체가 해시 digest 이거나 봉인 봉투(`account_ref_sealed`, 암호문)인 것만 건너뜀 — 둘 다 평문 계좌를 실을 수 없는 모양임.
// 숫자 값(pid·지연·수량)은 보지 않음 — 계좌 참조는 기록에서 늘 문자열임(`Entry.AccountRef` 등).
func recordStringsContaining(t *testing.T, data []byte, needle string) []string {
	t.Helper()
	var hits []string
	var walk func(key string, v any)
	walk = func(key string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, e := range x {
				if strings.Contains(k, needle) {
					hits = append(hits, "key "+k)
				}
				walk(k, e)
			}
		case []any:
			for _, e := range x {
				walk(key, e)
			}
		case string:
			if key == "account_ref_sealed" || hashDigestShape.MatchString(x) {
				return
			}
			if strings.Contains(x, needle) {
				hits = append(hits, key+"="+x)
			}
		}
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 1<<16), 1<<24)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var v any
		if err := json.Unmarshal(line, &v); err != nil {
			// 못 푸는 줄은 건너뛰지 않음 — 원 바이트로 판정해 놓치는 일이 없게 함.
			if bytes.Contains(line, []byte(needle)) {
				hits = append(hits, "undecodable line")
			}
			continue
		}
		walk("", v)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning the record: %v", err)
	}
	return hits
}
