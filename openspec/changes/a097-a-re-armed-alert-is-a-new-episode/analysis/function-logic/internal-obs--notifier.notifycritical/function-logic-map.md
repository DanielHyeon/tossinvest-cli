# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (L170-207)
- AST evidence: `ast.json` (4 branches, 3 returns)
- Risk scan: `risk-pattern-report.md`

**a097 초판은 이 함수를 편집 대상에서 뺐다. 2판이 편집한다.** proposal-freeze 리뷰가
"승격을 금지하면 재시작이 유일한 대응을 지운다"는 P1을 냈고, 그 승격이 들어갈 자리가
`B3@195`다 — `n.mu` 밖이어야 하므로 `claimAndDeliver`가 아니라 여기다.

산출물이 필요한 두 번째 이유는 proposal R2가 이 함수의 분기를 근거로 삼는다는 것이다 —
특히 `B4@199`가 claim 실패 시 도달 불가라는 주장.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `e` | critical 등급으로 판정된 이벤트 | `Notify@124-132`의 등급 판정 | — |
| `n.Journal` | nil 가능 | 조립 지점 | B1: 경고 후 best-effort로 강등 |
| `n.Log` | nil 가능 | 조립 지점 | B2가 nil을 거른다 |

**불변식**: durable first. 기록이 먼저이고 전송이 나중이다 — 메모리에만 있는 기록은
그것이 경고하려던 crash를 넘기지 못한다.

**불변식 (교착 회피)**: `escalate@204`는 `claimAndDeliver`가 뮤텍스를 **놓은 뒤** 불린다.
이 순서는 tidiness가 아니라 필수다(`:200-203` 주석).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@171 | `n.Journal == nil` | 경고 로그 + `publishBestEffort@180` | `nil` @181 | 기존 (a074) |
| B2@175 | `n.Log != nil` (B1 안) | 경고 한 줄 | — | 기존 |
| B3@195 | `claimAndDeliver` 오류 | **a097: `n.escalate`** — 내구적 차단 시도 | `wrapped error` @196 | **a097 2.4·2.5** |
| B4@199 | `owed && !sent` | `escalate@204` — 운영 모드 승격 | — | 기존 (a074/a096b) |

정상 반환: `nil` @206.

**B4가 claim 실패에 도달할 수 없다는 근거는 지금도 유효하다.** `claimAndDeliver`의 claim
오류 반환은 `false, false, err`이므로 `owed=false`이고, B3이 먼저 `return`하므로 B4는
평가조차 되지 않는다.

**그래서 a097은 승격을 B3 안에 직접 넣었다.** 이 문서의 초판은 "claim 실패에 대한 승격
경로는 없으며 그것이 의도"라고 적었다. **뒤집혔다** — proposal-freeze 리뷰의 P1이
그 의도를 기각했다: `EntryGate`의 래치는 메모리에만 있고 claim이 실패하면 원장에 행조차
없으므로, 재시작 한 번이면 차단도 증거도 사라진다. 구현은 `n.escalate(ctx, e)`를
B3에서 부르며(`notifier.go:213`), 이 문단은 구현을 따른다.

정리하면 claim 실패의 결과는 셋이다 — `claimAndDeliver`의 구조화 로그와 gate 래치
(`n.mu` 안), 그리고 여기의 승격 시도(`n.mu` 밖).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Log.Warn@176` | journal 미배선 경고 | 오류 없음 | AST calls |
| `n.publishBestEffort@180` | durable 없이라도 보낸다 | 실패는 로그만 | AST calls |
| `n.eventKey@185` | 중복 제거 key | 오류 없음 | AST calls |
| `encodeFields@190` | payload JSON | 오류 없음 | AST calls |
| `n.claimAndDeliver@194` | claim + send (배타 구간) | 오류는 B3 | AST calls |
| `n.escalate@204` | 운영 모드 승격 | **잠금 밖에서** 불린다 | AST calls |

AST calls 목록에 `n.Gate`가 **없다**는 것이 R2의 근거다.

## State mutations and fallbacks

- `record` 조립(`:184-191`)만 지역 변경이다.
- fallback: B1의 best-effort 강등. 조용하지 않게, 경고와 함께 한다.

## Safety conclusion

- Safe edit boundary: **`B3@195` 분기 본문에 `n.escalate` 한 줄.** 반환 계약, 오류 래핑,
  다른 세 분기, 그리고 `record` 조립은 불변이다.
- High-risk impact: yes (알림 전달 + 운영 모드 승격)
- a097 이후에도 B3의 반환값은 그대로다. 달라지는 것은 반환 **전에** gate가 잠기고(claimAndDeliver)
  내구적 차단이 시도된다(여기)는 것이다.
- `n.mu` 밖이라는 것이 이 위치의 이유다. `escalate`는 `ModeAnnouncer`로 알리고, 이 Notifier에
  배선된 announcer는 `Notify`로 재진입해 같은 뮤텍스에서 교착한다. `claimAndDeliver` 안에
  넣었으면 그 교착을 새로 만들었을 것이다.
