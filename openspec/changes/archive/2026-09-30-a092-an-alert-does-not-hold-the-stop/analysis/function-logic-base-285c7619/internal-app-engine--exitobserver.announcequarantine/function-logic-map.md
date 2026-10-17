# Function Logic Map: `ExitObserver.announceQuarantine`

- Source: `internal/app/engine/exit_quarantine_announce.go` (59-90)
- AST evidence: `ast.json` — branches 2, returns 1, calls 4, assignments 3,
  defers 0, go_statements 0
- Risk scan: `risk-pattern-report.md`

**`analysis/notify-reach.md`는 이 자리를 3판에서 이미 열거했다(P4 호출자 표 7행).
그런데 `analysis/delivery-latency.md`는 그것을 쓰지 않았다** — publish 유발 이벤트
집합에서 `exit.snapshot_quarantined`가 빠졌고 판별자 규칙이 불완전했다.
**산출물이 쓴 사실을 문서가 안 쓴 것**이고, 3라운드 H2가 지적한 실패 방식과 같다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p journal.Position` | 관리 포지션 | 격리 생성 3곳 | — |
| `q journal.ExitSnapshotQuarantine` | 격리 레코드 | journal | — |
| `o.quarantineAnnounced` | `map[string]bool`, **nil 허용** | 이 함수(`:63`·`:70`) | nil이면 B1이 만든다 |

## Branches and early returns

| Branch | 위치 | 조건 | Mutation/side effect | Return | `Notify` 도달 |
|---|---|---|---|---|---|
| B1 | `:61` | `o.quarantineAnnounced == nil` | 맵을 지연 생성 `:63` | — | — |
| B2 | `:67` | `o.quarantineAnnounced[key]` — 이미 알림 | 없음 | `return` `:68` | ❌ |
| — | `:70` | — | `o.quarantineAnnounced[key] = true` | — | — |
| — | `:71` | — | **`o.alert(...)` — `Notify` 도달 (P4)** | 암묵 `:90` | ✅ |

**B2가 래치다.** 키는 `quarantineAnnouncementKey(p.ID, q.PositionGeneration, q.Version)`
(`:66`)이므로 **격리 세대·버전당 1회**다. 파일 헤더 주석(`:55-58`)이 말하듯 읽기 경로는
이 함수를 부르지 않는다 — 생성 3곳에서만 부른다.

## Calls and live bindings

| Callee | 위치 | 계약 | Evidence |
|---|---|---|---|
| `quarantineAnnouncementKey` | `:66` | 즉시 | AST calls |
| `o.alert` | `:71` | **동기·기한 없음.** critical(`event.go:296` `EventExitSnapshotQuarantined: true`) | `internal-app-engine--exitobserver.alert` FLM |
| `string`·`o.label` | `:73`·`:74` | 즉시 | AST calls |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `o.quarantineAnnounced` | `:63` | 지연 생성 — `NewExitObserver` 편집을 피하려는 의도(주석) |
| `o.quarantineAnnounced[key]` | `:70` | **래치** — 세대·버전당 1회 |

- goroutine 없음, defer 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: **yes** — §0.5. 격리는 "손절을 포함해 아무것도 평가되지 않는다"는
  상태이고, 이 알림이 그것을 사람에게 전하는 유일한 경로다.
- **a092가 여기서 얻는 사실 두 가지**
  1. `o.alert` 호출자는 `exitloop.go`에만 있지 않다. publish 유발 이벤트 타입은
     **16종**이고 `exit.snapshot_quarantined`가 그중 하나다.
     (프로덕션에서 실제로 publish하는 것은 12종 — flatten 4종은 `Notifier`가 nil이다.)
  2. 이 타입에는 로그 전용 생산자도 있다(`exitloop.go:576` `o.log(..., false, ...)` → INFO).
     `Notify`가 쓴 줄은 critical이라 WARN이므로 **레벨이 판별자다.**
