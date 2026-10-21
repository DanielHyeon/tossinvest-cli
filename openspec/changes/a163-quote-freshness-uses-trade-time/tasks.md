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
- [ ] 0.5 **proposal-freeze 리뷰** — 적대적 Eng 보이스 포함(High-risk), 결과 `review.md`.
      필수 질문: 「이 diff 의 어느 줄이 판정 자격을 바꿀 수 있나」
- [x] 0.6 `adaptPrices`·`Client.Prices`·`ExitObserver.observe`·`ObserveOnce`·`quoteUsable`·`stableObservationID` 의
      callers/callees/impact 를 HEAD 와 대조 → `analysis/code-context/{codegraph-baseline,codegraphcontext-context,evidence-reconciliation}.md`
      — 2026-10-21 codegraph 1.6.0 기존 색인(2,929 files), 불일치 0(이름 충돌 거짓 엣지만), CGC 는 이 체크아웃을 못 봄(advisory 미반영).
      `make sdd-sync` 는 이 문서 단계에서 돌리지 않았다 → 0.7
- [ ] 0.7 편집 착수 직전 `make sdd-sync` 후 0.6 질의 재실행·`evidence-reconciliation.md` 갱신(fingerprint 신선)

## 1. 함수 내부 증거 (편집 전, High-risk 면제 없음)

- [ ] 1.1 FLM/BTM 번들 — 편집 대상 기존 함수: `official.adaptPrices` · `engine.ExitObserver.observe` · `engine.ExitObserver.ObserveOnce`
      (`scaffold_analysis.py` → `ast.json` · `function-logic-map.md` · `branch-test-map.md` · `risk-pattern-report.md`)
- [ ] 1.2 **diff 0 이어야 하는** 판정 함수의 `revision: base` 고정 번들: `quoteUsable` · `judge` · `judgeRatchet` · `judgeLadder` ·
      `record` · `refreshObservation` · `stableObservationID` (편집 뒤 source_sha256 이 base 와 같아야 함)
- [ ] 1.3 Pre-Edit 선언(대상 심볼 · 호출부 · 기존 시험 · 불변식 · 실패 시험 · rollback)을 `review.md` 에 기록

## 2. RED — 현재 동작을 먼저 고정 (편집 전 GREEN 이어야 하는 고정 시험)

- [ ] 2.1 고정: 받은 시각 15 s 안 · 브로커 시각 9h30m 묵음 · 손절선 아래 → **같은 주기에** 보호 청산이 제출된다(`cycle.Judged`·제출 수)
- [ ] 2.2 고정: 브로커 시각 `null`·키 없음·해독 불가 → 판정·제출이 2.1 과 같다
- [ ] 2.3 고정: 브로커 시각이 미래(시계 차) → 판정·제출이 2.1 과 같다
- [ ] 2.4 고정: 관측 identity(`stableObservationID`)와 원장 `ObservedAt` 이 브로커 시각과 무관하게 `FetchedAt` 으로 만들어진다
- [ ] 2.5 RED: `adaptPrices` 가 `timestamp` 를 `SourceAt` 로 보존한다(실측 형식 `+09:00` 두 시장 · 밀리초 · null · 빈 문자열 · 깨진 문자열)
- [ ] 2.6 RED: 판정 순회 뒤 보고 — 브로커 시각 부재(`source_missing`) 포지션 1건당 일반 등급 알림 1건, 래치로 다음 주기 0건,
      시각이 생긴 뒤 재무장. 체결 나이(`source_old`·`source_ahead`)는 로그 필드만이고 알림 0건(H1)
- [ ] 2.7 RED: 순서 — 보고·나이 로그는 그 주기의 마지막 포지션 판정·제출 **뒤**에 일어난다(가짜 알림기가 보고 시점의 제출 수를 단언)
- [ ] 2.9 RED: 진입 래치·운영 모드 강화·outbox 행이 보고로 생기지 않는다(H2)
- [ ] 2.8 RED: 알림·로그 필드에 계좌 ref 가 없다

## 3. GREEN — 최소 구현

- [ ] 3.1 `domain.Quote.SourceAt`(H7 결과의 JSON 태그) + `adaptPrices` 해독(오류 없음, 실패 = 0)
- [ ] 3.2 `observedQuote` 에 `SourceAt` 운반(`observe` 한 자리, 분기 수 불변)
- [ ] 3.3 새 파일 `internal/app/engine/exit_source_age.go`: 주기 집계 + 순회 뒤 나이 로그 + `source_missing` 일반 등급 보고 + 래치
- [ ] 3.4 1.1·1.2 번들 최신화 — 1.2 의 일곱 함수 source 해시 불변 확인

## 4. VERIFY

- [ ] 4.1 변이 원장(`analysis/mutation/`): 최소 — ① 보고 사유가 있으면 `continue`(판정 건너뜀) ② 보고를 순회 안으로 이동
      ③ `SourceAt` 0 을 `FetchedAt` 으로 메움 ④ `FetchedAt` 에 `SourceAt` 대입 ⑤ 래치 제거 ⑥ 보고 등급을 critical 로 — 전부 CAUGHT, 무변이 대조군 GREEN 먼저
- [ ] 4.2 관련 시험: `go test ./internal/official/ ./internal/app/engine/ ./internal/domain/` + `make test-race` 의 엔진 동시성 목록 +
      `make test-seams` (선택 이유·미실행 범위 기록; `codegraph affected` 는 1,197/1,216 건이라 선택 근거로 쓰지 않음)
- [ ] 4.3 기존 exit 시험 무수정 통과(`exitloop_test.go` · `exit_e2e_test.go` · `exit_observation_identity_test.go` 등)
- [ ] 4.4 독립 적대 리뷰(High-risk) → `review.md`
- [ ] 4.5 `make sdd-sync` · `make sdd-check` · (사용자 명시 요청 시) `make gate CHANGE=a163-quote-freshness-uses-trade-time`

## 5. 잔여 기록 (이 change 에서 고치지 않음)

- [ ] 5.1 C7(진입 게이트 수신 시각) · C8·C9(편입 t0 나이 검사 0) · C10(비상 청산 지정가 나이 검사 0) · C12(CLI 바닥글 시각) ·
      a112 (k) 를 후속 Story 후보로 `issues.md` 에 기록(H3 = 잔여 기록) + H1 후속(`source_old` 임계·정규장 한정, H6 분포 뒤)

## 6. 확정

- [ ] 6.1 archive · Story 경로 갱신 · PM `--check`
- [ ] 6.2 검증된 학습만 memory retain(시세 값·계좌 정보 금지)
