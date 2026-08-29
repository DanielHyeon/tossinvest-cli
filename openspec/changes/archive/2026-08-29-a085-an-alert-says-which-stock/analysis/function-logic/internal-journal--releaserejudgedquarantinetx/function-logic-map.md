# Function Logic Map: `releaseReJudgedQuarantineTx`

- Source: `internal/journal/exit_snapshot.go` (lines 863–891)
- AST evidence: `ast.json` (`source_sha256: 0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

선택기가 교체된 격리를, **그 행이 실제로 재판정되었을 때** 닫는다. 가드가 세 겹이고
개정이 하나씩 더했다.

- **개정 3 — 사실.** 개정 2는 재판정 여부를 활성 행의 *사유*로 추론했고, 그래서 지금 도는
  선택기가 방금 쓴 격리를 다음 성공 판정이 `SELECTOR_REVISED` 근거와 함께 닫았다.
  일어나지 않은 재판정이고, 운영자만 풀 수 있어야 하는 행이다. 각인은 판별자가 될 수 없다 —
  `judge`가 이 트랜잭션보다 먼저 각인하므로 행은 어느 쪽이든 이미 지금 revision을 달고 있다.
  그래서 호출자가 아는 사실을 인자로 받는다.
- **개정 4 — 어느 행인가.** 사실만으로는 부족했다. 각인과 이 트랜잭션 사이에 운영자가
  각인된 행을 `HUMAN_REPAIR`로 풀고 병행 관측의 실패한 선택이 새 행을 열 수 있다. 버전을
  보지 않는 해제는 **그 새 행**을 닫는다 — 새 행을 쓴 것은 지금 도는 선택기 자신이므로
  개정 3이 없앤 것과 같은 종류의 거짓 근거다. "재판정이 있었다"와 "*이* 행이
  재판정되었다"는 다른 주장이고, 해제를 정당화하는 것은 두 번째뿐이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reJudgingVersion` | 0 = 재판정 아님, 그 외 = 재시도를 소비한 격리 버전 | `ExitJudgement.ReJudgingVersion` ← `reJudgingVersion(m)` ← `managed.reJudge`/`reJudgeVersion` | `<= 0`이면 즉시 nil — 아무것도 닫지 않는다 |
| 활성 격리 행 | 세대별 최대 1행, 없을 수 있음 | `activeExitSnapshotQuarantineTx` | 없으면 no-op |
| `active.Reason` | 격리 사유 | 행 | `ambiguous_recovery`가 아니면 no-op — 복구 선택기가 판정한 적 없는 행은 이 경로가 풀 수 없다 |
| `active.Version` | `quarantine_version`, INSERT 이후 불변 | 행 | `reJudgingVersion`과 다르면 no-op. 판정은 그대로 커밋되고 격리는 선다 |

## Branches and early returns

세 개의 조기 반환이 모두 "닫지 않는다"로 수렴한다. 이 함수에는 실패로 열리는 경로가 없다.
분기별 조건과 측정된 커버리지는 `branch-test-map.md`에 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `activeExitSnapshotQuarantineTx` | 활성 행 조회 | 오류를 그대로 반환 — 판정 트랜잭션이 롤백된다 | AST `calls` |
| `releaseExitSnapshotQuarantineTx` | `SELECTOR_REVISED`로 닫기 | 오류 반환 | AST `calls` |

## State mutations and fallbacks

- `exit_snapshot_quarantines.released_at/release_kind/release_evidence` — `SELECTOR_REVISED`.
  판정과 같은 트랜잭션이라 둘이 갈라질 수 없다.
- `ReleaseExitSnapshotQuarantine`(운영자 경로)은 `SELECTOR_REVISED`를 거절한다. 이 근거는
  판정 트랜잭션 안에서만 쓰인다.
- 버전 불일치로 건너뛴 경우: 대체 행은 서 있고 `selector_revision`이 현재이므로
  `NeedsReJudgement()`가 false다. 포지션은 운영자가 풀 때까지 판정 대상에서 빠진다.
  보수 방향이고 개정 3의 원칙과 같다 — 지금 도는 선택기가 쓴 행은 기계가 풀지 않는다.
  다음 사이클의 `EventExitJudgementRefused`가 운영자에게 알린다.

## Safety conclusion

- Safe edit boundary: 세 가드. `reJudgingVersion` 가드를 빼면 운영자 전용 행이 스스로 풀리고
  (3차 리뷰가 재현), 버전 가드를 빼면 재판정을 번 행이 아니라 그것을 대체한 행이 풀린다
  (4차 리뷰가 재현). 5차 리뷰의 변이 테스트가 버전 가드 삭제로 새 테스트를 죽였다.
- High-risk impact: yes — 격리 해제는 포지션을 다시 판정 대상으로 만든다.
