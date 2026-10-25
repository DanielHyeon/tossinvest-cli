# Function Logic Map: `signalsView.freshness`

- Source: `internal/console/chrome.go` (323-334) — 편집 뒤(a173 rebase 후 main).
- Qualified: `signalsView.freshness`
- AST evidence: `ast.json` (`source_sha256` 367463d91837a4ab…) — AST branches 1 · returns 2
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 완료 게이트 직전(2026-10-25, 종결 에이전트). tasks 3.1 의 비례 원칙 면제는 required 7 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임(FLM 이 구현 판단에 쓰이지 않음).

**역할.** 발굴 저장소 시각을 표시줄 신선도로 읽는다(tone 없음).

a173 편집: 읽지 못한 갈래(B1)의 반환에 `Code: unmeasuredReason(v.Read.Code())` 추가. 판정 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v.Read` | Known 또는 사유 코드 | 발굴 저장소 읽기 | 미지 → unavailable + 사유 코드 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a173 |
|---|---|---|---|---|
| B1 | `324:2` | if | !Known → unavailable · 사유 코드 | a173: Code 추가. 거짓 갈래는 327행 cache 반환 |

Exact AST return positions: 325:3, 327:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `v.Read.Known` · `v.Read.Code` | 읽기 상태 · 사유 | 순수 | AST calls 324 · 325 |
| `unmeasuredReason` | 사유 형 변환(a173) | 순수 | AST calls 325 |
| `time.Duration` · `candidate.DefaultWatchInterval.String` | 나이 · Hold 문구 | 순수 | AST calls 330 · 331 |

## State mutations and fallbacks

- 값 반환만. 상태 변경 0.

## Safety conclusion

- Safe edit boundary: 반환 구조체에 열거 필드 하나 추가.
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).
