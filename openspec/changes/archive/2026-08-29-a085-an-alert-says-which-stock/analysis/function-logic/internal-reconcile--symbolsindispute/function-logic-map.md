# Function Logic Map: `symbolsInDispute`

- Source: `internal/reconcile/mismatch.go` (lines 434–453)
- AST evidence: `ast.json` (`source_sha256: 49fcb007cab394f144230db5b1bf330b95e2303dfef48648edc806501edd2a57`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

이 비교가 아직 이견을 가진 심볼의 집합. credit 만료를 비교 단위가 아니라 심볼
단위로 만드는 것이 이 함수다. 개정 2가 `diff.ExternalPos`를 더했다: `BlocksEntry()`가
세지 않더라도 외부 보유는 살아 있는 이견이다. 진입 게이트의 질문("이 비교가 신규 진입을
막는가")과 해제 규칙의 질문("이 비교가 이 심볼에 동의하는가")은 다른 질문이고, 엔진이
인스턴스를 갖지 않은 보유를 동의로 취급하면 재분류된 이견이 자기 블록을 푼다 (D8).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff.Quantities` | 수량 불일치 목록 | 대사 비교 | 심볼 대문자·trim 후 집합에 넣는다 |
| `diff.MissingOrders` | 누락 주문 목록 | 대사 비교 | 동일 |
| `diff.ExternalPos` | 외부 보유 목록 | 대사 비교 | 동일 — 개정 2가 추가 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (437) `range` — for _, mismatch := range diff.Quantities { | 본문 참조 | 아래 Branch Test Map |
| B2 | (440) `range` — for _, missing := range diff.MissingOrders { | 본문 참조 | 아래 Branch Test Map |
| B3 | (449) `range` — for _, external := range diff.ExternalPos { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToUpper` / `strings.TrimSpace` | 심볼 정규화 | 순수 | AST `calls` |

## State mutations and fallbacks

- 없음. 새 map을 만들어 반환한다.

## Safety conclusion

- Safe edit boundary: 세 목록. 하나를 빼면 그 종류의 이견이 동의로 읽히고 블록이 스스로 풀린다.
- High-risk impact: yes — 진입 차단 해제 판정.
