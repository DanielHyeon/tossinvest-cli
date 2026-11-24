# CodeGraph baseline (a199 1.4)

- 도구: CodeGraph CLI **1.6.0**, worktree 색인(`codegraph init .` → 56,265 nodes · 207,498 edges), HEAD `e4165fa3d`(코드 = `9738a0c32`).
- 원문: `codegraph-callers.txt`(수집 스크립트 `collect_callers.py`). `affected` 는 쓰지 않음(Go 시험 판별 결함 — 하네스 주석; 영향 시험은 grep + 타입 전수로).
- CodeGraph 는 함수 단위로 답하므로 **호출 자리 수**는 호출자 함수의 AST calls 로 셈(`../cited-ast/call_sites.py`, 원문 `../cited-ast/call-sites.txt`).

| 대상 | 생산 호출자(CodeGraph) | 호출 자리(AST) |
|---|---|---|
| `verifyGate` | `runInterlock` | 1 (call@400) |
| `runInterlock` | `NewContext` | 1 (@581) |
| `refuseStartup` | `NewContext` · `runInterlock` | 6 (@502·509·519·528·547·570) + 1 (@401) |
| `logGateDecision` | `runInterlock` | 3 (@396·402·419) |
| `NewContext` | `assembleEngine` · `New` (+ 시험 1) | 1 (@53) · 1 (@440) |
| `runEngineRun` | `newEngineRunCmd` | 1 (cobra RunE 클로저) |
| `Context.ExitObserver` | `engineRuntime` (+ 시험 19) | 1 (@671) |
| `Context.ReconcileDriver` | `engineRuntime` (+ 시험 5) | 1 (@662) |
| `judgeAttestation` | `attestationWatch.prepare` (+ 시험 6) | 1 (@323) |
| `newAttestationWatch` | `Context.AttestationWatch` (+ 시험 1) | 1 (@242) |
| `Context.AttestationWatch` | `engineRuntime` | 1 (@749) |
| `productionStrategyWorker` | `NewPairedStrategyEntryProductionAssembly` | 1 (@399, 시장 루프 안 — 실행 2회) |
| `execgw.AllReasonCodes` | 비시험 0 (시험 3; CodeGraph 가 `risk.AllReasonCodes` 시험 1 을 이름으로 섞음) | — |
| `att.Verify`(같은 술어) | `verifyGate` · `judgeAttestation` | 1 (@536) · 1 (@106) — 공용 판정 함수가 대체할 두 자리 |

불일치: `AllReasonCodes` 의 CodeGraph 결과에 다른 패키지(`internal/risk`)의 같은 이름 시험이 섞임 — 이름 기반 해소의 한계, grep 으로 해소(execgw 판은 비시험 호출자 0).
