package strategyhandoff

// 경계 패키지의 생산 소스를 **gofmt 정본 바이트 digest** 로 동결한다(Manager 판정 2026-10-01 — 리뷰 수리 4차, 안 (b)).
//
// **왜 동결인가.** 「경계 값은 두 문(Admit · AdmitEachOwnerScope)과 Deliver 에서만 만들어진다」를 모양으로 세는 검사는 세 라운드
// 연속 새 우회를 남겼다 — 주조 축의 추격은 구성상 끝나지 않는다(각 판이 한 모양을 닫으면 다음 판이 다른 모양을 찾는다). 이 패키지는
// ~270줄 · import 폐쇄(errors · strings · strategyflow — reflect · unsafe 없음) · 주문 경로의 안전 결정 자리이므로, 모양을 세는
// 대신 **바이트를 고정한다**: 생산 `.go` 파일 전부를 gofmt 정본으로 정규화해 이어 붙인 것의 SHA-256 이 아래 상수와 같아야 한다.
// 그러면 어떤 편집이든 — 모양이 무엇이든 — 이 시험을 깨고, 통과시키는 유일한 방법은 **같은 커밋에서 상수를 다시 적는 것**이다.
//
// **남는 약점과 그것이 문제가 아닌 이유.** 편집과 함께 상수를 갱신하는 변이는 통과한다. 그러나 그것은 diff 에 보이는 **의도적
// 재고정**이다 — 무음 우회가 아니다. 변이 배터리가 겨누는 과녁은 「리뷰어가 못 보는 통과」이고, 상수 갱신은 정의상 리뷰어가 보는
// 행위이므로 그 과녁 밖이다. 이 시험이 지키는 성질은 「이 패키지가 리뷰 없이 바뀌지 않는다」이고, 그 성질은 종결한다.
//
// **재고정 절차.** 이 패키지의 생산 소스를 바꾸는 커밋은 (1) 이 상수를 시험 실패 메시지가 알려 주는 새 값으로 **같은 커밋에서**
// 바꾸고, (2) 커밋 메시지 · 리뷰에 「strategyhandoff 소스 동결 재고정」과 바뀐 이유를 적고, (3) 독립 리뷰가 그 diff 를 본 뒤에만
// 착지한다. 주석 · 공백만 바꾸는 편집도 재고정이다(gofmt 가 지우지 않는 바이트는 전부 digest 에 들어간다 — 주석도 계약의 일부다).
// gofmt 가 정규화하는 차이(들여쓰기 · 정렬)만으로는 digest 가 바뀌지 않는다.
//
// **이 동결이 닫는 실증 우회 목록**(모양 census 가 못 본 것 — 전부 이 패키지의 생산 소스 편집을 요구하므로 여기서 멈춘다):
//   1. 비공개 별칭(`type hiddenDelivered = Delivered`)을 돌려주는 비공개 수신자 메서드 + `init()` 에서 공개 var 재대입
//      (codex 재확인 #3, 99ad897c).
//   2. 결과 없는 out-param 주조 함수(같은 재확인 — 1판 census 가 결과 없는 함수를 건너뜀).
//   3. 제네릭 주조 `func makeSeam[T ~struct{result strategyflow.Result}](r) T { return T{result: r} }` + `any` 반환 메서드 +
//      매개변수 포인터 경유 재대입 `replaceError(&ErrNoDelivery)`(codex 3차, 618b1002 — 타입 동일성 census 도 못 봄).
//   4. 같은 필드 모양의 비공개 쌍둥이 구조체 변환 `Delivered(twin{r})` — 리터럴을 세는 어떤 census 에도 안 걸리는 주조
//      (구현자 자체 점검, 2026-10-01). 문서상의 규칙으로 적는다: **경계 타입으로의 변환도 주조다 — 문 밖에서 금지.**
// 이 목록 밖의 모양도 같은 이유로 닫힌다 — 이 목록은 「동결이 필요했던 증거」이지 동결의 범위가 아니다.
//
// mint_census_test.go · escape_test.go 의 모양 검사는 「왜 이 두 문뿐인가」의 설명으로 남는다. **그들의 완전성 주장은 철회됐다** —
// 종결은 이 동결이 진다.

import (
	"crypto/sha256"
	"encoding/hex"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// frozenSourceDigest 는 이 패키지 생산 소스의 gofmt 정본 digest 다. 재고정은 위 절차대로만.
const frozenSourceDigest = "sha256:1ad1078735a477c6f9ce61fa4eacd1e8cd093e5581f58ce1144e9d5e74584b22"

func productionSourceDigest(t *testing.T) (string, []string) {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	sort.Strings(paths)
	hash := sha256.New()
	var files []string
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		canonical, err := format.Source(raw)
		if err != nil {
			t.Fatalf("gofmt %s: %v", path, err)
		}
		files = append(files, path)
		// 파일 이름을 digest 에 넣는다 — 파일을 나누거나 새로 더하는 것도 재고정이다.
		_, _ = hash.Write([]byte(path + "\x00"))
		_, _ = hash.Write(canonical)
		_, _ = hash.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), files
}

func TestTheSeamSourceIsFrozen(t *testing.T) {
	got, files := productionSourceDigest(t)
	if len(files) == 0 {
		t.Fatal("no production source was hashed, so this freeze proves nothing")
	}
	if got != frozenSourceDigest {
		t.Fatalf("strategyhandoff production source (%v) changed: digest %s, frozen %s.\n"+
			"This package is frozen: re-pin frozenSourceDigest to the new value IN THE SAME COMMIT as the edit, name the re-pin "+
			"in the commit message and review, and land only after an independent review has read the diff (see the header).",
			files, got, frozenSourceDigest)
	}
}

// 양성 대조: digest 는 gofmt 가 정규화하는 차이에 둔감하고, 그 밖의 한 바이트에는 민감하다.
func TestTheFreezeDigestIgnoresOnlyWhatGofmtNormalises(t *testing.T) {
	source := []byte("package p\n\nfunc F() int {\n\treturn 1\n}\n")
	reindented := []byte("package p\n\nfunc F() int {\n        return 1\n}\n")
	edited := []byte("package p\n\nfunc F() int {\n\treturn 2\n}\n")
	canonical := func(b []byte) string {
		out, err := format.Source(b)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(out)
		return hex.EncodeToString(sum[:])
	}
	if canonical(source) != canonical(reindented) {
		t.Fatal("a gofmt-only difference changed the canonical digest")
	}
	if canonical(source) == canonical(edited) {
		t.Fatal("a one-byte edit did not change the canonical digest — the freeze is blind")
	}
}
