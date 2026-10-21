# a163 tasks

> **High-risk**(청산 관측 루프). 판정 자격 변경 0 이 이 change 의 계약이다 — 그 경계를 FLM·고정 시험·변이로 먼저 못 박고
> 나서 편집한다. 사용자 방향 결정 A(2026-10-21, design D0). 사람 결정 H1~H7 은 2026-10-21 「권장대로 착수 승인」으로 확정됐고
> (design 「사람 결정 기록」), 같은 결정이 동시 change 상한 예외를 승인했다(proposal 「착수 승인과 동시 change 상한 예외」).
> production 편집은 0.5 freeze 리뷰 · 0.7 · §1 · §2 뒤에만 한다.

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
- [ ] 0.7 편집 착수 직전 `make sdd-sync` 후 0.6 질의 재실행·`evidence-reconciliation.md` 갱신(fingerprint 신선)

## 1. 함수 내부 증거 (편집 전, High-risk 면제 없음)

- [ ] 1.1 FLM/BTM 번들 — 편집 대상 기존 함수: `official.adaptPrices` · `engine.ExitObserver.observe` · `engine.ExitObserver.ObserveOnce` ·
      `cmd/tossctl.engineRuntime`(`SourceAgeLog` 배선 — freeze 리뷰 P1-3)
      (`scaffold_analysis.py` → `ast.json` · `function-logic-map.md` · `branch-test-map.md` · `risk-pattern-report.md`).
      편집 뒤 분기·반환 수 불변 단언: `observe` 15·6 · `ObserveOnce` 8·5(defer 2) · `adaptPrices` 1·1 (design D7)
- [ ] 1.2 **바뀌면 안 되는** 판정 함수 열 개의 고정 — `source_sha256` 은 파일 해시라 쓰지 않음(freeze 리뷰 P1-2, design D7):
      (a) 커밋된 Go 시험이 수신자 `ExitObserver` 의 `quoteUsable` · `judge` · `judgeRatchet` · `judgeLadder` · `record` · `refreshObservation` ·
      `stableObservationID` · `snapshotContext` · `requiresFullJudgement` · `noteHeldProposal` 을 `go/parser`(주석 미적재)+`go/printer` 로
      렌더한 sha256 을 시험 안 상수와 비교 — 기대값 `analysis/render-hash/render-hash-base.txt`(동명 `judge`·`record` 는 다른 수신자이므로
      수신자로 고를 것). (b) `SourceTimestamp` 금지 census 시험(`internal/app/engine`·`internal/official` 비시험 파일, 허용 자리·자리별 수는
      design D7 2). 두 시험 모두 **편집 전 GREEN**(census 는 편집 전 출현 0 = 허용 밖 0)
- [ ] 1.3 Pre-Edit 선언(대상 심볼 · 호출부 · 기존 시험 · 불변식 · 실패 시험 · rollback)을 `review.md` 에 기록

## 2. RED — 현재 동작을 먼저 고정 (편집 전 GREEN 이어야 하는 고정 시험)

- [ ] 2.0 고정(freeze 리뷰 P1-1, 안전 3): `tossctl quote --json` 경로의 JSON 바이트 불변 — `output.WriteQuote(FormatJSON, …)` 에
      (a) WTS 갈래 생산자 `internal/client/quote.go:115` 의 Quote(httptest WTS 응답 또는 그 생산 함수 경유) · (b) `adaptStocks` 의 Quote 를 넣어
      편집 전에 굳힌 골든 바이트와 같다(`FetchedAt` 은 고정 시각으로 덮어 비교), `source_timestamp` 키가 **없다**. 편집 전 GREEN
- [ ] 2.1 고정(e2e — freeze 리뷰 P2-1): `exit_e2e_test.go` 의 httptest 브로커(`newE2EBroker`·`servePrices`)로 **공식 어댑터를 거쳐**
      `ExitObserver.opts.Prices` 를 구동. `/api/v1/prices` 본문 `"timestamp":"2026-08-27T19:59:59.000+09:00"`(9h30m 묵음) + 받은 시각
      15 s 안 + 손절선 아래 → **같은 주기에** 보호 청산이 제출된다(`cycle.Judged`·제출 수). 편집 전에는 필드가 무시되므로 GREEN.
      `servePrices` 가 `map[string]string` 이라 `null`·키 없음은 원문 본문 변형이 필요함. `e2eClockDomainPrices` 는 `FetchedAt` 만 덮으므로
      편집 뒤 `SourceTimestamp` 는 보존됨(확인할 것)
- [ ] 2.2 고정(같은 e2e): 브로커 시각 `null` · 키 없음 · `""` · `"garbage"` · `+0900` 형식 → 판정·제출이 2.1 과 같다(행 탈락 없음 · 오류 없음)
- [ ] 2.3 고정(같은 e2e): 브로커 시각이 미래(시계 차) → 판정·제출이 2.1 과 같다
- [ ] 2.4 고정: 관측 identity(`stableObservationID`)와 원장 `ObservedAt` 이 브로커 시각과 무관하게 `FetchedAt` 으로 만들어진다
- [ ] 2.5 RED: `adaptPrices` 가 `timestamp` 원문을 `SourceTimestamp` 에 **그대로** 보존한다(실측 형식 `+09:00` 두 시장 · 밀리초 · null → `""` ·
      키 없음 → `""` · 깨진 문자열 → 원문 그대로). 해독하지 않음, 오류·행 탈락 0
- [ ] 2.6 RED: 판정 순회 뒤 보고 — 구간 `missing`·`unparseable` 포지션 1건당 `source_missing` 일반 등급 알림 1건, 래치로 다음 주기 0건,
      시각이 생긴 뒤 재무장. 나머지 구간은 로그만이고 알림 0건(H1). 정리(freeze 리뷰 P2-4): 이번 주기 `held` 에 있으나 판정 안 된 포지션은
      래치 불변, `held` 밖 id 는 지워짐 — 「포지션 청산 뒤 같은 id 재등장 없음 → 맵 크기 0」 단언. 상한(P2-5): 한 주기 `missing` 12건 →
      알림 8건, 다음 주기 4건
- [ ] 2.7 RED: 순서 — 보고·나이 로그는 그 주기의 마지막 포지션 판정·제출 **뒤**에, `settleUnobserved` 앞에 일어난다(가짜 알림기가 보고
      시점의 제출 수를 단언)
- [ ] 2.8 RED: 알림·로그 필드에 계좌 ref 가 없다(필드 목록 = design D4 「필드」)
- [ ] 2.9 RED: 진입 래치·운영 모드 강화·outbox 행이 보고로 생기지 않는다(H2)
- [ ] 2.10 RED(freeze 리뷰 P1-4): 보고기 내부 패닉 주입(시험 seam 또는 nil 맵 강제) → 같은 `ObserveOnce` 가 정상 반환, `settleUnobserved` 가
      돌고, 다음 주기의 제출 수가 패닉 없을 때와 같다. 이 시험 이름을 `Makefile` `RACE_ENGINE_TESTS`, 파일을 `RACE_ENGINE_FILES` 에 넣음
      (`tools/sdd/test_race_detector_actually_runs.py` 완전성 가드 통과)
- [ ] 2.11 RED(freeze 리뷰 P1-3): 생산 배선 — (a) `cmd/tossctl` 구조 시험: `engineRuntime` 의 `engine.ExitObserverOptions` 리터럴에
      `SourceAgeLog` 키가 있고 값이 엔진 로거(선례 `cmd/tossctl/a129_engine_runtime_exit_order_log_test.go`) · (b) 행동: `Context.ExitObserver`
      조립을 거친 관측자가 그 로거에 구간 전이 줄을 쓴다(시험 로거를 `opts.Log` 가 아니라 `SourceAgeLog` 로만 주입)
- [ ] 2.12 RED(freeze 리뷰 P2-3·P3-3·P3-4): 로그 빈도와 분류 — 같은 구간이 연속 3 주기 → 줄 1(첫 주기만), 구간 변화 → 줄 1 더.
      `missing` 과 `unparseable` 이 다른 `bucket` 값으로 남음. 나이는 `observedQuote.FetchedAt` 기준 — `domain.Quote.FetchedAt` 이 0 인
      대체 경로(`exitloop.go:867-870`)에서 `ahead` 로 잘못 분류되지 않음

## 3. GREEN — 최소 구현

- [ ] 3.1 `domain.Quote.SourceTimestamp string` · json `source_timestamp,omitempty`(design D2) + `adaptPrices` 대입 한 줄(해독 없음, 분기 0 추가)
- [ ] 3.2 `observedQuote` 에 `SourceTimestamp` 운반(`observe` 리터럴 한 자리, 분기 수 불변)
- [ ] 3.3 새 파일 `internal/app/engine/exit_source_age.go`: recover 경계 진입 함수 + a090 집계 읽기(지우지 않음) + 해독·구간 분류 +
      전이 줄(`SourceAgeLog`) + `source_missing` 일반 등급 알림(주기당 8건 상한) + 래치(`held` 기준 정리). `ObserveOnce` 에는
      `settleUnobserved` 앞 호출 한 줄만(design D4)
- [ ] 3.4 `ExitObserverOptions.SourceAgeLog *obs.Logger` 필드 + `cmd/tossctl/engine.go` `engineRuntime` 배선 `SourceAgeLog: logger`
- [ ] 3.5 1.1 번들 최신화(revision 대조) · 1.2 의 렌더 해시·census 시험 GREEN 확인

## 4. VERIFY

- [ ] 4.1 변이 원장(`analysis/mutation/`): 무변이 대조군 GREEN 먼저, 그다음 전부 CAUGHT —
      ① 보고 사유가 있으면 `continue`(판정 건너뜀) ② 보고를 순회 안으로 이동
      ③ 빈 `SourceTimestamp` 를 받은 시각 문자열로 메움(어댑터) ④a 어댑터에서 `FetchedAt` 을 브로커 시각으로 ④b `observe` 리터럴의
      `FetchedAt` 에 브로커 시각 대입 ⑤ 래치 제거 ⑥ 보고 등급을 critical 로(`criticalEvents` 등재) ⑦ 어댑터가 빈·깨진 `timestamp` 행을
      `continue`(행 탈락 → `unobservedNoQuote`, 2.2 e2e 가 잡음) ⑧ 어댑터가 빈·깨진 `timestamp` 에 오류 반환(→ `observe` B1, 2.2 가 잡음)
      ⑨ `observe` B7 필터의 `effectiveAt` 을 브로커 시각 기준으로 ⑩ `snapshotContext` 가 `stableObservationID` 에 브로커 시각을 넘김
      (1.2 렌더 해시·census) ⑪ 보고기 recover 경계 제거(2.10) ⑫ `engine.go` 의 `SourceAgeLog` 배선 삭제(2.11) ⑬ JSON 태그에서
      `omitempty` 삭제(2.0) ⑭ 전이 대신 매 주기 줄(2.12) ⑮ `held` 기준 정리 삭제(2.6 맵 크기) ⑯ 주기 상한 삭제(2.6)
- [ ] 4.2 관련 시험: `go test ./internal/official/ ./internal/app/engine/ ./internal/domain/ ./internal/output/ ./cmd/tossctl/` +
      `make test-race`(2.10 을 `RACE_ENGINE_TESTS` 에 넣은 뒤) + `make test-seams` (선택 이유·미실행 범위 기록; `codegraph affected` 는
      1,197/1,216 건이라 선택 근거로 쓰지 않음 — 쓸 때는 `--filter '*_test.go'`)
- [ ] 4.3 기존 exit 시험 무수정 통과(`exitloop_test.go` · `exit_e2e_test.go` · `exit_observation_identity_test.go` 등)
- [ ] 4.4 독립 적대 리뷰(High-risk) → `review.md`
- [ ] 4.5 `make sdd-sync` · `make sdd-check` · (사용자 명시 요청 시) `make gate CHANGE=a163-quote-freshness-uses-trade-time`

## 5. 잔여 기록 (이 change 에서 고치지 않음)

- [ ] 5.1 C7(진입 게이트 수신 시각) · C8·C9(편입 t0 나이 검사 0) · C10(비상 청산 지정가 나이 검사 0) · C12(CLI 바닥글 시각) ·
      a112 (k) 를 후속 Story 후보로 `issues.md` 에 기록(H3 = 잔여 기록) + H1 후속(`source_old` 임계·정규장 한정, H6 분포 뒤 — 전이 줄로는
      구간 내부 분포를 얻지 못하는 한계 포함) + freeze 리뷰 P3-7(기존 위험: `apiPrice.Timestamp` 디코드 타입이 바뀌면 `Prices` 전체 오류 →
      판정 0, a163 이 넓히지 않음) + H2 보충(보고 끄기 토글 없음 — 사용자 확인 대기)

## 6. 확정

- [ ] 6.1 archive · Story 경로 갱신 · PM `--check`
- [ ] 6.2 검증된 학습만 memory retain(시세 값·계좌 정보 금지)
