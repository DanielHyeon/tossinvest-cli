# codex — 독립 적대 리뷰(세 축 통합)

아래 공통 브리프를 따른다(이 파일 아래에 붙어 있다). 너는 읽기 전용 샌드박스에서 돈다 — 시험 실행이 안 되면 코드 인용으로 판정하고 그렇다고 적어라.
작업 디렉터리는 `git archive 00e1b9bd` 트리다(git 이력 없음). 커밋 diff 는 공통 브리프의 파일 목록으로 재구성하라.

필수 판정(각각 한 줄 + 근거):

1. **오늘-동등성 핀의 격리 논증.** `internal/app/engine/a112_owner_scope_handoff_test.go` `TestTwoOwnerScopesStillPlaceNothingBecauseTheFirstLegGuardRefuses` 가
   범위 하나 대조로 「거절은 B2(`strategy_account_first_leg_authority.go:217` `len(proposal.entries) != 1`)의 개수 조건에서 왔다」고 귀속한다.
   두 실행 사이에 개수 말고 다른 차이가 있어 B2 의 다른 조건 · 다른 관문이 거절했을 가능성, 그리고 B2 를 지워도 주문 0 이 다른 관문으로 유지되는지를 판정하라.
2. **RED 생략 시험들의 변이 대체.** review 5.2.2.1 절이 RED 를 거치지 않았다고 적은 시험들 각각을 유일하게 잡는 변이가 원장
   (`openspec/changes/a112-run-four-strategy-families-independently/analysis/measurements/lot-5.6.2-5.2.2/mutation-5.2.2.1.tsv`)에 있는가.
   없는 것, 그리고 원장 밖에서 살아남을 변이(구조 못 `TestTheProductionCycleDeliversEveryOwnerScopeHandoff` 의 우회 포함)를 제시하라.
3. **census 소스 유도의 완전성.** `internal/app/engine/strategy_dispatch_handoff_guard_test.go` `strategyHandoffAdmitDoors` 의 유도 규칙
   (수신자 없는 공개 함수, 결과 타입 문자열에 `Handoff`)이 경계 값을 만드는 모든 문을 보는가 — 동결 표면(`internal/strategyhandoff/escape_test.go`)이 먼저 막는 모양과
   둘 다 못 보는 모양을 구분하라.
4. 그 밖에 5.6.2.1(`runMarket` B13 · `EntryGate`) · 5.2.2.1(`dispatchHandoffs` · `deliverEachStrategyHandoff` · `AdmitEachOwnerScope`)에서 토글 OFF ≠ upstream,
   손절 · 청산 경로 약화, 안전 루프 정지 가능성이 있으면 P0 로.

---
# a112 5.6.2.1 · 5.2.2.1 다각 적대 리뷰 — 공통 브리프

## 안전 규칙(먼저 읽는다)

- **리뷰 트리는 `/tmp/claude-1000/a112-review-00e1b9bd`**(`git archive 00e1b9bd` 전체). 읽기는 여기서 한다.
- 실제 저장소 `/mnt/D/Axipient/workspace/TossOS` 는 **읽기 전용**이다. git 은 `git -C /mnt/D/Axipient/workspace/TossOS show|log|diff` 만 쓴다.
  commit · checkout · stash · reset · add · init · config 등 상태를 바꾸는 git 명령 금지. 그 저장소 안에 파일을 만들거나 고치지 마라.
- 시험 · 변이를 돌리려면 **자기 전용 사본**을 만든다: `set -euo pipefail; D=$(mktemp -d /tmp/claude-1000/a112-rev-XXXX); cp -a /tmp/claude-1000/a112-review-00e1b9bd/. "$D"/; cd "$D"`.
  사본 안에서도 `git init` 하지 마라. 변이는 그 사본에서만, 한 번에 하나, 끝나면 되돌린다. `GOFLAGS=-trimpath` 를 쓰고 전용 `GOCACHE` 는 선택.
  리뷰 트리(`a112-review-00e1b9bd`) 자체는 고치지 마라 — 다른 리뷰어와 공유한다.
- LIVE 주문 · 운영 토글 · 엔진 기동 · `mutating: true` 명령 금지. 운영 원장(`~/.config/tossctl/`) · 자격 증명 · `~/.codex` 열지 마라.
- `go test ./internal/app/engine/` 전체는 7~8분 걸린다. 필요하면 `-run` 으로 좁혀라.

## 대상 — 커밋 셋(브랜치 `feat/a112-four-family-runtime`, 푸시됨)

- `3260f4eb` — 사유 코드 `execgw.ReasonStrategyCentralIntegrity`(`strategy_central_integrity`): `reason.go` · `failclosed.go` `AllReasonCodes` · 골든 · `retry.go` `latchOrder` · a098 census · 순서 시험.
- `36ade9b2` — **5.6.2.1**: `internal/app/engine/strategy_entry_supervisor.go` `runMarket` B12(권한 갱신 전용 worker 갈래 — 오늘 생산의 유일한 구성) 안에
  새 B13: 중앙 무결성 오류면 `blockEntryOnCentralIntegrity` 로 `EntryGate.Block` 후 `continue`(루프 생존), 게이트 없으면 기존 `signalCentral`.
  옵션 `EntryGate StrategyEntryBlocker`, 생산 생성자 `NewRefreshingPairedStrategyEntrySupervisor` 는 `c.Entry == nil` 거절 + 전달.
  시험 `a112_central_integrity_entry_gate_test.go` · `…_internal_test.go` · `a112_fault_scope_test.go`.
- `00e1b9bd` — **5.2.2.1**: `internal/strategyhandoff/handoff.go` `AdmitEachOwnerScope` · `ownerScope(Of)`,
  `internal/app/engine/strategy_dispatch_handoff.go` `dispatchHandoffs`(서명 활성화 시에만 범위별) · `deliverEachStrategyHandoff`(첫 오류에서 멈춤),
  `strategy_entry_supervisor.go` `runProductionStrategyMarketCycle` 마지막 문장, census(`strategy_dispatch_handoff_guard_test.go`) · 동결 표면(`escape_test.go`) · import 허용 목록(`dependency_closure_test.go`),
  시험 `internal/strategyhandoff/owner_scope_test.go` · `internal/app/engine/a112_owner_scope_handoff_test.go`(태그 `tossos_testseams`).

읽을 것: 위 diff(`git -C … show <sha>`), `openspec/changes/a112-run-four-strategy-families-independently/` 의 `tasks.md` 5.2.2.1 · 5.2.2.2 · 5.6.2.1 · 5.6.2.2 절,
`review.md` 끝의 두 절(「2026-09-30 태스크 5.6.2.1」 · 「2026-09-30 태스크 5.2.2.1」), `HANDOFF.md` 의 「결정 (1)(5)(6) 기록」,
`analysis/harness/a112_lot_mutate.py` · `analysis/measurements/lot-5.6.2-5.2.2/`(RED 로그 · 변이 원장 · 편집 전 번들),
`analysis/function-logic/internal-app-engine--context.runproductionstrategymarketcycle/` · `…--strategyentrysupervisor.runmarket/`,
`analysis/goldens/four-family-runtime-v1.json`(queue 블록), `specs/` 의 해당 요구.
**`analysis/review-5621-5221/` 의 다른 파일은 열지 마라**(이 브리프와 자기 프롬프트만 예외).

## Manager 판정(배경 — 이것 자체를 다시 다툴 필요는 없다, 어긋난 구현만 지적)

- 결정 (6): fail-closed 수단은 프로세스 정지가 아니라 `EntryGate.Block`. Q1 = B12 삼킴만 바꾼다(B14 effective worker · 장부 고장 넷은 이월).
  Q2 = 새 사유 코드, 해제는 재시작뿐(해제 명령 부재는 의도).
- 5.2.2 분할: 5.2.2.1 = 소유자 범위별 Admit 경계 + 활성화 시장 한정 상한 상향, 하류 권한 무접촉. 5.2.2.2 = 하류를 범위 단위로(6.2 봉인 뒤).
  1차 레그 다섯 줄(`strategy_account_first_leg_authority.go` :217 · :221–:225)은 6.2 봉인까지 무변경(결정 (1)).
- HANDOFF 편차 승인: `Single` 서명 변경 대신 handoff 복수화.

## 판정 기준

- 모든 주장은 파일:줄 증거와 함께. 가능하면 **실행으로** 보여라(사본에서 시험 · 변이). 추측은 추측이라고 적어라.
- 등급: **P0** = 안전 불변식 위반 · 생산 동작 변화(토글 OFF ≠ upstream) · 주문/손절 경로 약화 · **P1** = 착지 상태로 둘 수 없는 결함(거짓 증거 · 뚫린 핀 포함) ·
  **P2** = 고칠 것 · (T) = 문서만.
- 안전 불변식: 손절 · 비상 청산 즉시성 약화 금지 · 토글 OFF = upstream · 사람 승인 없는 LIVE side effect 금지 · 엔진 정지 = 손절 없음(안전 루프를 내리는 변경 금지).

## 출력

맨 위 **판정 한 줄: APPROVE / BLOCK**. 발견 표(`# | 등급 | 주장 | 증거(파일:줄 · 실행 명령/출력) | 권고`), 그 아래 자기 프롬프트의 필수 항목별 한 줄 판정.
마지막 줄 `Recommendation: <action> because <reason>`. 만든 사본 경로를 적고, 끝나면 그 사본을 지워라.
