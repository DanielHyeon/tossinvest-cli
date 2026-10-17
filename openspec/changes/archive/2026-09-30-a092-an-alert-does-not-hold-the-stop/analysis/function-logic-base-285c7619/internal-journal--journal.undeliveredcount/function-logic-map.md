# Function Logic Map: `Journal.UndeliveredCount`

- Source: `internal/journal/outbox.go` (408-415)
- AST evidence: `ast.json` — branches 1, returns 2, calls 3, assignments 1,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판의 배달 루프가 "남은 미전달 수"를
운영 모드 승격과 게이트 해제의 입력으로 쓰기로 했으므로, 그 수가 무엇인지를
여기서 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 호출자의 것 | 17판에서는 배달 루프의 것 | 취소되면 B1 |
| 세는 대상 | **`state = PENDING`인 행 전부** | `:411` | — |
| 범위 | **계좌·심볼·종류 구분 없음 — 전역이다** | 같은 쿼리 | — |

**전역이라는 것이 계약이다.** `internal/obs/measurement_test.go:136`이 이미
그 사실에 의존한다: 미전달 알림 하나가 남아 있으면 그것이 어느 조건에서 왔든
이 수가 0이 아니다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:410` | `QueryRowContext`/`Scan` 실패 | 없음 | 오류 `:412` |
| — `:414` | — | 없음 | `n, nil` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.QueryRowContext` `:410` | `COUNT(*)` | 로컬 SQLite | AST calls |
| `Scan` `:410` | int 추출 | 같은 오류로 합쳐짐 | AST calls |
| `fmt.Errorf` `:412` | 감싸기 | — | AST calls |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| — | — | **없음. 읽기 전용이다** |

- fallback 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: 간접적. 이 수 자체가 안전을 만들지는 않고,
  **이 수를 읽고 게이트를 여는 코드**가 만든다.
- **측정한 사실 하나를 남긴다.** `UndeliveredCount`의 프로덕션 호출자는
  `notifier.go:460`(`Flush`)과 `notifier.go:506`(`Acknowledge`) 둘뿐이고,
  **그 둘 다 프로덕션 호출자가 없다**(`rg`로 `internal/`·`cmd/` 전체 확인).
  따라서 **오늘 프로덕션에서 이 함수는 한 번도 불리지 않는다.**
  `:407`의 doc comment *"the number the entry gate reacts to"*는
  **의도이지 현재 동작이 아니다.**
- 17판이 그것을 바꾼다: 감독 아래의 배달 루프가 주기마다 이 함수를 부르고,
  그 수로 승격과 해제를 판정한다. **그러므로 이 함수의 첫 프로덕션 호출자를
  a092가 만든다.**
