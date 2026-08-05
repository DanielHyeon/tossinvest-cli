# Function Logic Map: `TestTheReloadCellAndTheMetaTagAreOneFact`

- Source: `internal/console/status_strip_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 a080이 고치는 대상이 아니라, a080이 **눈을 뜨게 해야 하는** 테스트다.
독립 리뷰 F2가 지적한 결함이 여기 있었다. 편집 범위는 재로드 주기를 단정하는
가지 하나이고 나머지 가지는 손대지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `consoleScreens` | 콘솔이 셸을 그리는 모든 화면과 그 재로드 주기 | 같은 파일의 테이블. a080이 두 라인 화면을 `lineRefreshSeconds()`로 바꿨다 | 테이블에 없는 화면은 이 테스트가 보지 않는다 |
| `screen.reload` | 0이면 스스로 재로드하지 않는 화면, 양수면 초 단위 주기 | `RefreshSeconds()`가 만드는 meta 태그와 같은 값이어야 한다 | 0인데 meta가 있으면 B2가 잡는다 |
| 렌더된 `data-reload` 셀 | `<n>초마다` 또는 `걸려 있지 않음` | `templates.go`의 `{{.RefreshSeconds}}초마다` | 숫자가 없으면 `reloadPeriodIn`이 `t.Fatalf` |
| 불변식 | **표시줄이 말하는 주기와 meta 태그의 주기는 한 값이다** | 셀이 `.RefreshSeconds`를 직접 읽는다 | 두 값이 갈라지면 화면은 멎었는데 여전히 갱신 중인 것처럼 보인다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range consoleScreens` — 화면마다 1회 | 없음 (`h.get`은 GET 렌더) | 없음 | 자기 자신 (테이블 전체를 돈다) |
| B2 | meta 태그 유무와 셀의 `data-reload="on"`이 다르다 | 없음 | `t.Errorf` | 자기 자신 |
| B3 | 재로드가 없는데 셀이 `걸려 있지 않음`을 안 쓴다 | 없음 | `t.Errorf` | 자기 자신 |
| B4 | 재로드가 있다 → 셀이 말하는 주기를 **읽는다** | 없음 | 숫자가 없으면 `reloadPeriodIn`이 `t.Fatalf` | 자기 자신 (모든 재로드 화면) |
| B5 | 읽어낸 주기 ≠ 선언된 주기 | 없음 | `t.Errorf` | 변이 M-F2 (템플릿을 `15초마다` 고정) |

**B4·B5가 이 change의 편집이다.** 이전 형태는
`if on && !strings.Contains(cell, strconv.Itoa(screen.reload)+"초마다")` 한 줄이라
두 가지가 분리되어 있지 않았고, 포함 검사는 접미사에 눈이 멀었다 — a080이 라인
화면을 `5초마다`로 만든 뒤에는 `15초마다`를 그린 셀이 `screen.reload == 5`를
만족했다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newHarness` / `h.authenticate` | 셸을 그리는 최소 콘솔 | 실패 시 `t.Fatal` | AST calls |
| `h.get` / `body` | 화면을 렌더해 HTML을 얻는다 | httptest, 네트워크 없음 | AST calls |
| `cellOf` | 표시줄에서 `data-reload` 셀만 잘라낸다 | 셀이 없으면 `t.Fatalf` | AST calls |
| `reloadPeriodIn` | **a080 신규** — 셀이 말하는 주기를 정수로 읽는다 | 숫자가 없으면 `t.Fatalf` | `line_cadence_test.go` |
| `strings.Contains` | on/off와 문구 판정에만 남는다 | 없음 | AST calls |

## State mutations and fallbacks

- 이 함수는 상태를 바꾸지 않는다. 브로커·journal·엔진에 닿지 않고 GET 렌더만 한다.
- fallback 없음. 실패는 전부 `t.Errorf`(계속)와 `reloadPeriodIn`의 `t.Fatalf`(중단)다.

## Safety conclusion

- Safe edit boundary: **B4·B5 안쪽뿐.** B1의 테이블 순회, B2의 on/off 대조, B3의
  문구 대조는 글자 그대로 보존한다. 셀을 자르는 `cellOf`도 손대지 않는다 — 그
  함수를 건드리면 이 change와 무관한 표시줄 테스트 전체가 이 change의 증거
  대상이 된다. `reloadPeriodIn`을 `line_cadence_test.go`에 둔 것도 같은 이유다.
- High-risk impact: **no.** 테스트 전용 경로이고 production 코드에 도달하지 않는다.
  다만 이 테스트가 눈이 멀면 **운영자가 보는 주기 표시가 거짓이어도 통과**하므로,
  a080처럼 주기를 바꾸는 change에 대해서는 이 함수가 마지막 방어선이다.
