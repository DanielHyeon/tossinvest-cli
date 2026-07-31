## ADDED Requirements

### Requirement: 승인 후보는 독립 lane의 순수 결정으로 변환된다
strategy lane는 ApprovedCandidate와 versioned market inputs를 받아 EntryDecision 또는 명시적 refusal을 반환해야 하며 broker·journal·운영 토글을 직접 변경해서는 안 된다 (MUST NOT).

#### Scenario: 진입 결정
- **WHEN** 활성 lane가 유효 ApprovedCandidate를 평가해 진입 조건을 충족한다
- **THEN** candidate ID, lane ID/version, stop/target와 RiskIntent 입력을 가진 결정이 생성된다

#### Scenario: lane OFF
- **WHEN** lane desired/effective state가 OFF다
- **THEN** 신규 EntryDecision과 buy mutation은 0건이고 기존 exit loop는 계속된다

### Requirement: 첫 lane는 frozen KRX Parker VWAP conservative v1이다
첫 lane는 StockOS commit `d75113d3c338148606d86c8aedbbeb7ed446c0b8`와 source-set digest `09260ac29e50ed4d2a43d0e274f9a17465e00ee36fb61d759127f158985c23bd`의 `parker_vwap_trend_v1` conservative gate를 KRX regular-session closed 5-minute input에만 적용해야 한다 (SHALL). server-owned immutable constants와 gate order를 바꾸려면 새 lane version과 activation manifest 승인이 필요하다 (SHALL).

#### Scenario: StockOS golden accept
- **WHEN** frozen fixture가 VWAP above/slope, EMA9 bullish pullback, LVN forward space, untangled/band/RR, age/drift를 모두 통과한다
- **THEN** `krx_parker_vwap_conservative_v1`은 source와 같은 entry, 0.7% stop, 3R target, expected RR와 accept provenance를 반환한다

#### Scenario: 지원하지 않는 시장
- **WHEN** US 또는 pre/after-market candidate가 첫 lane를 요청한다
- **THEN** typed unsupported-scope refusal을 반환하고 EntryDecision과 broker call은 0건이다

#### Scenario: source digest 변경
- **WHEN** runtime lane source/constants digest가 manifest의 frozen digest와 다르다
- **THEN** desired state와 무관하게 effective entry는 OFF이고 새 manifest 승인 없이는 dispatch하지 않는다

#### Scenario: 불완전한 5분봉
- **WHEN** official 1분봉에 KST 정규장 밖 minute, 중간 누락 또는 아직 닫히지 않은 bucket이 있다
- **THEN** lane은 해당 5분봉을 만들지 않고 typed bar-integrity refusal과 broker call 0건을 반환한다

#### Scenario: 종목 상태 권위 부재
- **WHEN** HALT/LIMIT/MANAGED를 판정할 authoritative 상태가 없거나 30초보다 stale이다
- **THEN** quote나 price limit로 추측하지 않고 effective entry를 OFF로 유지한다

### Requirement: strategy entry는 공식 LIVE 경로만 사용한다
승인된 strategy entry는 Guardian, durable journal과 official Open API gateway를 순서대로 통과해야 하며 paper/shadow/canary order path를 가져서는 안 된다 (MUST NOT).

#### Scenario: 운영자 LIVE 승인
- **WHEN** 전체 gate가 통과하고 운영자가 lane와 automation을 명시적으로 승인한다
- **THEN** 다음 유효 결정은 공식 LIVE gateway를 사용한다

#### Scenario: Guardian refusal
- **WHEN** Guardian이 첫 실패 단계에서 거부한다
- **THEN** broker request는 0건이고 refusal과 provenance가 journal에 기록된다
