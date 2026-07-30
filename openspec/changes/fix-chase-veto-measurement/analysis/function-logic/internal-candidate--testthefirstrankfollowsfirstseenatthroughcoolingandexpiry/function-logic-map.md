# Function Logic Map: `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`

- Source: `internal/candidate/store_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1066–1148, 분기 18개)
- Risk scan: `risk-pattern-report.md`

이 테스트의 편집은 **`NoteFirstRank`의 시그니처 변경 반영**이다 — 위치 인자
`(rank, total, at, source)`가 `storedFirstRank(...)`가 만드는 `FirstRank` 하나로 합쳐졌다.

`storedFirstRank`는 두 자격 사실을 **평범한 측정된 경우**로 채운다(소스가 직전 읽기를 갖고
있었고 이 심볼이 거기 있었다, 그리고 읽기는 온전했다). zero value로 두면 이 파일의 모든
sighting이 미측정이 되어 테스트가 **아무것도 재지 않으면서 통과**한다.

재는 것: 최초 순위가 `first_seen_at`의 생애를 따라간다 — 냉각을 지나 보존되고, 재진입에
덮이지 않으며, **만료에 초기화된다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 생애 | 승격 → 냉각 → 재승격 → 만료 → 재승격 | 주입 clock | 각 단계에서 `FirstRank`를 확인 |
| `storedFirstRank(...)` | 두 자격을 측정된 경우로 | `veto_test.go` | zero면 마지막 단언이 미측정이 되어 통과 |
| 기대 백분위 | `96.666666666666` (5 of 150) | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Promote` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `NoteFirstRank` 오류(첫 기록) | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | `Cool` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 | 냉각 후 순위가 148이 아니다 | 없음 | `t.Errorf` — 냉각은 보존한다 | 자체 실행 |
| B5 | 재승격 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 | `first_seen_at`이 바뀌었다 | 없음 | `t.Errorf` | 자체 실행 |
| B7 | 재진입 시 `NoteFirstRank` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 | `FirstRank` 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B9 | 재진입이 저장된 순위를 덮었다 | 없음 | `t.Errorf` — write-once | 자체 실행 |
| B10 | 상태 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B11 | 만료 상태가 아니다 | 없음 | `t.Fatalf` | 자체 실행 |
| B12 | 만료 후 재승격 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B13 | 만료 후 `FirstRank` 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B14 | 만료가 순위를 지우지 않았다 | 없음 | `t.Errorf` — **`Promote`의 reset 절** | 자체 실행 |
| B15 | instant·source도 지워졌는가 | 없음 | `t.Errorf` | 자체 실행 |
| B16 | 새 생명의 `NoteFirstRank` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B17 | 새 순위가 5가 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B18 | 새 생명의 sighting이 측정되지 않거나 백분위가 다르다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.Promote` / `s.Cool` / `s.NoteFirstRank` / `s.FirstRank` / `s.Candidate` | 생애 전이 | 오류는 `t.Fatal` | ast.json calls |
| `MeasureFirstSighting(...)` | 마지막 단언 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — `NoteFirstRank` 호출 3곳의 인자 형태.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — `first_seen_at`과 그 옆의 write-once 칼럼이 생애를 따라가는지가 이 패키지의 존재 이유다.
