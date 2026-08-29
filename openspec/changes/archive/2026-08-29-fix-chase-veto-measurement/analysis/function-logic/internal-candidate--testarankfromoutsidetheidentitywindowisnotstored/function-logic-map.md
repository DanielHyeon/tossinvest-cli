# Function Logic Map: `TestARankFromOutsideTheIdentityWindowIsNotStored`

- Source: `internal/candidate/store_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1159–1190, 분기 6개)
- Risk scan: `risk-pattern-report.md`

이 테스트의 편집은 **`NoteFirstRank`의 시그니처 변경 반영**이다 — 위치 인자
`(rank, total, at, source)`가 `storedFirstRank(...)`가 만드는 `FirstRank` 하나로 합쳐졌다.

`storedFirstRank`는 두 자격 사실을 **평범한 측정된 경우**로 채운다(소스가 직전 읽기를 갖고
있었고 이 심볼이 거기 있었다, 그리고 읽기는 온전했다). zero value로 두면 이 파일의 모든
sighting이 미측정이 되어 테스트가 **아무것도 재지 않으면서 통과**한다.

재는 것: `first_seen_at`의 ±TTL 창 **밖** 읽기는 오류가 아니라 **조용히 저장되지 않는다**.
순위를 싣지 않는 소스가 처음 올린 후보가 나중에 순위 목록에 나타나는 것은 일상이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `late` | 창 밖 instant | 이 테스트 | 저장되지 않고 오류도 아니다 |
| `inside` | 창 안 instant | 이 테스트 | 저장된다 |
| `storedFirstRank(...)` | 두 자격을 측정된 경우로 | `veto_test.go` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Promote` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 창 밖 `NoteFirstRank`가 오류를 냈다 | 없음 | `t.Fatalf` — 오류가 아니어야 한다 | 자체 실행 |
| B3 | 창 밖 호출이 무언가를 돌려줬다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | 저장소에 실제로 쓰였다 | 없음 | `t.Errorf` | 자체 실행 |
| B5 | 창 안 `NoteFirstRank` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 | 창 안 위치가 저장되지 않았다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.Promote` / `s.NoteFirstRank` / `s.FirstRank` | 저장 경로 | 오류는 `t.Fatal` | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출 2곳의 인자 형태.
- High-risk impact: no (테스트 전용).
