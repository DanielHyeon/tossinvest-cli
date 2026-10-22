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
- [x] 1.3 RED — `analysis/red/red-3-tool-compile.txt`(시험만 · 구현 0). 도구 시험(mock 브로커 · mock 달력): 생산 `evidence.db` 경로 거절(직접 · 심링크) · 주문/계좌 이름 금지(AST) ·
      ARMED 에서만 호가 재조회 · setup 당 PROPOSED 1 회 · 결손 행 · 라벨(손절 선착 · 한 봉 동시 = 손절 · 마감 R) · 보고서 n<30 보류 ·
      그룹별 GET 계수
- [x] 1.4 GREEN — 도구 9 파일(main · source · guard · evaluate · sizing · label · ledger · report · shadow), 시험 21. 리뷰 반영 규칙 §2.1 전부(ARMED 5 분 재판정 · 틱 고정 · 앞부분 평가 · 폴 하한 · 재시작 복원 · MARKET_DATA 정지). 경량 변이 14/14 CAUGHT(`analysis/red/mutation-tool.txt`, M13 은 시험 추가 뒤). `tools/a140-breakout-signal-shadow/`: 랭킹 → 종목 집합(D8) · 분 단위 `PollClosedBars` · `SealBarSeries` ·
      `breakoutsignal` · `breakoutlane.Evaluate(prior=nil)` · `PollQuoteL1`(시작 · ARMED) · 원장 · 라벨러 · 보고서
- [x] 1.5 VERIFY — `analysis/red/verify.txt`: gofmt 빈 · vet · 태그 vet rc=0 · 38 PASS / 0 FAIL · -race ok · `make lint` rc=0. `go vet ./internal/breakoutsignal/... ./tools/a140-breakout-signal-shadow/...` · 태그 vet · `gofmt -l` 빈 출력 ·
      `rtk proxy go test -count=1` 두 패키지 · `-race` · `make lint`(gofmt 범위가 `./internal` 을 포함)
- [x] 1.6 팀메이트 커밋(푸시 금지) — 허용 경로 전수 확인 후 `git commit -- <paths>`

## 2. 리뷰 · 착지

- [x] 2.1 경량 코드 리뷰(독립 1 — review §3: BLOCK → P1 3 · P2 6 · P3 4 전부 수리, 변이 2 판 생존 0) 및 Manager 검증 완료. 검증 근거는 1.5의 focused/race/vet 영수증과 2.2의 경로 전수·시험 재실행·생산 import 0 대조다. **3절 측정 완료를 뜻하지 않는다.**
- [x] 2.2 Manager 푸시 — 2026-10-15 Manager 검증: `go test` breakoutsignal·도구 45 PASS 재실행 · 생산 import 0(grep) · marketSource = 읽기 GET 다섯(Rankings·TypedMarketCalendar·StrictMinuteCandles·StrictOrderbookTop·StrictLastPrice) 직접 판독 · D3 opening-range ATR 고정(소급 뒤집힘 방지)과 D21 기본값은 Manager 수용(그림자 측정 한정, 실측값 원장 병기)

## 3. 측정 (D9 — 기본 KR 5 세션)

- [x] 3.1 KR 세션 실행(에이전트 가능 — read-only GET). 실행 전 호출 예산 기본값 사람 승인(D11)
      — 유효 KR 세션 **5/5**(10-16 · 19 · 20 · 21 · 22): 5 세션 모두 공식 달력 출처 · 판(lot) 하나(`…3ff46b8b`) · 종목 30 ·
      세션 끝 요약 행 · 제안 전부 라벨 · MINUTE_GAP 0(`analysis/report-2026-10-22-supplement.txt` 「세션 유효성」). 10-19 의 두 프로세스
      (`/tmp/a140-kr-1019.log` 09:55:33 KST `context canceled` → `1019b` 09:57:50 재시작)는 design §2.1-6 재시작 경로로 세션 키 · universe 행 ·
      요약 행이 각 하나인 **한 세션**(정지 ≈2.5 분은 폴 하한 5 분이 소급 적재, 정지 자체의 결손 행은 없음). D11 승인 = review §1 일괄 승인
      「권장대로」 D1~D15(커밋 `526637460`, 2026-10-15 16:26 KST — 첫 실행 10-16 08:40 KST 전). D11 단독 영수증은 없다(10-20 reconciliation 의
      「receipt 없음」은 그것). 실측 봉 GET 평균 0.27~0.45/s — D11(0.5/s) 안.
- [x] 3.2 라벨러 완주 확인(마지막 PROPOSED 의 마감까지) — PROPOSED 22 · 라벨 22 · 미라벨 0 · 고아 0. 5 세션 모두 요약 행이 마감 + 라벨 유예
      30 분(06:59Z ≈ 15:59 KST)에 있고 10-22 프로세스는 종료(pgrep 0).
- [x] 3.3 원장 커밋(계좌 데이터 0 확인 후, evidence DB 는 커밋하지 않고 digest 만) — 원장 sha256 `0b802a92…40c`(654 행). 계좌 · 잔고 · 보유 ·
      주문 ID · 토큰 · 키 패턴 0 건(행 종류는 universe · signal · label · gap · budget · session 뿐, 종목 · 시장 가격 · 가상 수량 · 설정 한도 출처 경로만).
      evidence DB sha256 `772e4ff8…6765`(72,622,080 B) — `analysis/report-2026-10-22-tool.txt` 머리에 기록, 커밋 안 함.

> **2026-10-22 measurement close.** 위 3.1~3.3. 2026-10-20 note 는 그 시점(2/5 세션) 기록으로 남긴다.

> **2026-10-20 read-only reconciliation.** Current canonical ledger has closed KR sessions 2/5 plus one ongoing key, 14 completed labels total, and no preserved D11 approval receipt. Tasks 3.1–3.3 remain open. Details: `analysis/reconciliation-2026-10-20.md` and the paired JSON. This does not substitute for a128 market-by-market measurement or a human D16 decision.

## 4. 판정 · 종결

- [x] 4.1 보고서 — 세션별 퍼널 · 거절 분포 · RVOL 반사실 · 제안 수 · 결과 분포(n<30 보류) · 호출 예산 실측
      — 도구 보고서 원문 `analysis/report-2026-10-22-tool.txt`(`go run ./tools/a140-breakout-signal-shadow report`) + 보조 집계
      `analysis/report-2026-10-22-supplement.txt`(`analysis/report_supplement.py` — 세션 유효성 · 예산 · R 구간 · RVOL 종목-세션 기준).
      요지: 종목-세션 150 → 돌파 74 → PROPOSED 22(세션당 2~7) · 종결 timed_out 44 · invalidated 6 · armed_expired 2. 결과 close 17 · stop 3 ·
      target 2 — n=22<30 판 합계 판정 보류. RVOL 1.2× 근접 0(낮춰도 표본 불변) · 2.0× 이면 제안 13 · 2.5× 이면 7. 결손 55 행(봉 401 41 ·
      호가 401 4 · 교차 호가 bid>ask 거절 10). design §6 의 「수집 시간 > 간격 절반 WARNING」은 도구에 구현 없음(10 분 창 199 중 4 초과) — 기록만.
- [x] 4.2 사람 결정 기록 — S4(엔진 증거 생산) 착수 여부 · breakout 지표 판본 유지/수정 · 다른 가족 검토 여부
      — review §4(2026-10-22 사용자 원문 「a140: 중단」 · 「a140 승인 확인: 조회 예산(D11)과 호가 한도 기본값(D21)을 지금대로 둬」):
      **S4 착수 안 함 · breakout 가족 측정 중단 · 지표 판본 수정 없음**. D11 · D21 사람 확인 빈칸 닫힘. 다른 가족 검토는 이번 결정에 없음.
      근거 관측 = 가상 손익(supplement 「가상 손익」 절, 원장에서 재측정 일치): 제안 22 · 순이익 7 · 평균 순 −0.70% · 합계 −65,780 KRW /
      명목 9,267,200 KRW, 양수 세션은 10-19 하나. KR 타이머 `tossos-a140-kr.timer` 는 Manager 가 ≈19:55 KST disable(재확인 disabled · inactive).
      서술형 `analysis/report-2026-10-22.md` 는 도구가 쓰기를 거부해 없음(review §4).
- [ ] 4.3 sdd-sync → gate → archive → Story 경로 갱신 — 2026-10-22 시도: 게이트 5 단계(base 창) FAIL, base 재고정은 사람 승인 대기(review §5)
