# Branch Test Map: `Recovery.stableSnapshot`

Source: `internal/reconcile/recovery.go` (366-400). AST 기준 branches **7** / returns 5.

## 편집 전 — 무엇이 비어 있었나 (측정값)

`go test ./internal/reconcile/ -count=1 -coverprofile`(exit 0 · **130건 통과**)을
편집 전 범위 `recovery.go:333-359`로 잘라 읽은 것이다. 당시 분기는 5개였다.

| 편집 전 분기 | 조건 평가 | 본문 실행 | 근거 블록 |
|---|---|---|---|
| `:338` 루프 | yes | yes | `338.61,340.17` count=1 |
| `:340` `Collect` 오류 | yes | **no** | `340.17,342.4` count=**0** |
| `:344` `Stable` | yes | yes | `344.54,346.4` count=1 |
| `:347` 상한 도달 | yes | yes | `347.36,348.9` count=1 |
| `:350` `Sleep` 오류 | yes | **no** | `350.57,352.4` count=**0** |

**수집 오류 팔의 본문은 이 저장소의 어떤 테스트도 실행한 적이 없었다.** "429가 나면
복구가 죽는다"는 동작은 테스트로 고정된 적이 없고, 2026-08-13 02:03:30.545Z에 운영에서
처음 관측됐다. a102 §1의 RED가 정확히 그 칸이다.

## 편집 후 — 재측정 (측정값)

`go test ./internal/reconcile/ -count=1 -coverprofile`(exit 0 · **143건 통과**,
coverage 86.6%)을 `recovery.go:366-400`으로 자른 것이다.

| Branch | 위치 | 조건 평가 | 본문 실행 | 근거 블록 | 지는 테스트 |
|---|---|---|---|---|---|
| B1 | `:371` 루프 | yes | **yes** | `371.51,373.17` count=1 | 전부 |
| B2 | `:373` `Collect` 오류 | yes | **yes** | `373.17,374.48` count=1 | `TestNonRateLimitedCollectStillFailsImmediately` |
| B3 | `:374` 429가 **아님** | yes | **yes** | `374.48,376.5` count=1 | `TestNonRateLimitedCollectStillFailsImmediately` |
| B4 | `:377` 백오프 실패 | yes | **yes** | `377.68,379.5` count=1 | `TestRateLimitWaitBudgetExhaustionFailsClosed` · `TestRateLimitBackoffStopsOnContextCancel` |
| B5 | `:385` `Stable` | yes | **yes** | `385.54,387.4` count=1 | `TestRateLimitedCollectDoesNotEndRecovery` |
| B6 | `:388` 상한 도달 | yes | **yes** | `388.35,389.9` count=1 | `TestRecoveryFailsClosedWhenTheAccountWillNotSettle` (`recovery_test.go:302`) |
| B7 | `:391` `Sleep` 오류 | yes | **no** | `391.57,393.4` count=**0** | — (아래 참조) |

**편집 전에 비어 있던 두 칸 중 하나(수집 오류)가 채워졌다.** 429 경로(B3의 반대편)와
백오프 실패(B4)는 새로 생겼고 처음부터 테스트가 있다.

### ⚠ 남은 미덮임 하나 — B7 (`:391` 안정화 간격 대기의 실패)

편집 전에도 count=0이었고 지금도 count=0이다. **a102가 만든 공백이 아니라 물려받은
공백이다.** a102의 대기는 B4(429 백오프)이고 그쪽의 `ctx` 취소는
`TestRateLimitBackoffStopsOnContextCancel`이 진다. 안정화 간격 대기의 취소를 고정하는
것은 이 change의 범위가 아니다 — **조용히 넘기지 않고 이름을 붙여 남긴다.**

## 뮤테이션 정산 (통과는 실패시켜 본 뒤에만 증거다)

| 뮤테이션 | 실제 편집 | 죽은 테스트 | 원복 확인 |
|---|---|---|---|
| (a) 429가 attempt를 소모한다 | `:380-381`의 `continue` 앞에 `attempt++` | `TestRateLimitedCollectDoesNotEndRecovery` · `TestRateLimitDoesNotConsumeAStabilisationAttempt` · `TestRateLimitDefaultsMatchTheSurveyDiscipline` · `TestRateLimitWaitBudgetExhaustionFailsClosed` | `recovery.go` sha256 = `e0d5690f…` (GREEN과 동일) + `grep -n "Deliberately no attempt"` 1건 |
| (e) 모든 오류를 429로 본다 | `:374`의 조건에 `&& false` | `TestNonRateLimitedCollectStillFailsImmediately` | 같은 sha256 + `grep -c "errors.Is(err, official.ErrRateLimited)"` = 1 |
| (d) 백오프가 취소를 무시한다 | `ratelimit.go`의 `clk.Sleep(ctx, …)` → `context.Background()` | `TestRateLimitBackoffStopsOnContextCancel` (2.19s 만에 실패) | `ratelimit.go` sha256 = `d5146716…` + `grep -c "clk.Sleep(ctx, backoff)"` = 1 |

(b)·(c)는 각각 `ratelimit.go`·`snapshot.go`의 것이라 그 번들에 적었다.

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 7, returns 5) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/reconcile/ -count=1 -coverprofile` exit 0 · 143건 통과 · 86.6%
- 편집 전 기준선: 같은 명령, 130건 통과 (HEAD `03139000`)
