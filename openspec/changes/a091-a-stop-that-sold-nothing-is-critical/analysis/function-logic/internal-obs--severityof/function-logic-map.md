# Function Logic Map: `SeverityOf`

- Source: `internal/obs/event.go` (309-314)
- AST evidence: `ast.json` — branches 1, returns 2
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t EventType` | 임의 문자열 | 호출자 | 미등록이면 `SeverityNormal` |
| `criticalEvents` | `map[EventType]bool` (18종) | `event.go` 리터럴 | **등급의 유일한 정본** |

**production 호출자 2곳**: `obs/log.go:185`(구조화 로그의 `severity` 필드),
`obs/notifier.go:108`(critical만 outbox·게이트 경로로 보낸다).

> **도구 한계 기록**: `codegraph callers SeverityOf`는 **테스트 5개만** 돌려주고
> 위 두 production 호출자를 빠뜨린다. hard evidence(4단계)만으로는 이 함수의 영향 범위를
> 얻을 수 없었다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | 결과 |
|---|---|---|---|---|
| B1 `:310` | `criticalEvents[t]` | 없음 | `:311` `SeverityCritical` | outbox 기록 + 전달 재시도 + 실패 시 게이트 |
| — `:313` | 그 외 | 없음 | `SeverityNormal` | best-effort 발송만, **durability 없음** |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 순수 map 조회 | — | AST (calls 0) |

## State mutations and fallbacks

없음. 순수 함수. 기본값이 `SeverityNormal`이라 **누락은 조용히 강등된다.**

## Safety conclusion

- **Safe edit boundary**: `criticalEvents`에 항목을 더하는 것은 이 함수를 편집하지 않는다
- **High-risk impact**: **yes (간접)** — 이 map이 outbox·게이트 진입을 결정한다
- **주의**: `EventExitProposalCapped`를 map에 그냥 추가하면 **부분 캡까지 critical**이 된다.
  8/2가 보여준 결함은 **0주 캡**이므로, 등급을 나누려면 map 항목이 아니라
  **호출자가 사건을 구분**해야 한다
