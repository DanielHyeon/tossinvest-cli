# Function Logic Map: `Client.accountsLocked`

- Source: `internal/official/reads.go` (38-57)
- Qualified: `Client.accountsLocked`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` c7a24ed74925c48e…) — AST branches 5 · returns 3
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 공식 `/api/v1/accounts` 를 읽어 계좌 목록을 만들고 암묵 accountSeq 를 고정/검증한다. 계좌 번호가 프로세스에 들어오는 유일한 입구(호출자 `Accounts` · `ensureAccountSeq`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | — | 호출자 | c.get 이 존중 |
| `c.accountSeq`·`accountSeqExplicit` | 0 또는 양수 | 클라이언트 상태 | 표류면 오류 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 40:2 | `if err := c.get(ctx, "/api/v1/accounts", nil, &raw); err != nil {` | 무편집 |
| B2 | if at 44:2 | `if selected == 0 && len(raw) > 0 && raw[0].AccountSeq > 0 {` | 무편집 |
| B3 | else at 46:9 | `} else if selected > 0 && !c.accountSeqExplicit {` | 무편집 |
| B4 | if at 46:9 | `} else if selected > 0 && !c.accountSeqExplicit {` | 무편집 |
| B5 | if at 47:3 | `if len(raw) == 0 \|\| raw[0].AccountSeq <= 0 \|\| int64(raw[0].AccountSeq) != selected {` | 무편집 |

Exact AST return positions: 41:3, 48:4, 56:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.get` | 브로커 읽기 | 오류면 그대로 반환(B1) | AST calls |
| `adaptAccounts` | 도메인 변환 | 순수 | AST calls |
| `redact.Register`(a131 신설) | 받은 계좌 번호를 프로세스 치환 규칙에 등록 | 메모리 append — 오류 없음 | design D2 R0 |

## State mutations and fallbacks

- `c.accountSeq.Store`(B2). a131: 성공 반환 직전 `redact.Register(accountNo…)` — 반환값·오류·브로커 호출 불변.

## Safety conclusion

- Safe edit boundary: 마지막 `return adaptAccounts(raw), nil` 앞에 등록 한 줄. 분기 B1~B5 무편집.
- High-risk impact: no — 계좌 조회 경로지만 반환·오류·호출 수 불변. 등록은 메모리 쓰기 하나.
