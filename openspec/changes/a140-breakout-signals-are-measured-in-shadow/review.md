# a140 review

## 1. 사람 결정 (2026-10-15)

Manager 전달, 사용자 「권장대로」(2026-10-15):

- **D1~D15 · D17~D20**: design §9 권고 그대로.
- **D16 변경**: 「no default veto threshold. When the S3 seam is first turned on it runs RECORD-ONLY (Jev judges, logs to ledger, never
  blocks). The veto threshold is set later by a human, per market (KR and US separately), from the a128 report after ≥5 sessions per market.」
  근거(Manager 실측, a128-s2 원장 시장마다 J2 990 건): p ≥ 0.5 면 KR 16.0% · US 20.2% 거부, KR 거부군이 60 분 +0.14% 대 −0.39% 로 더 나았음 —
  하루 · 상관 표본이라 결론 아님, 고정 0.5 의 근거가 없다는 증거.
- **D17 후속**: 「a128 is moving to state rev a128-s3 (1.8: WTS news, KR per-stock investor flows, US indices) — v1 Jev state for S3 =
  whatever a128 state rev the report calibrates (s3), not s2.」

반영: design §9 D16 · D17 행, §8 S3 행(RECORD-ONLY 먼저), §8 S3 개요(켜는 순서), §9 「사람 결정 기록」.

### 승인 뒤 구현 정밀화 (Manager 확인 대상 — 승인 범위 안이라고 판단, 이견 있으면 되돌림)

- D3: ATR 창을 opening range 15 봉으로 고정(매 분 전체 봉 ATR 은 이미 판정된 setup 을 소급 재해석 — design §3 「구현 정밀화」).
- D4: opening range 봉은 RVOL 이 정의되지 않음(요청 시 거절, lane 입력 칸 0 — 평가기는 읽지 않음, evaluateFresh B4@52 `bars[15:]`).
- 입력 봉의 연속성(개장 분부터 빈 분 없음) 요구 추가.
- **D21(새 결정 — 기본값으로 진행, 확인 요청)**: breakout v1 설정 중 생산 값이 없는 넷 — `RetestTolerancePPM = 250,000` · `MaxQuoteAgeMS = 5,000` ·
  `MaxSpreadPPM = 3,000` · `MaxEntryDriftPPM = 5,000`. 뒤의 셋은 실측값을 원장에 함께 적어 재계산 가능.

## 2. proposal-freeze 경량 리뷰 (2026-10-15)

- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only, HEAD `044b7b94`). 비례 원칙상 경량(Low — 새 파일만). 저자 컨텍스트와 분리.
- 판정: **APPROVE-WITH-FIXES**. 코드 사실 주장 전부 참(production.go:439~440 · strategy_proposal_authority.go:353 · 370~371 · production.go:321~322 ·
  328~329 · 생산 호출 0 셋 · machine.go:111~117 · :162 · :52~70 · :101 재확인). 안전 경계 위반 0, a138 · a139 파일 겹침 0.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P1 | ARMED 뒤 거절이 정의 안 됨 — 평가기가 ARMED 에서 끝나 setup 이 세션 내내 ARMED, 첫 시도만 보면 엔진보다 PROPOSED 가 적게 잡힘 | **수용** — design §2.1-1 · spec(R2 문장 + 시나리오): 최대 5 분 분마다 재판정, 시도별 실측값 기록 |
| 2 | P1 | `TickMinor` · `FinalCap` · FX `AsOf/FreshUntil` 미정 | **수용** — §2.1-2: 틱 = opening range 저항의 KRX 틱(세션 고정), FinalCap = 최대, FX = 개장~마감 |
| 3 | P1 | 빈 분 하나가 세션 전체를 거절(VI 단일가 미실측) | **수용** — §2.1-3: 개장부터 이어진 앞부분만 평가, 3 분 넘게 안 채워지면 gap 행, 세션 끝 미충족 = 퍼널 `GAP` |
| 4 | P2 | 폴 하한이 개장이라 200 분 뒤 종목당 2 GET(≈1 GET/s) | **수용** — §2.1-4 하한 = 가장 새 저장 봉 − 5 분, §6 예산 문장에 전제로 명기 |
| 5 | P2 | ARMED 봉 위치를 알 수 없음 | **수용** — §2.1-5 앞부분 길이를 늘려 처음 ARMED 인 n 을 찾음 |
| 6 | P2 | 재시작 결정성은 ARMED 전까지만 | **수용** — §2.1-6 원장 setupID 로 중복 방지, §2 문장 고침 |
| 7 | P2 | 봉 ID 가 판에 의존하면 정정으로 setup 이 둘 | **수용** — §2.1-7 · spec R1 문장 + 시나리오, `LaneBarID` = 세션:여는 ms |
| 8 | P2 | 왕복 비용 정수화 · RR 검사 공허 | **수용** — §2.1-8 ceil 규칙, 보고서에 「RR 거절은 구성상 공허」 |
| 9 | P2 | D3 문구 불일치 | **수용** — D3 행에 정밀화 · 사람 확인 대기 표시(tasks 0.4) |
| 10 | P2 | 경로 가드가 하드링크를 못 막음 · 생산 경로 출처 미명시 | **수용** — §2.1-9 이름 · 이웃 journal.db · `os.SameFile`, 출처 = `journal.DefaultPath()` 디렉터리 · 설정 디렉터리 |
| 11 | P2 | 키 공유 — 호가 GET 이 엔진 보호 경로와 같은 그룹 | **수용** — §2.1-10 MARKET_DATA 429 한 번이면 호가 5 분 정지 |
| 12 | P2 | 쓰기 목록에 토큰 캐시 갱신 누락 · 설정 저장 금지 | **수용** — §2.1-11, AST 시험이 config 저장 이름 금지 |
| 13 | P3 | spec 「자기 전용 저장소에만 써야」가 원장 · 토큰 쓰기와 충돌 | **수용** — 「evidence 쓰기는」 |
| 14 | P3 | 빠진 시나리오(ATR 0 · 첫 봉 · 분모 0 · 올림) | **일부 수용** — ATR 0 · 정정 시나리오 추가. 첫 봉 · 분모 0 · 올림은 단위 시험이 덮음(signal_test.go) |
| 15 | P3 | 마지막 봉 미적재 · 마감가 출처 | **수용** — §2.1-12 |
| 16 | P3 | SealBarSeries 비용 | **수용** — §2.1-13 |
| 17 | P3 | `internal/config` import 가 a139 편집 패키지에 컴파일로 묶임 | **수용(위험 기록)** — 파일 겹침 0, a139 가 config 를 바꾸면 도구 빌드만 다시 확인 |
| 18 | P3 | 원장에 위험 한도 출처 기록 — 커밋 전 확인 | **수용** — 3.3 커밋 전 확인 항목 |

FREEZE: 위 P1 셋을 문서로 닫은 뒤 착수(같은 커밋). 남은 사람 확인: D3 정밀화 · D21 기본값(review §1).
