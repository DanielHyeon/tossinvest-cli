# Function Logic Map: `TestTheSoakSpawnCarriesThisProfile`

- Source: `cmd/tossctl/soakproc_test.go`
- AST evidence: `ast.json` (revision: current)
- Change: a060-soak-restart-uses-this-profile
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 도메인 입력 없음 | as before | as before | as before |

## Branches and early returns

a058의 테스트. 콘솔이 자기가 spawn한 엔진을 찾는다는 계약과, 그 발견이 다른 프로필의 엔진까지 넓어지지 않는다는 계약을 함께 고정한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 259 | 없음 — 테스트 코드 | unchanged behaviour | `TestTheSoakSpawnCarriesThisProfile` (this function is the test) |
| B2 | if at line 265 | 없음 — 테스트 코드 | unchanged behaviour | `TestTheSoakSpawnCarriesThisProfile` (this function is the test) |
| B3 | range at line 269 | 없음 — 테스트 코드 | unchanged behaviour | `TestTheSoakSpawnCarriesThisProfile` (this function is the test) |
| B4 | if at line 274 | 없음 — 테스트 코드 | unchanged behaviour | `TestTheSoakSpawnCarriesThisProfile` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 11 call sites, contract unchanged by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- 없음 — 테스트 코드.
- 새 브로커 호출·config key·audit 레코드·POST 라우트 없음.

## Safety conclusion

- Safe edit boundary: 자식 argv와 프로세스 발견·소유 판정.
- 이 change는 주문 제출·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결 어느 경로도 바꾸지 않는다. 대상은 조회 전용 서베이이며, 바꾸는 것은 자식 argv 두 플래그와 프로세스 발견·소유 판정이다. 방향은 두 갈래 모두 보수적이다 — 콘솔이 자기 서베이를 띄우고 찾게 되는 쪽은 지금 100% 실패하는 버튼을 동작시키고, 소유 판정 쪽은 다른 기록의 서베이에 SIGINT가 가지 않게 막는다.
