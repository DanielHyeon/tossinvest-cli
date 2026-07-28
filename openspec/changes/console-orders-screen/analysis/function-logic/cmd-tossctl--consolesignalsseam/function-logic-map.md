# Function Logic Map: `consoleSignalsSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L851–857, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `137cc8d0` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

/signals 화면이 받는 능력의 생성 지점. 경계를 넘는 것은 **값**이다.

- **넘어가는 것**: `console.SignalsReader` — 메서드 하나(`Signals`). 반환은 verdict·tally·panel 보고 값.
- **넘어가지 않는 것**: `*candidate.Store`. 그 타입은 promote/cool/prune을 할 수 있고 internal/console에서 도달 불가하다. 따라서 "발굴 화면은 아무것도 바꾸지 않는다"는 핸들러가 아니라 **배선의 사실**이다.
- **원천 호출 없음**: `open`은 저장소를 열 뿐이고 `candidate.Assess`는 이미 저장된 것만 읽는다. 15초마다 다시 로드되는 탭이 두 번째 발굴자가 되면 미문서화 RANKING 한도를 `tossctl candidate watch`와 다투게 되고, 랭킹은 WTS 폴백이 없어 429 한 번이 원천 전체를 없앤다(D14 결정 2).
- **계좌·주문 능력**: 전혀 관여하지 않는다. internal/candidate의 의존 폐포는 {internal/clock}이며 주문 경로를 타입으로도 볼 수 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | nil 허용 | `--config-dir` 등 | 저장소 열기 실패는 첫 `Signals` 호출에서 에러 문장으로 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 — closure 하나를 담은 `*consoleSignals` 할당 | `console.SignalsReader` | `TestTheSignalsSeamReadsTheStoreAndCallsNoSource` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidateStoreFactory` (closure 안) | 요청 context로 저장소를 열고 release를 함께 받는다 | 호출 시점은 매 렌더. release는 계약의 일부 | `TestTheSignalsSeamDoesNotHoldTheStoreOpenAcrossReads`, `TestTheSignalsSeamOpensTheStoreUnderTheCallersContext` |

## State mutations and fallbacks

- 상태 변이 없음.
- 저장소를 **보관하지 않는다**: 오래 사는 리더가 있으면 `wal_checkpoint(TRUNCATE)`가 아무것도 회수하지 못하고, 그 WAL은 주문 원장이 쓰는 파일시스템 위에서 자란다(D16). 렌더 한 번 길이만 열어 watch 루프의 청소를 살려둔다.

## Safety conclusion

- Safe edit boundary: closure 하나. 저장소를 필드로 승격하는 변경은 원장 파일시스템의 WAL 청소를 조용히 끈다.
- High-risk impact: no (발굴 읽기 경로 — 주문·손절·사이징·인증 어디에도 닿지 않는다). 다만 저장소를 붙잡는 편집은 **주문 원장과 같은 파일시스템**의 WAL 회수를 막는다는 점에서 원장 인접 위험이 있다.
