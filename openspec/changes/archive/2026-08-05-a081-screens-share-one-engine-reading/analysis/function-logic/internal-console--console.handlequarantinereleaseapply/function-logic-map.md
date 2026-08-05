# Function Logic Map: `Console.handleQuarantineReleaseApply`

- Source: `internal/console/exit_quarantine.go`
- AST evidence: `ast.json` (`source_sha256` 22c049b50d5e… (ast.json), lines 192–223)
- Risk scan: `risk-pattern-report.md`
- **개정 2026-08-05.** 초안은 성공 경로에 캐시 무효화 1건을 더했다. 독립 리뷰가
  그 근거가 사실이 아님을 보여 **철회**했다 — 이 함수의 편집은 이제 "왜 무효화하지
  않는가"를 적은 주석뿐이다. 증거는 그 판단의 기록으로 유지한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.exitQuarantines()` | 엔진이 a079 이전이면 미발견 | `c.opts.PositionPolicies`의 타입 단언 | 미발견이면 refuse, return (B1) |
| `capability` (POST) | 비어 있지 않은 engine capability | 브라우저 폼 | 빈 값이면 403, 엔진 미도달 (B2) |
| `ReleaseQuarantine`의 err | — | 엔진 | `writeQuarantineError`로 분류 후 return (B3) |

**불변식 1 — 격리 상태는 엔진 정책 캐시를 지나지 않는다.** 해제는
`exit_snapshot_quarantines`만 쓴다. `positionpolicy.State`에는 격리 필드가 없고
lifecycle SELECT도 그것을 join하지 않는다. 화면의 격리 배지는
`positionRow.Quarantine`이고 그것은 콘솔 자신의 journal 읽기에서 렌더마다 새로
온다 — 즉 **이미 신선하다**.

그래서 여기서 캐시를 버리는 것은 엔진 읽기 1회를 더 사고, 존재하지 않는 data
flow를 주장하는 주석과 그것을 단정하는 테스트를 남기는 일이었다. 무효화 목록은
**실제 data flow를 따라야 한다**는 것이 이 함수가 남기는 교훈이고, 코드 주석이
그 이유를 들고 있다.

**불변식 2 — 이 경로는 타입 단언으로 발견된다.** 커맨더 인터페이스가 아니라
`ExitQuarantineCommander`로의 단언이 성공해야 도달한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 격리 커맨더 미발견 (194) | 없음 | refuse, return (197) | 기존 |
| B2 | `capability == ""` (200) | 없음 | 403 refuse, return (203) | 기존 |
| B3 | `ReleaseQuarantine` err (208) | 없음 | 분류된 refuse, return (210) | 기존 |
| — | 성공 경로 | **없음** — 무효화는 철회했다 (8.4) | 303 redirect (222) | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `commander.ReleaseQuarantine` | 격리 해제 1회 | 에러는 B3가 분류. 재시도 없음 | ast line 205 |

## State mutations and fallbacks

- **없다.** 저장된 기준선도, 격리 원장도, 콘솔 캐시도 이 함수가 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 성공 경로의 주석뿐. 실행되는 문장은 편집 전과 같다.
- High-risk impact: **no.** 격리 해제 자체의 계약(기준선 보존, 재격리 가능성)은
  무변경이고 a081은 이 경로의 동작을 바꾸지 않는다.
