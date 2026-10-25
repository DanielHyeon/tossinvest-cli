# Function Logic Map: `ordersSnapshot.freshness`

- Source: `internal/console/chrome.go` (287-305) — 편집 뒤(a173 rebase 후 main).
- Qualified: `ordersSnapshot.freshness`
- AST evidence: `ast.json` (`source_sha256` 367463d91837a4ab…) — AST branches 6 · returns 4
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 완료 게이트 직전(2026-10-25, 종결 에이전트). tasks 3.1 의 비례 원칙 면제는 required 7 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임(FLM 이 구현 판단에 쓰이지 않음).

**역할.** 주문 캐시를 같은 방식으로, 자기 TTL 로 읽는다.

a173 편집: unavailable 반환 셋에 열거 사유 `Code` 를 붙임. 분기 · 순서 · Reason 문구 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Wired` · `s.Error` · `s.Present` · `s.Held` | bool · 문자열 | 주문 캐시 스냅샷 | 순서대로 첫 참 갈래가 이김 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a173 |
|---|---|---|---|---|
| B1 | `288:2` | switch | 상태 switch | 무변경 |
| B2 | `289:2` | case | !Wired → unavailable · seam_unwired | a173: Code 추가 |
| B3 | `291:2` | case | Error → unavailable · broker_read_failed | a173: Code 추가 |
| B4 | `293:3` | if | Error 이면서 Present → LastAt | 무변경 |
| B5 | `297:2` | case | !Present → unavailable · never_fetched | a173: Code 추가 |
| B6 | `301:2` | if | Held → Hold 문구 | 무변경 — 측정 시험 0(기존 공백, a173 무관) |

Exact AST return positions: 290:3, 296:3, 298:3, 304:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.TakenAt` | 마지막 성공 · 캐시 시각 | 순수 | AST calls 294 · 300 |

## State mutations and fallbacks

- 값 반환만. 상태 변경 0.

## Safety conclusion

- Safe edit boundary: 반환 구조체에 열거 필드 하나 추가. 갈래 판정 무변경.
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).
