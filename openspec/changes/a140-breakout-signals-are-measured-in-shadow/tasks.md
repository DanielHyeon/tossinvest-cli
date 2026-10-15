# a140 Tasks

Story: `STORY-TOS-a140` · **Low**(새 파일만 · High-risk 경로 무접촉 · 주문 side effect 0) — Function Logic Map: not-applicable
(기존 함수 편집 0 · 게이트 판정 변경 0, docs/WORKFLOW.md 「비례 원칙」). 시험 + lint + 경량 gstack 리뷰.

> 허용 경로: `internal/breakoutsignal/`(신규) · `tools/a140-breakout-signal-shadow/`(신규) · 이 change 디렉터리.
> a138 · a139 가 편집 중인 파일(`risk.Policy` · `guardian_wiring.go` · `interlock.go` · config · 콘솔 `settings_limits`)은 만지지 않는다.

## 0. 계약 · 증거 · 사람 결정

- [x] 0.1 proposal · design · spec 델타 · tasks · Story 등록(FEAT-TOS-013)
- [x] 0.2 분기 주장의 AST 근거 — `analysis/ast-evidence/`(11 함수, HEAD `050d092f`)
- [x] 0.3 CodeGraph hard evidence · 증거 조정 — `analysis/code-context/`(codegraph 1.6.0, R1~R8; 브리프의 「4/8」을 8/8 인스턴스 · 0/8 생산자로 정정)
- [x] 0.4 사람 결정 기록(review §1, 2026-10-15 「권장대로」 · D16 변경 · D17 후속). 남은 확인: D3 정밀화(ATR 창 고정) · D21 기본값 — Manager 에 보고
- [x] 0.5 CodeGraphContext — not-applicable 유지(새 파일만; import 할 심볼은 CodeGraph callers + 리뷰어 재확인으로 닫음, codegraphcontext-context.md)
- [x] 0.6 proposal-freeze 경량 리뷰 → review §2 APPROVE-WITH-FIXES, P1 셋 · P2 아홉 문서 반영
- [x] 0.7 base 고정 — `5266374` (`capture_change_base.py`, READY)

## 1. 구현 (RED → GREEN → REFACTOR, 새 파일만)

- [x] 1.1 RED — `analysis/red/red-1-compile.txt` · `red-2-behaviour.txt`(구현 초안이 먼저 있었음을 영수증에 기록). `internal/breakoutsignal` 시험: 네 지표의 경계값(직전 봉 14/15 · ATR 13/14 봉 · high = low · 분모 0 · 거래량 0 봉) ·
      결정성(같은 묶음 = 같은 바이트) · 올림/내림 방향 · typed 거절(0 대체 금지) · float 미사용(AST 시험)
- [x] 1.2 GREEN — `analysis/red/green-breakoutsignal.txt`(17 PASS · -race ok · vet · gofmt). 리뷰 반영: lane 봉 ID = `LaneBarID`(판 무관). `internal/breakoutsignal`: `strategyevidence.BarSeries` → `breakoutlane.ClosedBarInput` 목록 + `ATRMinor` + 지표 digest
- [ ] 1.3 RED — 도구 시험(mock 브로커 · mock 달력): 생산 `evidence.db` 경로 거절(직접 · 심링크) · 주문/계좌 이름 금지(AST) ·
      ARMED 에서만 호가 재조회 · setup 당 PROPOSED 1 회 · 결손 행 · 라벨(손절 선착 · 한 봉 동시 = 손절 · 마감 R) · 보고서 n<30 보류 ·
      그룹별 GET 계수
- [ ] 1.4 GREEN — `tools/a140-breakout-signal-shadow/`: 랭킹 → 종목 집합(D8) · 분 단위 `PollClosedBars` · `SealBarSeries` ·
      `breakoutsignal` · `breakoutlane.Evaluate(prior=nil)` · `PollQuoteL1`(시작 · ARMED) · 원장 · 라벨러 · 보고서
- [ ] 1.5 VERIFY — `go vet ./internal/breakoutsignal/... ./tools/a140-breakout-signal-shadow/...` · 태그 vet · `gofmt -l` 빈 출력 ·
      `rtk proxy go test -count=1` 두 패키지 · `-race` · `make lint`(gofmt 범위가 `./internal` 을 포함)
- [ ] 1.6 팀메이트 커밋(푸시 금지) — 허용 경로 전수 확인 후 `git commit -- <paths>`

## 2. 리뷰 · 착지

- [ ] 2.1 gstack 경량 리뷰 + Manager 검증(경로 전수 · 시험 재실행 · 생산 import 0: `rtk proxy grep -rn "breakoutsignal" internal cmd` 비시험이
      새 패키지 자신뿐)
- [ ] 2.2 Manager 푸시

## 3. 측정 (D9 — 기본 KR 5 세션)

- [ ] 3.1 KR 세션 실행(에이전트 가능 — read-only GET). 실행 전 호출 예산 기본값 사람 승인(D11)
- [ ] 3.2 라벨러 완주 확인(마지막 PROPOSED 의 마감까지)
- [ ] 3.3 원장 커밋(계좌 데이터 0 확인 후, evidence DB 는 커밋하지 않고 digest 만)

## 4. 판정 · 종결

- [ ] 4.1 보고서 — 세션별 퍼널 · 거절 분포 · RVOL 반사실 · 제안 수 · 결과 분포(n<30 보류) · 호출 예산 실측
- [ ] 4.2 사람 결정 기록 — S4(엔진 증거 생산) 착수 여부 · breakout 지표 판본 유지/수정 · 다른 가족 검토 여부
- [ ] 4.3 sdd-sync → gate → archive → Story 경로 갱신
