# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go` (편집 전 `437`–`607`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 52b6273fe6b3d08d…) — branches 15 · returns 11 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** 엔진 Context 조립 — 설정 · 감사 로그 · 계좌 · 저널 · 훅 바인딩 · 게이트웨이 · Guardian · 인터록. 첫 실패에서 거절.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 `assembleEngine`(cmd/tossctl/engine_assembly.go:14, `Options.Logger = logger`) · `New`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Logger` | engine.log 로거(생산) · nil(시험) | engine_assembly.go:24 | nil 이면 무출력 |
| `opts` | Options | 호출자 | 각 단계 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:439` resolveEnginePaths 오류 | — | `:440` | 무편집 |
| B2 | `:446` 감사 로그 열기 오류 | — | `:447` | 무편집 |
| B3 | `:450` clk nil | System 시계 | — | 무편집 |
| B4 | `:464` Publisher 주입 | — | — | 무편집 |
| B5 | `:472` 감사 기록 오류 | — | `:477` | 무편집 |
| B6 | `:487` 계좌 해석 오류 | — | `:488` | 무편집 |
| B7 | `:494` 저널 열기 오류 | — | `:495` | 무편집 |
| B8 | `:503` bindApplyHooks 오류 | jrn.Close | refuseStartup `:505` | a130 편집 자리 — 호출 하나를 `bindApplyHooksLogged(jrn, opts.Logger)` 로 |
| B9 | `:522` 게이트웨이 오류 | — | `:524` | 무편집 |
| B10 | `:537` Guardian 생성 조건 | — | — | 무편집 |
| B11 | `:539` factory nil | — | — | 무편집 |
| B12 | `:545` Guardian 오류 | — | `:547` | 무편집 |
| B13 | `:565` 인터록 오류 | — | `:567` | 무편집 |
| B14 | `:569` 미검증 | — | — | 무편집 |
| B15 | `:574` 오류 | — | `:576` | 무편집 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `bindApplyHooks` → `bindApplyHooksLogged` | 훅 바인딩 | 오류 → B8 | AST `:503` |

Live config binding: 없음.

## State mutations and fallbacks

- 없음(조립). a130: B8 의 조건식 안 호출 하나가 로거를 넘긴다.

## a130 편집 계획

B8 조건식의 호출 `bindApplyHooks(jrn)` → `bindApplyHooksLogged(jrn, opts.Logger)`. 분기 · 순서 · 반환 무변경. 구조 핀(go/parser)이 NewContext 의 바인딩 호출이 `opts.Logger` 를 넘김을 잰다.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 거절이던 성과가 행이 되는 방향, 거절 기록 추가.
