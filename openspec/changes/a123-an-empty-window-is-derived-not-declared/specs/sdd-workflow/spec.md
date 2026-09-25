## ADDED Requirements

### Requirement: 좁힐 수 없는 창은 역사에서 유도한다

착지 지점을 얻지 못하는 change 의 함수 분석 창은 저자가 고르는 값이 아니라 역사에서 유도되어야 한다(SHALL).

base 가 그 change 의 Go 작업 뒤로 옮겨진 change 는 창의 끝을 base 로 유도해야 한다(SHALL). 그 판정은 다음
셋이 **전부** 역사에서 참일 때만 성립한다(SHALL): 모든 `revision: current` 번들이 base 의 소스를 적는다 ·
base 뒤에 그 change 에 귀속된 Go 작업 커밋이 없다 · `base-commit.txt` 의 현재 값을 쓴 비병합 커밋이 그 change
의 번들도 함께 다시 뽑았다. 하나라도 빠지면 유도해서는 안 된다(SHALL NOT). 유도는 그 재기준화 커밋을
이름으로 말해야 한다(SHALL).

`revision: current` 번들이 하나도 없는 change 의 요구 집합은 창 전체가 아니라 그 change 에 귀속된 커밋이
고친 기존 Go 함수여야 한다(SHALL). 귀속은 착지 판정과 같은 규칙 — 그 change 디렉터리를 만지면서 Go 파일도
고친 비병합 커밋 — 이어야 하며(SHALL), 귀속 커밋이 병합뿐이면 창 전체를 요구한다(SHALL).

두 유도 모두 판정 줄에 근거를 적어야 한다(SHALL) — 재기준화 커밋 · 귀속 커밋 목록 · 요구 수. 근거 없는 0 은
면제와 구분되지 않으므로 내서는 안 된다(SHALL NOT).

#### Scenario: 재기준화된 change
- **WHEN** 번들 전부가 base 소스를 적고, base 뒤 귀속 Go 커밋이 0 이고, base 를 옮긴 비병합 커밋이 번들도 함께 다시 뽑았다
- **THEN** 5단계는 요구 0 으로 통과하며 판정 줄에 그 재기준화 커밋을 적는다

#### Scenario: base 만 옮긴 커밋
- **WHEN** `base-commit.txt` 만 바꾼 커밋이 있고 번들은 다시 뽑히지 않았다
- **THEN** 유도되지 않고 워킹트리까지의 창이 그대로 요구된다

#### Scenario: 번들이 없는 change
- **WHEN** `revision: current` 번들이 하나도 없다
- **THEN** 요구 집합은 귀속 커밋이 고친 기존 Go 함수이고, 그 집합이 비면 판정 줄에 "귀속 Go 작업 0" 을 적고 통과한다

#### Scenario: 귀속 커밋이 병합뿐이다
- **WHEN** 번들이 없고 그 change 디렉터리를 만진 Go 커밋이 병합 커밋뿐이다
- **THEN** 창 전체가 요구된다

#### Scenario: 위조된 번들
- **WHEN** 번들 하나가 base 와 다른 소스를 적는다
- **THEN** 이 요구는 적용되지 않고 기존 착지 규칙이 판정한다
