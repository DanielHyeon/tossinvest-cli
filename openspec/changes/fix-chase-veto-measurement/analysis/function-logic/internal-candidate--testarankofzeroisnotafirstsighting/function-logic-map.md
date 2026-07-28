# Function Logic Map: `TestARankOfZeroIsNotAFirstSighting`

- Source: `internal/candidate/store_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1195–1223, 분기 7개)
- Risk scan: `risk-pattern-report.md`

이 테스트의 편집은 **`NoteFirstRank`의 시그니처 변경 반영**이다 — 위치 인자
`(rank, total, at, source)`가 `storedFirstRank(...)`가 만드는 `FirstRank` 하나로 합쳐졌다.

`storedFirstRank`는 두 자격 사실을 **평범한 측정된 경우**로 채운다(소스가 직전 읽기를 갖고
있었고 이 심볼이 거기 있었다, 그리고 읽기는 온전했다). zero value로 두면 이 파일의 모든
sighting이 미측정이 되어 테스트가 **아무것도 재지 않으면서 통과**한다.

재는 것: `NoteFirstRank`의 경계 거부들 — 0·음수·목록보다 큰 순위, instant 없음, source 없음,
후보 없음.

이 change가 **거부를 하나 더 만들었지만**(요청 행 수가 음수) 이 표에는 넣지 않았다. 그것은
`negative_request_test.go`가 따로 잡는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 불량 위치 5종 | {0,150} {12,0} {-1,150} {0,0} {151,150} | 이 테스트 | 전부 오류여야 한다 |
| instant 없음 / source 공백 / 없는 후보 | 각각 거부 | 이 테스트 | 마지막은 `ErrNoCandidate` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Promote` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `for _, bad := range {...}` — 불량 위치 5종 | 없음 | — | 자체 실행 |
| B3 | 불량 위치가 수락됐다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | instant 없는 위치가 수락됐다 | 없음 | `t.Error` | 자체 실행 |
| B5 | source 없는 위치가 수락됐다 | 없음 | `t.Error` | 자체 실행 |
| B6 | 없는 후보에 대한 기록이 `ErrNoCandidate`가 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B7 | 없는 후보의 `FirstRank`가 뭔가를 돌려줬다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.NoteFirstRank` / `s.FirstRank` | 거부 경로 | 오류를 단언한다 | ast.json calls |
| `errors.Is(err, ErrNoCandidate)` | sentinel 확인 | 순수 | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출 8곳의 인자 형태.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk 인접 — write-once 칼럼의 경계다.
