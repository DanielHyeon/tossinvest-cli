# Function Logic Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go` (`50`–`176`)
- Qualified: `StartPositionPolicyCommandServer`
- AST evidence: `ast.json` (`source_sha256` a48cca236c41e02e…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 18

**역할.** 엔진 제어 endpoint 를 연다. a094: park 해동 route 를 capability 발견으로 등록(3줄, a066 블록 옆).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `commands` | 명령 서비스 | engine run | attemptThawCommands 를 구현하면 route 추가 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:52` `if commands == nil {` | 아니오 |
| B2 | if | `:56` `if dir == "" {` | 아니오 |
| B3 | if | `:59` `if err := positionpolicyrpc.ValidateEngineDirectory(dir); err != nil {` | 예 |
| B4 | if | `:64` `if err := os.Mkdir(controlDir, 0o700); err != nil {` | 예 |
| B5 | else | `:68` `} else {` | 예 |
| B6 | if | `:65` `if !errors.Is(err, os.ErrExist) {` | 아니오 |
| B7 | if | `:71` `if err := positionpolicyrpc.ValidateControlDirectory(controlDir); err != nil {` | 예 |
| B8 | if | `:72` `if createdControlDir {` | 아니오 |
| B9 | if | `:78` `if createdControlDir {` | 아니오 |
| B10 | if | `:107` `if err != nil {` | 아니오 |
| B11 | if | `:112` `if _, err := rand.Read(tokenBytes); err != nil {` | 아니오 |
| B12 | if | `:123` `if r.Method != http.MethodGet {` | 아니오 |
| B13 | if | `:130` `if r.Method != http.MethodGet {` | 아니오 |
| B14 | if | `:135` `if err != nil {` | 아니오 |
| B15 | if | `:148` `if quarantines, ok := commands.(exitQuarantineCommands); ok {` | 예 |
| B16 | if | `:153` `if relaxations, ok := commands.(riskRelaxationCommands); ok {` | 예 |
| B17 | if | `:157` `if thaws, ok := commands.(attemptThawCommands); ok {` | 예 |
| B18 | if | `:167` `if err := writePositionPolicyDescriptor(server.descriptor, descriptor); err != nil {` | 아니오 |

## Calls and live bindings

`registerAttemptThawRoute`(신설) 외 종전.

## State mutations and fallbacks

없음(등록).

## Safety conclusion

- capability 없는 빌드는 route 집합 불변. 콘솔은 route 를 부르지 않음(시험). High-risk: no(등록) — 명령 자체는 attempt_thaw_command.go.
