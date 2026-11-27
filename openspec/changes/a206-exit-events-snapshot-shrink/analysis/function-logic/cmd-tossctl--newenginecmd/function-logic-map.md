# Function Logic Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go` (115-136, 편집 뒤 — base 115-134)
- Qualified: `newEngineCmd`
- Revision: `current` (편집 뒤, `source_sha256` 5332df2c43843d08…; base 파일 cf8e80ad009ec7b3…)
- AST evidence: `ast.json` — AST branches 0 · returns 1(135:2)
- Risk scan: `risk-pattern-report.md`

**역할.** `tossctl engine` 묶음 명령을 만들고 하위 명령을 붙여 돌려줌. 생산 호출자 1(`newRootCmd`).
a206 편집은 `cmd.AddCommand(newEngineExitEvidenceCompactCmd(root))` 한 문장 추가(+ 주석 한 줄)뿐. 분기 · early return · 상태 변경 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | `*rootOptions`(nil 허용 — 하위 명령 생성자가 그대로 받음) | `newRootCmd` | 생성 시점에는 읽지 않음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 직선 함수(AST branches 0) | 지역 `cmd` 에 하위 명령 등록 | 135:2 `return cmd` | `TestA206CompactCommandIsMutatingAndAttached` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.AddCommand` | 하위 명령 등록 | 오류 없음 | AST calls |
| 기존 생성자 7(run · reconcile-resolve · alerts · risk relaxation · mode-release · attempt-resolve · heartbeat-check) | 기존 하위 명령 | 생성만 | AST calls — 무변경 |
| `newEngineExitEvidenceCompactCmd`(a206 추가) | `engine exit-evidence-compact` | 생성만. 실행 시 엔진 잠금을 잡고 저널을 엶(--apply 때만 쓰기) | AST calls · `a206_exit_evidence_compact.go` |

## State mutations and fallbacks

- 지역 `cmd` 만 만듦. fallback 없음.

## Safety conclusion

- Safe edit boundary: 하위 명령 하나를 맨 뒤에 추가. 기존 하위 명령의 생성 · 순서 · 표지 무변경.
- High-risk impact: 등록 자체는 no. 새 leaf 는 `mutating=true`(원장 행을 고침) — 자동 실행 금지 표지는 `TestMutatingAnnotationOnTradeCommands` 가 고정함.
