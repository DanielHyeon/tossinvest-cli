# Function Logic Map: `TestTheCommandsRefuseWhenNoEngineIsRunning`

- Source: `cmd/tossctl/a098_the_operator_command_names_a_person_test.go` (337-359)
- AST evidence: `ast.json` — AST 분기 5 · return 0 · defer 0
  (source_sha256 는 ast.json 이 정본, a109 §2-fix F6 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 §2-fix 편집 대상: 단정 한 줄(B5)** — `strings.Contains(err.Error(), "엔진이 없다")`
  → `errors.Is(err, errEngineAlertsUnavailable)`. 나머지 구조(임시 디렉터리·두 명령 반복·
  성공 금지)는 a098 이 세운 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 임시 엔진 디렉터리 | 0700, descriptor 없음 | `os.MkdirTemp` + `os.Chmod` | 생성 실패면 `t.Fatal` (B1·B2) |
| 두 명령 | `alerts list` · `alerts ack --operator …` | 반복(B3) | 하나라도 성공하면 `t.Fatalf` (B4) |
| 거부의 정체 | `errEngineAlertsUnavailable` | `dialAlertControl` 의 sentinel | 다르면 `t.Errorf` (B5) |

**a098 이 지는 요구**: 엔진 없이 이 두 명령이 「성공」하면 원장만 고친 것이고 진입은
재시작까지 막힌 채다(design D7.1). 그리고 거부는 **운영자가 자기 경로를 의심하게 만들면
안 된다** — 날것의 `no such file` 이 새어 나가는 것이 그 사고다.

**a109 §2-fix F6 이 고친 것**: 그 요구를 **문구의 부분 문자열**로 적어 둔 탓에, a109 가
같은 문장의 단정을 조건문으로 바꾸자 이 테스트가 회귀로 신고했다. 이제 거부의 **정체**를
잰다. 문구가 무슨 말을 해야 하는지는 소유자인 a109 가 잰다
(`TestTheAlertsCLIDoesNotAssertTheEngineIsAbsent`).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:339) | 임시 디렉터리 생성 실패 | 없음 | `t.Fatal` | — (자기 자신) |
| B2 (:343) | `Chmod 0700` 실패 | 없음 | `t.Fatal` | — |
| B3 (:346) | 두 명령 반복 | 명령 실행(디스크 읽기만) | — | — |
| B4 (:351) | 명령이 **성공**했다 | 없음 | `t.Fatalf` | — |
| **B5 (:354)** | 거부가 그 sentinel 이 아니다 | 없음 | `t.Errorf` | 뮤테이션 M27b |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a098RunAlerts` | 진짜 명령 트리를 돌린다 | 오류를 그대로 돌려준다(래핑 없음) | AST · :350 |
| `errors.Is` | 거부의 **정체** 판별 | 순수 | AST · :354 |

## State mutations and fallbacks

- 테스트 지역 임시 디렉터리뿐. 실계좌·운영 경로에 닿지 않는다.

## Safety conclusion

- Safe edit boundary: **단정 한 줄**. fixture 와 반복 구조는 a098 소유다.
- High-risk impact: **no** — 테스트다. 다만 이 테스트가 지는 요구(부재 안내)는 운영자
  행동에 직결되므로 단정을 **약화**하면 안 된다.
- 금지: 이 단정을 문구의 부분 문자열로 되돌리는 것(F6 이 지운 결합), 오류 종류를 안 보고
  「오류이기만 하면 통과」로 만드는 것(M27b 가 그 판을 잡는다).
