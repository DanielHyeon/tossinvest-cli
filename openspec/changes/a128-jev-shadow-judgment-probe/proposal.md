# a128 — Jev 의미 판단 shadow 프로브 (jev-shadow-judgment-probe)

## Why

사용자 결정(2026-10-07): 매매 판단에 TypeSafe System One 모델(Jev)의 확률 판단을 쓰되,
"p ≥ 90% 매매" 임계는 **측정 없이 생산에 배선하지 않는다**. Jev 의 p 는 "주어진 state 에
대한 질문의 답이 yes 일 확률"이지 시장 방향의 실현 확률이 아니므로(TypeSafe confidence 문서,
2026-10-07 읽음), 임계의 의미는 이 도메인의 원장으로만 확정할 수 있다. 벤더 문서 스스로
"임계값은 자기 데이터로 측정 후 결정"을 요구한다.

이 change 는 그 측정 도구 — **생산 배선 0 의 shadow 프로브** — 하나만 만든다. 측정 창은
사용자 지시로 **2일**(KR 2세션 + US 2세션, 양 시장 적용 지시 2026-10-07)이며, 2일 표본의
통계 한계(임계 "확정"이 아니라 예비 판정)는 보고서에 명기한다. 원장 포맷은 측정 연장이
가능하게 설계한다.

## What Changes

1. `tools/a128-jev-shadow-probe/` 신규 Go 도구(새 파일만, 생산 패키지 무접촉):
   - **수집 루프**: 세션 중 간격 샘플링(기본 10분) — official rankings 상위 K(기본 10,
     시장별)에서 종목 후보를 얻고, 종목별 state 를 공개 시장 데이터로만 조립
     (시세 /prices·호가 /orderbook·랭킹 맥락·stocks 메타, 가능 시 WTS 뉴스 요약).
   - **판단 2종**(TypeSafe HTTP API, Noul):
     J1 가설 — "이 state 에서 h분(기본 60) 내 양(+)수익이 날 것인가"(방향 예측 — 반증 대상),
     J2 veto — "이 state 에 매수를 보류할 명백한 악재가 있는가"(보수 방향).
   - **원장**: JSONL append — 시각·시장·심볼·질문 id·state digest·p·기준가.
   - **라벨러**: 판단 h분 뒤 /prices 재조회로 실현 수익률을 같은 원장에 라벨 행으로 append.
   - **보고서 생성기**: p 구간별(0.5~1.0, 0.05 단위) 실제 적중률 표 + 표본 수 + 2일 한계 명기.
2. spec 델타: 신규 capability `semantic-judgment-shadow` — 주문 side effect 0,
   외부 전송에 계좌 데이터 금지, 임계값의 생산 배선은 shadow 측정 영수증을 전제.
3. 측정 실행(2일) + calibration 보고.

## What Does NOT Change

- 생산 코드·엔진·주문 경로·콘솔: 무접촉. 이 change 로트에 `internal/`·`cmd/` 편집 0.
- 수동 매매 보조(콘솔/CLI 판단 표시)와 엔진 veto gate 배선: **후속 change** — 이 측정의
  결과 보고가 그 change 의 입력이다. 사용자 결정(1=둘 다, 2=둘 다)은 단계 순서로 수용:
  shadow(a128) → 수동 보조 → 엔진 veto. J1(방향 예측)은 calibration 이 예측력을 보여 주기
  전에는 어떤 발주 경로에도 배선하지 않는다.

## Human Gates

- TypeSafe API 키 조달(`TYPESAFE_API_KEY` env — 저장소·기억에 키 저장 금지).
- WTS 뉴스 소스를 쓰려면 `tossctl auth extend`(세션 만료 관측 2026-10-07). 뉴스 없이도
  프로브는 동작해야 한다(state 에 news 필드 부재 명기).
- 프로브 자체는 read-only GET + 외부 판단 API 호출뿐이라 mutating 이 아니다 — 에이전트
  실행 가능. 비용이 드는 외부 API 호출량(세션당 샘플 수)은 기본값을 사람이 승인한다.

## Impact

- Affected specs: `semantic-judgment-shadow`(신규).
- Affected code: `tools/a128-jev-shadow-probe/`(신규)만.
- 안전: High-risk 경로 무접촉. 외부로 나가는 state 는 공개 시장 데이터만 — 계좌·보유·
  시크릿·세션 토큰 금지(조립기에서 차단을 시험으로 고정).
