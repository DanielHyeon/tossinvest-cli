# 계좌 결속·엔진 인터록 읽기 전용 감사 — 2026-10-21

## 범위

실주문, 서비스 재시작, 설정 변경, attestation 재발급, 토큰 갱신은 실행하지 않았다. 읽기 전용 `GET /api/v1/accounts`를 최초 확인과 보완 확인으로 총 2회 수행했다. 응답 본문·토큰·계좌 식별자는 저장하거나 출력하지 않았다.

## 경로 정정

이전 기본 경로 판정은 이 엔진 프로필에 적용되지 않는다. `internal/app/engine/engine.go:365,375`는 `opts.ConfigDir`를 `apppaths.OpenAPI`에 넘기고, `internal/app/paths/paths.go:48-60`은 명시된 config directory에서 credentials와 token cache를 모두 찾는다. 실제 engine container는 그 config directory를 사용한다. 따라서 이번 token cache는 `~/.config/tossctl/openapi-token.json`이다. 기본 `DefaultPaths`의 cache directory 경로는 이 실행에 사용하지 않았다.

## 단발 조회 결과

- 시각: `2026-10-20T18:29:19.138262Z`
- 대상: source 상수 `https://openapi.tossinvest.com`의 `GET /api/v1/accounts`
- 요청 수: 1
- HTTP 상태: 200
- 소요: 279ms
- TLS 검증: 유지
- redirect follow: 없음
- refresh, retry, cache write: 모두 없음
- token cache SHA-256 전후: 동일
- 현재 artifact account binding과 **반환 목록 중 하나**의 정규화 비교: 일치
- artifact가 `internal/app/engine/interlock.go:219-228`의 필수 endpoint 8개를 덮는지: 일치

첫 조회는 목록 내 일치 항목만 확인했다. 보완 조회는 첫 record만 검사했다. 두 조회의 누적 요청 수는 2다. 보완 조회에서는 result count 1, first account binding 일치, first accountSeq 양수, attestation endpoint coverage 일치, HTTP 200을 확인했다. engine process memory의 SelectedAccountSeq는 읽지 않았으므로 그 비교는 `unverified_no_engine_memory`로 유지한다.

기계 판독 결과는 응답 본문 없이 다음 두 파일에 보존했다.

- `/tmp/tossos-verify-readonly-audit-20261021/accounts-readonly-result.json` — SHA-256 `318ad9d2b8720ed7bd8327447c0f392e80e0ffb1aec0119ec51383dfb10175bd`
- `/tmp/tossos-verify-readonly-audit-20261021/accounts-readonly-first-result.json` — SHA-256 `036c98ddb05782a37c450d04adeb647304ccd1eb74af571fd51515bd82de570e`

두 번째 결과는 token/attestation expiry 메타데이터와 SHA, 캐시 전후 동일성도 포함한다.

## 확인한 계약과 한계

`internal/attest/attest.go:211-255`는 만료, 정규화한 account binding, endpoint coverage를 확인한다. 계좌 비교는 ASCII 숫자만 추출해 형식 차이를 허용한다. 최초 조회는 목록 내 일치만 확인했고, 보완 조회는 **현재 cache token으로 반환된 첫 계좌가 artifact와 일치하며 첫 accountSeq가 양수임**을 확인했다. `resolveAccountRef`는 추가로 client의 SelectedAccountSeq와 첫 accountSeq의 일치를 검사한다. 실행 중 engine의 메모리 상태는 조회하지 않았으므로 그 일치와 실제 기동 통과는 증명하지 않는다.

확인하지 못한 항목:

- cache token이 현재 credentials에서 발급됐는지
- 실행 중 engine이 같은 token을 사용했는지
- 현재 실행 세대가 startup interlock를 통과했는지
- live account/engine의 최신성 또는 주문 권한

컨테이너 시작 시각 이후 제한된 stdout을 읽었지만 interlock 이벤트는 발견하지 못했다. 이는 해당 로그 범위의 부재일 뿐 guard 통과·실패 증거가 아니다.

## 가짜 인터록 회귀

실제 매칭을 `-v -count=1`로 확인한 fake-only tests:

- `TestGateOffStartsAndDoesNoGateWork`
- `TestGateOnWithEverythingInPlaceStarts`
- `TestGateDecisionIsLoggedStructurally/{verified,refused,gate_off}`
- `TestGateOnRefusals/{no_attestation_at_all,attestation_expired,attestation_is_for_another_account,attestation_misses_an_endpoint_the_engine_calls,attestation_predates_the_price_read}`

모두 PASS였다. 기존 비verbose raw receipts는 `/tmp/tossos-verify-readonly-audit-20261021/attest-tests.log` 및 `engine-interlock-tests.log`에 보존돼 있다. 완료된 verbose 실행 출력에서 옮긴 발췌본은 같은 폴더의 `engine-interlock-tests-verbose-transcript.log`다. 이 테스트들은 fake server·임시 경로 계약검증이며 실제 기동이나 주문 검증이 아니다. 전체 테스트는 실행하지 않았다.

## 남은 실측 및 순서

| 순서 | 남은 승인·실행 | 격리 조건 |
|---|---|---|
| 1 | 실제 engine startup generation의 structured interlock decision과 그 generation의 resolved first account를 비식별 방식으로 보존·대조 | 재시작은 운영 승인 후, 주문 기능이 비활성인 격리 조건에서만. task 1.4는 미완료 유지. |
| 2 | 기존 거래 영수증을 읽기 전용으로 수집해 전량·비용·거절 상태를 대조 | 증거가 부족하면 최소 주문별 계획과 별도 사용자 승인을 먼저 받는다. |
| 3 | conditional trigger의 market, symbol, quantity, loss threshold, session window, cleanup을 승인해 최소 검증을 수행 | TTL 경계는 선택 사항이며 trigger 선행조건이 아니다. |
| 4 | 담보·미결제·OCO·잔량·동시성의 sellable semantics를 별도 검증 | 조건주문/일반 매도 충돌을 포함한 별도 최소 수량 계획이 필요하다. |
| 5 | 정책 승인 뒤 범위가 정해진 검증을 수행 | 전체 gate는 별도 명시 승인 없이는 실행하지 않는다. |

주문 권한, 주문 제출, 실거래 결과는 이 감사 범위 밖이다.

## Manager 최종 검토

소스 경로 resolver, 첫 계좌 선택 계약, 계좌 정규화, 두 조회의 비식별 결과 및 관련 테스트 기록을 대조했다. 기본 캐시 경로 오판과 최초 any-account 비교의 증명 한계를 지적해 보완했다. 보완 조회 시각은 2026-10-21 03:31:08 KST이며, 이후의 계좌·토큰 상태를 보장하지 않는다. task 1.4는 실제 기동 증거가 없어 미완료로 유지한다. 이번 작업은 문서와 임시 검증 기록만 추가하며 커밋·아카이브하지 않는다.

스킬 적용성: research-first는 기존 경로 해석·계좌 조회·attestation 계약을 재사용했다. 신규 기능이나 명세 변경이 없어 grill-me 범위 재결정은 불필요하다. impeccable: N/A — UI 변경 없는 읽기 전용 검증·문서 작업.
