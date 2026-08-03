# Function Logic Map: `enginePIDsForJournal`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a060-soak-restart-uses-this-profile
- Risk scan: `risk-pattern-report.md`

편집 **전에** 작성했다 (tasks 1.1). a059가 만든 함수이고, 이 change는 그것을 두
프로세스가 공유하는 헬퍼로 **추출**한다. 동작은 바뀌지 않아야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `lines` | `pgrep -a` 출력 (`pid cmdline`) | `engineListProcesses` | 파싱 불가는 버린다 |
| `journalDir` | 콘솔의 journal 디렉터리 | `engineJournalDir(root)` | 빈 값이면 아무것도 반환하지 않는다 |
| `defaultDir` | 기본 프로필의 journal 디렉터리 | `engineJournalDir(nil)` | 해석 실패면 flag-free 후보를 배제 |
| `engineProcessMatcher` | 엔진 후보 판정 | 패키지 변수 | 불일치는 배제 |

불변식: 증명할 수 없는 것은 전부 배제한다. 이 목록이 SIGTERM 대상이 되기 때문이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `journalDir`가 공백 | 없음 | `nil` | `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags` |
| B2 | `lines` 순회 | `pids` 누적 | — | `TestOnlyThisProfilesEngineIsFound` |
| B3 | 줄 파싱 실패 **또는** 매처 불일치 | 건너뜀 | — | `TestUnparsableProcessLinesAreDropped` |
| B4 | `--config-dir` 부재 → 기본값 대입 | — | — | `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags` |
| B5 | 해석 결과가 비었거나 `want`와 다름 | 건너뜀 | — | `TestOnlyThisProfilesEngineIsFound` |

### 이 change가 바꾸는 것

추출뿐이다. 시그니처가 `pidsOwnedBy(lines, matcher, want, resolve)`가 되고,
`defaultDir` 인자는 사라진다 — `resolve("")`가 기본값을 담당한다. 분기 다섯 개의
**판정 내용은 그대로**이고, 엔진 호출부가 얇은 wrapper로 남는다.

| | 변경 전 | 변경 후 |
|---|---|---|
| 정체 해석 | `engineJournalDir` 고정 | 호출부가 넘기는 `resolve` |
| 기본값 | `defaultDir` 인자 | `resolve("")` |
| 매처 | `engineProcessMatcher` 고정 | 인자 |
| 엔진의 판정 결과 | — | **동일해야 한다** (tasks 5.4) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `splitProcessLine` | `pid cmdline` 분리 | 실패는 배제 | AST |
| `engineProcessMatcher.MatchString` | 후보 확인 | — | AST |
| `engineCommandConfigDir` | `--config-dir` 되뽑기 | 없으면 빈 문자열 | AST |
| `filepath.Clean` | 경로 비교 정규화 | — | AST |

## State mutations and fallbacks

- 순수 함수. side effect 없음.
- fallback은 전부 배제 방향이다.

## Safety conclusion

- Safe edit boundary: 추출. 판정 규칙은 한 글자도 바뀌지 않는다.
- 검증 방법: a059가 만든 엔진 소유 판정 테스트 전부가 추출 뒤에도 green이어야 한다
  (tasks 5.4). 하나라도 빨개지면 추출이 아니라 변경이다.
