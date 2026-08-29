# Branch Test Map: `newEngineCmd`

**GREEN 칸은 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이고 **비어 있다**: 분기 0 · 이탈 1 (편집 전에도 0 · 1).

## 분기가 없다 — 표는 **정상 경로 한 줄**이다

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | **분기 없음 — 정상 경로.** 세 생성자를 부르고 조립한 명령을 반환한다 (`:123`) | `cmd/tossctl/a098_the_operator_command_names_a_person_test.go:134` (`TestTheAlertCommandsAreWiredUnderEngine`) | **yes** — `newEngineAlertsCmd` 가 없어 build failed | **yes (2026-08-13)** — 두 leaf 를 트리에서 찾는다 |

> **⛔ 이 `B1` 은 AST 의 분기가 아니다 — 분기가 **없기** 때문에 요구되는 자리다.**
> `check_analysis` 가 *"branchless function still needs one happy-path row"* 로
> 요구한다. 조건 없는 함수도 **하는 일**이 있고, 그것을 아무도 안 재면 「분기가
> 없으니 잴 것도 없다」가 되어 등록이 통째로 빠져도 초록이다.
> 그것이 첫 판이 이 파일을 비워 뒀다가 검사에 잡힌 이유다.

## 그럼 이 편집의 GREEN 은 어디서 재나

이 함수가 하는 일은 **등록** 하나다. 등록됐다는 것은 명령 트리에서 잰다.

| 성질 | 어디서 | 뮤테이션 |
|---|---|---|
| `tossctl engine alerts list`·`ack` 가 트리에 있다 | `cmd/tossctl/a098_the_operator_command_names_a_person_test.go:134` (`TestTheAlertCommandsAreWiredUnderEngine`) | **T** — 등록 인자 제거 → 두 leaf 를 못 찾아 FAIL |
| 기존 둘이 그대로 있다 | 같은 테스트가 `engine reconcile-resolve` 를 **대조군**으로 찾는다 | 같은 뮤테이션이 이것도 잡는다 |
| `ack` 에 확인 플래그가 없다 | 같은 테스트 | **U** — `--confirm` 추가 → FAIL (R6) |

> **⛔ 대조군을 같은 테스트 안에 둔 이유.** 「`ack` 에 `--confirm` 이 없다」는
> **조회가 고장 나도 통과한다.** 그래서 `reconcile-resolve` 에는 **있다**는 것을
> 같은 방법으로 먼저 확인한다 — 그것이 없으면 이 단언은 측정이 아니라 오타다.

## 덮이지 않은 것을 이름으로 적는다

- **`newEngineCmd` 를 직접 부르는 테스트**는 있다 — 위 테스트가 `newRootCmd()` 를
  통해 부른다. 4.2·4.4b-1 의 `runEngineRun` 과 달리 여기는 하니스가 필요 없다:
  이 함수는 **아무것도 실행하지 않고** 조립만 하기 때문이다.
- **하위 명령이 실제로 동작하는 것**은 이 함수의 성질이 아니다. 같은 파일의 나머지
  여섯 테스트가 그것을 재고, 그 GREEN 은 `alerts` 명령 자신의 것이다.
