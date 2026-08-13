# Branch Test Map: `Stabilisation.withDefaults`

Source: `internal/reconcile/recovery.go` (86-109). AST 기준 branches **5** / returns 1.

## 커버리지는 주장이 아니라 측정값이다

`go test ./internal/reconcile/ -count=1 -coverprofile`을 편집 전(`:75-86`, 130건 통과)과
편집 후(`:86-109`, **147건 통과** · 86.6%)에 각각 돌려 블록 카운트를 잘라 읽었다.

| Branch | 위치 | 조건 평가 | 본문 실행 | 근거 블록 (편집 후) | 편집 전 |
|---|---|---|---|---|---|
| B1 | `:87` `Interval <= 0` | yes | **no** | `87.21,89.3` count=**0** | count=0 |
| B2 | `:90` `Required <= 0` | yes | **no** | `90.21,92.3` count=**0** | count=0 |
| B3 | `:93` `MaxAttempts <= 0` | yes | **no** | `93.24,95.3` count=**0** | count=0 |
| B4 | `:102` `RateLimitBackoff < 15s` | yes | **yes** | `102.50,104.3` count=**1** | (없던 분기) |
| B5 | `:105` `MaxRateLimitWait <= 0` | yes | **yes** | `105.29,107.3` count=**1** | (없던 분기) |

**a102가 더한 두 분기는 본문까지 실행된다.** 물려받은 세 분기는 편집 전과 똑같이 죽어
있다 — 이 패키지의 모든 recovery 테스트가 `recoveryOptions`(`recovery_test.go:130`)와
`a102Stabilisation`(`a102_recovery_rate_limit_test.go:126`)에서 셋을 **명시로** 넣기 때문이다.

> ⚠ **B1·B2·B3의 공백은 a102가 만든 것이 아니라 물려받은 것이다.** a102는 자기가 더한
> 칸만 메운다. 기존 셋의 기본값을 새로 고정하는 것은 이 change의 범위가 아니다 — 그것은
> §1이 손대지 않기로 한 안정화 판정의 의미까지 테스트로 굳히는 일이 된다.
> **침묵하지 않고 이름을 붙여 남긴다.**

| 분기 (재인용) | 무엇을 지는가 | 지는 테스트 |
|---|---|---|
| B4 `:102` | zero → 15s이고 그 값이 **서베이의 첫 백오프와 같다** | `TestRateLimitDefaultsMatchTheSurveyDiscipline` |
| B4 `:102` (하한) | 0보다 크지만 15s보다 짧은 노브도 15s로 올라간다 | `TestRateLimitBackoffNeverGoesBelowTheSurveyInterval` (§1.9 F6) |
| B5 `:105` | zero → 5m | `TestRateLimitDefaultsMatchTheSurveyDiscipline` |

그 테스트는 상수 비교(`soak.ReadRetryBackoff(0)`)와 **배선 실측**(zero 값으로 만든 복구가
실제로 15초를 요청하는지)을 둘 다 본다. 상수만 맞고 배선이 없으면 소용없기 때문이다.

§1.9의 하한 테스트는 5초짜리 노브로 만든 복구가 **실제로 15초를 요청하는지**를 잰다.
RED 실측(구현 원복 상태): `RateLimitWaited = 5s, want the floor 15s` ·
`waits = [5s 2s], want the first one raised to 15s`.

## 뮤테이션 정산

이 번들이 지는 뮤테이션은 (b)다 — `ratelimit.go`의 예산 검사를 지우면
`TestRateLimitWaitBudgetExhaustionFailsClosed`가 죽는다(err = nil이 되어 복구가 완주한다).
자세한 표는 `internal-reconcile--recovery.stablesnapshot/branch-test-map.md`에 있다.

B4·B5 자체의 반증은 (a)의 부수 효과로도 관측됐다 — 뮤테이션 (a) 아래에서
`TestRateLimitDefaultsMatchTheSurveyDiscipline`이 함께 죽었다.

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 5, returns 1) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/reconcile/ -count=1 -coverprofile` exit 0 · **147건 통과** · 86.6% (§1.9)
- 호출자 전수: `rg -n 'withDefaults' internal/reconcile/` → `recovery.go:86`(선언) · `:176`(유일 호출)
