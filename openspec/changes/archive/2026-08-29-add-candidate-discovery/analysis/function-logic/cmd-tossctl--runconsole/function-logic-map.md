# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L162–270, 분기 8개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: `console.Options` 리터럴에 `Orders`·`GateLimits`·`Signals` 세 seam 배선 추가 + 읽기 seam이 공유하는 계좌 해석기(`newConsoleBroker`) 1개 (revision=current)

콘솔의 실행 본체이자 **콘솔이 받는 모든 능력의 조립 지점**이다. 이 파일이 바이너리에서 internal/console을 import하는 유일한 파일이라는 규칙(`TestOnlyConsoleGoReachesTheConsolePackage`)이 여기를 리뷰 가능하게 만든다.

경계를 넘어가는 능력의 전부:

| 필드 | 넘어가는 것 | 넘어가지 않는 것 |
|---|---|---|
| `Holdings` | `broker.Holdings` method value | 브로커·주문 메서드·`*consoleBroker` |
| `Orders` | `Orders` 메서드 하나 | 브로커·`official.Client`·주문 메서드·`*consoleBroker` |
| `GateLimits` | float64 5 + 통화 | `*config.Service`·게이트 타입·쓰기 |
| `Settings` | `Load`/`Save` (편입 블록만) | 다른 config 블록·주문 능력 |
| `Signals` | verdict·tally 값 | `*candidate.Store` |
| `StartVerify` | verify 러너 실행 (웹 confirmer로 게이트됨) | 러너 구성의 변경 수단 |
| `Relaunch`/`Handoff`/`RestartSoak`/`StartEngine`/`StopEngine` | 실행 trigger | 실행 판단 (콘솔이 게이트만 통과시킨다) |

두 읽기 seam(`Holdings`, `Orders`)은 `reads` **하나**를 받는다 — `newConsoleBroker(root)`가 만든 공유 계좌 해석기다. 공유되는 것은 해석이지 능력이 아니다: 경계를 넘는 값은 여전히 seam마다 메서드 하나이고, `*consoleBroker` 자체는 `console.Options`의 어떤 필드에도 들어가지 않는다.

**여전히 배선되지 않은 것**: 브로커 자체, `--confirm-each`, 비루프백 바인딩, 세션 토큰 프리셋. `console.Options` 리터럴에 `broker`/`client`/`place`/`cancel`을 이름에 담은 필드가 생기면 테스트가 막는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil이면 `context.Background()` |
| `verifyRecord`/`verifyRecordUS`/`soakRecord`/`attestation` 경로 | 해석 가능해야 함 | `--config-dir`/기본 경로 | 해석 실패는 **치명** — 에러 반환, 콘솔이 뜨지 않는다 |
| `journalPath` | 해석 실패 허용 | `journal.DefaultPath()` | 실패 시 빈 문자열 + stderr 안내. 포지션·이력 화면이 원장 없이 뜬다 |
| `engineMarkerPath` | 해석 실패 허용 | `engineJournalDir` + `enginelock.MarkerPath` | 실패 시 빈 문자열 + stderr 안내. 엔진 상태 표시가 비어 있다 |
| `opts.port` | 0 또는 포트 | `--port` | 0이면 OS 선택 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ctx == nil` | `ctx = context.Background()` | 계속 | cobra 실행 경로 (`runCandidateStreams`류 테스트가 항상 context를 준다) |
| B2 | `resolveVerifyRecord` 실패 | 없음 | 에러 — 콘솔 미기동 | `verify` 경로의 record 해석 테스트 |
| B3 | `resolveVerifyRecordFor(US)` 실패 | 없음 | 에러 — 콘솔 미기동 | 동일 |
| B4 | `resolveSoakRecord` 실패 | 없음 | 에러 — 콘솔 미기동 | soak record 해석 테스트 |
| B5 | `resolveSoakAttestationPath` 실패 | 없음 | 에러 — 콘솔 미기동 | 동일 |
| B6 | `journal.DefaultPath()` 실패 | stderr 안내, `journalPath=""` | 계속 — **치명 아님** | `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes` |
| B7 | `engineJournalDir` 성공 | `engineMarkerPath` 설정 | 계속 | 엔진 상태 패널 렌더 테스트 |
| B8 | `engineJournalDir` 실패 | stderr 안내, 마커 경로 빈 문자열 | 계속 — **치명 아님** | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `signal.NotifyContext(ctx, os.Interrupt)` | Ctrl-C가 진행 중 검증의 정리를 끝내게 한다 | `defer stop()`; internal/console이 소켓을 닫기 전에 기다린다 | L169 |
| `newConsoleBroker(root)` | 읽기 화면 전체가 공유하는 라이브 클라이언트의 보관함 — **정확히 1회** 호출된다 | 아무것도 구축하지 않는다(lazy). 첫 화면이 계좌 해석 1회를 치른다 | L216, `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` |
| `newConsoleHoldings` / `consoleOrdersSeam` / `consoleGateLimitsSeam` / `consoleSettingsSeam` / `consoleSignalsSeam` | 다섯 seam 배선 — 앞 둘은 `reads`를, 뒤 셋은 `root`를 받는다(브로커를 쓰지 않으므로) | 각 seam의 실패 계약은 해당 함수의 map이 소유 | L232–249 |
| `consoleVerifyStarter(root)` | verify 러너 — **주문을 낼 수 있는 유일한 경로**. 공유 resolver를 받지 않는 유일한 곳이며 이유는 그 함수의 주석에 있다 | 웹 배치 confirmer로 게이트, 건별 confirmer는 항상 거절. 실행마다 자기 계좌 해석 1회 | L220, `consoleMutationConfirmer` |
| `verifyRunLockPath(verifyRecord)` | 실행 마커 — soak/발굴과 같은 경로 | 다른 답이 나오면 아무도 쓰지 않는 마커를 보는 콘솔이 된다 | `TestTheVerifyAndSoakSidesAgreeOnTheMarkerPath` |
| `console.ListenAndServe` | 127.0.0.1 바인딩 서버 | ctx 취소까지 블록 | L218 |

## State mutations and fallbacks

- 프로세스 레벨: signal 핸들러 등록(defer 해제).
- 원장·엔진 마커 해석 실패는 **화면의 문장**이 되고 콘솔 기동을 막지 않는다. 반대로 검증 기록 경로 실패는 치명이다 — 콘솔의 존재 이유가 그 검증이기 때문.
- **계좌 해석의 범위**: 읽기 화면 전체가 `reads` 하나를 공유하므로 세션은 화면을 몇 개 열든 `/api/v1/accounts`를 **1회** 읽는다(`TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`). 이전에는 seam마다 하나씩이라 두 화면을 여는 세션이 2회 읽었고, 그 읽기가 2026-07-26에 429를 세 번 받은 호출이다(measurements.md M4).
- **의도적 예외 1건**: 검증 실행(`consoleVerifyStarter`)은 실행마다 자기 클라이언트를 만든다. 기록에 적히는 계좌는 **실주문 직전에 그 실행이 확인한** 계좌여야 하고, 읽기 화면이 언젠가 해석해 둔 값(자격증명 교체 이전일 수도 있다)을 물려받으면 기록이 확인되지 않은 계좌를 이름 붙이게 된다. 비용은 실행당 1회로 묶여 있다.

## Safety conclusion

- Safe edit boundary: `console.Options` 리터럴의 필드 집합. 여기에 브로커나 주문 함수를 넣는 것이 콘솔에 주문 능력을 주는 유일한 방법이다.
- High-risk impact: yes (주문 경로) — 콘솔이 받는 모든 능력이 이 함수의 `console.Options` 리터럴에서 결정된다. `Holdings: newConsoleHoldings(reads)`를 `Broker: reads`로 바꾸는 **한 번의 편집**이면 콘솔이 주문 능력을 갖는다. 현재는 각 seam이 메서드 값 단위로만 넘어가고, 공유는 해석에만 적용된다.
