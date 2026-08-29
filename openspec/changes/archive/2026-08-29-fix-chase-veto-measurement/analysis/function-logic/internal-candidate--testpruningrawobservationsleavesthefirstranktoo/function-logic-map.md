# Function Logic Map: `TestPruningRawObservationsLeavesTheFirstRankToo`

- Source: `internal/candidate/store_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1229–1269, 분기 9개)
- Risk scan: `risk-pattern-report.md`

이 테스트의 편집은 **`NoteFirstRank`의 시그니처 변경 반영**이다 — 위치 인자
`(rank, total, at, source)`가 `storedFirstRank(...)`가 만드는 `FirstRank` 하나로 합쳐졌다.

`storedFirstRank`는 두 자격 사실을 **평범한 측정된 경우**로 채운다(소스가 직전 읽기를 갖고
있었고 이 심볼이 거기 있었다, 그리고 읽기는 온전했다). zero value로 두면 이 파일의 모든
sighting이 미측정이 되어 테스트가 **아무것도 재지 않으면서 통과**한다.

재는 것: D11의 두 계층 보존 — 원시 관측이 정리돼도 `candidates`의 최초 순위는 남고,
그 위치로 sighting이 여전히 **측정된다.**

마지막 단언(B9)이 `storedFirstRank`의 자격 채움에 직접 의존한다. zero value였다면
`NEW_ENTRANT_UNKNOWN`으로 미측정이 되어 이 테스트가 아무것도 재지 않으면서 통과했을 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 관측 | 정리 대상 instant | `RecordObservations` | 정리 후 0행 |
| 최초 순위 | 148 of 150 | `NoteFirstRank` | 정리 후에도 남는다 |
| `storedFirstRank(...)` | 두 자격을 측정된 경우로 | `veto_test.go` | **B9가 여기 의존한다** |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `RecordObservations` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `Promote` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | `NoteFirstRank` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 | `PruneObservations` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 | 정리 후 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 | 관측이 남아 있다 | 없음 | `t.Errorf` | 자체 실행 |
| B7 | `FirstRank` 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 | 최초 순위가 사라졌다 | 없음 | `t.Errorf` — **D11의 두 번째 계층** | 자체 실행 |
| B9 | 정리 후 sighting이 미측정이다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.PruneObservations` | 원시 행 정리 | `observations`만 건드린다 | ast.json calls |
| `MeasureFirstSighting(...)` | 정리 후에도 측정 가능 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출 1곳의 인자 형태.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — 보존 계층이 뒤섞이면 `first_seen_at`이 사라진다.
