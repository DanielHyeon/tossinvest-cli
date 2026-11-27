# a206 분기 주장의 AST 근거

proposal · design 이 근거로 쓰는 **기존 함수 내부 분기 · 호출**은 이 디렉터리의 `tools/logic-map` 산출물로 열거함
(.claude/CLAUDE.md 필수 진입 5). 이 단계는 기존 함수를 편집하지 않음 — FLM 번들이 아니라 주장의 근거라서
`analysis/function-logic/` 이 아니라 이 자리에 둠(선례 a140 · a203 `analysis/ast-evidence/`).

- 기준: worktree HEAD `e9266bfb6`(origin/main), 2026-11-27.
- 생성: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수> > <패키지>--<함수>.ast.json`
- `source_sha256` 이 그 시점 소스 지문임. 소스가 바뀌면 이 주장은 다시 뽑아야 함.

| 산출물 | 파일:줄 | source_sha256(앞 12) | 분기 | return |
|---|---|---|---|---|
| `journal--hydrateExitEventEvidence` | internal/journal/exit_snapshot.go:392-417 | 19ebe0eeb987 | 6 | 7 |
| `journal--decodeExitEventSnapshot` | exit_snapshot.go:529-541 | 19ebe0eeb987 | 2 | 3 |
| `journal--validateExitEventArmSuppression` | exit_snapshot.go:451-527 | 19ebe0eeb987 | 20 | 17 |
| `journal--scanExitEvent` | exit_snapshot.go:333-351 | 19ebe0eeb987 | 2 | 3 |
| `journal--ReadOnly.AccountExitEvents` | internal/journal/account_views.go:303-358 | 26032b49b93b | 6 | 5 |
| `journal--ReadOnly.BrokerOrderExitLinks` | account_views.go:370-588 | 26032b49b93b | 28 | 12 |
| `journal--Journal.ExitEvents` | internal/journal/exit_state.go:870-899 | ea9f939d35d3 | 4 | 4 |
| `journal--appendExitEventTx` | exit_state.go:792-851 | ea9f939d35d3 | 7 | 8 |
| `journal--appendExitEventFromHook` | exit_state.go:977-998 | ea9f939d35d3 | 3 | 4 |
| `engine--ExitObserver.submit` | internal/app/engine/exitloop.go:1513-1638 | 7626cf237058 | 18 | 10 |
| `console--attachOrderExitEvidence` | internal/console/orders.go:737-761 | 4d80478264e7 | 3 | 2 |

## 주장

- **C1 — 판정 증거 열이 하나도 없으면 행은 오류 없이 `legacy_event` 로 읽힘.** `hydrateExitEventEvidence` B1(:393 `!evidence.any()`)
  → :397 `return nil`(Saved=`no_saved_evaluation`, Recomputed · Effective=`legacy_event`). `any()` 는 20 열(세 JSON 포함) 중 하나라도
  Valid 면 참임(exit_snapshot.go:376-383, 손 읽기 — 단일 return 식).
- **C2 — 필수 증거 열 가운데 하나라도 빠진 채 다른 증거 열이 있으면 `ErrExitSnapshotCorrupt`.** 같은 함수 B2(:399 `!evidence.full()`)
  → :400 오류. `full()` 은 20 열 중 필수 부분집합(generation · policy 셋 · snapshot/decision/observation id · observation source/at ·
  projected · state_only · `RecomputedJSON` · `EffectiveJSON` · effective_source)의 Valid 를 요구하고 나머지(next target/protection ·
  ratio · suppressed · saved JSON · arm suppression)는 NULL 이어도 됨(:385-390). 따라서 **JSON 셋만 NULL 로 바꾸고 평탄화 열을 남기면 그 행은
  손상으로 읽힘**(S1 실측: `Journal.ExitEvents` 가 닫힌 포지션 124/129 에서 오류 — `a206_reader_check.out.txt`).
- **C3 — 읽기 전용 계좌 이력은 한 행의 손상을 목록 오류로 키우지 않음.** `ReadOnly.AccountExitEvents` B5(:348) 가
  `hydrateExitEventEvidence` 오류를 받으면 그 행의 Effective 를 비우고 `invalid_event_evidence` 로 표시한 뒤 계속함(return 없음).
  오류를 돌려주는 return 은 :326(쿼리) · :344(Scan) · :355(rows.Err) 뿐.
- **C4 — orders 화면 연결도 같음.** `ReadOnly.BrokerOrderExitLinks` B23(:555) 가 hydrate 오류를 `invalid_event_evidence` 로
  바꾸고 계속함. 연결 대상 행은 `exit_events.proposed_intent_id = intents.id` 이고 그 intent 는 확정 PLACE attempt 의
  `broker_order_id` 에서 옴(SQL 문자열 :443-468, 손 읽기). 같은 intent 를 가리키는 exit_events 가 2 행 이상이면
  B15 switch 의 B21 갈래(:545 `eventCount > 1`) 가 `ambiguous_exit_evidence` 로 표시함.
- **C5 — 화면이 스냅숏을 쓰는 곳은 연결된 1 행뿐.** `attachOrderExitEvidence` B2(:744 Ambiguous) 는 스냅숏 없이 return(:748),
  B3(:755 `effective.Snapshot != nil`) 일 때만 스냅숏 line 을 씀. 개요 · 포트폴리오(`overview.go:580-595`, `portfolio.go:738-759`)
  는 `Event.Evaluation` 을 읽지 않음(손 읽기 — 필드 열거: CreatedAt · ObservedPrice · HighWater · BaselineAfter · LevelAfter ·
  Action · ProposedIntentID).
- **C6 — 판정 행은 세 JSON 을 매번 따로 직렬화해 씀.** `appendExitEventTx` B2(:801, Effective · Recomputed 둘 다 있을 때) 안에서
  `encode` 를 Saved · Recomputed · Effective 에 한 번씩 부름(:809 · :812 · :815 의 오류 갈래 B4~B6). `encode` 는 snapshot 이 없으면 NULL 을
  돌려주므로(B3 :803-804) Saved 가 없는 첫 평가는 둘만 씀. 있는 것은 같은 값이어도 따로 씀(운영 사본 1차 ad-hoc 집계: saved NULL 127 / 28,326 — harness 출력 밖).
  훅 행(`appendExitEventFromHook`)은 9 열만 INSERT 하고 증거 열을 쓰지 않음(:978-983).
- **C7 — 거절된 청산 발의는 해제되어 다음 관측에서 다시 무장됨(속도 제한 없음).** `ExitObserver.submit` 의 해제 호출 5 곳:
  :1530(B2 하한 0주) · :1549(B3 Guardian 거절) · :1566(B5 intent 무효) · :1607(B12 `ReasonSymbolInFlight` → `ProposalCancelled`) ·
  :1634(B15 default — 브로커 확정 거절 등 → `ProposalRefused`). 각 해제 뒤 return 하며 대기 · 재시도 횟수 상한 분기는 없음.
  해제의 의도는 B2 안 주석(:1526-1529, 「Release the proposal so the same level identity is proposed again」)이 적음 —
  손절을 늦추지 않는다는 §0.3 과 같은 방향이라 이 change 는 재무장 빈도를 건드리지 않음.
  다음 주기 재무장은 시험 `TestARefusalReArmsTheLevel` 이 고정함(손 읽기 — 이름만 인용).

## 손 읽기로 남긴 것(AST 아님, 표시)

- SQL 문자열 안의 조인 · 조건(C4 의 :443-468) — logic-map 은 Go 분기를 열거하고 SQL 은 열거하지 않음.
- `any()` · `full()` 의 단일 return 식(C1 · C2).
