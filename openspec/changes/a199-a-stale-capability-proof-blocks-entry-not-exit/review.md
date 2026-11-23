# a199-a-stale-capability-proof-blocks-entry-not-exit Review

## 계약 작성 (2026-11-24)
- 작성: Manager(Fable). 사전 증거 조사: 읽기 전용 조사원(HEAD `9738a0c32`) — change 안 사본·HEAD 재확인은 tasks 1.3.
- 사용자 승인: 2026-11-24 「1번 올린다로 진행해줘」 「지금부터 작성하고 구현해줘」.
- 위험 등급 High-risk(기동 인터록 · 주문 권한 · 손절 경로 가용성).

## 1단계 — 등록 · 증거 · 제안 동결 리뷰 (팀메이트 Opus, 2026-11-24)

worktree `~/tossos-work/a199-impl`, 브랜치 `feat/a199-stale-proof-exit-only`. production·시험 Go 편집 0.

### 1.1 등록
- 계약서 넷을 옮김(`spec-delta-intent.md` 는 `analysis/`). MODIFIED 2건은 `analysis/modified-delta/generate.py` 로 생성 — 치환마다 1회 적중 단언, 역치환으로 정본 Requirement 바이트 동일 단언, 정본 Scenario 이름 보존 단언(A 12→17, B 18→20). 실행 원문 `analysis/modified-delta/run.txt`.
- 사실 정정(뜻 불변): intent A 의 추가 Scenario 자리 「`#### Scenario: dry-run` 앞 또는 뒤」는 다른 Requirement(Flatten Saga)의 Scenario 다 → 「자동화 게이트 기동 인터록」의 마지막 Scenario(「정수가 아닌 최대 보유 포지션 수」) 뒤에 둠. proposal 「영향」의 「ADDED 2~3」 → 「ADDED 없음」(intent C — A2 문단이 Requirement 안에서 strict 통과).
- B3 자리 수: 아래 「B3 자리표」.
#### B3 자리표 — Requirement 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 안의 「재시작이 거절」 계열

정확한 문자열 「재시작이 거절될 상태」는 **2자리**, 같은 뜻을 다른 꼴로 말하는 자리까지 **6자리**(grep `거절|거부`, 정본 2213-2297).

| # | 정본 줄 | 원문 | 처리 | 뜻 |
|---|---|---|---|---|
| 1 | 2216 | `**재시작이 거절되기 시작할 시각(기한)**` | `**재시작의 결과가 정상이 아니게 되는 시각(기한)**` | 기한의 값·계산 무변, 이름만 새 귀결에 맞춤(B3 과 같은 꼴). **intent 밖 — Manager 확인 Q** |
| 2 | 2220 | `재시작이 거절될 상태의 알림은 …` | B2 의 새 문장 전체 | intent B2 |
| 3 | 2221 | `재시작이 거절될 상태(만료 · 범위 상실 · 계좌 불일치 · 무효)는` | `재시작의 결과가 정상이 아닌 상태(…)는` | intent B3, 괄호·주기·키 무변 |
| 4 | 2221 | `거절 상태가 같은 한(` | `그 상태가 같은 한(` | 3 의 집합을 가리키는 줄임말 — 뜻 불변 |
| 5 | 2221 | `다시 거절 상태가 되면` | `다시 그 상태가 되면` | 같음 |
| 6 | 2297 | Scenario 「범위 상실 알림은 결과와 할 일을 말한다」 THEN 「기동이 거부된다 … 자동 손절이 돌지 않게 된다」 | 「새 진입이 막힌 청산 전용으로 뜬다 … 손절·청산은 돌고 새 진입만 막힌다」 | B2 가 정한 범위 상실의 새 귀결을 옮김. intent 는 「만료됨」「갱신이 주문 권한을 잃음」만 점검 대상으로 적었고 이 Scenario 를 빠뜨림 — 그대로 두면 Requirement 가 자기모순. **Manager 확인 Q** |

「만료됨」「갱신이 주문 권한을 잃음」 THEN 에는 귀결 문장이 없어 무편집(intent 대로).

- Story `STORY-TOS-a199`(FEAT-TOS-003 「Engine interlocks and Guardian controls」 — a196 과 같은 Feature), `_registry.yaml`, tracker 재생성 + `--check`.
- 그 밖의 intent 메모: A1 새 문장의 「아래 「낡은 증명」 절」은 A2 문단의 굵은 첫 문장(「capability attestation 의 낡음 … 진입 허가 조건이다」)을 가리키지만 그 이름의 머리는 없다(intent 문자열 그대로 둠 — 뜻 불변, Manager 가 원하면 문구만 조정). A3 의 THEN 에서 정본의 「재검증 안내가 출력된다」가 빠진다(intent 대로 — D8 의 기동 출력 한 줄이 대신함).

### 1.2 base
- `capture_change_base.py` → `base-commit.txt` = `e4165fa3dbc6e0b20dcec95f9e66f64cc5c21039`(등록 커밋 — 코드는 `9738a0c32` 와 같음).
- **a198 미포함**: base 고정 직전 `git fetch origin` 에서 `origin/main` = `9738a0c32`(a198 아카이브 커밋 없음). 지시대로 그때의 origin/main 위에서 고정했다. a198 이 `Context.ExitObserver` 를 편집하므로(edit-surface §3) a198 착지 뒤 base 재고정이 필요할 수 있다 — Manager Q1.

### 1.3 사전 증거 재확인
- `analysis/evidence-phase0.md`(사전 증거 원문 + 머리의 재확인 · 정정 3), AST `analysis/cited-ast/`(33개, `extract.py`, 요약 `summary.txt`). 사전 증거 AST 12개는 sha · 줄 · 분기 전부 같음.

### 1.4 CodeGraph
- `analysis/code-context/`(1.6.0, worktree 색인) 세 파일 + `codegraph-callers.txt`. 호출 자리 수는 호출자 AST 로 셈(`cited-ast/call-sites.txt`). 불일치 1(`AllReasonCodes` 이름 섞임) grep 으로 해소.

### 1.5 `Verified` 전수
- `analysis/verified-census/`(go/types 전수 — 133 패키지 타입 검사, 55 자리). `Verified` 읽는 자리 **8**: 청산 쪽 4(= design D3 의 넷 전부) · 진입 1(strategy worker, 엄격 유지) · 보고 3(logGateDecision 461·479, 전략 스냅숏 392). 판단 불가 0 → 멈춤 조건 아님. `EntryPermitted` 읽는 자리 2(보고).

### 1.6 · 1.7
- `analysis/d1a-d4-sites.md`. Verify 와 같은 계좌 비교의 공개 진입점 **없음**(`SameAccountMasked` 는 더 관대한 다른 규칙) → 새 파일의 새 메서드가 비공개 `sameAccount` 를 부르는 안. 래치 자리 권고 = `runInterlock` 안(+ `gateFacts` 칸, NewContext 리터럴 한 줄). 새 사유의 닫힌 목록 5곳(AllReasonCodes · golden · a098 등록 변수 · latchOrder · 순서 시험), 정본 닫힌 목록 없음(MODIFIED 불요).

### 1.8 편집 표면
- `analysis/edit-surface.md`. 편집할 기존 함수 확정 10(P1~P10) + 조건부 2(P11 `logGateDecision`, P12 soak.go 출력), 뒤집힐 기존 시험 함수 6. census 적중 1(`NewContext`). 겹침: **a198 이 `Context.ExitObserver` 를 편집**(순차 — a198 착지 뒤).

### 1.9 FLM/BTM pre-edit
- `analysis/function-logic/` 17 번들(생산 11 · 시험 6). 분기 표는 AST 열거로 생성(`analysis/harness/build_bundles.py`), GREEN 칸은 base 선택 실행 coverprofile:
  - `go test -count=1 -covermode=set -run 'TestGateOnRefusals|TestRefusedStartIsAudited|TestGateDecisionIsLoggedStructurally|TestNothingElseRefusesTheOperatorConfiguration|TestGateOffStartsAndDoesNoGateWork|TestEachInterlockClauseHasALine|TestA135|TestA196|TestTheDriverRefusesAnUnverifiedGate|TestA092ExitObserver|TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver|TestEntryIsPermittedOnlyWhenProtectionIsWired|TestTheGateStartsWithoutBrokerSideProtection|TestTheGateOffPathIsUntouched|TestA136|TestA138|TestA139' ./internal/app/engine` → ok 39.7s(`analysis/implementation/base-selected-engine.txt`)
  - `./cmd/tossctl -run 'TestAGateOffEngineRefusesWithoutEnumeratingClauses|TestAnUnmetInterlockIsEnumerated|TestANonInterlockFailureIsReportedAsItself|TestTheLoopSetIsTheSpecifiedThree'` → ok · `./internal/execgw -run 'TestReasonCodeEnumIsStable|TestTheSenderDownReason|TestTheStrategyCentralIntegrityLatch'` → ok.
  - 칸의 count 는 선택 집합 전체의 실측이고 시험별 귀속이 아님 — 시험 이름 확정은 2단계 RED. 이 선택에서 `ReconcileDriver` 성공 경로는 count 0(선택 밖).
- `python3 tools/logic-map/check_analysis.py --change a199-…` 원문: 「base e4165fa3dbc6 → working tree … required 0 function(s)」 · 「evidence complete or diff-proven exempt」(rc 0). Go 미편집이라 required 0 — 번들 내용은 2단계(required > 0)에서 판정된다.
- P12(soak.go) 번들은 만들지 않음 — 조건부(결정 뒤).

### 1.10 claim check
- `analysis/claim-check.md`. 거짓 · 부분 · 미확인은 아래 「Manager 에게 묻는 것」으로. 사실 정정 3을 design 에 반영(「가림」 예시, 「두 사례」→셋, §5 verifyGate 이름).
