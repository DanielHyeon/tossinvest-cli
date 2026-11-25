# Pre-Edit 선언 — time namespace 로트 2단계

- 기준: 브랜치 `feat/a166-time-namespace` `13be09fbd`(base `6ccc809d3`). Manager 결정 2026-11-25(6건, `review.md` · `investigation.md` §9).
- 위험 등급: **High-risk 취급**(결정 1) — 엔진 폴링 hint 복원(보호 가격 조회 보류와 묶임, design 73행).
- 편집 직전 재계수: `rg 'Namespace !=|Namespace =='` 생산 코드 4자리 — `internal/clock/bootstamp.go` L67 · L70 · L73 · L86(1단계와 같음).
- 편집할 기존 함수 4개(FLM/BTM 번들 `analysis/function-logic/internal-clock--*`, 편집 전 AST `pre-edit-ast/`):
  1. `RemainingBootHint`(생산) — 네 조건의 `Namespace !=` 항만 삭제. B8 분기는 남김(결정 2).
  2. `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` — ns 판을 수락 단언으로, uptime 역행 판 유지.
  3. `TestBootZeroDurationDoesNotHideInvalidClock` — ns 판을 수락으로, 같은 boot uptime 역행 + duration 0 거절 판 추가.
  4. `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` — 앞선 anchor 거절 유지, 다른 ns anchor 판을 수락 `{9s}` 로.
- 바꾸지 않는 것: `Snapshot` 의 표본 안 검사(L145 · L152 · L168) · `validStamp` · `BootStamp.Namespace` 필드 · `event.state` 형식 ·
  accountbudget 생산 코드 · census(`shapes.json`, clock/observation 키 0).
- 새 파일: `internal/clock/bootstamp_crossns_test.go`, `internal/accountbudget/observation_crossns_linux_test.go`(기존 시험 파일에
  넣지 않음 — 인접 함수가 gate ⑤ 번들로 잡히지 않게).
- 시험 범위: 선택 패키지만(`internal/clock` · `internal/accountbudget` · `internal/accountbudgetruntime` · `tools/a166-budget-state` ·
  `internal/execgw -run 'A166|EnginePoll'` 무태그/`tossos_testseams`), `make lint`. 전체 시험 · gate · sdd-check 미실행.
- 운영 · docker · 브로커 접근 0. 실제 커널 ns 재현은 `run.sh`(결정 3, 기본 스위트 밖).
- search-first: 새 헬퍼 없음 — 기존 `a166Store` · `a166Boot` fake 와 `RemainingBootHint` 만 쓴다. 새 의존성 0.
- UI 비해당(backend). 주석은 한글 「~함/~음」체(규칙 8).
