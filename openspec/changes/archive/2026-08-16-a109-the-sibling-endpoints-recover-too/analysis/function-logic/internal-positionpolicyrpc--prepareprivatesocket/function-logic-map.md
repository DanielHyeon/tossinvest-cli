# Function Logic Map: `PreparePrivateSocket`

- Source: `internal/positionpolicyrpc/private_endpoint.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**a109가 진단하는 병이 여기 있다.** 이 함수는 alert control 기동이 잔재를 치우는 유일한
수단이었고 두 가지를 못 한다:

1. `statPrivateSocket`의 **정확-0600** 요구 때문에 pre-chmod 0700 잔재를 영구 거부한다
   (design 병 표, A1 F3 실측).
2. **생존 probe가 없다.** 유효한 0600 socket이면 주인이 살아 있어도 `os.Remove` 한다 —
   산 주인 위에 두 번째 서버가 올라선다(A1 F4).

a109는 이 함수를 **고치지 않고 기동 경로에서 뺀다**: 회수는 신규
`positionpolicyrpc.ReclaimStalePrivateEndpoint`(이름-독립)가 맡는다. 함수 자체는 기존
공개 API와 기존 테스트를 위해 남고 호출자만 사라진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | control 디렉터리 안의 socket 경로 | 호출자 | `statPrivateSocket` error 전파 |
| 경로 부재 | 정상(치울 것이 없음) | `errors.Is(err, os.ErrNotExist)` | `nil` — 성공으로 읽는다 |
| 잔재 모양 | Unix socket · 비symlink · **정확 0600** | `statPrivateSocket` | error — pre-chmod 0700이 여기서 영구 거부된다 |
| 소유 | 우리 uid | `validateOwnerAndLinks(info, false)` | error |
| 주인의 생사 | **검사하지 않는다** | — | 산 주인의 socket도 제거된다(a109가 없애는 동작) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `errors.Is(err, os.ErrNotExist)` | 없음 | `nil` | 첫 기동(잔재 없음) |
| B2 | `err != nil` (그 밖의 stat 실패) | 없음 | error 그대로 | 정규 파일이 놓여 있으면 거부하고 **지우지 않는다** |
| B3 | `validateOwnerAndLinks(info, false) != nil` | 없음 | error | 남의 uid 소유 socket 제거 금지(비root 테스트로는 재현 불가) |
| 분기 밖 종단 | 위 셋 통과 | **`os.Remove(path)`** — 생존 확인 없는 제거 | `os.Remove`의 결과 | a109 §1.2 RED가 이 줄의 결과(탈취)를 고정한다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `statPrivateSocket` | 디렉터리 위생 + 모양/정확-0600 | error 전파 | AST calls |
| `errors.Is` | ErrNotExist를 성공으로 | — | AST calls |
| `validateOwnerAndLinks` | 소유 uid | error 전파 | AST calls |
| `os.Remove` | 잔재 unlink | error 전파, 재시도 없음 | AST calls |
| `strings.TrimSpace` | 경로 정규화 | — | AST calls |

## State mutations and fallbacks

- 유일한 상태 변경은 마지막 줄의 `os.Remove`다. **fallback이 없다** — 지우거나 실패하거나다.
- "주인이 살아 있으면 거부"라는 상태가 이 함수에는 없다. a109가 더하는 것이 그 상태다.

## Safety conclusion

- Safe edit boundary: 본문을 편집하지 않는다. 호출부(engine `StartAlertControlServer`)에서
  **제거**하고 회수 기계로 대체한다.
- High-risk impact: yes — 산 socket 제거는 두 번째 서버가 운영자 표면을 가로채는 상태를 만든다.
