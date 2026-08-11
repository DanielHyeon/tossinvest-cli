# Function Logic Map: `TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport`

- Source: `internal/protection/dormant_test.go` (L14-88)
- AST evidence: `ast.json` — 분기 20, return 4
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

design D6이 이 가드를 **최소로 연다**. 테스트 파일도 FLM 대상이다(프로젝트 기억
`tossos-logic-map-scope-creep`). 그리고 이 함수는 a071이 만든 봉인의 **본체**이므로, 어디를
얼마나 여는지는 문장이 아니라 이 열거로 정해야 한다.

## 이 가드는 네 겹이다

AST 20분기를 층으로 정리하면 이렇다.

| 층 | 분기 | 검사 대상 | a100의 영향 |
|---|---|---|---|
| 1. 프로필 | B1 (L15) | `execgw.ProfileProtection == ProtectionUnwired` | **영향 없음** — a100은 `Wired`를 생산하지 않는다 |
| 2. core 순수성 | B3~B7 (L25-39) | `internal/protection`의 non-test 파일이 `net/http`·`/internal/official`·`/internal/trading`을 import하지 않는다 | **영향 없음** — a100은 core에 transport를 넣지 않는다 |
| 3. import 경계 | B8~B14 (L45-72) | `cmd/`와 `internal/app/` 전체를 걸어 `internal/protection`을 import하는 파일이 **허용 목록에 있는지** | **여기가 열린다** |
| 4. 조립 내용 | B17~B20 (L78-87) | `gateway.go` 본문에 필수 2문자열이 있고 금지 4문자열이 없다 | **금지 4개 중 1개가 필요해진다** |

## 3층 — 허용 목록은 **파일 단위**다

L61이 정확히 이렇다.

```go
allowed := map[string]bool{"internal/app/engine/gateway.go": true}
```

⇒ **`internal/protection`을 import할 수 있는 파일은 `gateway.go` 하나뿐이다.**
수렴 워커를 `internal/app/engine/`의 **새 파일**에 두고 그 파일이 `protection.Scope` 같은
타입을 import하면 B14가 즉시 실패한다 — "exposes a second protection assembly path".

선택지는 둘이다.

| 선택 | 봉인 개방 폭 | 결과 |
|---|---|---|
| (a) `protectionofficial` gateway를 `gateway.go` 안에서 만들고, 워커 파일에는 **인터페이스만** 넘긴다 | 허용 목록 **불변** | 3층을 전혀 열지 않는다 |
| (b) 워커 파일을 허용 목록에 추가한다 | 항목 1개 추가 | 두 번째 조립 경로가 생긴다 — 가드가 막으려던 바로 그것 |

⇒ **(a)를 택한다.** design D6의 "봉인 가드는 최소로 연다"가 여기서 갖는 구체적 의미는
「허용 목록을 건드리지 않는다」이다. 워커는 `internal/protection`을 import하지 않고,
`gateway.go`가 만든 값을 좁은 인터페이스로 받는다.

## 4층 — 금지 4개 중 정확히 1개

L83의 목록:

| 금지 문자열 | a100에 필요한가 | 사유 |
|---|---|---|
| `protection.NewSupervisor` | **아니다** | supervisor는 a105 |
| `protectionofficial.New` | **필요하다** | 브로커 조건주문 어댑터를 만들어야 한다 |
| `protection.db` | 아니다 | 별도 DB를 도입하지 않는다(D4) |
| `GatewayFactory` | 아니다 | 팩토리를 쓰지 않는다 |

⇒ **금지 목록에서 빼는 것은 `protectionofficial.New` 하나.** 나머지 3개는 남긴다.
그리고 L78의 필수 2개(`protectionreadiness.NewProductionProvider`,
`protection.NewPairedReadinessAdapter`)는 **그대로 유지해야 한다** — a100은 readiness 조립을
바꾸지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `moduleRoot(t)` | 모듈 루트 | dormant_test.go:143 | 실패 시 `t.Fatal` |
| `packageDir` 목록 | `internal/protection`의 `.go` non-test | 파일시스템 | 읽기 실패 = B2 `t.Fatal` |
| walk 대상 | `cmd/`, `internal/app/` 전체 | 파일시스템 | walk 에러 = B15 `t.Fatal` |
| `gateway.go` 본문 | 문자열 검사 | 파일시스템 | 읽기 실패 = B16 `t.Fatal` |

**불변식 — 4층은 문자열 검사다.** AST가 아니라 `strings.Contains`이므로, 금지 문자열이
**주석 안에 있어도 실패한다.** a100이 `protectionofficial.New`를 주석으로 설명하려 해도
목록에서 빼기 전에는 실패한다.

## Branches and early returns

| Branch | 조건 (L) | 결과 |
|---|---|---|
| B1 | 프로필 ≠ UNWIRED (15) | `t.Fatalf` |
| B2 | `ReadDir` 실패 (22) | `t.Fatal` |
| B3 | `range entries` (25) | 순회 |
| B4 | 디렉터리·비-go·테스트 파일 (26) | skip |
| B5 | 파싱 실패 (31) | `t.Fatal` |
| B6 | `range file.Imports` (34) | 순회 |
| B7 | 금지 import (36) | `t.Errorf` |
| B8 | `range [cmd, internal/app]` (45) | 순회 |
| B9 | walk 에러 (47) | 반환 |
| B10 | 디렉터리·비-go·테스트 파일 (50) | skip |
| B11 | 파싱 실패 (54) | 반환 |
| B12 | `range file.Imports` (57) | 순회 |
| B13 | `internal/protection` import (59) | 허용 목록 대조 |
| B14 | **허용 목록에 없음 (62)** | `t.Errorf` — **a100이 마주칠 분기** |
| B15 | walk 반환 에러 (69) | `t.Fatal` |
| B16 | `gateway.go` 읽기 실패 (74) | `t.Fatal` |
| B17 | `range required` (78) | 순회 |
| B18 | 필수 문자열 부재 (79) | `t.Errorf` |
| B19 | `range forbidden` (83) | 순회 |
| B20 | **금지 문자열 존재 (84)** | `t.Errorf` — **a100이 마주칠 분기** |

**측정 note.** 테스트 파일은 커버리지 계측 대상이 아니므로 분기별 실행 여부를 커버리지로
말할 수 없다(`not-applicable`). 대신 측정한 것은 **이 테스트가 현재 통과한다**는 사실이다 —
`go test ./internal/protection` exit 0.

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `moduleRoot(t)` | 모듈 루트 해석 | 실패 = `t.Fatal` | AST L19, dormant_test.go:143 |
| `os.ReadDir` | core 패키지 파일 목록 | 실패 = `t.Fatal`(B2) | AST L21 |
| `parser.ParseFile(..., ImportsOnly)` | import만 파싱 | 실패 = `t.Fatal`/반환(B5·B11) | AST L30, L53 |
| `filepath.WalkDir` | `cmd/`·`internal/app/` 전수 순회 | walk 에러 = `t.Fatal`(B15) | AST L46 |
| `os.ReadFile(gateway.go)` | 조립 본문 문자열 검사 | 실패 = `t.Fatal`(B16) | AST L73 |
| `strings.Contains` | 필수·금지 문자열 판정 | — | AST L79, L84 |

**전수 순회라는 점이 중요하다.** `cmd/`와 `internal/app/` **전체**를 걸으므로, a100이 어디에
파일을 두든 이 가드를 지나간다. 새 워커 파일이 조용히 빠져나갈 경로는 없다.

## State mutations and fallbacks

- 이 함수는 **프로덕션 상태를 바꾸지 않는다.** 파일시스템 읽기와 문자열 판정뿐이다.
- 실패 방식이 두 가지다 — `t.Errorf`(계속 진행, 위반을 모두 열거)와 `t.Fatal`(즉시 중단,
  검사 자체가 불가능). 위반은 전자이므로 **한 번의 실행이 모든 위반을 보고한다.**
- fallback 없음. 허용 목록도 금지 목록도 하드코딩이고, 환경변수나 플래그로 완화할 수 없다.
  **이 가드를 넘기는 유일한 방법은 목록을 편집하는 것이며, 그 편집이 diff에 남는다.**

## Safety conclusion

- Safe edit boundary: **L83 금지 목록에서 `protectionofficial.New` 한 줄만 제거**하고,
  제거하는 대신 그 자리에 「a100이 허용한 이유와 무엇이 여전히 금지인지」를 주석으로 남긴다.
  L61 허용 목록과 L78 필수 목록은 **건드리지 않는다.**
- High-risk impact: **yes.** 이 가드가 a071의 「구조적으로 UNWIRED」 보장을 실행 가능한
  형태로 들고 있다. 필요 이상으로 열면 그 보장이 문서로만 남는다.
- **RED 테스트 의무:** 목록에서 한 줄을 빼기 전에, **나머지 3개가 여전히 거부되는지**를
  확인하는 테스트를 남긴다. 빼는 행위 자체가 가드를 약화시키므로 약화의 폭을 고정한다.
