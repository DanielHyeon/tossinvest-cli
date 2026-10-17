# Function Logic Map: `Context.AlertDeliverer`

- Source: `internal/app/engine/auxiliary.go` (156-186)
- Revision: current — a124 구현 로트(2026-09-27) 편집 뒤 재추출; source_sha256 `b1e02baf829f2acd0401d47228ecee7f059f9d4c8c34a94f3284b840cbe8a7e7`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 3 · returns 2 · calls 3
- Exact AST return positions: 158:3, 179:2
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Journal` · `c.Entry` | 둘 다 필수 | 엔진 조립 | 없으면 `ErrRuntimeUnavailable` |
| `c.AccountRef` | 엔진이 기동 때 푼 계정 | `engine.go` | 빈 값이면 실행자는 승격하지 않음(Notifier.escalate 와 같은 규칙) |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 157:2 | `if c == nil \|\| c.Journal == nil \|\| c.Entry == nil {` |
| B2 | if at 161:2 | `if clk == nil {` |
| B3 | if at 165:2 | `if c.Notifier != nil {` |

### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)

| Base branch | Base anchor | Base source | Current branch |
|---|---|---|---|
| b1 | if at 157·2 | `if c == nil \|\| c.Journal == nil \|\| c.Entry == nil {` | B1 |
| b2 | if at 161·2 | `if clk == nil {` | B2 |
| b3 | if at 165·2 | `if c.Notifier != nil {` | B3 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `c.clock` | 시계 기본값 | — |
| `blockEntryOnDeliveryStop` | 실행자 정지 → `ReasonAlertSenderDown`(기존) | OnStop |

## State mutations and fallbacks

- a124 배선: `Gate: c.Entry` · `AccountRef: c.AccountRef` 두 필드. Notifier 는 넘기지 않음(실행자는 n.mu 와 무관 — D5 · D7).

## Safety conclusion

- Safe edit boundary: 생산 조립의 유일한 자리(`cmd/tossctl/engine.go:659`). 게이트와 계정을 넘기는 것 외 변화 없음.
- High-risk impact: yes — 실행자가 진입 게이트를 잠글 수 있게 된다(보수 방향).
