# Function Logic Map: `Notifier.AnnounceOperatingMode`

- Source: `internal/obs/mode.go` (`49`–`74`)
- Qualified: `Notifier.AnnounceOperatingMode`
- AST evidence: `ast.json` (`source_sha256` 95d6a3417808a527…) — branches 2 · returns 2 · calls 7
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **a090 3판 설계 전**(codex 1라운드 N2 — 모드 강화 공지의 동기 전송). 이 함수의 모양을 근거로 쓰는 3판 문서보다 먼저 만들었다.

**역할.** 커밋된 운영 모드 전이 하나를 critical 알림으로 공지한다(`journal.ModeAnnouncer` 구현). 생산에서 관측자의 `Announcer` 는 이 Notifier 다
(`cmd/tossctl/engine.go:639`), 그래서 관측자가 부르는 `EscalateOperatingMode` 의 공지는 `n.Notify` 로 **동기 전송**된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n` | nil 가능 | 배선 | nil 이면 무동작(B1) |
| `previous`·`rec` | 커밋된 전이 | `journal.TransitionOperatingMode` | — |

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 결과 |
|---|---|---|---|
| B1 | if | `:50` `if n == nil {` | **return nil** `:51` |
| B2 | if | `:54` `if journal.MoreConservativeMode(previous, rec.Mode) == previous && previous != rec.Mode {` | `direction = "relaxed"` |

마지막 **return** `:57` `n.Notify(ctx, Event{…})` — key `"operating_mode:" + rec.AccountRef + ":" + rec.Mode`(`:61`), 전이 id 는 **의도적으로 뺐다**
(`:58-60` "a re-announcement of the same transition deduplicates … The transition id would be unique per row and defeat that").

## Calls and live bindings

`journal.MoreConservativeMode` · `n.Notify`(동기: `n.mu` 아래 적재·전송·재시도 — 최악 54초, `obs/alert_lease.go:57`) · `fmt.Sprintf` · `modeLabel` · `permission`.

## State mutations and fallbacks

outbox 행(Notify 경유). 그 밖 없음.

## Safety conclusion

- **Safe edit boundary(a090 3판)**: `Event` 를 만드는 부분(`:57-71` 의 구조체)만 순수 함수 `OperatingModeEvent(previous, rec) Event` 로 **옮겨** 이 함수와
  a090 의 enqueue-only 공지자가 같은 내용을 쓴다(내용을 두 곳에 두지 않는다). 이 함수의 분기·key·`Notify` 호출은 **무변화**.
- **High-risk impact**: yes(운영 모드 공지) — 동작 무변화 리팩터만.
