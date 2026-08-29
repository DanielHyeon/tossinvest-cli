# Function Logic Map: `TestFixedFSProberIsTestOnly`

- Source: `internal/testenv/static_test.go`
- AST evidence: `ast.json` (revision=current, L81–141, 분기 9개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: 정의 allowlist 2개화 + 정의 파일 안의 *호출* 금지 (revision=current)

`FixedFSProber`는 파일시스템 내구성 가드가 호출자가 말하는 대로 답하게 만든다 — tmpfs 위의 테스트에서는 정확히 옳고 그 밖에서는 정확히 틀리다. 가드가 존재하는 이유는 어떤 마운트에서 fsync가 fsync가 아니기 때문이고, 그것을 스텁으로 바꾼 프로덕션 경로는 내구성 계약을 주석으로 바꿔 놓은 것이다.

**이 branch range의 변경**: 정의가 **두 개**가 됐다. `internal/candidate`가 자기 사본을 갖는 이유는 `internal/journal`을 import하면 주문 원장의 API가 발굴 패키지의 타입 우주 안으로 들어오고, 그것이 발굴 격리 요구사항("발굴은 주문 경로에 도달할 수 없어야 한다")을 깨기 때문이다. 그래서 검사가 단일 `definition`에서 **명시적 allowlist**로 바뀌었고, allowlist 항목은 실제로 `func FixedFSProber(`를 선언해야 자격을 얻는다 — 그러지 않으면 이 목록이 "부르기만 하는 파일"을 면제하는 수단이 된다.

두 번째 추가: 정의 파일은 **선언에 대해서만** 면제된다. 같은 파일 안의 *호출*은 다른 곳의 호출과 같은 결함이고, 사실 더 있을 법한 결함이다 — `CheckFilesystem`의 nil-prober 분기에 대한 구미 당기는 "수정"이 바로 기본값으로 fixed prober를 넣는 것이고, 그러면 prober를 생략한 **모든** 호출자에 대해 가드가 꺼진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `definitions` | journal/fsguard.go, candidate/fsguard.go | 이 테스트 | 선언이 없는 항목은 `t.Fatalf` |
| `productionFiles(t)` | 非 _test.go 파일 목록 | 저장소 워크 | 읽기 실패는 `t.Fatalf` |
| `docMention` | journal/journal.go | 이 테스트 | 주석에만 있으면 통과 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 정의 allowlist 순회 | `isDefinition` 채움 | — | 이 테스트 |
| B2 | 정의 파일 읽기 실패 | — | `t.Fatalf` | 동일 |
| B3 | allowlist 항목이 `func FixedFSProber(`를 선언하지 않음 | — | `t.Fatalf` — allowlist가 use를 면제 중 | 동일 |
| B4 | 프로덕션 파일 순회 | — | — | 동일 |
| B5 | 파일 읽기 실패 | — | `t.Fatalf` | 동일 |
| B6 | `FixedFSProber` 미언급 | `continue` | — | 동일 |
| B7 | 정의 파일 | — | 선언은 허용 | 동일 |
| B8 | 정의 파일 안의 **호출** | — | `t.Errorf` — 모든 호출자의 가드가 꺼진다 | 동일 + `callsFixedFSProber` |
| B9 | doc 언급 파일이고 주석에만 존재 | `continue` | — | 동일 |
| (fallthrough) | 그 외 프로덕션 사용 | — | `t.Errorf` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `callsFixedFSProber` | 선언과 호출을 구분 | 주석 제거 후 `func ` 접두 여부로 판정 | static_test.go L145 |
| `onlyInComments` | doc 언급 예외 | 주석에만 있으면 use가 아니다 | 같은 파일 |
| `productionFiles` | 검사 대상 수집 | 非 _test.go | 같은 파일 |

## State mutations and fallbacks

- 테스트 — 파일 읽기만.
- 두 allowlist의 표류는 `internal/candidate/fsguard_drift_test.go`가 원장 소스를 읽어서 막는다.

## Safety conclusion

- Safe edit boundary: allowlist 두 항목과 정의/호출 구분. allowlist를 넓히는 것은 내구성 가드를 끄는 파일을 하나 더 허용하는 것이다.
- High-risk impact: yes (원장 경로 — 내구성) — 이 테스트가 약해지면 프로덕션 코드가 파일시스템 내구성 가드를 스텁으로 대체할 수 있고, 그 가드는 주문 원장이 fsync를 신뢰할 수 있는지에 대한 유일한 검사다. 테스트 자체는 실계좌 부작용이 없다.
