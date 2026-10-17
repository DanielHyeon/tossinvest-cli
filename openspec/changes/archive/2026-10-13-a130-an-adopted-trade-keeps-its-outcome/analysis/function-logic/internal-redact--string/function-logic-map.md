# Function Logic Map: `String`

> **출처 — 겹친 change.** a133(`a133-an-alert-carries-no-account`)이 편집한 함수가 a130 창
> (base `af0c0ce7` → landed-commit `727a841e`) 안에 끼어서 요구됨 — 겹친 change. 편집 커밋은
> a133 의 `2c2daae7`(창 안). a130 은 이 함수를 편집하지 않음.
> AST 는 a130 쪽에서 `go run ./tools/logic-map --file internal/redact/redact.go --func String` 으로
> **다시 추출**함. 추출한 소스 = `727a841e` 의 파일(워킹트리 · HEAD `272dba00` 와 바이트 동일,
> 파일 sha256 `aff3b9a6e1a809b7abddacba2557c1f65d1d84f70668be42418e128c0f8ff31e`).
> 재추출한 `ast.json` 은 a133 번들 `openspec/changes/a133-an-alert-carries-no-account/analysis/function-logic/internal-redact--string/ast.json`
> 과 바이트 동일(`cmp` 일치, ast.json sha256 `9bcf6b9bf087d465…`). 분기 · 시험 대응은 a133 번들을 출처로 삼되
> 각 행을 아래 Branch Test Map 의 시험별 coverprofile 로 다시 잼.

- Source: `internal/redact/redact.go` (79-109)
- Qualified: `String`
- AST evidence: `ast.json` (`revision: current` = `727a841e`, `source_sha256` aff3b9a6e1a809b7…) — AST branches 7 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 문자열 안의 계좌 정보를 치환함 — 등록 원문 → 등록 숫자열 → 마스크 꼴 → 유도 id(고정점까지 되풀이, a133).
호출자: obs 로거 · audit · 기록 봉인 · 전송 경계 `redactForTransport`(a133 FLM 기준). a130 의 편입 거래 성과 경로는
이 함수를 부르지 않음 — 창 안에 있어서 요구될 뿐임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | 임의 문자열 | 호출자 | 빈 값은 그대로(B1) |
| `reg.sorted` · `reg.digits` | 프로세스가 등록한 계좌 | `Register` | 읽기 잠금(`reg.mu.RLock`) 아래에서만 읽음 |

## Branches and early returns

| Branch | AST kind | Source line | 의미 |
|---|---|---|---|
| B1 | if at 80:2 | `if s == "" {` | 빈 문자열 → 그대로 반환(81:3) |
| B2 | range at 84:2 | `for _, ref := range reg.sorted {` | 등록 원문 순회 |
| B3 | if at 85:3 | `if strings.Contains(s, ref) {` | 등록 원문 치환 |
| B4 | if at 89:2 | `if len(reg.digits) > 0 {` | 등록 숫자열 규칙 적용 여부 |
| B5 | if at 91:4 | `if reg.digits[onlyDigits(run)] {` (클로저) | 숫자열 일치 → `Placeholder`(92:5), 불일치 → 원문(94:4) |
| B6 | for at 102:2 | `for {` | 유도 id 치환 되풀이(a133) |
| B7 | if at 104:3 | `if next == s {` | 고정점 → 반환(105:4), 아니면 `s = next` 후 다시 |

Exact AST return positions: 81:3, 92:5, 94:4, 105:4

가운데 둘은 `ReplaceAllStringFunc` 클로저 안의 반환임(함수 자체의 조기 반환은 81줄 하나).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reg.mu.RLock` · `reg.mu.RUnlock` | 등록부 읽기 잠금 | 97줄에서 해제(조기 반환은 잠금 전 81:3 뿐) | AST calls |
| `strings.Contains` · `strings.ReplaceAll` | 등록 원문 | 순수 | AST calls |
| `digitRun.ReplaceAllStringFunc` · `onlyDigits` | 등록 숫자열 | 순수 | AST calls |
| `maskPattern.ReplaceAllString` | 마스크 꼴 | 순수 | AST calls |
| `derivedIDPattern.ReplaceAllString` | 유도 id | 순수 — 대체 문자열 `-[redacted]` 는 다시 일치하지 않아 B6 이 끝남 | AST calls |

Live config binding: 없음(전역 등록부만).

## State mutations and fallbacks

- 지역 `s` 만 바꿈. 전역 등록부는 읽기만 함.

## Safety conclusion

- a130 편집 없음 — 이 번들은 창 안 남의 편집(a133)을 기술하는 증거임.
- High-risk impact: no — 로그 · 기록 · 알림 바이트만. 주문 · 손절 · 원장 판정 입력 아님.
