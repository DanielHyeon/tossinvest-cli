# Function Logic Map: `newRootCmd`

- Source: `cmd/tossctl/root.go`
- AST evidence: `ast.json` (revision=current, L52–189, 분기 7개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: `newCandidateCmd(opts)` 등록 한 줄 추가 (revision=current)

바이너리의 명령 조립 지점. 이 branch range의 변경은 **`newCandidateCmd(opts)` 등록 한 줄**이다.

**이제 배선된 것**: `tossctl candidate scan`과 `tossctl candidate watch`. 둘 다 `mutating: false`이고 `source: both`다. 발굴이 볼 수 있는 인터페이스(`internal/candidate.Source`)는 시장을 받아 행을 돌려주는 메서드 하나뿐이라 주문을 표현할 방법이 없고, internal/candidate의 의존 폐포는 {internal/clock}이다.
**여전히 배선되지 않은 것**: 발굴이 브로커·주문 경로에 닿는 어떤 배선도 없다. 저장소 경로 플래그도 없다 — 읽기 전용 명령의 경로 플래그는 저장소를 둘로 쪼개 이력을 반씩 나누는 방법이 하나 더 생기는 것이다.

이 함수의 나머지(`PersistentPreRunE`의 세션 만료 경고·config 경고·온보딩 힌트)는 이 change 묶음에서 바뀌지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.outputFormat` | table/json/csv | `--output` | 파싱 실패는 PersistentPreRunE 에러 |
| `opts.configDir` / `sessionFile` / `backend` | 빈 문자열 허용 | persistent flag | 해석 실패는 개별 헬퍼가 흡수 |
| 등록되는 서브커맨드 집합 | 고정 목록 | 이 함수 | 누락은 명령 부재; 잘못된 annotation은 자동 실행 금지 규칙을 무력화 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `output.ParseFormat` 실패 | 없음 | 에러 — 명령 미실행 | `TestResolveBackend*` 계열과 별개로 cobra 실행 경로 |
| B2 | `resolveUpdateCachePath` 성공 | gate/mark 콜백 설정 | 계속 | `TestExpiryWarningRespectsBackoffGate` |
| B3 | `loadConfigStatus` 성공 | config legacy 경고 시도 | 계속 | `TestConfigLegacyWarningOnLegacyFields`, `...SilentWhenNoConfig` |
| B4 | `format == FormatTable` | 온보딩 힌트 블록 진입 | 계속 | `TestConfigLegacyWarningSilentInJSONAndSkipCommands` |
| B5 | `resolveOpenAPIPaths` 성공 | 자격증명 확인 시도 | 계속 | 온보딩 힌트 테스트 |
| B6 | `official.LoadCredentials` 성공 && creds != nil | `hasOfficialCreds = true` | 계속 | 동일 |
| B7 | `shouldHintOnboarding(...)` | stderr 한 줄 | 계속 | `shouldHintOnboarding` 단위 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.AddCommand(... newCandidateCmd(opts))` | 발굴 표면 등록 — 이 change의 실제 편집 | 읽기 전용 명령 2개(scan/watch) | `TestTheDiscoveryCommandsDeclareThemselvesReadOnly`가 `newRootCmd()`에서 경로로 찾아 annotation을 검사 |
| `newConsoleCmd(opts)` | 콘솔 등록 (`mutating: true`) | 실주문 가능 — 타이핑 승인 뒤 | `TestConsoleIsRegisteredAndAnnotated` |
| `newOrderCmd` / `newFlattenCmd` | 주문 가능 명령들 (기존) | 이 change에서 무변경 | 무변경 |
| `writeExpiryWarningIfNeeded` / `writeConfigLegacyWarningIfNeeded` | stderr 안내 | JSON 모드·skip 명령에서 침묵 | root_test.go 경고 테스트 12종 |

## State mutations and fallbacks

- 전역 상태 변이 없음. flag 바인딩과 명령 트리 구성만.
- 발굴 명령 등록은 기존 명령을 하나도 바꾸지 않는다 — 새 파일·새 명령.

## Safety conclusion

- Safe edit boundary: `AddCommand` 목록에 읽기 전용 명령 하나 추가. annotation을 잘못 달면 도구 계층의 `mutating: true` 자동 실행 금지가 적용되지 않는다.
- High-risk impact: yes (주문 경로 — 조립 지점) — 이 함수가 `order`·`flatten-all`·`console` 같은 주문 가능 명령을 등록한다. 이번 편집 자체는 `mutating: false` 명령 한 개의 추가이며 주문 능력을 늘리지 않는다. 그러나 여기에 잘못된 annotation을 단 명령을 추가하는 편집은 안전 게이트를 우회시킨다.
