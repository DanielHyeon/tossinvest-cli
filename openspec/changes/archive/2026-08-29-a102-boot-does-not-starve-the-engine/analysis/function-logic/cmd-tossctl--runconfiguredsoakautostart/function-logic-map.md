# Function Logic Map: `runConfiguredSoakAutostart`

> **이 change는 이 함수를 편집하지 않는다.** D7이 무편집을 명시로 못박았다. 이 문서는
> **proposal 근거용 분석**이고, 분기 열거는 전부 `ast.json`이 낸 것이다.

- Source: `cmd/tossctl/soakautostart.go` (87-117)
- AST evidence: `ast.json` — AST 기준 branches **7** / returns 6 / calls 7
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256 (현재 HEAD): `937b0c68762523ff02a39f0554371572c636308c431d9d19e42a7b7f66adab1c`
- 작성 사유: a102 proposal이 "자동 시작이 켜져 있으면 부팅 서베이가 엔진 복구와 같은 rate
  예산을 때린다"를 주장하면서 **이 함수의 판정**(승인 여부 → start 호출)을 근거로 썼다.
  a102의 대기는 이 함수가 받는 `start` seam을 감싸므로, 그 seam의 계약이 곧 안전 경계다.

## 이 함수가 하는 일

부팅 시점의 soak 자동 시작 **판정 전부**다. `runConsole`이 0.0%로 측정되므로 판정을 그 안에
쓰면 어떤 테스트도 닿지 못한다 — 그래서 두 seam(`load`, `start`)을 인자로 받는다.
**오류를 돌려주지 않는다.** 결과는 언제나 운영자가 읽는 한 줄이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `load func() (bool, error)` | **nil 가능** | `consoleSoakBoot.Load` | B1 — nil이면 아무것도 안 한다(승인 기록이 없다) |
| `load()`의 오류 | — | config 파일 | B2 — fail-closed. 읽을 수 없는 설정은 승인이 아니다 |
| `on` | bool | `soak.autostart` 키 | B3 — false면 빈 문자열(출력도 없다) |
| `start func() (string, error)` | **nil 가능** | `runConsole`의 `bootSurveyIfAbsent` | B4 — nil이면 "배선이 없다" 한 줄 |
| `start()`의 `(note, err)` | 임의 | a101의 `bootSurvey` | B5·B6·B7 — 실패도 빈 note도 문장이 된다 |

> **관통 불변식**: **이 함수는 error를 반환하지 않는다.** 서베이는 선택 기계장치이고,
> 여기서 나온 어떤 것도 운영자 화면이 없는 이유가 되어서는 안 된다(a101).
> **a102의 대기는 이 불변식을 지켜야 한다** — 대기 결과도 문자열로만 나온다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:91` | `load == nil` | — | `:92` `""` |
| B2 | `:95` | `load()` 오류 | — | `:96` "설정을 읽을 수 없다: …" |
| B3 | `:98` | `!on` | — | `:99` `""` |
| B4 | `:101` | `start == nil` | — | `:102` "이 빌드에는 soak 기동 배선이 없다" |
| B5 | `:106` | `start()` 오류 | — | `:111` "soak 자동 시작 실패: …" |
| B6 | `:108` | 실패 note가 비어 있지 않음 | 문장 이어붙임 | (B5의 return으로) |
| B7 | `:113` | 성공 note가 공백뿐 | `note = "서베이를 시작했다."` | `:116` "soak 자동 시작: …" |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `load()` `:94` | 승인 조회 | 오류는 fail-closed | ast.json |
| `start()` `:105` | 기동 | **오류가 문자열이 된다** | ast.json |
| `strings.TrimSpace` `:108`·`:113` | 빈 note 판정 | — | ast.json |
| `fmt.Sprintf` `:116` | 성공 문장 | — | ast.json |

live binding — 유일한 호출자는 `runConsole` (`console.go:378`). **a102가 바꾸는 것은 이
함수가 받는 `start` 인자**다: `bootSurveyIfAbsent` 대신 그것을 대기로 감싼 closure가 온다.
시그니처·본문·반환 계약은 전부 그대로다.

## State mutations and fallbacks

- 프로세스 밖 상태를 직접 바꾸지 않는다. `start`가 서베이를 띄운다.
- fallback이 이 함수의 전부다 — 다섯 실패 방향이 전부 **문자열 한 줄**로 끝난다.

## Safety conclusion

- Safe edit boundary: **없음 — a102는 이 함수를 편집하지 않는다.**
- High-risk impact: yes(attestation 갱신 프로세스의 기동 판정). a102는 위험을 **줄이는**
  쪽으로만 작동한다 — 같은 판정 뒤에 "엔진이 준비될 때까지"라는 지연을 하나 넣는다.
- a102가 이 함수에 의존하는 사실: **B5가 `start`의 오류를 문자열로 강등한다.** 콘솔 종료로
  대기가 끊겼을 때 a102가 돌려주는 sentinel도 여기서 "soak 자동 시작 실패: …" 한 줄이 된다
  — 콘솔이 내려가는 중이므로 그것이 맞는 문장이다.
