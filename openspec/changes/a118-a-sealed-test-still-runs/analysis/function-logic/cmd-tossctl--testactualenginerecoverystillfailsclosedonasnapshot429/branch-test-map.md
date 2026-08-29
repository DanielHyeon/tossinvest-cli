# Branch Test Map: `TestActualEngineRecoveryStillFailsClosedOnASnapshot429`

근거는 커버리지가 아니라 실행이다 — `go test` 는 `_test.go` 를 계측하지 않는다.
"GREEN observed" 는 2026-08-29 `make test-seams` 통과 실행(96 패키지 ok)을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | fake broker 가 경로를 가른다 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` `engine_account_seq_recovery_test.go:170` | no | yes |
| B2 | 토큰 발급 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | yes |
| B3 | 계좌 1회 조회 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | yes |
| B4 | 주문 조회가 429 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | yes |
| B5 | 그 외 경로 404 — 방어 팔, 미진입 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no |
| B6 | 조립 실패 — 미진입 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no |
| B7 | Recovery 생성 실패 — 미진입 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no |
| B8 | 오류가 ErrRecoveryIncomplete 가 아니면 실패 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no (통과 = 미진입) |
| B9 | 부분 스냅샷을 complete 로 보고하면 실패 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no (통과 = 미진입) |
| B10 | 부분 스냅샷을 보존하면 실패 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no (통과 = 미진입) |
| B11 | accounts 호출이 1이 아니면 실패 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no (통과 = 미진입) |
| B12 | orders 호출이 유도값과 다르면 실패 — 뮤테이션 M1·M2·M3 이 이 팔로 죽었다 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | yes (뮤테이션) | no (통과 = 미진입) |
| B13 | 소진액이 wantWaits × backoff 와 다르면 실패 — M3·M4 가 이 팔로 죽었다 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | yes (뮤테이션) | no (통과 = 미진입) |
| B14 | 예산을 넘겨 기다리면 실패 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no (통과 = 미진입) |
| B15 | 기록된 대기를 순회한다 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | yes |
| B16 | 어떤 대기라도 백오프와 다르면 실패 — M5 가 이 팔로 죽었다(짧은 수면) | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | yes (뮤테이션) | no (통과 = 미진입) |
| B17 | 사유가 rate limit 을 안 지목하면 실패 | `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | no | no (통과 = 미진입) |
