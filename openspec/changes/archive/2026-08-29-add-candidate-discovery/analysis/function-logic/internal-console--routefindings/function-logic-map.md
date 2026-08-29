# Function Logic Map: `routeFindings`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

라우트 **한 건**에 대한 계좌 변조 판정, 레코드의 순수 함수.

**두 주장, 두 루프**: ① 표 어디에도 주문을 내거나 취소·정정하거나 게이트를 열거나 자격증명에 닿는 이름이 없다 — `consoleStateChanging`의 아홉도 포함해서(그것들은 무언가를 바꿔도 되지만 계좌를 건드리지 않는다). ② 그 목록 밖의 모든 라우트는 읽기다.

예외는 **두 루프 모두에서** 참조된다. 두 번째 루프의 `actVerbs`가 `accountVerbs`를 포함하므로 `/orders`가 양쪽에 걸리고, 틀린 수리는 `/orders`를 `consoleStateChanging`에 넣어 두 번째를 조용하게 만드는 것이다 — 그 목록은 CSRF 게이트를 요구하고, CSRF 게이트 뒤의 읽기 라우트는 아무도 열 수 없는 페이지다.

**잡지 못하는 것(기록된 경계)**: 이것은 경로 문자열의 느슨한 부분 문자열 검사다. **금지 동사를 이름에 쓰지 않고 계좌에 닿는 경로**는 아무 발견도 내지 않는다. console-operator-overview 리뷰 P1-4가 초안의 반대 주장을 정정했다 — '경로 allowlist가 주문 티켓 이식을 막는다'는 틀렸고, 티켓은 `/ticket`으로 POST할 수 있으며 그 경로는 CSRF 게이트에 걸릴 때만 실패한다. **기계적 보증은 `Options` 능력 열거가 한다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r` | 라우트 1건 | 추출기 | 해당 없음 |
| `consoleStateChanging` | 상태변경 허용 9경로 | static_test.go:606 | 여기에 없는데 행위처럼 읽히면 발견 |
| `accountVerbs` | `sharedAccountVerbs`(order·sell·buy·cancel·modify·amend·flatten) + `routeOnlyAccountVerbs`(gate·credential·secret·token·adopt·enroll) | static_test.go:647 | 경로에 포함되면 발견 — 단, `reads`면 면제 |
| `actVerbs` | start·stop·approve·abort·restart·reset·delete·save·include·enable·config + `accountVerbs` | 함수 본문 | config 쓰기 어휘가 여기 있는 이유는 미래의 `/settings/anything`이 그냥 지나가지 못하게 하기 위함이다(console-adoption-controls 리뷰 P2-7) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, path := range consoleStateChanging` | `allowed` 맵 구성 | 없음 | 표 판정 전체 |
| B2 | `for _, verb := range accountVerbs` | 없음 | 없음 | 같은 위 |
| B3 | `strings.Contains(lowered, verb) && !reads` | 발견 추가 | 없음 | `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`(`/order/place`, `/flatten`, `/gate/open`) |
| B4 | `allowed[r.Path] || reads` | 없음 | 여기까지의 발견만 반환 | `/verify/start` 등 아홉 + `/orders` |
| B5 | `for _, verb := range actVerbs` | 없음 | 없음 | 표 판정 전체 |
| B6 | `strings.Contains(lowered, verb)` | 발견 추가 | 없음 | `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`(`/verify/reset`) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `routeReadsTheAccountRecord` | 예외 판정 | 세 사실 전부 | static_test.go:679 |
| `strings.ToLower` / `strings.Contains` / `fmt.Sprintf` | 느슨한 문자열 검사 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수).
- 발견 개수를 주석에 적지 않는다 — 한때 '다섯'이라고 적혀 있었는데 목록에는 아홉이 있었다.

## Safety conclusion

- Safe edit boundary: 신설(테스트 본문에서의 추출). 판정 규칙 자체는 이전 인라인 루프 둘과 같고, 더해진 것은 `reads` 예외가 **두 루프 모두에서** 참조된다는 것이다.
- High-risk impact: yes (계좌 라우트 부재 보증 — 이 판정이 넓어지면 주문·게이트·자격증명 경로가 표에 조용히 들어온다)
