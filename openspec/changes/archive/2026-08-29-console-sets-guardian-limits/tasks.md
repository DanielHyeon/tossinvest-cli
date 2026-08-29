# Tasks: console-sets-guardian-limits

위험도 **High-risk** — Guardian 한도는 §0.5가 열거한 High-risk 경로이고, 이 change는
콘솔에 그 편집 표면을 연다. 1.x는 전부 RED 선행이다.

## 1. 계약 확인 (구현 전)

- [x] 1.1 §0.1 확인 — 기본값 설치가 게이트 ON 기동을 열지 않음을 인터록 6조항으로
  확인하고 근거를 design §0.1과 FLM에 적는다 (attestation·ProtectionReady 잔존)
- [x] 1.2 §0.2 확인 — 한도가 진입 경로 상한이고 청산(RISK_REDUCING)은 한도 적용
  대상이 아님을 코드·스펙으로 확인 (§0.4 무접촉)
- [x] 1.3 Function Logic Map + Branch Test Map — 기존 함수 내부를 고치는 전 대상
- [x] 1.4 Pre-Edit 선언 기록 (review.md)

## 2. 티어 레지스트리와 상한 (internal/config)

- [x] 2.1 RED — 네 티어가 StockOS `risk_profiles.py` 값과 일치하고, 기본 티어가
  `risk-management` 승인 다섯 숫자와 정확히 같다
- [x] 2.2 RED — 통화별 상한이 그 통화 등록 티어의 필드별 최대이고, 미등록 통화는
  fail-closed 오류다
- [x] 2.3 RED — 상한 초과 값은 필드명과 상한을 담은 위반 목록을 돌려준다
- [x] 2.4 GREEN — `GuardianTier` 레지스트리 + `GuardianCeiling` + 위반 검사.
  각 수치에 StockOS 출처 주석

## 3. 암묵 기본값 부재의 고정 (internal/config)

- [x] 3.1 RED — 파싱은 기본값을 주입하지 않는다: 한도 없는 게이트는 로드 후에도
  `LimitsSet() == false`다 (D1 — 기존 `TestAutomationGateDefaultsOff`가 이미
  주장하는 바를 이 change의 이름으로 한 번 더 고정한다)
- [x] 3.2 RED — 현재 값이 등록 티어와 정확히 일치하면 그 티어 ID를, 아니면 빈
  문자열을 돌려주는 대조 함수 (D12)
- [x] 3.3 GREEN — 대조 함수. `mergeEngine`은 **수정하지 않는다**

## 4. 한도 전용 writer (internal/config)

- [x] 4.1 RED — 저장이 `enabled: true`를 `false`로도, 그 반대로도 바꾸지 않는다.
  파일의 `enabled` 바이트가 그대로다 (D6)
- [x] 4.2 RED — `attestation_file`과 블록 밖 미지 키가 바이트 단위로 보존된다
- [x] 4.3 RED — 인터록이 거부할 블록(0·음수·NaN·Inf·비율>1·통화 없음)은 저장 거부
- [x] 4.4 RED — 상한 초과는 저장 거부
- [x] 4.5 RED — `automation_gate` 키가 없는 파일, `engine` 키가 없는 파일, 파일
  자체가 없는 경우 각각 올바르게 삽입/생성된다
- [x] 4.6 RED — 유효하지 않은 JSON 파일은 덮어쓰지 않는다
- [x] 4.7 GREEN — `LoadRawEngineGate` + `SaveEngineGateLimits` (키 단위 splice)

## 5. 콘솔 편집 표면

- [x] 5.1 RED — `/settings/limits/preset`·`/settings/limits`가 세션+CSRF 게이트 뒤에
  있고, CSRF 없는 요청은 거부되며 config가 바뀌지 않는다
- [x] 5.2 RED — 프리셋 적용이 다섯 한도와 통화를 한 번에 쓰고 `enabled`를 안 건드린다
- [x] 5.3 RED — 프리셋 카드는 확인 대화상자 1회 외에 입력을 요구하지 않는다
  (타이핑 확인 부재 — 사용자 결정 2026-07-27)
- [x] 5.4 RED — 개별 기입은 고급 접힘 안에 있다
- [x] 5.5 RED — 통화가 바뀌는 적용의 응답이 닫히는 시장을 명시한다 (D8)
- [x] 5.6 RED — 응답이 현재형 보장을 쓰지 않고 `effectNotice`에 위임한다 (D9)
- [x] 5.7 RED — seam 미배선이면 501, 상한 초과·무효 값은 사유와 함께 거부
- [x] 5.8 GREEN — seam·핸들러·라우트·템플릿

## 6. 화면의 정직성

- [x] 6.1 RED — 한도가 파일에 없으면 화면이 미설정이라고 말하고 기본 티어 숫자를
  적용된 것처럼 그리지 않는다 (D1)
- [x] 6.2 RED — 부분 설정 상태의 안내가 인터록 거부와 프리셋 복구를 말한다
- [x] 6.3 RED — 현재 값이 티어와 일치하면 티어 이름, 아니면 "사용자 지정값"
- [x] 6.4 GREEN — 배선. `internal/console/overview.go`의 `limitReading`과 개요
  패널은 **수정하지 않는다**

## 7. 정적 가드 (스펙이 요구하는 동반 갱신)

- [x] 7.1 RED — 라우트 수 하한·열거 주석·`stateChanging`·`consoleStateChanging`이
  두 새 라우트를 알고 있다. 하나라도 빠지면 실패
- [x] 7.2 RED — `actVerbs`가 `limit`·`preset`을 안다 (미승인 `/settings/limits/*`가
  조용히 통과하지 못한다 — A3 선례)
- [x] 7.3 RED — 라우트 이름에 게이트 어휘가 없다 (`routeOnlyAccountVerbs` 유지)
- [x] 7.4 RED — 새 seam이 정확히 2메서드다
- [x] 7.5 GREEN — 가드 갱신

## 8. audit 배선

- [x] 8.1 RED — 저장이 필드별 전후 값을 audit에 남긴다
- [x] 8.2 GREEN — `cmd/tossctl` seam에 audit append (console 패키지는 쓰지 않는다)

## 9. 검증

- [x] 9.1 `make test` / `make vet` / `make validate` — 상속 테스트 회귀 0
- [x] 9.2 review.md — 적대적 Eng 리뷰
- [x] 9.3 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=console-sets-guardian-limits`
- [x] 9.4 PM 등록
