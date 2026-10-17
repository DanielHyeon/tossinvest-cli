# Function Logic Map: `ExitObserver.sellIntent`

- Source: `internal/app/engine/exitloop.go` (1571-1597), 판정 HEAD `102d4e99` (= base, a087 task P1.1)
- AST evidence: `ast.json` — branches 4 (B1 `:1573` · B2 `:1576` · B3 `:1581` · B4 `:1585`), returns 4, calls 6
- Risk scan: `risk-pattern-report.md` (이 함수 안의 발견 0 — 파일 단위 `go-panic :1799` 는 범위 밖)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `observed` | 양의 십진 문자열 | 유일한 생산 경로: `record` `:1190` `judgement.ObservedPrice = snapshot.ObservedPrice` → `submit` `:1301` → `:1382` (재대입 없음) | 비면 B1 로 기준선 폴백 |
| `m.state.Baseline` | 십진 문자열 | 원장 `exit_states` | `observed` 도 비면 B2 거부 |
| `quantity` | `applyFloor` 뒤 수량 | `submit` `:1345` | B3 거부 |
| `m.position.Market` | `kr`/`us` | 원장 | `currencyFor` 로 통화 모드만 결정 |

**불변식(측정·AST 로 확인, 아래 「모집단」)**: 생산 경로에서 `observed` 는 언제나 양수다. `snapshot.ObservedPrice` 는
`EvaluateRatchetSnapshot`(`snapshot.go:227` ← `in.Input.ObservedPrice`)과 `EvaluateLadderSnapshot`(`:155` ← `eval.ObservedPrice`)이
채우고, 두 함수는 각각 `EvaluateRatchet`(`snapshot.go:189`) · `EvaluateLadder`(`:120`) 가 오류 없이 돌아온 뒤에만 스냅숏을 만든다.
그 두 평가기는 성공 반환 전에 `positive("observed price", …)`(`ratchet.go:347` · `ladder.go:333`)를 **무조건** 지난다 — AST 상
그 앞의 반환(`ratchet.go` 341·345, `ladder.go` 309·314·319·322·326)은 전부 오류 반환이다. `positive` 는 파싱 실패와 `Sign() <= 0`
을 거부한다(`ratchet.go:592-601`). 따라서 `strings.TrimSpace(observed) == ""` 는 생산 경로에서 참이 될 수 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:1573` `price == ""` (관측가 공백) | `price` 를 기준선으로 대체 | 계속 | 생산 도달 불가 — 직접 호출 단위 시험만 가능 |
| B2 | `:1576` 기준선도 공백 | 없음 | `:1577` 오류 "has no price to submit a liquidation at" → `submit` `:1384` `alertRefused` + `ProposalRefused` 해제 | 생산 도달 불가 (B1 이 선행 조건) |
| B3 | `:1581` 수량 `floatOf` 실패 | 없음 | `:1582` 오류 | 직접 호출 |
| B4 | `:1585` 가격 `floatOf` 실패 | 없음 | `:1586` 오류 (`positive` 는 통과하지만 float64 범위를 넘는 값) | 직접 호출 |
| 정상 | — | 없음 | `:1588` LIMIT·sell·`Price = observed` | 기존 e2e 시험 다수 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` ×2 | 공백 정규화 | 없음 | AST `:1572` `:1574` |
| `fmt.Errorf` | B2 거부문 | 없음 | AST `:1577` |
| `floatOf` ×2 | 십진 → float64 | 오류 반환 | AST `:1580` `:1584` |
| `currencyFor` | 시장 → 통화 모드 | 없음 | AST `:1595` |

브로커 호출·원장 쓰기·설정 읽기 없음. CodeGraph 1.6.0: 호출자 `submit` 1 (`codegraph callers sellIntent`), `submit` 의 비시험
호출자 `record` 1, `record` 의 호출자 `judgeRatchet`·`judgeLadder`.

## State mutations and fallbacks

- 상태 변경 없음(순수 함수). 폴백은 B1 하나(관측가 → 기준선)이고 **생산에서 도달하지 않는다**.
- 커버리지 실측(2026-09-30, `go test ./internal/app/engine/ -count=1 -coverprofile`, 157 s, 70.7%):
  `1573.17-1575.3` **0**, `1576.17-1579.3` **0**, `1581.16-1583.3` 0, `1585.16-1587.3` 0, 정상 경로 `1588.2-1596.8` 1.
  저장소 전체에서 거부문 문자열을 단언하는 시험 0(`grep "no price to submit"` → 정의 1줄뿐).

## Safety conclusion

- Safe edit boundary: B2 앞에 하한가 단(D2a)을 넣는 편집은 **도달 불가 분기 뒤**에 놓인다 — 생산 동작 변화 0. 편집은 중단하고
  Manager 판단을 요청했다(`issues.md` I-P1).
- High-risk impact: yes(손절 제출 경로) — 단 현재 HEAD 에서 B1·B2 는 우연이 아니라 평가기의 `positive` 검사가 닫아 둔 문이다.
