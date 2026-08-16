# Function Logic Map: `PositionPolicyCommandServer.Close`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

세 Close 중 **유일하게 socket을 지우지 않는다** — TCP endpoint이므로 지울 파일이
descriptor와 control 디렉터리뿐이다. 그래서 a108의 late-unlink 논거(design D2b)가 여기에는
적용되지 않는다: listener가 unlink할 경로 자체가 없다. a109는 **이 함수를 편집하지 않는다.**

편집 대상이 아닌 이유를 남기는 것이 이 맵의 목적이다: 세 Close를 "같으니까 같이 고친다"로
묶으면 근거 없는 High-risk 편집이 하나 늘어난다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | nil 허용 | 호출자 defer | nil이면 `nil` |
| `s.once` | 한 번만 | `sync.Once` | 두 번째 Close는 무동작 |
| Shutdown 시한 | 2초 | 본문 상수 | 초과 error가 `result` |
| 제거 대상 | descriptor, 그 부모 디렉터리 | 본문 | ErrNotExist만 관용 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s == nil` | 없음 | `nil` | nil 수신자 안전 |
| B2 | `os.Remove(s.descriptor)` err && !ErrNotExist && result==nil | descriptor 제거 | 첫 오류 보존 | Close 후 descriptor 부재 |
| B3 | `os.Remove(filepath.Dir(s.descriptor))` err && !ErrNotExist && result==nil | control 디렉터리 제거 | 첫 오류 보존 | Close 후 디렉터리 부재 |
| 분기 밖 종단 | — | `Shutdown` 결과가 기본값 | `result` | 정상 Close는 nil |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `context.WithTimeout` | 2초 시한 | `defer cancel()` | AST defers |
| `s.server.Shutdown` | 요청 정리 | 시한 초과 error 보존 | AST calls |
| `os.Remove` ×2 | descriptor · 디렉터리 | ErrNotExist 관용 | AST calls |
| `filepath.Dir` | 디렉터리 경로 | — | AST calls |
| `errors.Is` | 관용 판정 | — | AST calls |

## State mutations and fallbacks

- descriptor와 디렉터리가 사라진다. **staging 잔재는 지우지 않는다** — 정상 종료에서는
  `writePositionPolicyDescriptor`의 defer가 이미 치웠고, crash에서는 이 함수가 아예
  불리지 않는다. 그 잔재를 받는 것은 다음 부팅의 위생(a109 §1.4)이다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** 디렉터리 제거가 `ENOTEMPTY`로 실패할 수
  있는데(staging 잔재가 있으면), 그것은 첫 오류로 보존될 뿐 다음 부팅을 막지 않는다 —
  위생이 다음 부팅에서 치운다.
- High-risk impact: yes — 격리 해제 표면의 수명주기다.
