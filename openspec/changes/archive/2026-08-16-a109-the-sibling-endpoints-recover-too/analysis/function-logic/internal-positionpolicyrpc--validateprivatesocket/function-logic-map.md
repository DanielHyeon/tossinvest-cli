# Function Logic Map: `ValidatePrivateSocket`

- Source: `internal/positionpolicyrpc/private_endpoint.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**클라이언트·발행 확인용 검증**이다. a109의 회수 완화(`perm&0o077 == 0`)는 여기에
**오지 않는다**(design D2, freeze P1-3): 발행 후 최종 이름은 항상 0600을 지난 뒤에만
나타나므로, 이 함수가 정확-0600을 요구하는 것은 계약이지 관성이 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 유효한 control 디렉터리 안의 socket 경로 | 호출자(engine `StartAlertControlServer` 발행 후 확인) | `statPrivateSocket`의 error 그대로 |
| socket 모양·권한 | Unix socket · 비symlink · **정확 0600** | `statPrivateSocket` | error — 완화 금지 대상 |
| 소유·hard link | 우리 uid | `validateOwnerAndLinks(info, false)` (nlink 요구 없음) | error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `statPrivateSocket(path)`가 error | 없음 | 그 error 그대로 | 0700 socket을 **거부**해야 한다(a109 뮤테이션 원장의 클라이언트 정확-0600 항목) |
| 분기 밖 종단 | stat 통과 | 없음 | `validateOwnerAndLinks(info, false)` | 우리 uid 소유 socket은 통과 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `statPrivateSocket` | 디렉터리 위생 + socket 모양/정확-0600 | error 전파 | AST calls |
| `validateOwnerAndLinks(_, false)` | 소유 uid 확인(hard link 수는 요구하지 않음) | error 전파 | AST calls |

## State mutations and fallbacks

- 상태 변경 없음. 제거도 생성도 하지 않는다 — 그것이 `PreparePrivateSocket`과의 유일한 차이다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** a109가 넓히는 것은 회수 전용 함수뿐이다.
  이 함수를 `perm&0o077 == 0`으로 바꾸면 클라이언트가 group/other 비트 없는 아무 socket을
  받아들이게 되어 발행 계약(0600)이 검증되지 않는다.
- High-risk impact: yes — 발행 직후 확인에 쓰이므로, 느슨해지면 잘못 발행된 endpoint가 통과한다.
