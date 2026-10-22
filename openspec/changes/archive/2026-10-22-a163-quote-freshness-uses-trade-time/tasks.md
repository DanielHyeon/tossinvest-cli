# a163 tasks

> **High-risk**(청산 관측 루프). 판정 자격 변경 0 이 이 change 의 계약이다 — 그 경계를 FLM·고정 시험·변이로 먼저 못 박고
> 나서 편집한다. 사용자 방향 결정 A(2026-10-21, design D0). 사람 결정 H1~H7 은 2026-10-21 「권장대로 착수 승인」으로 확정됐고
> (design 「사람 결정 기록」), 같은 결정이 동시 change 상한 예외를 승인했다(proposal 「착수 승인과 동시 change 상한 예외」).
> production 편집은 0.5 freeze 리뷰 · 0.7 · §1 · §2 뒤에만 한다.
> 사용자 결정 2026-10-22(design 「사람 결정 기록 — 사용자 결정 2026-10-22」): 보고 토글 필요(a163 안 · 기본 OFF, design D8) ·
> H7 형태 `source_timestamp` 승인 · 구간 경계 5분/1시간 · 주기당 알림 8건 승인 → 설계 동결(0.8).

## 0. 게이트 선행

- [x] 0.1 PM: `STORY-TOS-a163` 등록(Feature `FEAT-TOS-009`) · Feature `stories` · `_registry.yaml` 연결 →
      `python3 tools/pm/generate_master_tracker.py --check` 통과 — 2026-10-21, HEAD `f7d80c34` 의 격리 worktree 에서 등록·재생성,
      `--check` rc=0(「hierarchy and generated trackers are current」). 다른 세션의 미커밋 PM 편집(a162)과 섞지 않으려고 공유 트리에는
      registry·tracker 를 index 에만 적용
- [x] 0.2 `python3 tools/sdd/capture_change_base.py --change a163-quote-freshness-uses-trade-time` 로 base commit 고정
      — 2026-10-21 `f7d80c348f612cb7d981cea0a958c81b40d8b7c8`(`feat/a112-four-family-runtime` HEAD)
- [x] 0.3 `openspec validate a163-quote-freshness-uses-trade-time --strict --no-interactive` — 2026-10-21 rc=0 「is valid」
- [x] 0.4 **사람 결정 H1~H7** 기록(design 「사람 결정」 표에 날짜와 선택) — 사용자 결정 2026-10-21 「권장대로 착수 승인」,
      design 「사람 결정 기록」 표. 결과: 이 change 의 알림 사유는 `source_missing` 하나, 체결 나이는 로그 필드만
- [x] 0.5 **proposal-freeze 리뷰** — 적대적 Eng 보이스 포함(High-risk), 결과 `review.md`.
      필수 질문: 「이 diff 의 어느 줄이 판정 자격을 바꿀 수 있나」 — 2026-10-21 독립 적대 리뷰 APPROVE-WITH-CHANGES(P0 0 · P1 4 · P2 5 · P3 7,
      원문 `analysis/freeze-review-2026-10-21.md`). 같은 날 문서 로트로 P1 넷·P2·P3 를 반영하고 처분표를 `review.md` 에 둠(편집 전 조건 충족).
      처분 중 사용자 확인 대기 잔여 1(H2 보충 — 보고 끄기 토글 없음)
- [x] 0.6 `adaptPrices`·`Client.Prices`·`ExitObserver.observe`·`ObserveOnce`·`quoteUsable`·`stableObservationID` 의
      callers/callees/impact 를 HEAD 와 대조 → `analysis/code-context/{codegraph-baseline,codegraphcontext-context,evidence-reconciliation}.md`
      — 2026-10-21 codegraph 1.6.0 기존 색인(2,929 files), 불일치 0(이름 충돌 거짓 엣지만), CGC 는 이 체크아웃을 못 봄(advisory 미반영).
      `make sdd-sync` 는 이 문서 단계에서 돌리지 않았다 → 0.7
- [x] 0.7 편집 착수 직전 `make sdd-sync` 후 0.6 질의 재실행·`evidence-reconciliation.md` 갱신(fingerprint 신선) — D8 로 편집 대상이 된
      `config.mergeEngine` · `engine.recordGateSettings` · `engine.Context.ExitObserver` 의 callers 도 함께
      — 2026-10-22 격리 worktree(HEAD `1eab9b931`): `make sdd-sync` rc=0(CodeGraph 1.6.0 init · GBrain advisory `Sync BLOCKED`), 재질의 원문
      `analysis/code-context/codegraph-callers-0.7-20261022.txt`, `evidence-reconciliation.md` 「tasks 0.7 재질의」(불일치 2 → issues.md I1·I2)
- [x] 0.8 설계 동결 — 사용자 결정 2026-10-22 반영: H2 보충 「토글 필요」(Manager 범위 결정 — a163 안, 기본 OFF, 보고·로그·알림 경로만 막음,
      어댑터 `SourceTimestamp` 보존은 토글 밖) · H7 형태 변경 승인 · 구간 경계 5분/1시간 · 주기당 알림 8건 승인. design D8 · 결정표 ·
      spec exit-policy 「체결 나이 보고는 기본 꺼진 운영 토글 뒤에 있다」 · proposal · review 처분 반영. D8 의 함수 내부 주장은 이 로트의 AST 셋
      (`analysis/ast/config--mergeengine` · `engine--recordgatesettings` · `engine--context.exitobserver`, `enumerate-output.md` AST 32 · STALE 0)

## 1. 함수 내부 증거 (편집 전, High-risk 면제 없음)

- [x] 1.1 FLM/BTM 번들 — 편집 대상 기존 함수: `official.adaptPrices` · `engine.ExitObserver.observe` · `engine.ExitObserver.ObserveOnce` ·
      `cmd/tossctl.engineRuntime`(`SourceAgeLog` 배선 — freeze 리뷰 P1-3) · D8 토글: `config.mergeEngine` · `engine.recordGateSettings` ·
      `engine.Context.ExitObserver`
      (`scaffold_analysis.py` → `ast.json` · `function-logic-map.md` · `branch-test-map.md` · `risk-pattern-report.md`).
      편집 뒤 분기·반환 수 불변 단언: `observe` 15·6 · `ObserveOnce` 8·5(defer 2) · `adaptPrices` 1·1 (design D7) ·
      `recordGateSettings` 6·2 · `Context.ExitObserver` 6·4. 바뀌는 것은 `mergeEngine` 5→6 · 반환 2 (B4 조기 반환 **앞**에 `if raw.X != nil` 하나, design D8)
      — 2026-10-22: 번들 12(목록 7 + `engine.NewContext` · a074 시험 함수 넷 — issues.md I3), `check_analysis.py` 「evidence complete」.
      편집 뒤 분기·반환 수 단언과 revision 최신화는 3.5
- [x] 1.2 **바뀌면 안 되는** 판정 함수 열 개의 고정 — `source_sha256` 은 파일 해시라 쓰지 않음(freeze 리뷰 P1-2, design D7):
      (a) 커밋된 Go 시험이 수신자 `ExitObserver` 의 `quoteUsable` · `judge` · `judgeRatchet` · `judgeLadder` · `record` · `refreshObservation` ·
      `stableObservationID` · `snapshotContext` · `requiresFullJudgement` · `noteHeldProposal` 을 `go/parser`(주석 미적재)+`go/printer` 로
      렌더한 sha256 을 시험 안 상수와 비교 — 기대값 `analysis/render-hash/render-hash-base.txt`(동명 `judge`·`record` 는 다른 수신자이므로
      수신자로 고를 것). (b) `SourceTimestamp` 금지 census 시험(`internal/app/engine`·`internal/official` 비시험 파일, 허용 자리·자리별 수는
      design D7 2). census 는 보고 토글 식별자 `SourceAgeReport` 도 센다(`internal/app/engine` 비시험 파일 — 허용 = `ExitObserverOptions` 필드
      선언 1 · `exitwiring.go` 덮어쓰기 1 · `exit_source_age.go` 임의 수, design D8). 두 시험 모두 **편집 전 GREEN**(census 는 편집 전 출현 0 = 허용 밖 0)
      — 2026-10-22: `internal/app/engine/a163_judgement_pin_test.go` 편집 전 GREEN(렌더 해시 열 개 = `render-hash-base.txt` · census 허용 밖 0).
      편집 전 판은 허용 자리 0 도 받고 GREEN 로트에서 정확한 수로 조임. 설계와 다른 사실 둘은 issues.md I1(자유 함수)·I2(엄격 리더 넷)
- [x] 1.3 Pre-Edit 선언(대상 심볼 · 호출부 · 기존 시험 · 불변식 · 실패 시험 · rollback)을 `review.md` 에 기록
      — 2026-10-22 `review.md` 「1.3 Pre-Edit 선언」

## 2. RED — 현재 동작을 먼저 고정 (편집 전 GREEN 이어야 하는 고정 시험)

- [x] 2.0 고정(freeze 리뷰 P1-1, 안전 3): `tossctl quote --json` 경로의 JSON 바이트 불변 — `output.WriteQuote(FormatJSON, …)` 에
      (a) WTS 갈래 생산자 `internal/client/quote.go:115` 의 Quote(httptest WTS 응답 또는 그 생산 함수 경유) · (b) `adaptStocks` 의 Quote 를 넣어
      편집 전에 굳힌 골든 바이트와 같다(`FetchedAt` 은 고정 시각으로 덮어 비교), `source_timestamp` 키가 **없다**. 편집 전 GREEN
      — `internal/output/a163_quote_json_golden_test.go`(WTS) · `internal/official/a163_stocks_quote_json_golden_test.go`(adaptStocks), 편집 전 GREEN
- [x] 2.1 고정(e2e — freeze 리뷰 P2-1): `exit_e2e_test.go` 의 httptest 브로커(`newE2EBroker`·`servePrices`)로 **공식 어댑터를 거쳐**
      `ExitObserver.opts.Prices` 를 구동. `/api/v1/prices` 본문 `"timestamp":"2026-08-27T19:59:59.000+09:00"`(9h30m 묵음) + 받은 시각
      15 s 안 + 손절선 아래 → **같은 주기에** 보호 청산이 제출된다(`cycle.Judged`·제출 수). 편집 전에는 필드가 무시되므로 GREEN.
      `servePrices` 가 `map[string]string` 이라 `null`·키 없음은 원문 본문 변형이 필요함. `e2eClockDomainPrices` 는 `FetchedAt` 만 덮으므로
      편집 뒤 `SourceTimestamp` 는 보존됨(확인할 것)
      — `internal/app/engine/a163_source_time_e2e_test.go` `TestA163AStaleBrokerTimeStillLiquidatesInTheSameCycle` 편집 전 GREEN(원문 본문 브로커 `newA163Broker`)
- [x] 2.2 고정(같은 e2e): 브로커 시각 `null` · 키 없음 · `""` · `"garbage"` · `+0900` 형식 → 판정·제출이 2.1 과 같다(행 탈락 없음 · 오류 없음)
      — `TestA163AMissingOrMalformedBrokerTimeJudgesTheSame`(하위 5) 편집 전 GREEN
- [x] 2.3 고정(같은 e2e): 브로커 시각이 미래(시계 차) → 판정·제출이 2.1 과 같다
      — `TestA163AFutureBrokerTimeJudgesTheSame` 편집 전 GREEN
- [x] 2.4 고정: 관측 identity(`stableObservationID`)와 원장 `ObservedAt` 이 브로커 시각과 무관하게 `FetchedAt` 으로 만들어진다
      — `TestA163IdentityAndObservedAtComeFromTheReceiptNotTheBrokerTime` 편집 전 GREEN(e2e, 같은 브로커 시각 두 주기)
- [x] 2.5 RED: `adaptPrices` 가 `timestamp` 원문을 `SourceTimestamp` 에 **그대로** 보존한다(실측 형식 `+09:00` 두 시장 · 밀리초 · null → `""` ·
      키 없음 → `""` · 깨진 문자열 → 원문 그대로). 해독하지 않음, 오류·행 탈락 0
      — `internal/official/a163_adapt_prices_source_time_test.go`(받은 시각 창 단언 포함 — 변이 ④a), RED→GREEN `analysis/red-green/`
- [x] 2.6 RED: 판정 순회 뒤 보고 — 구간 `missing`·`unparseable` 포지션 1건당 `source_missing` 일반 등급 알림 1건, 래치로 다음 주기 0건,
      시각이 생긴 뒤 재무장. 나머지 구간은 로그만이고 알림 0건(H1). 정리(freeze 리뷰 P2-4): 이번 주기 `held` 에 있으나 판정 안 된 포지션은
      래치 불변, `held` 밖 id 는 지워짐 — 「포지션 청산 뒤 같은 id 재등장 없음 → 맵 크기 0」 단언. 상한(P2-5): 한 주기 `missing` 12건 →
      알림 8건, 다음 주기 4건
      — `a163_source_age_report_test.go` 넷. 「맵 크기 0」 은 D4 호출 자리상 도달 불가(보고는 판정이 있는 주기에만 돌고 판정된 포지션은 맵에 남음) → 닫힌 포지션 id 가 빠지고 남은 하나만 있음으로 단언(변이 ⑮ 가 잡음). RED→GREEN `analysis/red-green/`
- [x] 2.7 RED: 순서 — 보고·나이 로그는 그 주기의 마지막 포지션 판정·제출 **뒤**에, `settleUnobserved` 앞에 일어난다(가짜 알림기가 보고
      시점의 제출 수를 단언)
      — `TestA163TheReportComesAfterEveryJudgementAndBeforeTheUnobservedSettlement`(제출 수 + a090 줄과 한 버퍼 순서)
- [x] 2.8 RED: 알림·로그 필드에 계좌 ref 가 없다(필드 목록 = design D4 「필드」)
      — `TestA163NoReportLineOrAlertCarriesTheAccount`(줄 · 알림 키 ⊆ D4 목록 + 로거 자기 열, 계좌 ref 없음, 원문 ≤64 바이트 · unparseable 만)
- [x] 2.9 RED: 진입 래치·운영 모드 강화·outbox 행이 보고로 생기지 않는다(H2)
      — `TestA163TheReportNeverBlocksEntryOrEscalates`(실제 Notifier RecordOnly · 원장 outbox 행 0 · 모드 NORMAL · 진입 래치 없음). 골격에서도 통과 — 이빨은 변이 ⑥
- [x] 2.10 RED(freeze 리뷰 P1-4): 보고기 내부 패닉 주입(시험 seam 또는 nil 맵 강제) → 같은 `ObserveOnce` 가 정상 반환, `settleUnobserved` 가
      돌고, 다음 주기의 제출 수가 패닉 없을 때와 같다. 이 시험 이름을 `Makefile` `RACE_ENGINE_TESTS`, 파일을 `RACE_ENGINE_FILES` 에 넣음
      (`tools/sdd/test_race_detector_actually_runs.py` 완전성 가드 통과)
      — `a163_source_age_panic_test.go`(로거 패닉 · 알림기 패닉, Run goroutine), `RACE_ENGINE_FILES`·`RACE_ENGINE_TESTS` 등재, 가드 `tools/sdd/test_race_detector_actually_runs.py` OK, `-race` PASS. 골격에서도 통과 — 이빨은 변이 ⑪
- [x] 2.11 RED(freeze 리뷰 P1-3): 생산 배선 — (a) `cmd/tossctl` 구조 시험: `engineRuntime` 의 `engine.ExitObserverOptions` 리터럴에
      `SourceAgeLog` 키가 있고 값이 엔진 로거(선례 `cmd/tossctl/a129_engine_runtime_exit_order_log_test.go`) · (b) 행동: `Context.ExitObserver`
      조립을 거친 관측자가 그 로거에 구간 전이 줄을 쓴다(시험 로거를 `opts.Log` 가 아니라 `SourceAgeLog` 로만 주입)
      — (a) `cmd/tossctl/a163_engine_runtime_source_age_log_test.go` · (b) `TestA163TheAssembledObserverWritesItsLinesToTheDedicatedLogger`(검증 게이트 조립 + 원문 브로커, 묵은 시각이 gt_1h 로 도착)
- [x] 2.12 RED(freeze 리뷰 P2-3·P3-3·P3-4): 로그 빈도와 분류 — 같은 구간이 연속 3 주기 → 줄 1(첫 주기만), 구간 변화 → 줄 1 더.
      `missing` 과 `unparseable` 이 다른 `bucket` 값으로 남음. 나이는 `observedQuote.FetchedAt` 기준 — `domain.Quote.FetchedAt` 이 0 인
      대체 경로(`exitloop.go:867-870`)에서 `ahead` 로 잘못 분류되지 않음
      — `TestA163AnAgeLineIsWrittenOnlyWhenTheBucketChanges` · `TestA163EveryBucketIsClassifiedFromTheReceiptTime`(일곱 구간 + 경계 15 s · 5 분 · 1 시간 포함, 사이클 대체 받은 시각)
> 2.6 · 2.7 · 2.8 · 2.9 · 2.10 · 2.11(b) · 2.12 는 보고 토글 **ON**(`ExitObserverOptions.SourceAgeReport = true`)으로 돈다(design D8 — 보고는 ON 에서만).
- [x] 2.13 RED(사용자 결정 2026-10-22, design D8 — 기본 OFF): `internal/config` — engine 블록 없음 · 키 없음 · 명시 `false` → `ExitSourceAgeReport == false`,
      `true` → `true`. `automation_gate` 키 **없는** 파일의 `true` 도 `true`(merge 가 `mergeEngine` B4 조기 반환 앞). 꺼짐이면 `config.Status` JSON
      (`tossctl config show`)에 `exit_source_age_report` 키가 **없다**(omitempty — 편집 전 바이트와 같음). `schemas/config.schema.json` engine 블록에
      boolean · `default: false` 속성이 있고 `TestPublishedSchemaMatchesSchemaVersion` 통과
      — `internal/config/a163_exit_source_age_report_test.go` 셋(Status.Engine 골든 = 편집 전 바이트)
- [x] 2.14 RED(OFF = 오늘 동작): 같은 주기 입력(`missing` · `unparseable` · 9h30m 묵음 · 손절선 아래 하나)을 토글 OFF 와 ON 관측자에 각각 돌려
      — OFF: `SourceAgeLog` 줄 0 · `source_missing` 알림 0 · 래치 맵 크기 0 · a090 집계·`settleUnobserved` 결과 ON 과 같음. 두 관측자의
      `cycle.Judged` · 제출 수 · 제출 순서 · 원장 행이 같다(토글은 판정과 제출 사이에 서지 않음). ON: 2.6 · 2.12 의 줄·알림이 남음.
      `ExitObserverOptions` 0 값 = OFF 단언
      — `TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn`(주기 결과 · 제출 · 원장 행 문자열 대조, 0 값 옵션 = OFF)
- [x] 2.15 RED(배선 — design D8 「읽는 시점」): `Context.ExitObserver` 가 호출자 값과 무관하게 `opts.SourceAgeReport` 를
      `c.Config.Engine.ExitSourceAgeReport` 로 덮는다 — 설정 `true` · 호출자 `false` → 관측자 ON, 설정 `false` · 호출자 `true` → OFF
      (선례 `NotificationsEnabled` `exitwiring.go:355` 의 시험 모양). 설정은 기동 때 한 번만 읽힘(주기 중 설정 파일 변경이 관측자에 닿지 않음)
      — `TestA163TheProductionAssemblyReadsTheLoadedToggle`(설정 · 메모리 변경이 만들어진 관측자에 안 닿음)
- [x] 2.16 RED(audit — §0.5): `recordGateSettings` 가 `changes` 목록 **끝**(`per_trade_risk_budget` 뒤)에 action `audit.ActionExitSourceAgeReport`
      (`"exit.source_age_report"`) · setting `engine.exit_source_age_report` · 값 `"false"`/`"true"` 행을 쓴다. 꺼짐→켜짐 재기동은 old `false` · new
      `true` 한 행, 같은 값 재기동은 0 행(`RecordChange`). 기존 행 순서 불변(`a074_notification_audit_test.go` 의 앞부분 단언 무수정 통과 — 호출부 4곳은
      끝 인자 `false` 만 기계적으로 추가). 이 행의 기록 실패는 기존 행과 같은 기동 거부

      — `a163_source_age_audit_test.go`(끝 행 · 바뀐 기동만 · 같은 값 0 행) + `a163_source_age_startup_test.go`(NewContext 생산 경로, 세 번 기동 → 행 둘). 이 행만의 기록 실패 주입 시험은 없음 — 같은 루프 B6(FLM)이라 기존 실패 모양 그대로
## 3. GREEN — 최소 구현

- [x] 3.1 `domain.Quote.SourceTimestamp string` · json `source_timestamp,omitempty`(design D2) + `adaptPrices` 대입 한 줄(해독 없음, 분기 0 추가)
      — `internal/domain/models.go` · `internal/official/market_reads.go`
- [x] 3.2 `observedQuote` 에 `SourceTimestamp` 운반(`observe` 리터럴 한 자리, 분기 수 불변)
      — `exitloop.go` `observedQuote` 필드 + `observe` 리터럴(분기 15 · 반환 6 불변)
- [x] 3.3 새 파일 `internal/app/engine/exit_source_age.go`: recover 경계 진입 함수 + a090 집계 읽기(지우지 않음) + 해독·구간 분류 +
      전이 줄(`SourceAgeLog`) + `source_missing` 일반 등급 알림(주기당 8건 상한) + 래치(`held` 기준 정리). `ObserveOnce` 에는
      `settleUnobserved` 앞 호출 한 줄만(design D4)
      — `internal/app/engine/exit_source_age.go`(새 파일) + `ObserveOnce` 호출 한 줄(분기 8 · 반환 5 · defer 2 불변). 이벤트 상수 둘(`obs.EventExitSourceAge` · `obs.EventExitSourceMissing`, 등급표 밖)은 `internal/obs/event.go`
- [x] 3.4 `ExitObserverOptions.SourceAgeLog *obs.Logger` 필드 + `cmd/tossctl/engine.go` `engineRuntime` 배선 `SourceAgeLog: logger`
      — `ExitObserverOptions.SourceAgeLog` + `cmd/tossctl/engine.go` 배선
- [x] 3.5 1.1 번들 최신화(revision 대조) · 1.2 의 렌더 해시·census 시험 GREEN 확인
      — 번들 12 `revision: current` 로 재생성(`analysis/flm-gen/README.md` 편집 전·뒤 대조), `check_analysis.py` required 12 · evidence complete. census 를 정확한 수로 조임(I2 의 기존 넷 포함)
- [x] 3.6 보고 토글(design D8): `config.Engine.ExitSourceAgeReport bool` · json `exit_source_age_report,omitempty` + `rawEngine` `*bool` +
      `mergeEngine` B4 앞 merge 한 줄 묶음 · `schemas/config.schema.json` 속성 · `audit.ActionExitSourceAgeReport` 상수 · `recordGateSettings` 끝 인자 +
      끝 행 · `engine.go:472` 호출부 · `ExitObserverOptions.SourceAgeReport bool` · `Context.ExitObserver` 덮어쓰기 한 줄 · `exit_source_age.go` 진입 함수의
      recover 경계 바로 뒤 「꺼짐이면 반환」. `ObserveOnce` 와 판정 함수에는 토글 코드 0. config 를 **쓰는** 함수(콘솔 스위치·저장)는 더하지 않음

      — config · schema · audit 상수 · `recordGateSettings` 끝 인자/끝 행 · `NewContext` 호출부 · `Context.ExitObserver` 덮어쓰기 · 진입 함수 토글 판단. config 쓰기 코드 0
## 4. VERIFY

- [x] 4.1 변이 원장(`analysis/mutation/`): 무변이 대조군 GREEN 먼저, 그다음 전부 CAUGHT —
      ① 보고 사유가 있으면 `continue`(판정 건너뜀) ② 보고를 순회 안으로 이동
      ③ 빈 `SourceTimestamp` 를 받은 시각 문자열로 메움(어댑터) ④a 어댑터에서 `FetchedAt` 을 브로커 시각으로 ④b `observe` 리터럴의
      `FetchedAt` 에 브로커 시각 대입 ⑤ 래치 제거 ⑥ 보고 등급을 critical 로(`criticalEvents` 등재) ⑦ 어댑터가 빈·깨진 `timestamp` 행을
      `continue`(행 탈락 → `unobservedNoQuote`, 2.2 e2e 가 잡음) ⑧ 어댑터가 빈·깨진 `timestamp` 에 오류 반환(→ `observe` B1, 2.2 가 잡음)
      ⑨ `observe` B7 필터의 `effectiveAt` 을 브로커 시각 기준으로 ⑩ `snapshotContext` 가 `stableObservationID` 에 브로커 시각을 넘김
      (1.2 렌더 해시·census) ⑪ 보고기 recover 경계 제거(2.10) ⑫ `engine.go` 의 `SourceAgeLog` 배선 삭제(2.11) ⑬ JSON 태그에서
      `omitempty` 삭제(2.0) ⑭ 전이 대신 매 주기 줄(2.12) ⑮ `held` 기준 정리 삭제(2.6 맵 크기) ⑯ 주기 상한 삭제(2.6)
      — 보고 토글(사용자 결정 2026-10-22, design D8): ⑰ 토글 무시(진입 함수의 「꺼짐이면 반환」 삭제 → 2.14 OFF 줄·알림 0 이 잡음)
      ⑱ 기본값 ON 으로 뒤집기(`mergeEngine` 이 키 없음을 `true` 로 / 진입 함수 판단을 `!SourceAgeReport` 로 → 2.13·2.14)
      ⑲ merge 를 B4 조기 반환 뒤로 이동(2.13 `automation_gate` 없는 파일) ⑳ `exitwiring.go` 덮어쓰기 삭제(2.15)
      ㉑ audit 행 삭제 · 값 고정(2.16) ㉒ 토글 판단을 `ObserveOnce` 판정 순회 앞으로 옮겨 꺼짐이면 판정 건너뜀(2.14 판정·제출 동일 · census · `ObserveOnce` 분기 수)
      — `analysis/mutation/`(하네스 `harness.py` · `ledger.md` · `ledger.json`): HEAD `ff232b3cf` 사본, 무변이 대조군 26 묶음 rc=0, **32 CAUGHT / 32**
      (목록 ①~㉒ 에 ④a/④b · ⑱a/⑱b · ㉑a~c 분할 + 리뷰 처분 ㉓~㉘). 첫 판에서 ② 생존 → 순서 시험 수리(`aeeea2c02`). ⑧ 은 시한 초과로 잡힘(issues I9)
- [x] 4.2 관련 시험: `go test ./internal/official/ ./internal/app/engine/ ./internal/domain/ ./internal/output/ ./internal/config/ ./internal/audit/ ./cmd/tossctl/` +
      `make test-race`(2.10 을 `RACE_ENGINE_TESTS` 에 넣은 뒤) + `make test-seams` (선택 이유·미실행 범위 기록; `codegraph affected` 는
      1,197/1,216 건이라 선택 근거로 쓰지 않음 — 쓸 때는 `--filter '*_test.go'`)
      — 2026-10-22 HEAD `ff232b3cf`(Go 최종) 기준: 위 7 패키지 + `internal/obs` · `client` · `hybrid` · `flatten` 무태그 전부 ok(engine 296 s). `make test-race` rc=0
      (7 패키지 + 엔진 목록, `TestA163APanicking…` 포함). `make lint` rc=0. `make test-seams`(`-tags tossos_testseams ./...`) rc=2 — 114 ok · 실패 1:
      `internal/performance` `TestMaximumAttributionGenerationIntegrityScanMeetsP95Target` p95 2.019 s > 2 s(병행 부하 중 시간 상한 시험, 단독 재실행 ok,
      이 패키지는 a163 이 고친 패키지에 의존하지 않음 — `go list -deps`). 선택 이유: 편집 패키지 + 그 소비자(flatten · hybrid · client · obs). 미실행: `make test`(전체)
- [x] 4.3 기존 exit 시험 무수정 통과(`exitloop_test.go` · `exit_e2e_test.go` · `exit_observation_identity_test.go` 등)
      — 4.2 의 패키지 실행에서 `exitloop_test.go` · `exit_e2e_test.go` · `exit_observation_identity_test.go` 등 기존 시험 무수정 PASS(무편집 — 편집한 시험은
      `a074_notification_audit_test.go` 의 호출 인자 넷뿐, design D8)
- [x] 4.4 독립 적대 리뷰(High-risk) → `review.md`
      — `review.md` 「4.4 구현 리뷰」: 전문 · 적대(Claude) · Codex 적대 · 저자 critical pass. 수리 10(R1~R10) · 잔여 4(I5~I7) · 수용 3. 손절 경로 결함 0
- [x] 4.5 `make sdd-sync` · `make sdd-check` · (사용자 명시 요청 시) `make gate CHANGE=a163-quote-freshness-uses-trade-time`

      — `make sdd-sync` rc=2: CodeGraph · CGC 갱신 OK, GBrain advisory sync 가 1800 s 시한 초과(incomplete) — `check_index_freshness.py` 「CodeGraph hard-evidence
      index matches the worktree」. `make sdd-infra` 로 worktree-local `.sdd/.venv` 생성 뒤 `make sdd-check` rc=0(gbrain advisory WARN). `make gate` 는 사용자 요청 없음 — 미실행
## 5. 잔여 기록 (이 change 에서 고치지 않음)

- [x] 5.1 C7(진입 게이트 수신 시각) · C8·C9(편입 t0 나이 검사 0) · C10(비상 청산 지정가 나이 검사 0) · C12(CLI 바닥글 시각) ·
      a112 (k) 를 후속 Story 후보로 `issues.md` 에 기록(H3 = 잔여 기록) + H1 후속(`source_old` 임계·정규장 한정, H6 분포 뒤 — 전이 줄로는
      구간 내부 분포를 얻지 못하는 한계 포함) + freeze 리뷰 P3-7(기존 위험: `apiPrice.Timestamp` 디코드 타입이 바뀌면 `Prices` 전체 오류 →
      판정 0, a163 이 넓히지 않음) + 보고 토글의 콘솔 스위치 부재(사람 flip 은 config 편집 + 엔진 재기동, 콘솔 스위치가 필요하면 별도 Story —
      design D8 「표시」). H2 보충은 사용자 결정 2026-10-22 로 닫힘(잔여 아님)

      — `issues.md` I11(+ I5~I10)
## 6. 확정

- [x] 6.1 archive · Story 경로 갱신 · PM `--check`
      — 완료 게이트 PASS 뒤 같은 닫는 로트에서 archive · Story `openspec.path` 갱신 · tracker 재생성 · PM `--check`(관례: 체크 → tracker → sdd-sync → gate → archive; 결과는 `review.md` 「완료 게이트」)
- [x] 6.2 검증된 학습만 memory retain(시세 값·계좌 정보 금지)
      — Hindsight bank `tossos-dev-episodes` 에 교훈 넷(pre-flip 결정은 Story limits 에 · 순서 시험의 판정 순서 의존 · 원문 길이 상한 · gate 2단계와 6.x 체크 순서). 시세 값 · 계좌 정보 0
