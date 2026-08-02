# Function Logic Map: `splitProcessLine`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a059-console-finds-the-engine-it-owns
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 도메인 입력 없음 | as before | as before | as before |

## Branches and early returns

`pgrep -a` 한 줄을 pid와 명령줄로 나눈다. 명령줄이 없는 줄은 판정 근거가 없으므로 실패로 돌려준다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 380 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B2 | if at line 384 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |
| B3 | if at line 388 | 없음 — 순수 함수 | unchanged behaviour | 이 함수를 통과하는 `cmd/tossctl` 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 4 call sites, contract unchanged by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- 없음 — 순수 함수.
- 새 브로커 호출·config key·audit 레코드·POST 라우트 없음.

## Safety conclusion

- Safe edit boundary: 엔진 프로세스 발견과 소유 판정.
- 이 change는 주문 제출·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결 어느 경로도 바꾸지 않는다. 바꾸는 것은 엔진 프로세스 **발견**과 그 발견의 **소유 판정** 둘이며, 배타는 그대로 journal 디렉터리 flock이 담당한다. 방향은 두 갈래 모두 보수적이다 — 콘솔이 자기 엔진을 찾게 되는 쪽은 정지 버튼을 동작시키고, 소유 판정 쪽은 다른 프로필의 엔진에 SIGTERM이 가지 않게 막는다.
