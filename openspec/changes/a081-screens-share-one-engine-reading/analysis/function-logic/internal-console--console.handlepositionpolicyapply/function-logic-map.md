# Function Logic Map: `Console.handlePositionPolicyApply`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json` (`source_sha256` 5dda5ebbf9b5…, lines 350–375)
- Risk scan: `risk-pattern-report.md`
- a081이 더하는 것은 **성공 경로의 호출 1건**이다. 분기·early return·거부 사유는
  전부 무변경이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.PositionPolicies` | nil 허용 | `console.go` | nil이면 501로 거부 (B1) |
| `capability` (POST) | 비어 있지 않은 engine capability | 브라우저 폼 | 빈 값이면 403, **엔진에 닿지 않는다** (B2) |
| `confirm` (POST) | `"yes"` 여부 | 브라우저 폼 | 위험 action에서 미확인이면 엔진이 거부 |
| `Apply`의 err | — | 엔진 | `writePositionPolicyError`로 분류 후 return (B3) |

**불변식 — 무효화는 성공 뒤에만 일어난다.** B3의 return 아래로 내려온 실행은
엔진이 정말로 상태를 바꾼 경우뿐이다. 거부 경로에서 캐시를 버리면 거부된 요청을
반복하는 것만으로 a081의 상한을 무력화할 수 있다 (design D5).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 커맨더 미배선 (351) | 없음 | 501 refuse, return (353) | 기존 |
| B2 | `capability == ""` (356) | 없음 | 403 refuse, return (358) | 기존, 4.3 |
| B3 | `Apply` err (363) | 없음 | 분류된 refuse, return (365) | 기존, 4.3 |
| — | 성공 경로 | **`c.enginePolicy.invalidate()`** (371) | 303 redirect (374) | 4.1 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.opts.PositionPolicies.Apply` | 엔진에 capability-bound 변경 1회 | 에러는 B3가 분류. 재시도 없음 | ast line 360 |
| `c.enginePolicy.invalidate` | 방금 움직인 lifecycle을 다음 렌더가 보게 한다 | 에러 없음. nil 수신자 안전 | ast line 371 |

## State mutations and fallbacks

- a081 이전: 이 함수는 콘솔 프로세스 안의 상태를 바꾸지 않았다. 이후에도 바꾸는
  것은 캐시를 **비우는 것** 하나뿐이며, 비운 캐시는 다음 읽기에서 정상 경로로
  다시 채워진다. 실패해도 되는 mutation이다.

## Safety conclusion

- Safe edit boundary: line 371 한 줄. capability 검사·거부 분류·redirect는 불변.
- High-risk impact: **no** — 이 함수 자체는 기존에도 엔진 정책 mutation을 중계했고
  a081은 그 계약을 넓히지도 좁히지도 않는다. 추가된 호출은 표시 캐시를 비운다.
