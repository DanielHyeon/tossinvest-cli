# Function Logic Map: `candidateCycleOptions`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L346–357, 분기 0개)
- Risk scan: `risk-pattern-report.md`

스캔·watch가 쓰는 `CycleOptions`를 만든다. 이 change가 **임계 리터럴을 지웠다** —
`candidate.VetoThresholds{NearHighDistancePct: ...}`가 여기 있었고, 같은 리터럴이
`console.go`의 /signals seam에도 있었다.

둘을 묶는 것이 없었다. 임계를 바꾸는 평범한 방법이 한쪽을 편집하는 것이고, 그 결과는
**같은 저장소·같은 시각에 대해 두 화면이 다른 판정을 내는 것**이다 — 각자 내부적으로
일관되고, 어느 쪽도 실패하지 않는다. 이 저장소는 그 실패를 이미 한 번 겪었다(네 번째
그림자 밴드, `watch.go:686-692`).

이제 둘 다 `candidateVetoThresholds()`를 부르고, `vetothresholds.go`가 값과 그 근거를
갖는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | CLI 옵션 | 호출자 | `candidateEngineProbe`로만 쓰인다 |
| `market`/`sources`/`sched`/`backoff` | 호출자가 만든 것 | `candidateWiring` | 그대로 전달 |
| 임계 | `candidateVetoThresholds()` — **유일한 출처** | `vetothresholds.go` | near_high만 값이 있고 나머지 둘은 `THRESHOLD_ABSENT` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (무분기) 구조체 하나를 만들어 돌려준다 | 없음 | `candidate.CycleOptions` | `TestTheTwoSurfacesApplyTheSameThresholds` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidateVetoThresholds()` | 임계 단일 출처 | 순수, 값 복사 | ast.json calls |
| `candidateEngineProbe(root)` | 엔진 실행 여부 probe | — | ast.json calls |

## State mutations and fallbacks

- 없음 — 옵션 값 하나를 만든다.
- fallback 없음. seen_late·extended에 임계가 없는 상태는 **의도된 것**이고 `THRESHOLD_ABSENT`로 보고된다(D18).

## Safety conclusion

- Safe edit boundary: 리터럴 → 함수 호출. 필드 구성과 값 무변경.
- High-risk impact: **yes 인접** — chase veto의 입력이다. 이 change는 값을 바꾸지 않고 출처를 하나로 만들었다.
