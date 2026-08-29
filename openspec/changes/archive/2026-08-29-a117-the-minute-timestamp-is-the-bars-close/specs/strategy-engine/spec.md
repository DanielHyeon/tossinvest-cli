## MODIFIED Requirements

### Requirement: 첫 lane는 frozen KRX Parker VWAP conservative v1이다
첫 lane는 StockOS commit `d75113d3c338148606d86c8aedbbeb7ed446c0b8`와 source-set digest `09260ac29e50ed4d2a43d0e274f9a17465e00ee36fb61d759127f158985c23bd`의 `parker_vwap_trend_v1` conservative gate를 KRX regular-session closed 5-minute input에만 적용해야 한다 (SHALL). server-owned immutable constants와 gate order를 바꾸려면 새 lane version과 activation manifest 승인이 필요하다 (SHALL).
KRX open은 `09:00 KST`로 봉인하고 session open/close/evaluation은 같은 KST 거래일이어야 하며, caller가 이동한 session window를 받아서는 안 된다 (MUST NOT).
공식 1분봉의 `timestamp`는 봉이 **닫힌** 시각이므로 (2026-08-18 KR/US 라이브 실측), 라벨이 `t`인 봉은 `[t-1분, t)` 구간을 담아야 한다 (SHALL). 5분 버킷이 여는 시각은 첫 라벨보다 1분 이르며 5분 격자 정렬과 `closed_at`은 그 여는 시각에서 계산해야 한다 (SHALL). 정규장 편입 판정은 라벨 기준 `09:01`부터 `15:30`까지 양끝 포함이어야 하고 (SHALL), `09:00` 라벨(개장 전 `08:59`~`09:00`)을 정규장으로 받아들이거나 `15:30` 라벨(정규장 마지막 1분)을 버려서는 안 된다 (MUST NOT).

#### Scenario: StockOS translated accept arithmetic
- **WHEN** frozen fixture가 VWAP above/slope, EMA9 bullish pullback, LVN forward space, untangled/band/RR, age/drift를 모두 통과한다
- **THEN** `krx_parker_vwap_conservative_v1`은 source와 같은 entry, 0.7% stop, 3R target, expected RR와 accept provenance를 반환한다

#### Scenario: StockOS 세션 refusal 패리티
- **WHEN** frozen KRX calendar가 non-trading, 시가 동시호가, 종가 동시호가, 시간외, 시초 제외 또는 close-minus-45m cutoff를 나타낸다
- **THEN** lane은 source 순서와 같은 typed refusal reason을 반환하고 EntryDecision과 broker call은 0건이다

#### Scenario: 지원하지 않는 시장
- **WHEN** US 또는 pre/after-market candidate가 첫 lane를 요청한다
- **THEN** typed unsupported-scope refusal을 반환하고 EntryDecision과 broker call은 0건이다

#### Scenario: source digest 변경
- **WHEN** runtime lane source/constants digest가 manifest의 frozen digest와 다르다
- **THEN** desired state와 무관하게 effective entry는 OFF이고 새 manifest 승인 없이는 dispatch하지 않는다

#### Scenario: 불완전한 5분봉
- **WHEN** official 1분봉에 KST 정규장 밖 minute, 중간 누락 또는 아직 닫히지 않은 bucket이 있다
- **THEN** lane은 해당 5분봉을 만들지 않고 typed bar-integrity refusal과 broker call 0건을 반환한다

#### Scenario: 장 시작 첫 버킷
- **WHEN** 라벨 `09:01`~`09:05`인 다섯 개의 공식 1분봉을 `09:05` 이후에 집계한다
- **THEN** 여는 시각 `09:00`, 닫는 시각 `09:05`의 5분봉 하나를 만들고 정렬 거절은 0건이다

#### Scenario: 개장 전 1분을 담은 라벨
- **WHEN** 라벨 `09:00`인 봉이 5분 버킷에 포함된다
- **THEN** 그 봉은 `08:59`~`09:00`을 담으므로 typed outside-regular-session refusal을 반환하고 5분봉을 만들지 않는다

#### Scenario: 종목 상태 권위 부재
- **WHEN** HALT/LIMIT/MANAGED를 판정할 authoritative 상태가 없거나 30초보다 stale이다
- **THEN** quote나 price limit로 추측하지 않고 effective entry를 OFF로 유지한다
