# a170 review

생산 함수 편집 0. 바뀐 것은 시험 파일 `internal/campaignstateevidence/authority_guard_test.go` 의 가드 함수뿐이다(비례 원칙: 시험 전용).
게이트는 시험 파일의 기존 함수도 수정 집합으로 세므로 FLM/BTM 번들 둘을 만들었다(§4) — 처음의 not-applicable 표기는 그 실측으로 대체한다.

## 1. 원인 조사 · 판정 (2026-10-23)

- 사용자 결정(Manager 전달, 2026-10-23): a140 완료 gate 를 막는 이 실패를 작은 수리 change 로 고친다. 이 change 에 대해 `make test` 실행 승인.
- 판정: 의도된 검사 전용 · 권한 없는 사용(근거는 proposal 「Why」) → 허용 목록에 그 파일 하나를 이름으로 추가하고, 그 패키지의 비시험 import 를 거절.
  판단이 갈리지 않아 멈추지 않았다.

## 2. 경량 코드 리뷰 (2026-10-23)

- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only, base `2bbee345f` + 미커밋 diff). 판정: P1 0 · P2 0 · P3 2.
- 확인: 원인 서사 정확(a162 design:18 · review:34 · :118, `cd5f27219` 가 `3f8e34632` 의 조상, 비시험 importer 0). 조기 반환이 import 루프 **뒤**라
  이 파일에도 lane 점 import 거절 · `offlinecontinuation` import 거절이 그대로 적용. 생산 diff 0. 리뷰어 직접 실행 `go test ./internal/campaignstateevidence/` ok.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P3 | 조기 반환의 **위치**를 못 박는 시험이 없음 — 루프 앞으로 옮기면 평가 파일이 lane 점 import 로 캡처를 불러도 통과(추론) | **수용** — 시험 단언 추가 + 변이 M5(아래 영수증) |
| 2 | P3 | spec 문구가 가드 범위(`internal`·`cmd`·`tools`, 한정 selector)보다 넓음 | **수용** — 범위로 좁히고 a161 부터의 한계를 한 줄 명시 |

## 3. 검증 (2026-10-23)

- RED: `analysis/red/red.txt`(컴파일 RED → 상수만 더한 행동 RED: 기존 실패 + 새 시험 실패). GREEN: `analysis/red/green.txt`(두 패키지 · 태그 빌드 · gofmt, rc 0).
- 변이: `analysis/red/mutation.txt` — M1 · M3 · M4 · M5 CAUGHT, M2 는 1 판 생존(형제 파일이 evidence import 없이 캡처만 부르는 경우를 시험이 안 잼) →
  시험 추가 뒤 CAUGHT. 무변이 대조군 GREEN(두 번).
- `make test` rc=0(ok 115, FAIL 0) · `make lint` rc=0 — `analysis/red/verify.txt`. P3-1 단언은 그 뒤에 더했고 영향 패키지만 다시 돌렸다(green.txt receipt 3).
- 미실행: 완료 gate(`make gate CHANGE=a170-…`) — 승인 범위 밖(tasks 2.3).

## 4. 완료 gate 준비 (2026-10-23)

- 사용자 승인(Manager 전달, 2026-10-23): 완료 gate → archive, base 창에 형제 착지가 들어오면 base 재고정(귀속 실측 뒤).
- `check_analysis.py --change a170-…`(gate 5 단계와 같은 판정) 첫 실측: required 3 — `authority_guard_test.go:allowedImport` ·
  `…:TestGuardRejectsRuntimeCaptureUse` · `internal/official/stock_reads.go:adaptStocks`.
  - 앞의 둘은 a170 자신의 편집이다. `TestGuardRejectsRuntimeCaptureUse` 는 본문 불변이지만 바로 뒤의 순수 삽입 훅이 base 쪽 함수와 맞닿아 잡힌다.
    → 번들: `analysis/function-logic/internal-campaignstateevidence--allowedimport`(revision current, 분기 11) ·
    `…--testguardrejectsruntimecaptureuse`(revision base, base `2bbee345f` 소스 sha256 `79f219f5…a4f714`, 분기 3).
  - 번들 뒤 재실측: 남은 것은 `missing evidence for modified function internal/official/stock_reads.go:adaptStocks` 하나 — a169 의 `064f67297`
    (2026-10-23 06:21, base 뒤 형제 착지).
- 귀속 실측(`2bbee345f..a25db35f4`): a170 디렉터리를 만진 비병합 커밋 중 `.go` 편집은 `6d12ca6bb` 하나이고, 그것이 고친 기존 함수 둘은 위 번들로 덮였다.
  `adaptStocks` 는 a170 이 만지지 않았다(`git log 2bbee345f..HEAD -- internal/official/stock_reads.go` = `064f67297` 뿐).
- 처분: base 를 이 기록 커밋으로 재고정한다(다음 단독 커밋이 영수증).

## 5. 종결 (2026-10-28, 종결 에이전트 Opus 5.5)

- 사용자 승인 "1"(2026-10-28): 두 change(a172 → a170)에 FLM 번들 작성 · 복사, 커밋, `/tmp` 밖 격리 worktree 워밍업 뒤 전체 `make gate`, archive, Story 경로,
  tracker, push, 정리, retain. **base 재고정 금지**(사용자가 2번 안 기각). 코드 변경 0.
- 상황: §4 의 재고정(`be349895a`, 2bbee345f → 0d168e0a3) 뒤 a170 은 닫히지 않은 채 남았고, 그동안 착지한 a172 · a173 · a174 · a175 · a176(a178)이
  창(base `0d168e0a3` → 워킹트리)에 함수 26 개를 더했다. 귀속 실측: `0d168e0a3..HEAD` 에서 a170 디렉터리를 만진 커밋은 재고정 커밋 `be349895a` 하나이고 `.go` 편집 0.
- gate ⑤ 사전 실행(HEAD `8fdce8aaf`, a172 archive 뒤): required 26(a170 자기 번들 2 는 이미 충족) → rc 1.
- 처분: **복사 26 · 새로 작성 0.** 26 개 모두 가장 나중 아카이브인 `archive/2026-10-27-a172-console-anti-ai-redesign` 번들을 복사하고 출처 주석 한 줄을 닮.
  `revision: current` 22 개는 현재 소스 sha256, `revision: base` 4 개(`TestTheSummaryCostsNoBrokerCall` · `TestTheCurrencyConsequenceIsRenderedAndNotFolded` ·
  `limitCurrent` · `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`)는 `git show 0d168e0a3:<file> | sha256sum` 이 번들 sha 와 같음을 대조함
  (`0d168e0a3..fd6afe23b` 사이 `.go` 변경 0 이라 두 base 의 파일이 같다). 재실행 → `evidence complete or diff-proven exempt`, rc 0.
  번들 내용의 근거(시험 실행 · 커버리지 · 사후 작성 사유)는 a172 review.md 「gate ⑤」 에 있다.
- 쓰지 않은 길: 도구는 a170 에 `--record-landing`(창 끝을 착지 커밋으로 좁힘)을 권했다. 승인된 절차가 번들 복사라 기록하지 않았다.
- 완료 게이트 결과는 아래 「완료 게이트 실행 영수증」(archive 뒤 추가).
