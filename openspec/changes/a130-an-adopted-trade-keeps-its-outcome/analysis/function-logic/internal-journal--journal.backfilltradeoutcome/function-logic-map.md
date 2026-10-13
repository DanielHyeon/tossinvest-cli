# Function Logic Map: `Journal.BackfillTradeOutcome`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `843`–`873`, base `233d5d9c`)
- AST evidence: `ast.json`(편집 뒤 `source_sha256` ff51203717c0ed3c…) — branches 6 · returns 6 · defers 0 · go 0; 편집 전 `ast-base.json`(`source_sha256` 0bd43abf96107b99…, branches 6 · returns 6)
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** 성과 행이 없는 닫힌 포지션의 행을 계산해 쓴다. 덮어쓰기 거절.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 0(CodeGraph · grep). 시험 3. a130 의 `BackfillTradeOutcomes` 가 첫 생산 호출자가 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | CLOSED 포지션 | 호출자 | — |
| `model` | costs.Model | 호출자 | 미구성 → B4 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:846` 이미 행 있음 | 없음 | ErrTradeOutcomeExists `:847` | TestABackfillRecoversTheGapAndRefusesToRewriteIt |
| B2 | `:848` else | — | — | — |
| B3 | `:848` 조회 오류(NotFound 아님) | 없음 | err `:849` | 없음 |
| B4 | `:853` compute 거절 | 없음 | ErrInvalidRequest `:854` | TestAnAdoptedPositionWithNoEngineSellFreezesNothing · TestAPartlyEngineSoldAdoptedPositionFreezesNothing |
| B5 | `:857` INSERT 오류 | 없음 | — | 없음 |
| B6 | `:867` UNIQUE 위반 | 없음 | ErrTradeOutcomeExists `:868`; 그 밖 `:870` | 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `TradeOutcomeOf` | 존재 확인 | — | AST |
| `computeTradeOutcome` | 계산 | — | AST |
| `j.db.ExecContext` | INSERT | — | AST |

Live config binding: 없음.

## State mutations and fallbacks

- trade_outcomes INSERT(트리거가 UPDATE 금지).

## a130 편집 계획 (2판)

compute 반환형이 사유가 되므로 B4 조건은 `reason != ""`, 오류 문장에 사유 코드를 덧붙임(감싸는 ErrInvalidRequest 그대로). 그 밖 무변경. 생산 호출자 0 그대로(2판은 backfill 도구 없음 — design D4).

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 성과 산식 무변경, 거절 셋 추가(보수 방향), 거절 기록 추가.

## 편집 뒤 (a130 4판)

- 편집 뒤 `889`–`919` · `source_sha256` ff51203717c0ed3c… — branches 6 · returns 6
  (편집 전 6 · 6, 편집 전 AST 는 `ast-base.json`). 위 분기 표의 좌표는 편집 전, 아래 BTM 은 편집 뒤 좌표.
