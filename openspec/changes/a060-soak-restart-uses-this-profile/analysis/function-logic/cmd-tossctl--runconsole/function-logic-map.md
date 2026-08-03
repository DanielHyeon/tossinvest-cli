# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision: current)
- Change: a060-soak-restart-uses-this-profile
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 도메인 입력 없음 | as before | as before | as before |

## Branches and early returns

콘솔 조립. a060은 restartSoak 호출에 root를 넘기는 한 줄만 바꾼다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 204 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B2 | if at line 213 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B3 | if at line 218 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B4 | if at line 222 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B5 | if at line 226 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B6 | if at line 230 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B7 | if at line 234 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B8 | if at line 239 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B9 | if at line 247 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B10 | if at line 249 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B11 | else at line 252 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B12 | if at line 256 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B13 | else at line 259 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B14 | if at line 270 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B15 | else at line 273 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B16 | if at line 281 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B17 | else at line 284 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B18 | if at line 284 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B19 | else at line 286 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B20 | if at line 288 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B21 | else at line 296 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B22 | if at line 290 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B23 | else at line 292 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B24 | if at line 299 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B25 | if at line 303 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B26 | else at line 305 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B27 | if at line 312 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B28 | if at line 315 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B29 | if at line 331 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B30 | if at line 338 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B31 | if at line 346 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B32 | if at line 348 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B33 | else at line 360 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B34 | if at line 350 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B35 | else at line 352 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B36 | if at line 360 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 95 call sites, contract unchanged by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- 없음 — 순수 함수.
- 새 브로커 호출·config key·audit 레코드·POST 라우트 없음.

## Safety conclusion

- Safe edit boundary: 자식 argv와 프로세스 발견·소유 판정.
- 이 change는 주문 제출·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결 어느 경로도 바꾸지 않는다. 대상은 조회 전용 서베이이며, 바꾸는 것은 자식 argv 두 플래그와 프로세스 발견·소유 판정이다. 방향은 두 갈래 모두 보수적이다 — 콘솔이 자기 서베이를 띄우고 찾게 되는 쪽은 지금 100% 실패하는 버튼을 동작시키고, 소유 판정 쪽은 다른 기록의 서베이에 SIGINT가 가지 않게 막는다.
