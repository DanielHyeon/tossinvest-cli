# Function Logic Map: `processAlive`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (169-175)
- AST evidence: `ast.json` — **revision `base`** (커밋 `75df9bf9`). AST 분기 1
- Risk scan: `risk-pattern-report.md`

## 이 문서가 base revision인 이유

**이 함수는 a108에서 삭제됐다.** 현재 HEAD의 `transport_unix.go`에 `processAlive`는 없다.
AST는 비교 기준 커밋의 파일에서 뽑았고, 삭제 전 상태의 증거로 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pid` | `descriptor.PID`(잔재 endpoint.json에 적힌 값) | 디스크의 descriptor | `pid <= 0`이면 사망으로 읽는다(B1) |
| 커널의 PID 공간 | **이 프로세스와 같은 PID 네임스페이스** | `unix.Kill(pid, 0)` | 오류 없음 또는 `EPERM`이면 생존 |

**깨진 불변식이 삭제 이유다.** 이 판정은 "descriptor를 쓴 프로세스가 지금도 그 PID로
살아 있다"를 전제하는데, 컨테이너를 다시 만들면 PID 배정이 근사-결정적이라(a102 D4b-2
실측) 그 자리에 무관한 프로세스가 앉는다. 8/13 사고의 descriptor PID는 16이었고, 새
컨테이너에서 PID 16은 거의 확실히 점유돼 있다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `pid <= 0` | 없음 | `false`(사망) | 없음 — 삭제된 함수 |

분기 밖의 반환은 `err == nil \|\| errors.Is(err, unix.EPERM)`이며, `EPERM`을 생존으로 읽는
것은 "다른 uid의 프로세스도 존재는 한다"는 뜻이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `unix.Kill(pid, 0)` | 신호 없이 존재 여부만 묻는다 | `ESRCH`=부재, `EPERM`=존재(권한 없음) | AST calls |
| `errors.Is` | `EPERM` 구분 | — | AST calls |

호출부는 `reclaimStaleControlDirectory` 하나뿐이었다(`grep -rn processAlive --include=*.go`
= 정의 1 + 호출 1). 그래서 삭제의 영향 범위는 그 함수 안에서 닫힌다.

## State mutations and fallbacks

- 없다. 순수 조회 함수였다. 위험은 side effect가 아니라 **오답의 전파**였다.

## Safety conclusion

- Safe edit boundary: 삭제. 대체는 `projectionSocketAccepts`의 connect probe이며, 그것은
  "이 경로에서 수락하는 자가 있는가"라는 **다른 질문**에 답한다 — PID가 누구인지 묻지
  않으므로 PID 재배정에 면역이다.
- 판정에서만 빼고 로그 참고값으로 남기지 않은 이유: 남기면 다음 독자가 다시 판정에
  쓴다(design D2). 되돌아갈 자리를 없애는 것이 삭제의 목적이다.
- High-risk impact: yes — 이 함수의 오답이 8/13 사고에 숨어 있던 두 번째 사고 형태였다.
  뮤테이션 M5·M5c(`../../mutation-ledger-t1.md`)가 그 오답 두 방향을 각각 죽인다.
