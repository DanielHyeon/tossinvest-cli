# U2 순수 거래량 점수 로트

2026-10-31 Manager 제한 freeze READY. 기준 `2d67c226e`. 전체 a169 freeze·측정은 보류한다. 사용자 승인된 남은 개발 중 확정된 Q3 산식만 구현한다. U0/U1 로트와 다른 새 파일 `tools/a140-breakout-signal-shadow/a169_u2_score.go` 및 대응 시험만 소유하며 기존 함수·caller·CLI·HTTP·ledger·report는 수정하지 않는다.

## 확정 범위와 지원 제한

하나의 후보에 대한 V와 일봉을 받아 S20·ratio_ppm·increase_ppm·원본 수치 진단을 돌려주는 비공개 순수 함수다. `domain.Candle`과 stdlib math/big·sort·time을 재사용한다. 후보 선정/합집합/유형/랭킹 receipt 정합화/순위 tie/Top N은 범위 밖이다. 거래량 랭킹과 거래대금 랭킹의 공통 필터 결과가 다를 때 canonical 후보 선택은 아직 미정이므로 이 함수가 대신 고르지 않는다. 호출자가 전달한 V의 출처를 인증하지 않는다.

입력: V float64, 일봉 slice(최대25), KST 세션 날짜 자정, 공식 정규장 open/close, 실제 selected_at. 시계 호출·네트워크·파일 읽기 없음. 날짜 입력은 UTC+9 기준 정확 자정이어야 한다. open/close는 같은 KST 날짜이며 close>open, open<selected_at≤close여야 한다. 다른 날짜·0시각·역순·장 밖은 결과 전체 오류. 실제 U2의 selected_at=max(두 ranking receipt 수신시각), target=open+30분30초(D1), ready deadline은 caller 후속 계약이다. 수학적으로 close와 같은 selected_at을 허용하는 것이 U2 마감을 통과한다는 뜻은 아니다.

V와 사용되는 일봉 Volume은 유한·비음수·정수이고 ≤2^53−1만 지원한다. 이는 float64 원문 손실을 줄이는 지원 제한이며 API 수치의 정확성을 증명하지 않는다. 현재 타입의 누락/파싱오류와0 합침 한계를 유지한다. V==0은 유효하고 진단에 남긴다. V 오류 `numerator_invalid`.

일봉: Time=0은 기존 계약대로 직전 봉에서 제외한다. KST 세션 자정보다 작은 봉만 복사·정렬하여 최신20개를 취한다. 오늘/미래 봉은 분모에 넣지 않는다. 가장 늦은 직전 봉 시각 및 자정 이상에서 가장 이른 next_bar_at을 반환한다(없으면 zero). 과거 봉에 같은 KST 날짜가 둘 이상 있으면 `baseline_invalid`로 거부한다. 조용히 합산/중복제거/최신선택하지 않는다. 이는 broker 중복 여부가 확인된 사실이 아니라 보수적 지원 제한이다. 실제 거래일 달력·표기 검증은 caller/리허설 몫이다.

직전 봉20개 미만은 `history_lt_20`; 선택된20개의 Volume 유효성 오류는 `baseline_invalid`; S20=0은 `baseline_zero`. 최신20개 밖의 과거 Volume은 분모에 쓰지 않으므로 검사하지 않는다. 중복 날짜 판정은 전체 prior 입력에 적용한다는 제한을 명시한다. 입력 slice와 봉 순서를 변경하지 않는다. 원장/사용자 출력이 아닌 typed 결과와 사유만 돌려준다.

산식은 두 단계 floor다. `ratio_ppm=floor(V*20*1000000/S20)`, `increase_ppm=floor(ratio_ppm*(close-open)/(selected_at-open))`. 모든 곱·합·나눗셈은 math/big 정수로 계산한다. 결과 필드는 int64(V,S20,ratio_ppm,increase_ppm), prior_days, zero_volume_days, V==0, 경계 시각이다. int64 저장 범위를 넘는 결과는 `score_overflow` 명시 오류이며 wrap/clamp/float fallback하지 않는다. zero_volume_days는 선택된20개만 센다. 실패 결과를 정상 점수0으로 취급하지 않도록 반환 계약을 고정한다.

## 연구·검토·수용

source: a169 design §4.1–4.3, §13 D1, `domain.Candle`, a140 sizing.go big.Int 선례. 조사 `/tmp/a169-u2-contract-research.md`를 대조했다. repo 검색과 설치된 Go stdlib만으로 충분해 외부 registry/web/MCP 패키지 조사는 생략, 신규 의존성/라이선스 없음. grill-me/grilling 검토에서 확정 산식과 미확정 canonical 선택을 분리했으며 기존 정책을 재질문하지 않는다. impeccable N/A — backend 계산·시험만.

신규 함수만 작성하므로 기존 함수 FLM/BTM 대상 없음. 구현 전 영향/RED 계획 기록. literal expected fixtures와 RED/GREEN, 정확 선택 normal/tagged race, 유효 source-copy 변이, scoped gstack 수동 checklist 및 독립 검토. 전체 tests/gate·계좌·브로커·운영 없음. 기존 source caller0은 AST/한정 검색으로 확인하며 실제 전체 U2/기본 a140 digest를 검증했다고 확대하지 않는다.

필수 경계: 전일·오늘 KST 자정; unordered slice·input 불변;0Time;19/20/21~25봉; 중복 날짜; 미래 봉 제외; V0; 선택20일 중0; S20=0; NaN/Inf/음수/소수/2^53 경계; int64 score overflow; open/close/selected 오류; 두 floor가 한 번의 합성 나눗셈과 다른 literal 사례. 변이는 자정 경계→open, today 포함,20→19, floor 단계 생략, 정수/overflow guard 제거 중 의미 있는 것을 선택하며 동등변이는 검출로 세지 않는다.

독립 설계 검토 뒤 Manager 제한 READY가 있어야 구현한다. 개발 완료 후 Manager 승인 전 push 금지. a1692.2 전체·freeze1.3·측정·archive는 체크하지 않는다.

독립 검토: 새 P0/P1 없음, bounded READY. Manager는 이 순수 산식 로트만 승인한다. 오류 우선순위는 envelope(개수·시각)→numerator_invalid→prior 날짜 중복 baseline_invalid→history_lt_20→선택된 volume baseline_invalid→baseline_zero→score_overflow로 고정한다. int64 변환 전에 IsInt64 등으로 범위를 확인하는 경계 fixture를 둔다. 전체 a169 freeze는 그대로 미완이다.
