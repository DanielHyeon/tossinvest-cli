# Branch Test Map: `engineRuntime`

편집 전(base `2f698db6`). 분기 좌표는 a092 아카이브 번들(같은 source_sha256)과 같다 — 진입 귀속은 그 번들의 측정을 인용한다
(`openspec/changes/archive/2026-09-30-a092-an-alert-does-not-hold-the-stop/analysis/function-logic/cmd-tossctl--engineruntime/branch-test-map.md`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:639` 체결 감지 생성 실패 | 없음(a092 측정: 시험 0) — a090 무편집 | no | no |
| B2 | `:648` 대사 드라이버 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` (`cmd/tossctl/engine_runtime_branch_test.go:49`) | no | yes |
| B3 | `:659` exit 관측자 생성 실패 | 같음 | no | yes |
| B4 | `:664` 복구 생성 실패 | 같음 | no | yes |
| B5 | `:668` 전략 진입 외곽 생성 실패 | 없음(a092 측정: 시험 0) — a090 무편집 | no | no |
| B6 | `:680` 배달 실행자 생성 실패 | 없음(a092 측정: 시험 0) — a090 무편집 | no | no |

## 필요한 RED (a090)

- R17(tasks 2.17): 이 함수가 만드는 관측자 옵션에 `UnobservedLog: logger` 가 있고 `Log` 가 **없음**(구조 핀 — cmd 시험은 엔진 패키지의
  `OptionsForTest` 에 닿지 못하므로 `go/parser` 로 이 함수의 `ExitObserverOptions` 리터럴 키를 센다) + 엔진 패키지 시험에서 `Context.ExitObserver` 가
  `UnobservedLog` 를 통과시키고 그 로거에 새 줄이 나옴(행동).
