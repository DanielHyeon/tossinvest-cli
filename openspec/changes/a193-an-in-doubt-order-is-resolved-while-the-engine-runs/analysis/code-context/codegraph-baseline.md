# CodeGraph baseline (a193, 1단계)

- 도구: codegraph **1.6.0**(`codegraph --version`). 이 worktree 전용 색인을 새로 만듦: `codegraph init -y .` → 파일 3,223 · 노드 55,788 · 엣지 205,618
  (`codegraph status .`). 공유 체크아웃의 `.codegraph/` 는 쓰지 않음(`.codegraph/` 는 `.gitignore` 로 무시됨).
- 기준 commit: `791fb6ef1`(worktree HEAD = base, 작업 트리의 Go 변경 0).
- 원문: `codegraph-callers.txt`(15 심볼), `codegraph-affected.txt`.
- 주의: callers 는 **이름** 단위라 같은 이름의 다른 함수가 섞인다(`Resolve` 15건에 `main`·`ResolvePaths`·시험 등). 아래 production 호출자는 HEAD 의 `rg` 로 다시 좁힌 값.

## 질의와 결과 (production 호출자만, 시험 제외)

| 질의 | 결과 | 쓰임 |
|---|---|---|
| `Resolve`(execgw.Resolver) | `Recovery.Run`(recovery.go:300), `Saga.resolveStep`(flatten.go:428) — `rg '\.Resolve\(ctx'` 로 재확인 2건 | design §1.1 · D6 |
| `ResolveUnresolved` | `Resolver.park`(indoubt.go:376), `Gateway.ReplayInDoubt`(replay.go) | D3 — 세션 해소는 재생을 안 부르므로 park 만 |
| `ResolveConfirmed` | `Resolve`, `resolveCancel`, `ReplayInDoubt`, `confirmAcked`(acked_boot.go) | D5 |
| `ResolveFailed` | `Resolve`, `resolveCancel`, `resolveAmend` | D5 |
| `waitOrExpire` | `Resolve`, `resolveCancel`, `resolveAmend`, `settleAmendFromClosedOriginal` | D3 — Resolver 수리를 고르지 않은 이유 3(넷이 같이 바뀜) |
| `park` | 같은 넷 | D3 |
| `absenceCorroborated` | `Resolve` 1 | §1.3 |
| `PendingAttempts` | `RecoverPending`, `Recovery.Run`, `Gateway.unsettledFor`(a094_shared_judgements.go:30) | D1 — 세션 실행자가 넷째 호출자 |
| `UnresolvedAttempts` | `Gateway.checkSymbolFree` 1 | D5 |
| `noteHeldProposal` | `ExitObserver.judge`(exitloop.go:987) 1 | §1.2-3 |
| `submit` | `ExitObserver.record`(exitloop.go:1314) — 같은 이름의 `Gateway.place/Cancel/Amend` 안 호출은 다른 함수 | §1.2 |
| `ExitIntentAttempts` | `noteHeldProposal` 1 | §1.2-3 |
| `RecordCritical` | exit 관측(`recordEpisodeCritical`·`raiseUnobserved`·…), 기동(`alertAcked`·`catchUpExitProposals`), 해동(`recordThawReleaseFailure`) 등 | D4 — 같은 입구 |
| `engineRuntime` | production 은 `runEngineRun`(engine.go:196) 이 `engineRuntimeFactory`(call@277) 로 부름 + 시험 4 | D2 배선 |
| `Run`(이름) | 다수 — `Recovery.Run` 의 호출자는 `engineRecoverySequence` 클로저(engine.go) | §1.1 |

## 영향 시험 (`codegraph affected`)

⚠ 기본 실행은 Go 시험을 못 찾는다(하네스 주석). 원문 `codegraph-affected.txt`(대상 `cmd/tossctl/engine.go` — 2단계에서 고칠 production 파일 하나; 새 파일은
그래프에 아직 없음):

| 명령 | 결과 |
|---|---|
| `codegraph affected cmd/tossctl/engine.go`(기본 판별식) | **1건** — `auth-helper/tests/test_cli.py`(Python, `/tests/` 로 걸림) |
| `codegraph affected cmd/tossctl/engine.go --filter '*_test.go'`(깊이 5) | **1,280건** — internal/app/engine 212 · internal/journal 154 · cmd/tossctl 108 · internal/execgw 84 · … |
| `codegraph affected -d 2 cmd/tossctl/engine.go --filter '*_test.go'` | cmd/tossctl 79 · internal/app/engine 51 · 그 밖 8 |

선택 시험(2단계, 전체 시험 아님): `./internal/app/engine` 의 a193 시험 + 보조 실행자·런타임 시험(`-run 'A193|Auxiliary|Runtime'`), `./cmd/tossctl` 의 a098·
engine_runtime_branch·a193 배선 시험, `./internal/execgw` 의 indoubt 시험(무편집 확인용 — `-run 'Resolve|InDoubt'`), `./internal/journal` 의 resolution 시험.
