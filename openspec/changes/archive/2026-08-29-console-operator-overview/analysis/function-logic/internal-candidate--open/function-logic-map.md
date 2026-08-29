# Function Logic Map: `Open`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 함수다. 이 change가 넣은 것은 한 줄 — `Store`에 `prober`를 보관해서 여유 공간을 다시 물을 수 있게 한 것(D16)이다. Open의 FSInfo는 '여기 써도 되는가'라는 한 번 정해지는 판정이고, '남은 공간이 있는가'는 매 사이클 다시 물어야 하는 값이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Path` | 절대 경로 또는 빈 문자열 | 호출자. 비면 `DefaultPath()`가 `$TOSSOS_DATA_DIR > $XDG_DATA_HOME/tossos > ~/.local/share/tossos` 순으로 푼다 | 해석 실패는 에러, 아무것도 만들지 않는다 |
| `opts.FSProber` | nil이면 `SystemFSProber()` | 호출자. 프로덕션은 항상 nil이고 테스트만 주입한다(`internal/testenv`의 `TestFixedFSProberIsTestOnly`) | 허용목록 밖 파일시스템이면 **디렉터리를 만들기 전에** 거부한다 |
| `opts.BusyTimeout` | 0 이하면 `defaultBusyTimeout`(5s) | 호출자 | 미설정이 '무제한 대기'가 되지 않는다 |
| `opts.CoolingTTL` / `opts.StalenessTTL` | 0 이하면 각각 `DefaultCoolingTTL`/`DefaultStalenessTTL` | 호출자 | 미설정 필드가 경계를 끄는 것은 이 패키지가 반복해서 만난 실패 형태라 전부 기본값으로 떨어진다 |
| `opts.Clock` | nil이면 `clock.System()` | 호출자 | 시각은 주입된 clock 또는 호출자 인자에서만 온다. `TestNothingInThisPackageAsksTheWallClockWhatTimeItIs`가 `time` import를 **경로로** 해석해 `Now/Since/Until/After/Tick/NewTimer/NewTicker/AfterFunc/Sleep` 9개 이름과 alias import·dot import 두 형태, 합쳐 11가지 철자를 막는다. |
| DSN | 경로를 URL escape | `dsn()` | `SQLITE_OPEN_URI`라 escape하지 않은 `#`은 **다른 파일**을 연다(§1 7절). `_txlock=immediate`는 `Promote`의 read-then-write 때문에 load-bearing이다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `path == ""` | 없음 | `DefaultPath()` 결과 사용 | 직접 테스트 없음 — 프로덕션 경로(`openCandidateStore`가 configDir 없으면 빈 문자열) |
| B2 | `DefaultPath()` 에러 | 없음 | `nil, err` | 직접 테스트 없음 |
| B3 | `prober == nil` | 없음 | `SystemFSProber()` | 직접 테스트 없음 — 프로덕션 경로. 테스트는 항상 주입한다 |
| B4 | `CheckFilesystem` 에러 | **없음 — 거부된 마운트는 손대지 않는다** | `nil, err` | `TestTheStoreRefusesAFilesystemThatCannotPromiseAWrite` |
| B5 | `os.MkdirAll` 에러 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B6 | `busy <= 0` | 없음 | 기본 5s | `openStore` 헬퍼(모든 저장소 테스트) |
| B7 | `sql.Open` 에러 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B8 | `clk == nil` | 없음 | `clock.System()` | `openStore` 헬퍼 |
| B9 | `cooling <= 0` | 없음 | `DefaultCoolingTTL` | `TestAReEntryWithinTheCoolingTTLKeepsTheOriginalFirstSeenAt` |
| B10 | `staleness <= 0` | 없음 | `DefaultStalenessTTL` | `TestACandidateNobodyCooledDoesNotStayActiveForever` |
| B11 | `s.migrate` 에러 | `db.Close()` | `nil, err` | `TestAStoreFromANewerBuildIsRefused` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `CheckFilesystem` | 허용목록 판정 | 실패는 즉시 반환하며 디렉터리를 만들지 않는다 | `fsguard.go:155` |
| `os.MkdirAll` | `0o700` — 계좌 식별자를 담은 파일 옆이라 소유자 전용 | 실패는 wrap | ast.json calls |
| `sql.Open` + `db.SetMaxOpenConns(1)` | 단일 writer 설계. 연결을 늘리면 그럴 필요가 없게 설계된 파일에 경합을 산다 | — | `store.go` 주석 |
| `s.migrate` | 스키마 생성 + 사다리 | 한 트랜잭션. 실패 시 핸들을 닫고 에러 | `climb` |
| `clock.System` | 기본 시각원 | — | `internal/clock` — 이 패키지의 모듈 내부 의존 폐포는 `{internal/clock}` 하나다(`go list -deps ./internal/candidate`, `TestDiscoveryDependsOnNothingItHasNotArguedFor`). 주문 경로에 닿는 간선이 없고, 이 함수도 예외가 아니다. |

## State mutations and fallbacks

- 디스크 상태 변경: 데이터 디렉터리 생성(0700), `candidates.db` 생성, 스키마 DDL, `store_meta.schema_version` 스탬프, 필요 시 마이그레이션 사다리.
- 거부된 마운트에는 **아무것도 만들지 않는다**. `CheckFilesystem`이 `MkdirAll` 앞에 있는 것이 그 보장이고, `TestTheStoreRefusesAFilesystemThatCannotPromiseAWrite`가 `os.Stat`으로 확인한다.
- 마이그레이션 실패는 `db.Close()` 뒤 에러다. 사다리 자체가 한 트랜잭션이라 반쯤 오른 스토어가 남지 않는다.
- 이 change가 더한 상태는 `Store.prober` 필드 하나 — 인메모리이고 디스크에 닿지 않는다.

## Safety conclusion

- Safe edit boundary: `CheckFilesystem`을 `MkdirAll` 뒤로 옮기기, `_txlock=immediate` 제거, DSN 경로 escape 제거, 미설정 필드를 기본값 대신 0으로 두기는 모두 이미 재현된 결함의 복원이라 금지
- High-risk impact: no — 발굴 전용 SQLite 파일을 연다. 원장(`journal.db`)의 락도 파일도 건드리지 않고 (`TestAScanningStoreDoesNotBlockTheEngineFromStarting`, `TestTheStoreIsItsOwnFile`), 주문 경로에 닿는 간선이 없다. 다만 **원장과 같은 디렉터리·같은 파일시스템**에 파일을 만든다는 사실 자체가 D16의 출발점이다.
