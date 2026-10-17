# Function Logic Map: `EntryGate.Clear`

- Source: `internal/execgw/retry.go` (543-555)
- Revision: current — a124 구현 로트(2026-09-27) 편집 뒤 재추출; source_sha256 `a9612c2f35817471eee506ee2cfa3686fb297c3326bf1215eacbd162c2574f4c`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 2 · returns 0 · calls 4
- Exact AST return positions: (none)
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason` | 사유 코드 | 호출자 | 없는 래치면 삭제 없음 |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 547:2 | `if g.clearEpochs == nil {` |
| B2 | if at 551:2 | `if _, exists := g.latches[reason]; exists {` |

### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)

| Base branch | Base anchor | Base source | Current branch |
|---|---|---|---|
| b1 | if at 547·2 | `// when it is allowed.` | — (없어짐 · 하위 함수로 옮김) |
| b2 | if at 551·2 | `// rest of the account, and conflating the two is what made this gate wider than` | — (없어짐 · 하위 함수로 옮김) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `g.mu.Lock/Unlock` | 게이트 잠금 | map 연산뿐 |

## State mutations and fallbacks

- **a124**: 해제 요청마다 `clearEpochs[reason]++` — 래치 유무와 무관(반례 ㉣). 지연 생성.
- `revision` 은 그대로 실제 삭제 때만 오른다(전략 봉인 의미 불변).

## Safety conclusion

- Safe edit boundary: 세대 증가 한 줄 + 지연 생성 두 줄. 잠금 안에서 밖을 부르지 않는다.
- High-risk impact: yes — 진입 게이트(High-risk). 해제 동작 자체는 불변.
