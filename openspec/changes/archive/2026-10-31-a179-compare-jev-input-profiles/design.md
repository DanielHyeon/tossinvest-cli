# 설계 계약

## 비교 대상

`initial`은 과거 숫자를 재사용하지 않고 현재 스냅숏에서 최초 단순 입력의 데이터 종류를 재현함. 버전은 `initial-fields-v1`. current와 initial 모두 현재 J1~J4 질문 원문·질문 revision/digest와 모델을 동일하게 유지함. 새로운 수익성 기준을 만들지 않음.

## 1. 초기 입력 생성

PublicState를 allowlist 투영함. 유지: Market, Symbol, Name, ListingMarket, Currency, ObservedAt, HorizonMinutes, BasePrice, ReferencePrice, ChangePctVsReference, RankingType, RankingRank, TradingAmount, TradingVolume, TopOfBook, MinutesSinceOpen, MinutesToClose, SessionSource. 제거: Intraday, Daily, MarketContext, WarningTypes, PriceLimits, StockFlows, NewsHeadlines. News=absent. MissingInputs는 기본 필드의 실제 결측만 새 슬라이스로 구성하고 `news: intentionally omitted by initial-fields-v1`를 고정 문구로 추가함. 제거된 분봉·일봉·시장 정보·수급·뉴스·경고·가격제한 정보의 조회 오류가 모델에 유출되지 않아야 함. 알 수 없는 결측 문구는 무조건 복사하지 않음. source의 슬라이스·포인터가 후속 투영/편집으로 바뀌지 않는 격리를 보장함.

## 2. CLI와 출력 격리

사용자 후속 지시(2026-10-31)에 따라 run의 기본 profile은 compare임. 별도 옵션 없이 현재·초기 두 판단을 같은 회차에서 자동 수행함. `--input-profile current`는 기존 단독 측정으로 되돌리는 명시적 opt-out이며 기존 ID·질문·직렬화·보고를 보존함. `--input-profile initial`은 초기 입력 단독 모드임.

`--out`을 생략하면 current는 기존 `analysis/shadow-ledger`, compare는 형제 `analysis/shadow-ledger-compare`, initial은 형제 `analysis/shadow-ledger-initial`로 자동 저장함. 명시적 custom --out은 허용하되 initial/compare가 기존 shadow-ledger 또는 그 하위/별칭을 쓰는 경우 외부 호출·쓰기 전에 거부함. 경로 정규화와 가능한 symlink 해석을 적용함. 정상적인 run-probe.sh와 KR/US 서비스는 profile/out을 지정하지 않는 것이 읽기 전용 확인되었으므로 운영 파일 수정 없이 새 구현의 기본값을 따를 수 있음. 본 작업은 그 서비스를 재시작하거나 실제 API를 호출하지 않음.

report도 옵션 없이 새 비교 출력 경로를 읽고 profile별 calibration과 paired 요약을 보여야 함. `--ledger <기존 원장>`을 명시하면 옛 보고 형식을 보존함. 초기 단독 원장도 explicit ledger로 보고 가능함. 예전 사전등록 평가기의 기본 경로는 그대로이므로 새 비교 자료는 새 평가 진입점으로 읽는다고 사용법에 명시함. 자동 비교는 동일 스냅숏당 두 요청(기존 대비 한 요청 추가)을 사용하며 수집 API와 질문은 그대로임.

## 3. 동일 수집 결과와 실패

단일 cycle의 후보 추출·수집 결과를 공유하여 두 profile을 생성함. 기존 수집을 profile마다 반복하지 않음. 두 판단은 관측 시각·기준가·horizon·라벨 기한·모델·질문을 공유함. 관측 가격을 응답 완료 가격으로 바꾸지 않음. profile별 선별을 수행하고 screen/selection 집계를 섞지 않음. 한쪽 판단 실패에도 다른 쪽 결과를 기록하고 실패 profile과 원인을 gap에 남김. 중단/취소로 시도하지 못한 쪽도 성공으로 위장하지 않음. 요청 순서 및 지연의 한계는 문서화함; 무작위 교차 순서 같은 추가 실험은 범위 밖임.

## 4. 원장·라벨·재시작

`state_digest`는 실제 전송 상태의 저장 바이트 해시이고 `source_state_digest`는 전체 수집 상태의 저장 해시임. optional `input_profile`은 current/initial을 식별함; 기존 누락 행은 legacy current로 읽음. 비교 cohort를 구분하여 standalone current를 비교군으로 오인하지 않음. initial state_rev는 initial-fields-v1, current는 기존 state_rev임. 신규 initial/compare 판단 ID는 profile로 충돌을 방지함. 기본 current ID는 보존함. 동일 source로부터 나온 각 질문의 ID는 같은 가격 라벨 결과와 연결함. pending 복원도 두 profile을 보존하며 가격 재조회는 공유함. 결손은 null 또는 gap으로 남기며 성공값을 복제해 대체하지 않음.

해시는 기존 storeState 계약의 JSON payload 바이트를 기준으로 하며 파일 끝에 붙이는 개행은 제외함. 새 digest 알고리즘으로 옛 자료를 재해석하지 않음. 비기본 판단 ID에는 profile뿐 아니라 cohort도 포함하여 같은 관측 시각의 별도 실행이 한 출력에 존재해도 라벨이 충돌하지 않게 함.

원장은 additive optional 메타데이터만 추가하며 옛 행을 다시 쓰지 않음. current 상태는 source/input digest가 같을 수 있음. 롤백은 새 실행을 중지하고 명시적 `--input-profile current`로 실행함; 실험 출력은 보존하고 기존 원장으로 병합하지 않음.

실패 gap에도 cohort·profile·시장·종목·관측 시각·source digest를 기록함. 두 군 모두 실패해 judgment 행이 없더라도 하나의 시도 쌍과 두 실패를 추적할 수 있어야 함. 전체 source 상태는 판단 전에 저장하여 이후 판단 실패로 가격 관측 자료까지 유실되지 않게 함.

## 5. 보고와 오프라인 비교

기존 calibration 보고서는 profile을 로트 키에 포함하되 legacy current 표시는 기존 형식을 보존함. 비교 보고는 동일 cohort·source digest·시장·종목·관측 시각·질문 id/rev/digest·모델·horizon·기준가가 일치하는 응답만 짝지음. 불일치·중복·한쪽 실패를 정상 쌍으로 집계하지 않음. 입력군별 judgment 수/라벨 수, 양쪽 성공·한쪽 성공·실패 건수, J2 initial-current 확률 차이와 p>=0.6 문턱 교차 수를 표시함. 질문별 행 수를 모델 요청 수로 오인하지 않음.

비교 평가 진입점은 실제 입력 검토에 state_digest, 가격 평가에 source_state_digest를 사용함. KR 기존 오전 9시 J2 신호 창과 10:00→14:10 가격·왕복 비용 0.50%, US 기존 신호 창/60분 라벨·왕복 비용 2.10%를 재사용하되 기존 사전등록 파일/결과는 수정하지 않음. source가 누락되면 결측으로 보고하고 초기 입력에 분봉이 없다는 이유로 전체 자료가 없는 것으로 판정하지 않음. profile별 종목-일 중복 제거·공통 비교 대상·날짜 동일 가중 평균·비용 전후 평균·가격 결측 수를 표시함. 동일 관측 시각의 동일 종목 쌍은 profile별로 다른 가격·수익률을 만들지 않음. 미분류는 일반주식으로 간주하지 않음. 실표본 수익성 PASS 또는 초기 수익률 회복을 자동 선언하지 않음.

US 종목-일 성과는 기존 규칙인 각 profile의 첫 강한 경고 관측(경고가 없으면 첫 관측)을 선택함. 따라서 같은 종목-일이어도 두 profile이 다른 관측 시각을 선택하면 각각 다른 60분 구간의 수익률이 보고될 수 있음. 이는 같은 관측의 결과 가격을 바꾼 것이 아니라 입력 변화로 선택 시점까지 달라진 전략 결과임. 보고서는 이 규칙과 선택 시각이 다른 US 종목-일 수를 명시하여 고정 진입 시각의 순수 입력 비교로 오인하지 않게 함. 모델 p 차이·문턱 교차는 반드시 동일 관측 쌍으로만 계산함. 선택 시각이 다른 합성 US 사례와 각 시각의 공유 라벨을 시험으로 고정함. KR은 선택 시각과 무관하게 기존 고정 진입/청산 가격을 공유함.

KR 사후 가격은 오전 신호의 source 한 개에 미래 가격이 있다고 가정하지 않음. 같은 시장·날짜·종목의 후속 회차 source 분봉을 합치며 같은 봉은 가장 늦게 관측한 source를 사용함. 후속 판단이 실패해도 저장된 source와 gap 연결을 통해 해당 분봉을 사용할 수 있음. source digest 검증 후에만 평가에 사용하며 이후 회차에도 필요한 봉이 없으면 결측으로 남김. 미래 가격은 사후 평가에만 쓰고 모델 입력으로 되돌려 넣지 않음.

## 6. 구현 전 증거와 검토

이 문서는 목표 동작 계약이며 기존 함수 내부 분기에 대한 증명은 아님. 구현자는 CodeGraph/현재 HEAD·대상 테스트와 Go AST를 확인하고, 변경되는 기존 함수의 FLM/BTM 및 Pre-Edit 근거를 남긴 뒤 편집함. 도구 부재는 실행 영수증과 대체 검증 한계를 기록함. 새 수집기·모델 클라이언트·외부 의존성은 만들지 않음.

## 7. 검증 및 종료

HTTP fixture에서 실제 요청 필드 제거와 두 군 질문 동등성을 검사함. 원본 보존, profile/out 무옵션의 자동 compare 선택, 명시적 current 회귀, 모드별 자동 출력과 출력 가드, ID·digest·공유 라벨·재시작, profile별 screening, 한쪽 실패, 쌍 불일치·중복, source 분봉 사용, 기존 원장 호환을 대상 테스트로 확인함. 합성 자료로 기본 CLI 보고 결과를 확인함. Terra gstack 리뷰 후 Manager가 diff와 테스트 영수증을 독립 검토함. 해당 change validate/PM 검사 및 archive 절차를 수행하되 전체 gate가 실행됐다고 주장하지 않음. 신규 capability semantic-input-profile-comparison을 동기화하여 아직 활성인 a128 capability와 충돌하지 않게 함.
