# U2 순수 후보 합집합·순위 로트

2026-10-31 Manager 제한 freeze READY. 선행 U0/U1 selection 및 U2 score 로트가 검증·게시된 뒤 구현한다. 전체 a169 freeze·CLI·측정·archive는 계속 보류한다. 파일 소유: 새 `a169_u2_union.go`와 대응 시험, 이 analysis 기록만. 기존 함수 본문은 수정하지 않는다.

## 정본 해석

design §4.1은 각 랭킹의 공통 필터를 통과한 종목의 합집합이다. 따라서 양쪽 중 한쪽만 통과해도 후보에 든다. 한쪽 필터 실패를 전체 후보 실패로 바꾸지 않는다. §4.2는 분자 V만 거래량 랭킹에 존재하면 그 raw TradingVolume을 우선하고 없을 때 거래대금 값을 사용한다. 이 우선 규칙에 필터 통과 조건은 없다. tie도 raw volume rank→raw amount rank→symbol 순이다. 독립 source 재대조 `/tmp/a169-u2-union-resolution.md`는 이 해석을 확인했고 이전 충돌 전체 오류 권고를 철회했다.

단일 `last_price`의 canonical 선택은 정해져 있지 않으므로 이번 순수 결과는 양쪽 원본과 각각의 필터 사유를 모두 보존한다. 한쪽 가격을 정본으로 선택하거나 실제 ledger schema를 바꾸지 않는다. 원장 연결 전 이 projection은 별도 설계 대상이다.

## 입력·출력과 계산

입력은 amount와 volume 각각의 ranking receipt/count/caution, cap, N20/30, KST session date/open/close, symbol별 제공된 daily slice다. 실제 HTTP 수행/adjusted 여부·일봉 출처는 caller 책임이며 이 함수가 인증하지 않는다. map에 symbol이 없는 경우 baseline_unavailable; 존재하고 빈 slice면 history_lt_20이다. 입력 자료를 읽기만 한다.

기존 순수 a169Select(U0,U1)를 재사용해 receipt 검증과 최대100개 전체 Rows를 얻는다. 이때 Symbols의 앞N 목록으로 U2 후보를 자르지 않는다. Rows에서 Reason이 빈 행의 합집합이 후보(최대200)다. 두 receipt의 원문과 전체 필터 사유를 결과에 보존한다. result가 입력 slice를 공유하지 않게 이미 제공된 복사 계약을 사용한다.

selected_at은 두 ranking FetchedAt의 max다. 두 수신 시각 모두 같은 KST 세션 날짜이며 open<received≤close, 날짜 자정·정규장 창은 U2 score와 같은 지원 조건으로 먼저 검증한다. 후보0개라도 잘못된 시각을 성공으로 처리하지 않는다. 이 값은 final ready/일봉 완료 시각이 아니며 U2 deadline·late 판정은 caller 후속이다.

후보마다 raw amount/volume item의 존재 여부와 값, 각각의 필터 사유, V 출처, score 또는 제외 사유를 반환한다. 원본은 값 복사로 저장한다. volume item이 존재하면 필터 실패 여부와 관계없이 그 TradingVolume을 사용하며 invalid numerator이면 후보를 제외한다. amount 값으로 조용히 fallback하지 않는다. 기존 a169 U2 score 함수를 호출해 분모·20일·정수·시각·overflow 지원 제한과 정확 floor를 재사용한다. 오류는 명시 후보 제외로 남기고 점수0 성공으로 바꾸지 않는다. envelope 오류는 전체 결과 오류다.

유효 score 후보 정렬: ratio_ppm 큰 순 → volume 존재 우선/Rank 작은 순 → amount 존재 우선/Rank 작은 순 → ASCII symbol 오름차순. 존재와 rank를 별도로 비교해 maxint sentinel 충돌을 만들지 않는다. increase_ppm은 정렬 키가 아니다. V0·ratio0도 정상 score면 후보에 남는다. 앞N만 Selected=true, Symbols는 그 순서다. n<N은 Short=true이며 empty도 정상 short다. 전체 후보/제외 목록은 symbol 순 등 고정 순서로 반환해 map iteration에 따라 결과가 바뀌지 않게 한다.

## 수용 증거와 범위

기존 함수 재사용·Go sort/map만으로 충분, 새 의존성·공개 API 없음. repo 연구는 design §3–4와 두 선행 순수 함수 및 위 독립 조사다. 외부 registry/web 패키지는 조사하지 않았다. grill-me/grilling의 입력 신뢰·오류·시간·지원 경계를 명시했으며 확정된 ANY-pass/V/tie 계약을 새 정책 승인으로 포장하지 않는다. impeccable N/A — backend 순수 계산만.

구현 전 영향·RED 계획 기록, RED/GREEN·literal 결과 fixture·선택 normal/tagged race·유효 변이·수동 scoped gstack, 독립 검토와 Manager 판정. 기존 함수 FLM/BTM 편집 없음. tests는 selection+score+union 정확 접두사와 기존 제한 policy fixture만, 전체 tests/gate 금지. no production caller는 AST/한정 검색으로 확인한다.

필수 반례: 각각31위 이상이 U2 상위가 됨; 한쪽 필터 실패라도 반대쪽 통과하면 후보; volume 필터 실패라도 rawV 우선; volume invalidV를 amount로 대체 금지; 두 rankedAt이 아닌 FetchedAt max; 한쪽 수신 장외/다른 날짜; 두 순위 tie 및 symbol tie; 20/30/short/empty; missing vs empty daily; score 오류 사유; input/result 독립; shuffled map insertion 동일 결과. 변이는 ANY→BOTH, Rows→Symbols 후보 축소, V fallback, tie 순서, increase로 정렬 등 유효 반례를 선택한다.

본 로트는 U2 순수 계산만 마친다. 실제 수신 provenance·deadline·cache·profile·shared event prewire fence·engine 신원·영속 abort·ledger/report·리허설은 미완이다. 정본 tasks2.2 전체/1.3/측정 완료를 체크하지 않는다. 독립 설계 검토 뒤 제한 READY 전 구현 금지.

독립 재검토: P0/P1 없음. Manager 제한 READY. 후보 오류 우선순위는 V 유효성(numerator_invalid)→daily key 부재(baseline_unavailable)→score가 정한 순서다. V검사는 기존 a169U2Volume을 재사용한다. invalidV와 daily누락이 겹치는 literal fixture를 추가한다. 전체 envelope가 잘못되면 후보별 오류보다 먼저 전체 오류다.
