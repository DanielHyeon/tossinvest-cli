# Function Logic Map: `String`

- Source: `internal/redact/redact.go` (79-109, 편집 뒤 — 편집 전 79-100, HEAD 35936cc8 = base 와 같은 소스)
- Qualified: `String`
- AST evidence: `ast.json` (편집 뒤 재추출, `source_sha256` aff3b9a6e1a809b7… · 편집 전 4b78b4630bfdc61e…) — AST branches 7(편집 전 5) · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** a131 규칙 하나 — 문자열 안의 계좌 정보(등록 원문 → 등록 숫자열 → 마스크 꼴 → 유도 id)를 치환. 호출자: obs 로거 핸들러 · 엔진 Writer · audit · soak/검증 기록 봉인 ·
정리 명령(a131) · 전송 경계 `redactForTransport`(a133).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | 임의 문자열 | 호출자 | 빈 값은 그대로(B1) |
| `reg.sorted` · `reg.digits` | 프로세스가 등록한 계좌 | `Register`(공식 계좌 조회) | 읽기 잠금 아래 |

## Branches and early returns

| Branch | AST kind | Source line | a133 |
|---|---|---|---|
| B1 | if at 80:2 | `if s == "" {` | 무편집 |
| B2 | range at 84:2 | `for _, ref := range reg.sorted {` | 무편집 |
| B3 | if at 85:3 | `if strings.Contains(s, ref) {` | 무편집 |
| B4 | if at 89:2 | `if len(reg.digits) > 0 {` | 무편집 |
| B5 | if at 91:4 | `if reg.digits[onlyDigits(run)] {` (클로저) | 무편집 |
| B6 | for at 102:2 | `for {` — 유도 id 치환 되풀이 | **신설**(a133 I-1) |
| B7 | if at 104:3 | `if next == s {` → `return s` | **신설** — 고정점에서 반환 |

Exact AST return positions(편집 뒤): 81:3, 92:5, 94:4, 105:4 — 편집 전 마지막 99:2(유도 id 치환 한 번)가 B7 안 105:4(고정점)로 바뀜. B1~B5 는 같은 좌표.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ReplaceAll` | 등록 원문 | 순수 | AST calls |
| `digitRun.ReplaceAllStringFunc` | 등록 숫자열 | 순수 | AST calls |
| `maskPattern.ReplaceAllString` | 마스크 꼴 | 순수 | AST calls |
| `derivedIDPattern.ReplaceAllString` | 유도 id | 순수 — **결함(I-1)**: 뒤 경계 그룹 `([^0-9a-f]\|$)` 가 구분자를 먹어, 구분자 한 글자로 붙은 다음 id 의 앞 경계가 사라짐 | 구현 리뷰 I-1 · `analysis/red/red-redact.txt` |

## State mutations and fallbacks

- 지역 `s` 만. 전역 등록부는 읽기만(RLock).

## Safety conclusion

- Safe edit boundary: 마지막 return(99:2)만 — 유도 id 치환을 결과가 안 바뀔 때까지 되풀이. 대체 문자열 `-[redacted]` 는 hex 가 아니라 다시 일치하지 않으므로 판마다 일치가 줄고 끝남.
- High-risk impact: no(직접) — 로그 · 기록 · 알림 바이트만. 식별자 칸(`id`·`_id`·`digest`)은 호출자(`SealRecordLine`)가 애초에 이 함수에 안 넘김(a131 R-9).
