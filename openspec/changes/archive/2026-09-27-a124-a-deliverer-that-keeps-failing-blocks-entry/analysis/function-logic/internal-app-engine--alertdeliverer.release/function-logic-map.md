# Function Logic Map: `alertDeliverer.release`

- Source: `internal/app/engine/alertdelivery.go` (600-606)
- Revision: current — a124 구현 로트(2026-09-27) 편집 뒤 재추출; source_sha256 `5791a31af9d2407935ea71edbdbf9a14cbd2025fce5da9866b1d586f8e8da0cf`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 1 · returns 0 · calls 6
- Exact AST return positions: (none)
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` · `token` | 이 사이클이 끝낸 행의 임차 | `deliverOne` 하위 | 오류는 로그 |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 603:2 | `if _, err := d.led().ReleaseAlertClaim(relCtx, id, token); err != nil {` |

### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)

| Base branch | Base anchor | Base source | Current branch |
|---|---|---|---|
| b1 | if at 603·2 | `` | — (없어짐 · 하위 함수로 옮김) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `d.led().ReleaseAlertClaim` | 임차 반납(떼어 낸 ctx, 5 s) | 오류는 로그만 — 반납 Applied 는 기록 실패 계수를 지우지 않음(X3) |

## State mutations and fallbacks

- 편집은 한 줄 — `d.Journal` → `d.led()`(시험 결함 주입용 원장 래퍼). 생산 값은 같은 `*journal.Journal`.

## Safety conclusion

- Safe edit boundary: 동작 불변(원장 대상만 인터페이스로).
- High-risk impact: low — 임차 반납뿐.
