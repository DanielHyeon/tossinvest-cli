## Context

우선순위 5/9: Volume Profile 데이터 적합성 평가. 선행 change: a149.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** LVN/HVN을 계산하기 전에 데이터가 가격대별 실제 체결량을 재현할 수 있는지 확인해야 한다. 5분 OHLCV 분배 근사로 진짜 Volume Profile이라고 주장하지 않는다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

기존에 제공된 파일만 조사하며 신규 API 수집·구독·운영 예약은 하지 않는다. 분류는 exact_trades, documented_price_bins, ohlcv_proxy, unavailable 네 가지다. trade 자료는 가격/거래량/체결시각/수신시각/정정·취소 식별자, 세션 coverage를 요구한다. aggregate bin 자료는 bin 폭·경계·작성 방법·갱신시각을 요구한다. 호가 잔량은 체결량으로 대체하지 않는다.
산출물은 coverage·중복/정정 처리·틱/bin 단위·as-of 가능 여부·비용/접근 제약·지원 판정표다. 지원 가능한 경우 전일 고정 profile과 당일 prefix profile의 입력 계약을 각각 설계한다. LVN cutoff/bins의 최적화나 생산용 계산 엔진은 이 change에서 구현하지 않는다.
데이터 부재도 적합성 평가의 정상 완료 결과다. unsupported는 기능 승인이 아니며 a153에서 VP arm 제외 이유로 이어진다. 자료 확보와 계산 프로파일 구현이 필요하면 별도 범위를 승인받는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Offline input schema/quality report; immutable provided trade/price-bin data; existing LVN feature input boundary. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: Read-only offline data feasibility. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

사용자가 a152의 로컬 자료 검증과 완료 절차를 승인했다. 새로운 자료 수집이나 공급자 선택은 이번 범위가 아니다. 확인한 자료에는 세션 전체 tape/bin 증거가 없으므로 실제 자료 적합성은 unsupported로 기록한다. 아래 계약으로 구현하며 추가 사용자 결정은 없다.

## Manager contract: a152-v1

### Research-first and scope decisions

2026-10-18 기준 base `0f95adce191ca4c82a4abd2f29aa1da02a5c2c42`. Manager 직접 설계, Terra 읽기 전용 조사, 별도 독립 검토를 수행했다. `search-first`, `grill-me` 및 `grilling`을 적용했다. 결정 트리는 기존 계산기 재사용 여부 → 입력 의미 → 시간/coverage → 정정 지원 범위 → 판정 권위 순서다. 기존 승인으로 범위는 확정되어 재승인을 요구하지 않는다.

- 저장소 검색: a147 MFI와 a151 손절 비교는 OHLCV 소비 도구이며 profile qualifier가 아니다. 공식 Trade DTO는 price/volume/timestamp/currency의 최근 최대 50건이며 세션 coverage와 정정 ID 계약이 없다. API 호출은 하지 않았다. a128 자료를 수정하거나 신규 변환하지 않는다.
- 선택: 독립 `tools/a152-qualify-volume-profile-evidence/` Python stdlib leaf를 최소 구현. a151의 plain-decimal/중복키 거절/입력 hash/exclusive-output 패턴을 참고하되 a151 파일 수정이나 동적 import 결합은 하지 않는다. 작은 입력 검증 정책은 새 의존성보다 stdlib가 적합하다. 기존 저장소 라이선스 범위 내 코드이며 외부 코드는 추가하지 않는다.
- 도구: 로컬 rg/git/Python 사용 가능. 외부 패키지 registry/GitHub/MCP 검색은 필요하지 않아 생략했으며 검색 완료로 주장하지 않는다. 새 패키지·MCP·실행 프레임워크를 설치하지 않는다.
- `impeccable: N/A — backend/documentation-only`: 오프라인 JSON CLI, UI 변경 없음. DB migration/high-risk FLM: N/A, production 함수 수정 없음. 기존 함수 동작을 근거로 상세 주장할 때는 해당 AST 근거를 별도로 보존한다.
- 기존 r1/a128/a147, frozen lane, live/ops 유지. 소스/정책 활성화·인증·거래·수집·예약·재시작은 없다.

### Minimal bounded input

단일 로컬 JSON envelope: `version` = `a152-v1`, `synthetic` = boolean, `datasets` = 최대 100개. 입력 최대 16 MiB, 전체 records 최대 100,000개, 문자열 최대 512자, 숫자는 최대 64자 unsigned plain decimal string. bool-as-number, exponent, NaN/Inf, 중복 JSON key, 미지/누락 필드와 중복 dataset ID를 거절한다. CLI 경로 오류/잘못된 envelope는 비정상 종료. 개별 dataset 오류는 status `invalid`와 안정적인 reason으로 보고하여 다른 dataset은 계속 평가한다.

모든 dataset 필드:

`id, kind, market, symbol, currency, basis, profile_mode, session, cutoff_at, as_of, source, coverage, correction_policy, price_increment, records, bin_method, access`.

- `kind`: `trades`, `bins`, `ohlcv`, `orderbook`, `unavailable`. 분류 결과는 각각 `exact_trades`, `documented_price_bins`, `ohlcv_proxy`, `unavailable`, `unavailable`; book-only 이유는 별도로 보존한다.
- `market/currency`: KR/KRW 또는 US/USD. basis는 `unadjusted`만 지원. 시장 간 합산 없음.
- `session`: `date, timezone, open_at, close_at`. 명시적 offset과 KR Asia/Seoul, US America/New_York IANA offset/date를 검증한다. 같은 현지 날짜에서 open < close, 길이 최대 24시간. 거래소 정규장/휴장·전일 거래일 정합성은 인증하지 않는다. 임의 세션을 공식 거래소 일정으로 표시하지 않는다.
- `profile_mode`: `current_prefix` 또는 `prior_session_fixed`. offset-bearing `cutoff_at`는 데이터 구간 끝, `as_of`는 정보 이용 시점이며 서로 다르다. prefix는 open < cutoff_at <= as_of <= close; 요청 구간은 [open, cutoff_at). prior는 cutoff_at=close <= as_of이고 session.date < as_of의 시장 현지 날짜; 요청 구간은 [open, close). prior는 제공된 과거 세션이며 직전 거래일임을 인증하지 않는다.
- `source`: `name, version, reference, declared_kind`. declared_kind는 `executions`, `executed_price_bins`, `ohlcv`, `orderbook`, `unknown`. kind와 출처 선언이 맞지 않으면 unsupported. 출처 선언은 검증되지 않은 사용자 입력이다.
- `coverage`: `start_at, end_at, status, gaps, sequence_status, venues, trade_types`. status `complete|partial|unknown`, sequence_status `complete|gap|unknown`, gaps는 최대 100개 `{start_at,end_at}`. gap은 양의 길이이며 요청 세션 안에 있어야 한다. venues/trade_types는 비어 있지 않은 문자열 목록(각 최대 32개). complete 자격은 선언 start/end가 요청 구간과 정확히 같고 status/sequence_status complete, gaps가 빈 경우만. 일부 구간은 not_ready. 첫/마지막 체결 시각이나 건수로 coverage를 추론하지 않는다. coverage는 항상 `caller_asserted_unverified`로 표시한다.
- `correction_policy`: `none_declared|unresolved|unknown`. 이 최소 버전은 정정 replay를 구현하지 않는다. trades/bins의 none_declared 외 정책은 unsupported; 수정·취소 event도 unsupported로 거절하고 합산하지 않는다. none_declared는 정정이 없다는 증명이 아니다.
- `price_increment`: 양의 plain decimal; 시장 규칙으로 인증하지 않는 공급자 선언 grid. trades 가격과 bin 경계는 zero-origin grid에 정렬되어야 한다.
- `access`: `license, availability, cost_note`; availability `provided|restricted|unknown`. 세 필드를 보고서에 보존하되 라이선스 권한을 인증하지 않는다. provided 외는 unsupported.
- `bin_method`: trades/proxy/unavailable은 null. bins는 `{name, volume_basis, unit, boundary, width, price_min, price_max, zero_bins_explicit, built_at, received_at}`. volume_basis `executed`, unit `shares`, boundary `[lower,upper)`, 양의 width/grid 정렬, zero_bins_explicit=true를 요구한다. 다른 방식은 unsupported. price_min > 0, price_max > min. 요청 구간 종료 <= built_at <= received_at <= as_of, 따라서 장중 최종 profile을 과거 prefix로 사용할 수 없다.

### Records and corrections

모든 timestamp는 RFC3339 부분집합 `YYYY-MM-DDTHH:MM:SS[.1..6digits](Z|±HH:MM)`만 허용한다. 마이크로초보다 높은 정밀도, 소수 offset, offset 초, 생략된 초/다른 separator는 절삭/정규화하지 않고 invalid로 거절한다. Python datetime의 관대한 parser가 미래 1ns를 cutoff와 같게 만들지 않도록 parser 진입 전에 문법을 검사한다. 각 필드의 동일한 문법 적용과 +1ns 경계 반례를 검증한다.

- trades: `{event_id, action, reference_id, executed_at, received_at, price, volume}`. event_id는 dataset 내 유일, action `trade|correct|cancel`. trade reference_id는 null, correct/cancel은 nonempty 원본 참조. 이 버전은 correct/cancel이 하나라도 있으면 unsupported; 중복 ID를 임의 제거하거나 동일시하지 않는다. 모든 timestamp는 offset-bearing, 요청 구간 start <= executed_at < end, executed_at <= received_at <= as_of. received_at 비감소 순서, 체결 시간 역전은 지연 수신일 수 있으므로 금지하지 않는다. price/volume 양수, 가격 grid 확인. event ID의 진위/전역 유일성은 검증하지 않는다.
- bins: `{lower, upper, volume}`. 경계 양수, upper-lower=width, volume>=0, 정렬되어 price_min부터 price_max까지 끊김/겹침 없는 모든 bin(0 거래량 포함)을 제공해야 한다. 가격대 밖 체결의 부재는 공급자 선언에 의존한다. 개별 체결 복원 가능하다고 주장하지 않는다.
- ohlcv/orderbook/unavailable: records는 빈 배열, bin_method null. 이 qualifier는 원본 proxy payload를 재해석하지 않고 출처 메타데이터만 분류한다. OHLCV는 unsupported `ohlcv_is_not_execution_profile`, book은 unsupported `book_size_is_not_executed_volume`, unavailable은 unsupported `no_profile_evidence`. proxy 입력을 trade/bin으로 변환하지 않는다.
- 빈 trades/bins는 not_ready `empty_evidence`. zero-volume bins만 있으면 not_ready `no_executed_volume`. 자료 없음과 실제 무거래 인증을 구분한다.

### Output and authority boundary

분류와 지원 상태를 별도 필드로 출력한다: classification, status `eligible_unverified|unsupported|not_ready|invalid`, reasons, mode/requested interval, rows, supplied metadata, exact total_volume(가능한 경우). 부적합 판정에서 LVN/HVN 숫자나 부분 합산을 만들지 않는다. eligible_unverified는 제공 자료의 구조적 적합성만 의미한다.

항상 input SHA-256, synthetic, provenance=`caller-supplied, unverified`, authenticity_verified=false, coverage_verified=false, profitability=`inconclusive`, production_authorized=false를 출력한다. downstream은 `vp_input_eligible`와 `vp_exclusion_reasons`만 제공한다. synthetic=true이면 구조 적격이어도 실제 실증용 vp_input_eligible=false, 이유 synthetic_only. non_profile_research_blocked=false는 항상 유지한다. actual profile 생성/전략 연결/다른 arm 실행 여부 결정은 범위 밖이다. 라이선스, reference, coverage metadata는 주장으로 표시하며 원시 event/계좌 값 전체를 오류에 출력하지 않는다.

CLI는 입력 1개를 읽고 stdout JSON 또는 `--output`의 새로운 파일에 쓴다. 기존 a151 방식처럼 exclusive/no-follow 및 조상 symlink 방지로 덮어쓰기를 거절한다. 입력·기존 증거·운영 파일을 수정하지 않는다. 임의 input 코드 실행·subprocess·network 없음. 출력 경로 생성은 명시된 단일 파일뿐이며 crash-atomic durability를 주장하지 않는다.

### Verification and actual availability

합성 KR/US trades와 bins, OHLCV/book/unavailable, 미래 수신/bin, 과거 세션과 prefix, US DST offset, 중복 ID/key, 정정/취소/unknown policy, coverage gap/partial, 시간 범위, grid/겹침/누락/zero bins, 선언 불일치, oversized/numeric 입력, 독립 dataset 오류 격리, synthetic 제외, 입력 hash 불변 및 출력 덮어쓰기/심볼릭 링크 거절을 검사한다. 수량 합산 oracle은 2+3=5 shares, bin 2+0+3=5 shares처럼 독립 기대값을 사용한다. 정정은 재구성 결과가 아니라 unsupported가 기대 결과다.

별도 실제 가용성 보고서에는 확인한 소스/fixture 경로와 한계만 기록한다. 전체 사용자 디스크를 조사했다고 주장하지 않는다. 현재 확인한 적합 실증 tape/bin 없음 → unsupported; 합성 성공으로 이를 승격하지 않는다. a153에 데이터 부재 및 VP arm 제외 이유를 전달할 수 있으나 a153 구현·측정은 수행하지 않는다. Python 시험은 별도 명시 실행하고 전체 gate 포함 여부를 사실대로 기록한다.
