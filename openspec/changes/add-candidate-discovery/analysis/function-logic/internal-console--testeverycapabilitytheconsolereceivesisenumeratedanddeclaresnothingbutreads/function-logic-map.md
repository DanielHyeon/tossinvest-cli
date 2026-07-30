# Function Logic Map: `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

주입 표면 검사의 본체. 검사 단위가 **인터페이스가 아니라 `Options` 필드**라는 것이 전부다 — 불변식은 '브로커 seam이 좁다'가 아니라 '**이 콘솔이 받는 모든 능력이 여기 열거돼 있다**'이다. 열거에 없는 필드는 그것만으로 실패한다.

**잡는 것**: ① 열거되지 않은 새 `Options` 필드, ② 열거된 필드의 seam이 허용 밖 메서드를 선언하거나 약속한 메서드를 잃은 것, ③ 필드에서 **도달 가능한 모든 타입**의 이름이 금지 동사를 쓰는 것, ④ 도달 가능한 인터페이스의 embed, ⑤ `Options`의 구조체 embed, ⑥ 필드가 사라졌는데 남아 있는 allowlist 항목, ⑦ 패키지 어디든 `verifylive.Broker`라는 이름.

**잡지 못하는 것(측정된 경계)**: 동사 검사는 `mutationVerbs`에 **철자로 적힌** 것만 본다. `Liquidate`·`Execute`·`Unwind`·`Square`·`Dispose` 같은 이름은 통과한다. 그것이 이 가드의 실질 검사가 **메서드 집합**인 이유이고, 동사 검사는 그 위의 보조 장치라고 `capability.VerbExemptions`의 문서가 스스로 적고 있다. `Options` **밖**으로 오는 능력은 `TestNoCapabilityReachesTheConsoleAroundOptions`가 맡고, 그쪽에는 메서드 집합 검사가 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `consoleCapabilities` | 필드 이름 → 허용 메서드 집합 + 동사 예외 | static_test.go:842 | 필드가 없으면 실패, 항목이 남아 있으면도 실패(양방향) |
| `optionsFields(t, files)` | `Options`의 필드 전부 | 패키지 소스 | 0개면 `t.Fatal` |
| `packageTypes(files)` | 패키지가 선언한 모든 타입 | 같은 곳 | 비어 있으면 `methodless`가 false를 답해 소리 내어 실패한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `len(fields) == 0` | 없음 | `t.Fatal` | positive control |
| B2 | `for _, field := range fields` | 없음 | 없음 | 현재 26개 필드 |
| B3 | `for _, name := range field.Names` | `declared[name] = true` | 없음 | 같은 위 |
| B4 | `!enumerated` | `t.Errorf` + 동사 검사(예외 없이) | continue | 새 필드 추가 변이 — 이름이 무엇이든 열거되지 않았다는 이유로 실패한다 |
| B5 | `len(field.Names) == 0` | `t.Error` — 구조체 embed | 없음 | embed 삽입 변이 |
| B6 | `for name := range consoleCapabilities` | `stale` 수집 | 없음 | 필드 제거 변이 |
| B7 | `!declared[name]` | `stale` 추가 | 없음 | 같은 위 |
| B8 | `for _, name := range stale` | `t.Errorf` | 없음 | 같은 위 |
| B9 | `for name, fileSrc := range packageFiles(t)` | 없음 | 없음 | 패키지 전 파일 |
| B10 | `strings.Contains(code, "verifylive.Broker")` | `t.Errorf` | 없음 | 광폭 브로커 이름 사용 변이 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 **비테스트** `.go` 파일 전부 | 0개면 `t.Fatal` | static_test.go:36 |
| (런타임 무접촉) | 가드는 **선언**을 읽는다 — 존재하지만 호출되지 않는 메서드가 정확히 잡아야 할 대상이다 | reflect를 쓰지 않는다 | static_test.go:936 |
| `optionsFields` / `packageTypes` / `parsedPackage` | 검사 입력 조립 | 각자 positive control 또는 소리 내는 실패를 갖는다 | static_test.go:1355, 1383, 1397 |
| `checkVerbsExcept` / `checkVerbs` | 이름 필터 | 예외는 **철자 전체** 키 | static_test.go:1341, 1329 |
| `checkCapability` | 필드별 메서드 집합 검사 | 고정점까지 해석 | static_test.go:1047 |
| `nonCommentLines` | 주석 제외 문자열 검사 | 주석의 이름이 발견을 만들지 않는다 | static_test.go:1680 |

## State mutations and fallbacks

- 없음(판정 전용).
- `consoleCapabilities`는 26개 필드를 열거한다 — 평문 데이터 10, func 타입 seam 8, 인터페이스 seam 6(`Handoff`·`Holdings`·`Settings`·`GateLimits`·`Signals`·`Orders`), `Out`은 `io.Writer`.
- 동사 예외는 `Orders` 하나에만 있고 다섯 철자뿐이며, 각각에 근거 문장이 붙는다. `TestTheOrdersSeamIsTheOnlyCapabilityWithVerbExemptionsAndTheyAreEnumerated`가 그 크기를 고정한다.

## Safety conclusion

- Safe edit boundary: 신설(대체). 이전 가드의 `verifylive.Broker` 문자열 검사는 그대로 승계했고, 나머지는 검사 단위를 필드로 옮겨 새로 썼다.
- High-risk impact: yes (주문 능력 주입 차단 — `verifylive.Broker`는 `PlaceOrder`·`CancelOrder`·`ModifyOrder`와 조건주문 변조 셋을 갖고 있고, 그것을 읽기 화면에 건네면 '콘솔은 주문을 내지 않는다'가 타입의 사실이 아니라 핸들러의 습관이 된다)
