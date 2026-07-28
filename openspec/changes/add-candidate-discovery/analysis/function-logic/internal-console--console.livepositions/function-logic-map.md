# Function Logic Map: `Console.livePositions`

- Source: `internal/console/portfolio.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

console-operator-overview task 2.3이 `positions`에서 뽑아낸 **원장 절반**. 뽑아낸 이유 하나: `positions`는 `holdings.get`을 부르고 그것은 갱신한다. 개요 화면은 같은 관리/미관리 구분이 필요하면서 브로커 호출이 0콜이어야 하므로, 개요가 `positions`를 재사용했다면 콘솔에서 가장 오래 열려 있는 탭이 조용히 rate budget 위에 올라갔을 것이다 — 렌더마다 holdings 1콜, 영원히.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.JournalPath` | 빈 문자열 허용 | cmd/tossctl | 빈 값이면 `openJournal`이 `journalUnwired` 뷰를 답하고 `ro == nil` |
| 원장 스키마 | 바이너리와 호환 | `journal.OpenReadOnly` | 새로우면 '콘솔 업데이트 필요', 오래됐으면 '엔진 기동으로 마이그레이션 필요' — 어느 쪽도 빈 상태로 위장하지 않는다 |
| `ro.AccountRefs(ctx)` | 계좌 참조 목록 | 원장 | 실패하면 `journalFailed`로 표시하되 **계속 진행**한다(부분 답이 빈 화면보다 낫다) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ro == nil` | 없음 | `nil, jv, nil` — jv가 미배선/부재/스키마 사유를 나른다 | `TestThePositionsScreenRendersWithEitherSourceMissing`, `TestThePositionsScreenNamesBothSchemaDirections` |
| B2 | `refs, err := ro.AccountRefs(ctx); err != nil` | `jv.State, jv.Detail = journalFailed, err` | 없음 — 루프로 계속 | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` |
| B3 | `for _, ref := range refs` | 계좌별 누적 | 없음 | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`(1계좌), `positionsView.Multi()`(다계좌 열) |
| B4 | `got, err := ro.LivePositionExits(ctx, ref); err != nil` | `journalFailed` + `break` | 없음 — 그때까지 모은 행을 유지 | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` |
| B5 | `len(got) > 0` | `accounts = append(accounts, attest.Mask(ref))` | 없음 | `TestTheDashboardMasksTheAccountNumber`(마스킹), `TestAnEmptyJournalSaysSoRatherThanLookingLikeAMissingOne`(0행이면 계좌를 세지 않는다) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.openJournal(ctx)` | `journal.OpenReadOnly` — 디렉터리를 만들지 않고 마이그레이션하지 않는다 | 실패는 `journalView` 상태로 돌아온다 | portfolio.go:134 |
| `ro.AccountRefs` / `ro.LivePositionExits` | 원장 읽기 | 부분 실패는 사유를 이름 붙여 그때까지의 답을 유지 | internal/journal ReadOnly |
| `attest.Mask(ref)` | 계좌번호 마스킹 | 원문 계좌번호가 화면·로그에 나가지 않는다 | internal/attest |
| `defer ro.Close()` | 핸들 반납 | B1 이후에만 등록된다 — nil close 없음 | portfolio.go:410 |
| (금지 바인딩) | 원장은 `journal.ReadOnly`로만 연다. `journal.Open`·`journal.Journal`·`journal.Options{}`·쓰기 메서드 넷은 이 패키지에 존재할 수 없다 | `TestTheConsoleOpensTheJournalReadOnly` | static_test.go:1542 |

## State mutations and fallbacks

- 원장을 **읽기만** 한다. 파일·디렉터리를 만들지 않고 마이그레이션하지 않는다.
- 부분 실패 계약: `AccountRefs` 실패는 계속, `LivePositionExits` 실패는 break — 어느 쪽이든 `jv`가 사유를 나르고 그때까지의 행을 버리지 않는다.
- `accounts`는 **행이 있는** 계좌만 담는다 — 빈 계좌가 다계좌 열을 켜지 않는다.

## Safety conclusion

- Safe edit boundary: 신설(추출). `positions`가 인라인으로 하던 것과 동일한 순서·동일한 실패 처리이며, 달라진 것은 호출자가 둘이 됐다는 것뿐이다.
- High-risk impact: yes (원장 읽기 경로 — 이 함수의 `journalView`가 '관리 외(미편입)' 라벨을 결정하고, 원장을 못 읽은 것을 '관리 안 됨'으로 읽으면 화면이 보호받는 포지션을 보호받지 않는 것으로 표시한다)
