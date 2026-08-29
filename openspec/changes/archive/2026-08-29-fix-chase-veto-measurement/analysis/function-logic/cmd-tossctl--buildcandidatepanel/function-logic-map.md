# Function Logic Map: `buildCandidatePanel`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `base`, L463–488, 분기 5개)
- Risk scan: `risk-pattern-report.md`

**이 파일에서는 삭제되었다.** 함수는 `cmd/tossctl/candidatepanel.go`로 **옮겨졌고**(신규
파일이라 FLM 대상에서 면제된다), 여기 실린 `ast.json`은 base commit `b268593`의
`cmd/tossctl/candidate.go`에서 뜬 것이다(`"revision": "base"`).

## 왜 옮겼나 — 소비자 가드(리뷰 F5)

`internal/candidate/consumer_guard_test.go`는 chase verdict를 명명할 수 있는 **파일 목록**을
고정하고, 그 파일들이 주문 동사를 **함께** 명명하지 않는지 검사한다. 검사가 파일의 텍스트를
읽는 이유는 `cmd/tossctl`이 하나의 Go 패키지이고 이미 `execgw`·`orderintent`·`trading`을
import하므로 import 그래프가 아무것도 말해 주지 않기 때문이다.

selector 스캔이 볼 수 있는 것은 `official.New`다 — `client.CreateConditionalOrder(…)`는
그 시점에 패키지 이름이 사라진 메서드 호출이라 보이지 않는다. 그래서 **자격증명을 모든 쓰기를
가진 값으로 바꾸는 그 한 줄**이 verdict를 읽는 파일에 있어서는 안 된다. `candidate.go`는
veto 블록을 렌더하므로 여기서 나가야 했다.

옮기면서 본문에 바뀐 것은 **`clock.System()` 인자 하나**다 — `candidatesrc.Panel`이 시계를
받게 되었기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | CLI 옵션 | 호출자 | 경로 해석 실패는 오류 |
| Open API 자격증명 | 존재해야 한다 | `official.LoadCredentials` | 없으면 오류 — 공식 순위는 필수 소스다 |
| WTS 세션 | 선택 | `session.NewFileStore` | 없으면 패널이 저하될 뿐 멈추지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `resolveOpenAPIPaths` 오류 | 없음 | 오류 | **커버 없음** |
| B2 | `LoadCredentials` 오류 | 없음 | 오류 | **커버 없음** |
| B3 | `creds == nil` | 없음 | 'openapi login' 안내가 붙은 오류 | **커버 없음** |
| B4 | 세션이 읽히고 non-nil | `wts` 설정 시도 | — | **커버 없음** |
| B5 | config가 읽힘 | `wts = tossclient.New(...)` | — | **커버 없음** |

B4·B5가 중첩된 것은 의도다 — 세션이 있어도 config를 못 읽으면 WTS 없이 진행한다. 부재가
일상이라는 것이 spec Requirement 5의 hard half다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `official.LoadCredentials` | 자격증명 | 오류 래핑 | ast.json calls (base) |
| `official.New(*creds, tokenFile)` | **모든 쓰기를 가진 클라이언트** | — | ast.json calls (base) — 이 호출이 파일 이동의 이유다 |
| `tossclient.New(...)` | WTS 클라이언트 | — | ast.json calls (base) |
| `candidatesrc.Panel(market, client, client, wtsPopularityReader(wts))` | 소스 목록 | — | ast.json calls (base). 현재 판본은 `clock.System()`을 하나 더 넘긴다 |

## State mutations and fallbacks

- 없음 — 슬라이스 하나를 만든다. 클라이언트 생성이 유일한 side effect이고 그것은 네트워크 호출이 아니다.
- fallback: WTS 부재는 nil reader이고 `Panel`이 그 소스를 빼는 것으로 처리한다.

## Safety conclusion

- Safe edit boundary: **파일 이동**(candidate.go → candidatepanel.go) + `clock.System()` 인자 1개. 분기·오류·문구 무변경.
- High-risk impact: **yes 인접** — `official.New`가 주문·취소·정정을 가진 클라이언트를 만든다. 이 change의 편집은 그 줄을 **verdict를 읽지 않는 파일로 옮긴 것**이고, 방향은 격리를 강화하는 쪽이다.
