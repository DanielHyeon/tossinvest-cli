# Function Logic Map: `renderFields`

> **출처 — 겹친 change.** a133(`a133-an-alert-carries-no-account`)이 편집한 함수가 a130 창
> (base `af0c0ce7` → landed-commit `727a841e`) 안에 끼어서 요구됨 — 겹친 change. 편집 커밋은
> a133 의 `35936cc8`(창 안). a130 은 이 함수를 편집하지 않음.
> AST 는 a130 쪽에서 `go run ./tools/logic-map --file internal/obs/notifier.go --func renderFields` 로
> **다시 추출**함. 추출한 소스 = `727a841e` 의 파일(워킹트리 · HEAD `272dba00` 와 바이트 동일,
> 파일 sha256 `780cd20d29ecbc3d20a1e76d876449db3c6195f5df1332e1a8d4505b140f02fd`).
> 재추출한 `ast.json` 은 a133 번들 `openspec/changes/a133-an-alert-carries-no-account/analysis/function-logic/internal-obs--renderfields/ast.json`
> 과 바이트 동일(`cmp` 일치, ast.json sha256 `c43fa2fde02ccffb…`). 분기 · 시험 대응은 a133 번들을 출처로 삼되
> 각 행을 아래 Branch Test Map 의 시험별 coverprofile 로 다시 잼.

- Source: `internal/obs/notifier.go` (1029-1046)
- Qualified: `renderFields`
- AST evidence: `ast.json` (`revision: current` = `727a841e`, `source_sha256` 780cd20d29ecbc3d…) — AST branches 3 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 알림 문맥 필드(`Event.Fields`)를 키 순서대로 `k: v` 줄로 만들어 알림 본문 끝에 붙일 문자열을 돌려줌.
계좌 키(`redact.IsAccountKey`)는 건너뜀(a133 D3). 호출자 `notificationFor`(a133 FLM · CodeGraph 기준).
a130 의 편입 거래 성과 경로(journal · engine)는 이 함수를 부르지 않음 — 창 안에 있어서 요구될 뿐임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `fields` | map — 호출자가 비어 있지 않을 때만 부름 | 사건 생산자 | 값은 `%v` 로 펼침. 계좌 키 이름이면 줄 자체를 안 씀 |

## Branches and early returns

| Branch | AST kind | Source line | 의미 |
|---|---|---|---|
| B1 | range at 1031:2 | `for k := range fields {` | 키 모으기 |
| B2 | if at 1034:3 | `if redact.IsAccountKey(k) {` → `continue` | 계좌 키 건너뜀(a133 신설) |
| B3 | range at 1042:2 | `for _, k := range keys {` | 정렬된 키마다 `k: v` 줄 쓰기 |

Exact AST return positions: 1045:2 (반환 1 — 조기 반환 없음).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `redact.IsAccountKey` | 계좌 키 판정 | 순수 | AST calls |
| `sortStrings` | 안정 순서 | 순수 | AST calls |
| `fmt.Fprintf` | 줄 쓰기 | `strings.Builder` — 오류 없음 | AST calls |
| `strings.TrimRight` | 끝 개행 제거 | 순수 | AST calls |

Live config binding: 없음.

## State mutations and fallbacks

- 지역 슬라이스 · 빌더만. 모든 키가 계좌 키면 빈 문자열을 돌려줌.

## Safety conclusion

- a130 편집 없음 — 이 번들은 창 안 남의 편집(a133)을 기술하는 증거임.
- High-risk impact: no — 알림 본문 바이트만. 주문 · 손절 · 원장 판정 입력 아님.
