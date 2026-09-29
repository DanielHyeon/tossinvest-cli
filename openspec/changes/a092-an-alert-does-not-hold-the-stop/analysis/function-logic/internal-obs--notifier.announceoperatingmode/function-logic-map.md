# Function Logic Map: `Notifier.AnnounceOperatingMode`

- Source: `internal/obs/mode.go`
- AST evidence: `ast.json` — **편집 뒤**, :48–54, 분기 1 · 반환 2 · 호출 2, source_sha256 `ff7224f12ed7…`, 추출 커밋 `c6e2e3ac`.
  편집 전 번들(:49–74, 분기 2, `95d6a3417808…`, `8c390aa6`)은 `analysis/pre-edit/unit2/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집(24판 K1 · M10, 착지 단위 ②): 통지 사건 구성을 `operatingModeEvent(previous, rec)`(새 파일 `internal/obs/record_only.go`)로 뺐고,
  기록 전용 통지자(`RecordOnly.AnnounceOperatingMode`)가 같은 함수를 쓴다. 중복 제거 키에 전이 행 신원 `rec.ID`가 붙었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n` | nil 허용 | 배선 | B1 — nil이면 통지하지 않고 nil 반환 |
| `previous` | `""`(행 없음) · NORMAL · ENTRY_BLOCKED · HALT_ALL | `TransitionOperatingMode`의 `current.Mode` | `operatingModeEvent` 안의 `modeLabel`이 빈 값을 NORMAL로 |
| `rec` | 방금 커밋된 전이 행(`ID` · `AccountRef` · `Mode` · `Cause` · `Actor`) | `TransitionOperatingMode` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n == nil` (:49) | 없음 | `nil` (:50) | `TestAnnouncingWithoutANotifierIsSafe` |
| 종단 | — | `n.Notify(ctx, operatingModeEvent(previous, rec))` (:53) | `Notify`의 반환 | `TestA092BothAnnouncersBuildTheSameEvent` · `TestA092EachTransitionIsItsOwnAnnouncement` |

편집 전 B2(완화 방향 판정 `:54`)는 이 함수에서 사라져 `operatingModeEvent`(새 함수, 새 파일)로 옮겨졌다 — 그 판정의 시험
(`TestARelaxationIsAnnouncedAsARelaxation` · `TestTheTransitionLogLineIsCountable`)은 그대로 초록이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `operatingModeEvent` :53 | 사건 구성(종류 · 키 `operating_mode:계정:모드:rec.ID` · 제목 · 본문 · 필드) — 기록 전용 통지자와 공유 | 순수 | AST · `record_only.go` |
| `n.Notify` :53 | critical 통지 — 동기 경로(`claimAndDeliver` → `deliver`) | 기록 실패에서만 오류 | AST |

## State mutations and fallbacks

- 이 함수 자신은 상태를 바꾸지 않는다. 부작용은 `n.Notify`가 한다.
- 키에 `rec.ID`가 붙어 「강화 → 완화 → 재알림 창 안의 재강화」가 새 통지 행을 만든다(K1). 같은 전이의 재통지는 원장이 막는다(변화 없는 전이는 통지 전에 반환).
  RED 관측: 키 변경 전 `TestA092EachTransitionIsItsOwnAnnouncement` publish 2 / 행 2(기대 4) — 변경 뒤 4 / 4.

## Safety conclusion

- Safe edit boundary: 분기 B1과 `Notify` 호출은 그대로. 사건 구성만 공유 함수로 옮김.
- High-risk impact: yes — critical 통지의 중복 제거 신원. 통지 수는 상태를 바꾼 전이 수에 묶인다.
