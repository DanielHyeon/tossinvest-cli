# Function Logic Map: `routeFindings`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `consoleStateChanging` | 논증된 상태변경 라우트 전부 | 스펙 본문의 열거 | 누락은 "행위하는데 논증 없음" 실패 |
| `accountVerbs` | 계좌·게이트·자격증명 어휘 | 공유 목록 | 경로가 담으면 무조건 실패 |
| `actVerbs` | 행위 어휘 | 이 함수 | 목록 밖 어휘는 침묵 통과 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `consoleStateChanging` 순회 | `allowed` 구성 | 없음 | `TestNoRouteNamesAnAccountMutation` |
| B2 | `accountVerbs` 순회 | 없음 | 없음 | 자기 자신 |
| B3 | 경로가 계좌 어휘를 담고 읽기 예외가 아님 | 없음 | finding 추가 | `TestNoRouteNamesAnAccountMutation` |
| B4 | `allowed[r.Path] \|\| reads` | 없음 | findings 조기 반환 | `TestNoRouteNamesAnAccountMutation` |
| B5 | `actVerbs` 순회 | 없음 | 없음 | 자기 자신 |
| B6 | 목록 밖 라우트가 행위 어휘를 담음 | 없음 | finding 추가 | `TestNoRouteNamesAnAccountMutation` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `routeReadsTheAccountRecord` | `/orders` 읽기 예외 | 세 사실 전부 필요 | CodeGraph + AST |
| `strings.Contains` | 느슨한 부분문자열 판정 | 느슨함이 의도다 | AST |

## State mutations and fallbacks

- 이 change가 바꾼 것: `actVerbs`에 `"limit"`과 `"preset"` 추가.
- 근거는 `"exclude"`가 추가된 것과 같다. `/settings/limits`는 기존 어휘 중 **어느 것도** 담지 않으므로, 목록에 넣기 전에는 논증 없이 등록해도 이 가드를 통째로 비켜 갔다.
- `"preset"`은 우연히 `"reset"`을 부분문자열로 담아 RED에서 실제로 걸렸다. 그것은 운이지 가드가 아니다 — `/settings/limits` 단독이었다면 아무 일도 없었을 것이고, 그래서 `"limit"`을 명시적으로 넣었다.
- `routeOnlyAccountVerbs`의 `"gate"`는 **건드리지 않았다**. 이 change 이후에도 게이트를 여는 라우트는 없어야 하기 때문이다.

## Safety conclusion

- Safe edit boundary: `actVerbs` 목록. 새 상태변경 라우트를 만드는 change는 그 어휘를 여기에 더해야 다음번을 막는다.
- High-risk impact: yes(가드) — 이 목록이 새 config-write 어휘를 모르면 미승인 라우트가 침묵으로 통과한다.
