# Function Logic Map: `strategyProposalAuthorityLoader.loadFamilyActivation`

- Source: `internal/app/engine/strategy_family_activation.go`
- Source SHA-256: `50c775155bc8a12a5844ddbc30785f72e49b4fab7380146071e458f815d4c377`
- Signature: `strategyProposalAuthorityLoader.loadFamilyActivation(params=5, results=2)`
- Source range: `154:1`–`197:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.5-R).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- env 를 읽을 수 없으면 핀이 비었는지 모른다 — 미선언(기존 경로)으로 답하지 않고 Unavailable(되돌림)로 답한다(더 닫힌 쪽).
- 생산 생성자 `newStrategyProposalAuthorityLoader` 는 nil 을 `os.Getenv` 로 채운다 — B1 도달은 손으로 만든 적재기뿐.

## Branches and early returns

- Exact AST return nodes: `162:3, 190:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 161:2 | **(새)** getenv 가 nil → 맨 `ErrProductionFamilyActivationUnavailable` — 미선언이 아님(핀을 읽을 수 없음), 관문 되돌림(맨 sentinel 인 이유: 이 경로 호출은 `TestTheRollbackPathOnlyReads` 의 읽기 전용 허용 목록으로 묶임) |
| B2 | if | 165:2 | US 시장이면 US 활성화 digest env |
| B3 | if | 187:2 | US 시장이면 US 위험 정책 env |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyMarketCalibrationDigest` | 175:20 |
| `strategyrouter.LoadProductionFamilyActivation` | 190:9 |
| `strategyRouterMarket` | 191:40 |
| `strings.TrimSpace` | 192:19 |
| `loader.getenv` | 192:37 |
| `strategyRuntimeBuildDigest` | 194:60 |
| `strings.TrimSpace` | 195:21 |
| `loader.getenv` | 195:39 |

## State mutations and fallbacks

- 상태 변경 없음 — env 읽기와 strategyrouter 적재 호출.

## Safety conclusion

- B1 은 거절만 더한다(수락 집합 불변) — 공황 → 되돌림으로 바뀌어 같은 주기의 앞선 닫힘 사유(FX_NOT_READY 등)가 보존된다.
- B1 은 새 호출을 더하지 않는다 — 되돌림 경로 읽기 전용 허용 목록(`TestTheRollbackPathOnlyReads`) 불변(첫 판의 `fmt.Errorf` 는 그 시험이 막았다 — 변이 대조군 전체 스위트 실측).
