# a128 설계 — Jev shadow 프로브

## 비례 원칙 적용

새 코드가 새 파일(`tools/a128-jev-shadow-probe/`)에만 들어가고 High-risk 경로(주문·손절·
사이징·Guardian·원장·대사·인증·체결)를 만지지 않으므로, Function Logic Map·변이 원장·
다중 적대 리뷰는 `not-applicable`(기존 함수 편집 0, 게이트 판정 변경 0 — docs/WORKFLOW.md
「비례 원칙」). 기본 = 시험 + 경량 리뷰(gstack).

## 데이터 소스 — 2026-10-07 실측 근거 (문서가 아니라 호출로 확인)

| 소스 | 표면 | 상태 |
|---|---|---|
| 랭킹 `market rankings`(MARKET_TRADING_AMOUNT 등) | official | 작동 확인(100행 수신) |
| 시세 `/prices`·호가 `/orderbook` | official | a112 L1c 프로브 선례(leader 재사용 가능) |
| 종목 메타 `/api/v1/stocks` | official | 작동 확인(5심볼 이름 수신) |
| 뉴스 `market news` | **WTS-only** | 세션 만료 관측 — 사람 `auth extend` 후에만, 없어도 동작 |

주의: official 호출에 rate limit 간헐(실측 2026-10-07, 45초 대기로 해소) — 수집 루프는
호출 간 간격과 429/limited 재시도(지수 백오프, 포기 시 그 샘플만 결손 기록)를 가진다.

## TypeSafe 연동

- HTTP API 직접 호출(stdlib `net/http`) — Go SDK 없음, SDK 의존 추가 안 함(YAGNI).
- 구현 전에 https://docs.typesafe.ai/api.md 와 https://docs.typesafe.ai/primitives/noul.md 를
  읽고 요청/응답 계약을 **그 문서 원문 인용으로** `analysis/typesafe-api-contract.md` 에
  동결한다(계약 숫자·문장은 영수증에서 — 추측 금지). 엔드포인트·필드명을 이 설계가
  미리 지정하지 않는 이유: 아직 읽지 않은 계약을 지어내지 않기 위함.
- 인증: `TYPESAFE_API_KEY` env 만. 키가 없으면 프로브는 시작을 거부하고 사유를 말한다.

## 판단 설계

- J1(방향 가설, Noul): state = {market, symbol, name, 최근 체결가, 등락률, 호가 스프레드,
  거래대금 순위, (있으면) 뉴스 헤드라인 목록}. 질문: "이 맥락에서 이 종목이 다음 h분 안에
  기준가 대비 양(+)수익으로 마감할 가능성" — **반증 대상 가설**: calibration 에서 p 구간과
  실현 적중률의 관계가 평평하면 예측력 없음으로 기록하고 J1 은 배선 후보에서 제외.
- J2(악재 veto, Noul): 같은 state. 질문: "이 맥락에 매수를 보류해야 할 명백한 악재
  신호가 있는가". 뉴스 필드가 없으면 state 에 `news: absent` 를 명시해 모델이 결손을
  알게 한다(조용한 생략 금지 — 결손을 숨기면 p 의 의미가 달라진다).
- 질문 원문·버전은 원장 행에 박는다(질문이 바뀌면 p 의 모집단이 달라진다 —
  측정은 잰 순간을 달고 다닌다).

## 원장

- JSONL, append-only, 행 2종: `judgment`(시각 RFC3339·market·symbol·question_id·
  question_rev·state_digest(sha256)·p·기준가·수집 결손 목록), `label`(judgment 참조 id·
  라벨 시각·실현가·실현 수익률·라벨 결손 사유).
- 위치: `openspec/changes/a128-jev-shadow-judgment-probe/analysis/shadow-ledger/`
  (장중 산출물 — 2일 측정 종료 후 커밋; 계좌 데이터가 없으므로 커밋 가능).
- state 원문은 원장에 넣지 않고 digest 만 — 원문은 같은 디렉터리의 state 파일로
  별도 보존(재검증용). 어느 파일에도 계좌 식별자·토큰이 없음을 시험이 강제한다.

## 세션 경계

- KR 09:00~15:30 KST, US 22:30~05:00 KST(당일 영수증 기준, DST 는 market hours 재조회로).
  프로브는 `market hours` 를 시작 시 읽어 세션 밖이면 대기하거나 종료한다(플래그).
- 2일 = 달력 2일의 KR·US 세션 각 2회. 운영: 에이전트가 세션마다 실행(read-only 라 가능),
  장시간 실행은 백그라운드 + 로그 파일.

## 임계 판정 보고서

- p 구간(0.50~1.00, 0.05 폭)별: 표본 수, J1 실현 적중률, J2 는 라벨 정의(악재 판단 뒤
  h분 수익률이 음수였는가)로 적중률. **표본 수가 구간당 30 미만이면 그 구간 판정 보류**를
  보고서 양식에 박는다 — 2일 측정의 한계를 숫자가 직접 말하게.
- 90% 임계의 채택/기각/보류 권고와 근거를 보고서가 내고, 결정은 사람이 한다.

## 후속 (이 change 밖)

1. 수동 보조: 콘솔/CLI 에 J2 판단 표시(읽기 전용 UI) — calibration 보고 후 별도 change.
2. 엔진 veto gate: J2 를 기존 매수 신호의 AND 거부 조건으로 — High-risk change,
   사람 토글·audit·FLM 전 절차. J1 배선은 calibration 이 예측력을 증명한 경우에만 후보.
3. 손절·비상 청산 경로에는 어떤 Jev 판단도 배선하지 않는다(안전 불변식 4) — 이 금지는
   spec 델타에 요구로 박는다.
