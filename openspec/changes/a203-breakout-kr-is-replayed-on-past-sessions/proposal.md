# a203 — 돌파 KR 은 과거 세션에서 재생해 잰다 (breakout-kr-is-replayed-on-past-sessions)

## Why

사용자의 직접 판단 매수가 지고 있어 엔진에 코드로 있는 레인 중 대안을 찾는 중임. 과거 데이터로 성과를 재 볼 수 있는 레인은 돌파 KR
(`kr_short_breakout_retest_v1`) 하나뿐이었음 — 증거가 공식 닫힌 1분봉과 L1 호가뿐인 유일한 가족(a140 design §1).

그런데 그 레인의 측정은 a140 그림자 측정 5세션(2026-10-16 ~ 10-22) 22건뿐이고, n < 30 이라 a140 보고서는 판정을 보류함
(a140 archive `analysis/report-2026-10-22-tool.txt` · `report-2026-10-22-supplement.txt`: 순이익 거래 7, 건당 평균 순 −0.70%, 판정 아님).
장중 그림자 측정을 더 늘리면 세션당 1일씩 걸림. 공식 캔들 API 의 과거 1분봉으로 **같은 코드를 지난 세션에 다시 돌리면** 표본을 한 번에
키울 수 있음.

사용자 승인(2026-11-26): 「돌파 KR 과거 재생을 OpenSpec change 로 시작」. 첫 단계는 1분봉 소급 깊이 실측이며, 브로커 조회는 그때 다시 승인받음.

## What Changes

1. **깊이 실측(tasks 0, 사람 승인 뒤)**: 기존 CLI `tossctl ops call` 로 공식 GET ≤ 9 회 — 종목 1분봉을 며칠 전까지 받을 수 있는지,
   일봉 200개가 어디까지 닿는지, 과거 날짜 달력이 오는지. 결과로 표본 기간 분기를 기계적으로 정함(design §6 · §9).
2. **재생 도구(설계안 (가))**: `tools/a140-breakout-signal-shadow/` 에 새 파일과 `replay` 하위 명령을 더함. a140 의 `evaluateSetup` ·
   `labelProposal` · `planSizing` · `contiguousPrefix` 를 **복사 없이** 그대로 호출하고, 봉은 a140 과 같은 엄격 리더 →
   `officialbars.PollClosedBars` → 도구 전용 evidence DB 경로로 만듦. 응답은 저장소 밖 캐시에 두어 재실행은 다시 조회하지 않음.
   기존 함수 편집은 `dispatch` 의 case 1개뿐(design §2).
3. **사전등록**: 동결 v1 상수 · a140 판 설정 그대로(파라미터 탐색 금지), 호가 관문 주 가정 P0, 종목군 규칙 U-A, 표본 100 세션,
   PASS / FAIL / INSUFFICIENT 숫자 규칙을 수집 전에 커밋함(design §1 · §4 ~ §7).
4. **대조 검증**: a140 이 기록한 5세션을 재생해 a140 원장과 단계 · 제안 일치를 잼. ARMED 전 단계가 하나라도 어긋나면 주 표본을 수집하지 않음.
5. **보고서**: 첫 줄에 호가 관문 미재생 한계, 다음 줄에 종목군 편향, 그다음 판정 · 민감도(S1 · S2) · 결손 · 호출 수.

## What Does NOT Change

- 엔진 · 주문 · Guardian · 원장(journal) · 대사 · 손절 · 콘솔 · 운영 설정 — 무접촉. `internal/` · `cmd/` 생산 Go 편집 0 을 목표로 함
  (종목군 U-B 를 고르면 예외 — 열린 질문 H4).
- 레인 상수 · a140 판 설정 · 비용 모델 값을 바꾸지 않음. 값을 바꿔 가며 좋은 조합을 고르지 않음.
- 주문 · 서명 매니페스트 · 엔진 전달 · 토글 · Jev 호출 0.
- a140 `run` · `report` 경로와 a140 정책 digest 는 그대로.

## Capabilities

### Modified Capabilities

- 없음 — 생산 동작 · 요구사항 변경 0. `skip_specs: true`(design §10, 선례 a202).

## Impact

- 위험 등급: **Low**. 오프라인 측정 도구, High-risk 경로 · 게이트 판정 변경 0.
- 영향 파일(예정): `tools/a140-breakout-signal-shadow/main.go`(`dispatch` case 1개) · 같은 디렉터리 새 파일 `a203_*.go` ·
  `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`(`dispatch` 1건) · `sites.json`(새 배선 자리) — 둘 다 시험 데이터.
- 호출: 깊이 실측 ≤ 9 GET, 본 수집 약 6,800 ~ 10,100 GET(1초 간격 약 2 ~ 3 시간, 장외 시간대만).
- 비례 원칙: FLM 은 `dispatch` 번들 1개(gate ⑤). 변이 원장 · 다중 적대 리뷰는 **not-applicable** — 주문 경로 밖의 오프라인 측정 도구,
  High-risk 경로 · 게이트 판정 변경 0. 리뷰는 경량 1회.
- 분기 근거: `analysis/ast-evidence/claims.md`(logic-map 산출물 14개).

## 열린 질문

### 사람 결정

1. **H1 — 키 · 프로필**: 어느 `--config-dir`(엔진과 같은 Open API 키 vs 별도 앱 키)로 돌릴지. 같은 키면 a166 계정 예산에서 이 도구(보조
   우선순위)가 엔진 보조 읽기와 몫을 나눔 — 엔진이 도는 시간대는 design §8 에서 거절하지만, 장외 시간에도 엔진의 체결 · 대사 읽기와 겹침.
   (기억: 같은 키 공유 시 토큰 캐시 경합 전례가 있음 — `--config-dir` 필수.)
2. **H2 — tasks 0 실행 승인**: 깊이 실측 GET ≤ 9 회(design §9 명령 원문)와 실행 시각.
3. **H3 — 본 수집 승인**: 깊이 결과 · 확정된 세션 목록을 본 뒤, 약 6,800 ~ 10,100 GET 실행.
4. **H4 — 종목군 모집단**: U-A(지금의 6mo 거래대금 상위 100 — 코드 범위 안, 생존 편향 · 모집단 미래 정보 있음) 대 U-B(공식
   `/api/v1/stocks/all` ACTIVE + DELISTED — 편향이 크게 줄지만 `internal/official` 새 읽기 함수 · 일봉 약 2,500 GET 추가). 권고 U-A,
   대조 2 결과로 차이를 보고 후속 change 에서 U-B 를 고려.
5. **H5 — 사전등록 확인**: design §7 판정 숫자(최소 30건 · 제안 세션 20 · 결손 20% · 세 조건 · seed 203)와 §6 표본 규칙을 수집 전에 승인.

### Manager 결정

1. **M1 — 도구 배치**: (가) a140 패키지 확장(권고) 대 (나) 새 디렉터리 + 동일성 시험(design §2 비교표).
2. **M2 — a166 census 시험 데이터 갱신**: (가)면 `shapes.json` 의 `main.go||dispatch` 1건 갱신, 두 안 모두 `sites.json` 새 자리 추가를 허용할지.
3. **M3 — 시간대 거절 범위**: KR 평일 08:00~16:00 KST 에 미국 정규장까지 막을지(권고: 막음).
4. **M4 — 원장 커밋**: 재생 원장을 a140 처럼 change `analysis/replay/` 에 커밋할지(종목 코드 · 가상 수량 포함, 계좌 데이터 0). 크기가 a140 의
   약 20배(100 세션)일 수 있음 — 요약 보고서만 커밋하고 원장은 지문만 남기는 안도 가능.
5. **M5 — 최소 세션 수 40(깊이 분기)과 제안 세션 20(§7)의 값**: 근거는 design 에 적었으나 Manager 가 확정.
