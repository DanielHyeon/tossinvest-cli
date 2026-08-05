# Function Logic Map: `Tracker.Observe`

- Source: `internal/reconcile/mismatch.go` (lines 360–491, pre-edit)
- AST evidence: `ast.json` (`source_sha256: a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 진입 차단 상태기계. 면제 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff` | 한 비교의 결과. `AsOf`는 RFC3339 또는 빈 문자열 | `Comparer.Compare` (`compare.go:360`) | `AsOf`가 비면 수렴이 애초에 거부되므로 credit이 생기지 않는다 |
| `diff.BlocksEntry()` | `len(Quantities) > 0 \|\| len(MissingOrders) > 0` | `compare.go:313` | 비교 단위(계좌 전체)이며 심볼 단위가 아니다 |
| `t.blocks` | key → Block. `Refresh`/`Restore`가 원장에서 재구성 | journal `reconcile_states` | nil이면 B1이 초기화 |
| `t.adjusted` | **(편집 전)** 심볼 집합 · **(편집 후)** 심볼 → credit 비교 as-of | `AdjustmentApplied` (프로덕션 호출자는 `Converger` 하나) | 재시작 시 복원하지 않는다 (의도된 계약) |
| `t.failures` | 연속 실패 수 | 이 함수 | `>= MaxFailures`에서 영구 승격 |
| `t.permanent` | 계좌 범위 영구 차단 존재 여부 | `hasPermanentQuantityAccountBlock` | 영구는 운영자만 해제 |
| `t.mu` | 함수 전체를 덮고 487행에서 1회 Unlock | — | `persist`가 lock 안에서 I/O를 한다 (기존 성질) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (366) | `t.blocks == nil` | `t.blocks` 초기화 | — | 기존 |
| B2 (373) | `!diff.BlocksEntry()` — 비교 전체 일치 | `t.failures = 0` | — | 기존 |
| B4 (380) | `range t.blocks` (B2 안) | — | — | 기존 |
| B5 (381) | `Cause != QUANTITY_MISMATCH \|\| Release != adjusted_reconcile` | skip | — | 기존 — 다른 producer의 상태를 해제하지 않는다 |
| **B6 (388)** | `!t.adjusted[symbol]` | `out.AwaitingAdjustment` 추가 | — | **편집 site 1** — credit 사용 가능 판정을 as-of 기반으로 바꾼다 (3.1–3.5) |
| B3 (396) | else — 비교 불일치 | `t.failures++` | — | 기존 |
| B7 (398) | `range blocksFor(diff, ...)` | — | — | 기존 |
| B8 (399) | 같은 key의 차단 없음 | `t.blocks[key]` 추가, `pending=true`, `out.Added` | — | 기존 |
| B9 (405) | `t.failures >= max && !t.permanent` | 계좌 범위 영구 차단 생성 | — | 기존 — 건드리지 않는다 (3.8) |
| B10 (418) | 영구 key 미존재 | `t.permanent = true` | — | 기존 |
| B11 (440) | `range out.Added` | `added` set 구성 | — | 기존 |
| B12/B13 (443/444) | `block.pending && !added[key]` | `toPersist.Added`에 재시도분 추가 | — | 기존 |
| — (449/453/456) | persist 결과 반영 | `pending=false` · authoritative 교체 · released 삭제 | — | 기존 |
| **(462)** | `persistErr != nil && !diff.BlocksEntry()` | 살아남은 차단의 credit만 보존 | — | **편집 site 2** — D4의 untouched/answered/refuted 회계로 바꾼다 (3.10) |
| **(475)** | else | **`t.adjusted = nil`** | — | **결함의 지점** — credit을 무조건 전부 버린다 (3.1, 3.11) |
| 종료 (481–490) | — | `out.Failures/Permanent/Blocked/NextDueAt`, `syncGate` 2회 | `(Outcome, persistErr)` | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `diff.BlocksEntry()` | 비교가 진입을 막는가 | 순수 | `compare.go:313` |
| `blocksFor(diff, account, now)` | diff → 차단 목록 | 순수. `Detail` 문자열을 여기서 고정한다 | `mismatch.go:832` |
| `t.syncGate(active)` | entry gate 투영 | 무오류. persist 전 1회(보수) + 후 1회 | `mismatch.go:878` |
| `t.persist(ctx, toPersist)` | 원장 write-through | 오류 시 부분 결과 + err. 호출자는 차단 유지로 다룬다 | `mismatch.go:513` |
| `journal.EnterReconcile` | 차단 durable | 멱등 — 첫 관측의 evidence를 보존한다 | `persist` 내부 |
| `journal.ReleaseReconcile` | 해제 durable | `ExpectCause`로 producer 소유권 확인 | `persist` 내부 |
| `t.clock().Now()` | `now`, `NextDueAt` | — | `mismatch.go:951` |

## State mutations and fallbacks

- `t.lastAt`, `t.observed` — 항상 갱신 (재대사 최소 간격의 근거)
- `t.failures` — 전체 일치면 0, 불일치면 +1. 심볼 단위가 아니다
- `t.blocks` — 추가 · pending 해제 · authoritative 교체 · released 삭제
- `t.permanent` — persist 후 `hasPermanentQuantityAccountBlock`로 재계산
- **`t.adjusted`** — 편집 대상. 현재는 관측 끝에서 무조건 nil이거나, persist 실패 시
  "아직 살아 있는 차단의 credit"만 남긴다
- fallback: persist가 실패해도 in-memory 차단과 gate는 유지된다 (fail-closed)

## Safety conclusion

- **Safe edit boundary**: `t.adjusted`의 타입과 그것을 읽고 쓰는 세 곳(B6, 462블록,
  475블록). 차단 생성(B7–B10), 영구 승격(B9), gate 투영, persist 계약은 건드리지 않는다.
- **High-risk impact**: **yes**. 해제 판정을 바꾼다. 방향 판정은 design D3 —
  어떤 입력에서도 현재보다 먼저 해제되지 않으며, 새로 해제되는 입력은 전부 승인된
  SHALL이 해제하라고 적어 둔 것이다.
- **건드리지 않는 것**: 영구 차단은 `Release = ReleaseOperatorOnly`이므로 B5에서
  걸러진다. 청산 경로는 이 함수에 없다 (`ExitAllowed`는 항상 true).
