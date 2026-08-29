# Function Logic Map: `TestActualEngineRecoveryStillFailsClosedOnASnapshot429`

- Source: `cmd/tossctl/engine_account_seq_recovery_test.go` (lines 170-263)
- AST evidence: `ast.json` (17 branches)
- Risk scan: `risk-pattern-report.md`

이 함수는 테스트다. `go test`는 `_test.go`를 계측하지 않으므로 어떤 커버리지 프로파일도
이 안의 분기를 셀 수 없다. 따라서 아래 표의 근거는 커버리지 숫자가 아니라 **그 팔을 지나간 실행**이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `backoff` (:198) | `reconcile.DefaultRateLimitBackoff` = 15s | `internal/reconcile/ratelimit.go:51` | 상수이므로 실패 없음 |
| `budget` (:201) | `3*backoff + 5s` = 50s — **일부러 배수가 아니다** | 이 테스트가 명시 | 배수를 쓰면 예산과 소진액이 우연히 같아 혼동한 단언이 통과한다 |
| `wantWaits`/`wantReads` (:205-206) | `budget/backoff`, `+1` | 계약에서 유도 | 리터럴로 굳히면 상수 변경이 신호를 잃는다 |
| fake broker `/api/v1/orders` (:181-183) | 항상 429 | `httptest` 핸들러 | 실 API 접촉 0 |
| `clk` (:210) | 잠들지 않고 대기를 기록 | `a102Clock`(`cmd/tossctl/a102_engine_ready_test.go`) | runaway 가드가 상한 없는 뮤테이션을 오류로 끝낸다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 :174 | fake broker 경로 분기 | 없음(테스트 서버) | — | 자신 |
| B2 :175 | `/oauth2/token` | 토큰 응답 | — | 자신 |
| B3 :177 | `/api/v1/accounts` | `accountCalls++` | — | 자신 |
| B4 :181 | `/api/v1/orders` | `ordersCalls++`, 429 | — | 자신 |
| B5 :184 | 그 외 경로 | 404 | — | 미진입(방어 팔) |
| B6 :191 | 엔진 조립 실패 | — | `t.Fatalf` | 미진입(조립 성공) |
| B7 :221 | `Recovery` 생성 실패 | — | `t.Fatalf` | 미진입 |
| B8 :225 | `ErrRecoveryIncomplete`가 아님 | — | `t.Fatalf` | **안전 단언** — 미진입이 통과다 |
| B9 :228 | 부분 스냅샷을 complete로 보고 | — | `t.Fatal` | **안전 단언** |
| B10 :231 | 부분 스냅샷 보존 | — | `t.Fatalf` | **안전 단언** |
| B11 :234 | accounts 호출 ≠ 1 | — | `t.Fatalf` | **안전 단언** |
| B12 :239 | orders 호출 ≠ `wantReads` | — | `t.Fatalf` | 뮤테이션 M1·M2·M3가 이 팔로 죽었다 |
| B13 :247 | `RateLimitWaits` 또는 소진액이 `wantWaits × backoff` 와 불일치 | — | `t.Fatalf` | M3·**M4**가 이 팔로 죽었다 |
| B14 :251 | 실제 대기 횟수 ≠ `wantWaits` | — | `t.Fatalf` | 예산 초과 대기 금지 |
| B15 :255 | 기록된 대기를 순회한다 | — | — | 리뷰 지적으로 추가 |
| B16 :256 | 어떤 대기라도 백오프와 다르면 | — | `t.Fatalf` | **M5**가 이 팔로 죽었다(짧은 수면) |
| B17 :260 | 오류 문구가 rate limit을 안 지목 | — | `t.Fatalf` | 사유 오인 방지(a102 A1 F8) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `assembleAccountSequenceTestEngine` | fake broker 위에 엔진 조립 | 실패 시 B6 | AST calls |
| `ectx.Recovery` | 복구 시퀀스 생성 | 실패 시 B7 | AST calls |
| `recovery.Run` | 429 아래 복구 실행 | `ErrRecoveryIncomplete` 기대 | AST calls |
| `newA102Clock` | 잠들지 않는 시계 | runaway 시 오류 반환 | AST calls |
| `strings.Contains` | 사유 문구 확인 | — | AST calls |

## State mutations and fallbacks

- 프로세스 밖 상태를 바꾸지 않는다. `httptest` 서버와 임시 저널만 쓴다.
- 실 계좌·실 API·실 주문 경로 접촉 0. 브로커 응답은 전부 이 파일 안의 핸들러가 만든다.

## Safety conclusion

- Safe edit boundary: 이 테스트 함수 안. 생산 코드는 이 change 에서 한 줄도 바뀌지 않는다.
- High-risk impact: no — 다만 이 테스트가 지키는 대상(복구의 fail-closed)은 High-risk 이고,
  안전 단언 4개(B8~B11)는 이 change 에서 문구 하나 바뀌지 않았다.
