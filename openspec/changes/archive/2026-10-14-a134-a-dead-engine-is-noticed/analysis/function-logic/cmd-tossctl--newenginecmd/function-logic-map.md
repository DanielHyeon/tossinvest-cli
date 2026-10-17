# Function Logic Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go` (115-134, 편집 뒤 — base 115-132)
- Qualified: `newEngineCmd`
- AST evidence: `ast.json` (편집 뒤 재추출, `source_sha256` f311f5ad47438819… · 편집 전 446e27a347768b0b…) — AST branches 0 · returns 1
- Risk scan: `risk-pattern-report.md` (configured risk pattern 0)

**역할.** `tossctl engine` 묶음 명령을 만들고 하위 명령을 붙여 돌려줌. 생산 호출자 1(`newRootCmd`, root.go:52 — CodeGraph `callers`).
a134 편집은 `cmd.AddCommand(newEngineHeartbeatCheckCmd(root))` 한 문장 추가(+ 주석 한 줄)뿐. 분기 · early return · 상태 변경 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | `*rootOptions`(nil 허용 — 하위 명령 생성자가 그대로 받음) | `newRootCmd` | 생성 시점에는 읽지 않음 — 실행 시 각 하위 명령이 해석 |

## Branches and early returns

| Branch | AST kind | Source line | a134 |
|---|---|---|---|
| (없음) | — | 직선 함수 — AST branches 0 | 편집 전후 동일(0) |

Exact AST return positions(편집 뒤): 133:2 (편집 전 131:2)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.AddCommand` | 하위 명령 등록(6회) | 오류 없음 | AST calls |
| `newEngineRunCmd` | `engine run` | 생성만 | AST calls |
| `newEngineReconcileResolveCmd` | `engine reconcile-resolve` | 생성만 | AST calls |
| `newEngineAlertsCmd` | `engine alerts` | 생성만 | AST calls |
| `newEngineRiskRelaxationCmds` | 진입 잠금 · latch 해제 · 확인 | 생성만 | AST calls |
| `newEngineModeReleaseCmd` | `engine mode-release` | 생성만 | AST calls |
| `newEngineAttemptResolveCmd` | `engine attempt-resolve` | 생성만 | AST calls |
| `newEngineHeartbeatCheckCmd`(a134 추가) | `engine heartbeat-check` — 엔진 밖 감시 | 생성만, 실행 시 엔진 · 저널을 열지 않음 | AST calls · `engine_heartbeat_check.go` |

## State mutations and fallbacks

- 지역 `cmd` 만 만듦. fallback 없음.

## Safety conclusion

- Safe edit boundary: 하위 명령 하나 추가. 기존 하위 명령의 생성 · 순서 · 표지 무변경.
- High-risk impact: no — 새 leaf 는 `source=local` · `mutating=false`, 주문 · 원장 · 설정을 바꾸지 않음.
