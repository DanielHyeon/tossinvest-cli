# sdd-workflow Specification

## Purpose
Manager/Teammate 역할 분리와 리뷰·완료 게이트를 포함한 TossOS SDD 개발 워크플로 요구사항.
## Requirements
### Requirement: OpenSpec 기반 변경 관리
신규 기능·동작 변경·안전 경로(주문 실행, 위험관리, 원장, reconciliation) 수정은 OpenSpec change로 정의된 후에만 구현되어야 하며(SHALL), 구현 전 `openspec validate <change> --strict`를 통과해야 한다. 오탈자·주석·문서 수정, 동작 불변 의존성 patch, 테스트만 추가하는 변경은 change 없이 가능하다. 보안·자금 위험의 긴급 수정은 즉시 구현하되 24시간 내 사후 change로 문서화해야 한다(SHALL).

#### Scenario: 스펙 없는 기능 구현 시도
- **WHEN** openspec change 없이 신규 기능 코드가 제출되면
- **THEN** Manager 리뷰에서 반려된다

#### Scenario: 긴급 보안 수정
- **WHEN** 자금 위험이 있는 결함을 즉시 수정하면
- **THEN** 24시간 내 해당 수정을 기술하는 사후 change가 생성된다

### Requirement: 역할 분리
Manager(총괄 아키텍트)는 스펙 작성·검토·검증만 수행하고 구현 코드는 Teammate(구현 에이전트)가 작성해야 한다(SHALL). 구현을 생성한 컨텍스트와 이를 검증하는 컨텍스트는 별도 세션으로 분리되어야 한다(SHALL). Teammate는 blocking 수준 스펙 결함 발견 시 구현을 중단하고 `openspec/changes/<id>/issues.md`에 기록 후 Manager에 보고해야 한다(SHALL).

#### Scenario: blocking 스펙 결함 발견
- **WHEN** Teammate가 안전·동작에 영향을 주는 스펙 모순을 발견하면
- **THEN** 구현을 중단하고 issues.md에 기록하며, Manager가 스펙을 수정한 뒤 재개한다

### Requirement: 등급제 문서 리뷰 게이트
change의 proposal·design·spec 델타는 첫 구현 task 착수 전(proposal-freeze) gstack 리뷰를 1회 통과해야 하며(SHALL), 이후 Requirement 수준의 스펙 수정은 수정분에 대한 리뷰를 재실행해야 한다(SHALL). 리뷰 결과(일시, 보이스 구성, 발견 요약, 수용/거절과 근거)는 `openspec/changes/<change-id>/review.md`에 기록되어야 한다(SHALL). tasks.md 상태 갱신·오탈자·리뷰 결정 반영은 면제된다.

#### Scenario: 리뷰 기록 없이 구현 착수 시도
- **WHEN** review.md가 없는 change에 대해 `make gate CHANGE=<id>`를 실행하면
- **THEN** 게이트가 실패한다

### Requirement: 테스트 동반 구현
기능 구현은 해당 기능을 검증하는 테스트가 같은 change 안에 존재하고 통과하는 상태로만 완료될 수 있다(SHALL). 각 Requirement는 최소 1개의 테스트 또는 검증 명령으로 추적 가능해야 한다(SHALL).
여기서 "통과"는 **저장소 자신의 게이트가 실행한** 통과여야 한다(SHALL). build tag 뒤에 있어 완료 게이트도 CI도 실행하지 않는 테스트는 이 요구를 만족시키지 못한다(SHALL NOT). 따라서 완료 게이트와 CI는 무태그 실행과 `tossos_testseams` 태그 실행을 **둘 다** 수행해야 한다(SHALL).
테스트가 상수에서 유도되는 값을 단언할 때는 그 값을 상수에서 계산해야 하며(SHALL), 계산 결과를 리터럴로 굳혀서는 안 된다(SHALL NOT) — 상수를 바꾸는 change가 아무 신호도 받지 못하기 때문이다.

#### Scenario: 기능 커밋 리뷰
- **WHEN** 기능 커밋을 리뷰하면
- **THEN** 해당 기능의 테스트가 같은 change 안에 존재하고 통과한다

#### Scenario: 게이트가 실행하지 않는 테스트
- **WHEN** `tossos_testseams` 태그 뒤에만 존재하는 테스트가 실패하는 상태로 완료 게이트를 실행하면
- **THEN** 게이트가 실패한다

#### Scenario: 상수에서 유도되는 기대값
- **WHEN** 어떤 테스트가 두 런타임 상수의 비로 결정되는 호출 횟수를 단언하고 그 상수 중 하나가 바뀌면
- **THEN** 그 테스트는 새 상수로 계산한 값을 기대하고 통과하거나, 계약이 실제로 깨졌을 때 실패한다

### Requirement: 자동화된 완료 게이트
change 완료 선언은 `make gate CHANGE=<change-id>`(tasks.md 미완료 항목 0건 + review.md 존재 + test·test-seams·vet·validate 통과) 성공과 Manager의 diff 리뷰·독립 테스트 재실행 이후에만 가능하다(SHALL). task 완료 체크는 그 산출물을 만드는 커밋과 같은 커밋에서 수행해야 한다(SHALL). 완료된 change는 `openspec archive`로 확정 스펙에 반영한다.

#### Scenario: 미완료 task가 있는 완료 시도
- **WHEN** tasks.md에 미완료 체크박스가 남은 상태로 gate를 실행하면
- **THEN** 게이트가 실패하고 미완료 항목이 출력된다

#### Scenario: 태그 스위트가 깨진 상태의 완료 시도
- **WHEN** 무태그 스위트는 통과하지만 `tossos_testseams` 태그 스위트가 실패하는 상태로 gate를 실행하면
- **THEN** 게이트가 실패한다

### Requirement: 실계좌 보호의 기계적 강제
자동 테스트는 실계좌 주문을 발생시켜서는 안 되며(SHALL NOT), 이는 규칙이 아니라 테스트 인프라로 강제되어야 한다(SHALL): 테스트는 격리된 임시 config 디렉터리에서 실행되고, 실 endpoint 호출은 httptest 대체 없이는 구성될 수 없어야 한다. 실계좌 검증은 사용자 승인 하의 수동 절차로만 수행한다.

#### Scenario: 주문 로직 테스트 실행
- **WHEN** 주문 관련 테스트가 실행되면
- **THEN** 격리된 config 경로와 httptest 서버만 사용되며 실제 API 주문 호출은 발생하지 않는다

### Requirement: 최상위 안전 불변식
docs/WORKFLOW.md §0의 안전 불변식은 모든 방법론·스펙보다 우선한다(SHALL). 특히: 개발·테스트 중 승인 없는 LIVE 주문 side-effect 금지, 토글 OFF 시 upstream 동작 보존, 손절·비상 청산 즉시성 약화 금지, 손절·익절·사이징 변경은 보수 방향만 허용(불명확 시 변경 금지), 운영 토글 flip은 사람 승인 필수.

#### Scenario: 사이징 로직 완화 변경 제출
- **WHEN** 위험 기반 수량 계산을 더 공격적으로 바꾸는 변경이 명확한 근거 없이 제출되면
- **THEN** 안전 불변식 §0.9 위반으로 반려된다

### Requirement: High-risk Pre-Edit 선언
High-risk 경로(주문 제출·취소·정정, 손절/사이징, Guardian, journal·원장, reconciliation, retry matrix, 인증·세션, 체결 감지)의 기존 코드를 수정하기 전에 Teammate는 Pre-Edit 선언(change/task id, 대상 심볼, 기존 동작 근거, upstream 테스트 영향, 실패 테스트 선행 여부, §0 검토)을 구현 보고에 기록해야 한다(SHALL). 근거 없는 기존 함수 내부 수정은 금지된다(SHALL NOT).

#### Scenario: 선언 없는 High-risk 수정
- **WHEN** Pre-Edit 선언 없이 주문 경로 코드 수정이 보고되면
- **THEN** Manager 리뷰에서 반려되고 선언 후 재작업한다

### Requirement: 완료 보고 조건
Teammate의 완료 보고에는 실행한 테스트 명령과 실제 결과, 변경 파일 요약, DoD 충족 여부, High-risk 영향 여부, upstream 테스트 회귀 여부, 남은 위험이 포함되어야 한다(SHALL). 하나라도 없으면 완료로 취급하지 않는다.

#### Scenario: 테스트 결과 없는 완료 보고
- **WHEN** 테스트 실행 결과가 없는 완료 보고가 제출되면
- **THEN** 완료로 인정되지 않고 검증 후 재보고를 요구한다

### Requirement: Full SDD 권위 계층

모든 신규 기능·동작 변경은 OpenSpec 계약, 현재 HEAD와 CodeGraph hard evidence, CodeGraphContext 보조 문맥, 기존 함수 내부 변경 시 Go AST·ast-grep Function Logic Map, Superpowers TDD, gstack 게이트 순서로 수행되어야 한다(SHALL). CodeGraphContext·GBrain·기억·관측 그래프는 advisory이며 OpenSpec, 현재 HEAD, 테스트, gstack을 대체해서는 안 된다(SHALL NOT).
일반 변경의 함수 분석 비교 기준은 불변 `base-commit.txt`여야 한다(SHALL).

함수 분석 비교의 **대상 쪽 끝**은 그 change 의 작업이 착지한 지점이어야 하며(SHALL),
그 지점이 기록된 change 에 대해 워킹트리를 대상으로 삼아서는 안 된다(SHALL NOT).
착지 지점이 기록되지 않은 change 는 워킹트리를 대상으로 한다(SHALL) — 작업 중인
change 의 판정은 바뀌지 않아야 한다.

이 규칙이 필요한 이유는 병합이다. 대상이 워킹트리이면, base 와 워킹트리 사이에 들어온
다른 change 의 함수가 이 change 가 고친 함수로 집계되어, 어떤 change 도 답할 수 없는
질문이 된다. 특히 배포 후 실측 태스크는 정의상 다른 작업이 착지한 뒤에 닫히므로 그런
태스크를 가진 모든 change 가 이 상태에 빠진다.

착지 지점의 기록은 위조 가능해서는 안 된다(SHALL NOT). 그 값은 비교 대상을 정하는
유일한 손잡이이므로, 유효한 값의 조건과 기록 시점이 명시되어야 한다(SHALL).
유효성은 그 change 의 **증거**로 판정되어야 하며(SHALL), change 의 신원이 그 커밋에
존재하는지로 판정해서는 안 된다(SHALL NOT) — `base-commit.txt` 는 불변이므로 그 존재는
freeze 이후 모든 커밋에서 참이고 아무것도 가르지 못한다. 구체적으로, 착지 지점에서
그 change 의 모든 `revision: current` 증거 묶음의 source hash 가 일치해야 한다(SHALL).

그 대조만으로는 값이 **정해지지 않는다**. 활성 change 13건을 전수로 재면 고정을
통과하는 커밋이 12건에서 2~59개이고 그 수는 번들 수와 무관하다. 저자가 그중 가장 낮은
것을 고르면 창이 비어 요구 집합이 ∅ 이 된다. 따라서 착지 지점의 기록은 저자가 선언하는
값이 아니라 **도구가 그 change 의 증거로 계산해서 기록하는 값이어야 한다**(SHALL).
계산 규칙은 하나다 — 그 change 의 고정 번들이 역사에 들어온 지점 이후이면서 고정을
통과하는 **가장 낮은** 커밋. 계산이 그런 커밋을 못 찾으면 아무것도 기록해서는 안 되며
(SHALL NOT) 사유를 이름으로 말해야 한다(SHALL). 그 사유는 계산이 **걸은** 커밋에 대해서만
말해야 하며(SHALL) 걷지 않은 리비전을 두고 주장해서는 안 된다(SHALL NOT) — 걷기 실패
17건 전수에서 4건은 증거가 하한 아래 커밋을 전부 맞게 기술했다. 어느 후보도 "판정이 읽은
묶음을 들고 있지 않다"로 거절되지 않았으면, 무엇이 틀렸는지는 걸은 첫 커밋을 거절한 유효
조건의 문장으로 말해야 한다(SHALL).

착지 지점은 그 change 의 고정 번들이 가리키는 소스 중 **하나 이상**을 비교 base 와 다르게
가진 커밋이어야 한다(SHALL). 증거가 base 의 소스를 적은 채로 남으면 그 증거와 맞는 커밋은
전부 그 소스가 아직 base 와 같은 자리이고, 거기서 좁힌 창은 그 change 의 작업을 하나도 담지
못한다 — FLM 을 먼저 커밋하고 편집 뒤 번들을 갱신하지 않은 change 와, 작업 이전에서 갈라진
곁가지에 증거를 두고 병합한 change 가 그 모양이다. base 가 작업 **뒤로** 옮겨진 change 도 증거
내용으로는 같은 모양이라 착지 지점을 얻지 못한다 — 둘을 가를 정보가 저장소에 없고, 이것은 알고
치른 대가다(2026-09-14 전수: 착지를 얻던 76건 중 8건). 이 조건은 다른 유효 조건들 **뒤**,
계산값 대조 **앞**에 서야 한다(SHALL).

착지 지점 **뒤에** 그 change 자신의 Go 작업이 더 있으면 그 지점은 착지가 아니다(SHALL NOT).
구체적으로, 그 change 의 디렉터리를 만지면서 Go 파일도 고친 비병합 커밋이 후보 **뒤에** 있으면 그
후보를 받아서는 안 된다(SHALL NOT). 기록은 한 번 쓰이면 다시 대조되지 않으므로, 기록 뒤의 리뷰
수리는 고정 파일을 고쳐도 창 밖에 남아 통과한다 — 번들을 갱신하지 않은 저자가 통과하고 갱신한
저자가 막히는 역전이 그 결과다(2026-09-16 실측). 이 조건은 다른 유효 조건들 **뒤**, 계산값 대조
**앞**에 서야 한다(SHALL) — 앞에 서면 그 조건들의 거절 지점을 가로채서, 번들이 그 커밋을 기술하지
않는다는 진단이 "나중 작업이 앞선다"로 바뀐다.

이 신호는 후보를 **거절**하는 데만 써야 한다(SHALL). change 의 신원이 그 커밋에 존재하는지로
착지를 **받아들여서는** 안 된다는 위 규칙은 그대로다 — 이 판정은 창을 좁히지 못한다. 틀리는
방향은 둘이고 서로 다르다: 아닌 것을 **거절하면** 창이 넓어지고(안전한 쪽), 맞는 것을 **놓치면**
이 규칙이 없던 상태로 남는다(창이 그대로 좁다). 그래서 놓치는 자리는 전부 한계로 열거해야 한다
(SHALL). 내용으로는 가를 수 없다는 것이 전수 실측이다: 고정 소스가 착지와 지금 같아야 한다는
규칙은 착지를 얻는 68건 중 파일 단위 61건 · 함수 단위 55건을 거절했고, 원인은 거의 전부 이웃
change 의 커밋이었다.

이 신호를 재는 이력 조회는 **역사를 단순화해서는 안 된다**(SHALL NOT). 경로를 제한한 조회는
기본으로 단순화를 하고, 병합이 그 경로에 대해 한 부모와 같으면 반대편 가지를 통째로 버린다 —
그러면 곁가지에서 한 자기 수리가 안 보이고, 이 규칙이 닫으려는 구멍이 **병합**에서 다시 열린다
(2026-09-16 적대 리뷰가 픽스처 셋으로 재현). 판정은 그것을 실행하는 사람의 git 설정의 함수여서도
안 된다(SHALL NOT) — 이름 인용과 rename 탐지를 명령줄에서 고정한다. 신호를 못 재면(이력 조회
실패) 빈 신호로 물러나서는 안 되며(SHALL NOT) 판정이 되어야 한다(SHALL): 빈 신호는 "거절할 것이
없다"로 읽혀 가드를 조용히 끄기 때문이다.

두 모양은 이 규칙이 막지 못하며 **알려진 한계**다: Go 수리와 그 change 의 문서를 서로 다른 커밋으로
쪼갠 경우와, 병합 커밋 **자신의** 변경으로 고친 경우다(이력 조회가 병합의 변경을 읽지 않는다 — 위
하한과 같은 부류). 이것은 망각을 막는 가드이지 위조를 막는 가드가 아니다.

거절된 기록을 옮기는 경로를 사유 문장이 말해야 한다(SHALL) — 번들을 갱신하고, 기록을 지우는 커밋을
한 뒤, 다시 기록한다. 기록을 덮어써서는 안 된다(SHALL NOT): 덮어쓰면 그 값이 무엇이었는지가 아무
데도 안 남는다. 그 경로가 문장에 없으면 성실하게 번들을 갱신한 저자가 "이 증거가 기술하는 리비전이
아니다"와 "이미 있다 — 덮어쓰지 않았다" 사이에 갇힌다.

고정 번들을 역사에 처음 들인 커밋이 **병합 커밋뿐**이면 하한이 서지 않고 그 change 는 착지
지점을 얻지 못한다 — 하한을 찾는 이력 조회가 병합 커밋 자신의 변경을 읽지 않기 때문이다. 이것은
알려진 한계이고 막는 쪽으로 틀린다: 기록이 없으면 5단계는 워킹트리를 대상으로 판정한다. 도구와
5단계는 그 사유를 말할 때 증거가 커밋되지 않았다고 말해서는 안 되며(SHALL NOT), 병합이 아닌
커밋이 그 증거를 들이지 않았다고 말해야 한다(SHALL).

완료 게이트는 기록된 값이 그 계산값과 **같은지 확인해야 한다**(SHALL). 유효 조건을
통과하는 값이 여럿이면 저자는 그중에서 고를 수 있다 — 활성·아카이브 전수에서 착지를
얻는 76건 중 65건이 그렇고(최대 516개) 44건은 그 선택이 판정 입력을 바꾼다. 그 선택이
오늘의 역사에서 창을 넓히는 방향뿐이라 해도, 값을 만드는 주체를 도구로 옮긴 규칙은
게이트가 그 값을 확인할 때만 규칙이다. 이 확인은 다른 유효 조건들 **뒤에** 서야 한다
(SHALL) — 앞에 서면 그 조건들의 거절 지점을 가로챈다.

기록된 착지 지점이 그 change 의 고정 번들이 역사에 들어온 커밋보다 **앞서서는 안 된다**
(SHALL NOT). 저자는 오늘 만든 번들을 과거 커밋에 넣을 수 없으므로 그 지점이 저자가
고를 수 없는 유일한 하한이다. 고정 번들이 역사에 하나도 없으면 그 기록은 유효하지
않다(SHALL) — 게이트가 끝난 뒤 지울 수 있는 파일은 아무것도 고정하지 못한다.

아카이브가 증거 묶음을 **그대로** 옮기는 것이 그 하한을 올려서는 안 된다(SHALL NOT).
옮기기 전에 유효했던 기록이 아카이브만으로 무효가 되면 아래 "아카이브된 change 의
재검사"가 깨진다. 건너뛰는 것은 내용이 **바이트까지 같은** 이동뿐이어야 한다(SHALL) —
옮기면서 고친 증거는 새로 들어온 증거이고, 정규 파일과 심링크 사이의 전환도 내용
교체다. 하한 계산은 그것을 실행하는 사람의 git 설정에 의존해서는 안 된다(SHALL NOT).

판정이 **읽은** 증거를 착지 지점 커밋이 들고 있어야 한다(SHALL). 그 커밋에 그 change 의
`ast.json` 이 없거나 그 내용이 판정이 읽은 것과 다르면 그 기록은 유효하지 않으며(SHALL),
도구도 그런 커밋을 계산해 기록해서는 안 된다(SHALL NOT). 하한은 증거의 **경로**를 보고
판정은 그 **내용**을 읽으므로, 둘이 갈리는 자리마다 저자는 하한을 안 움직이고 증거를 바꿀
수 있다 — 자리표시자를 일찍 커밋해 두고 워킹트리에서 갈아 끼우기, 감시 경로 밖을 가리키는
심링크, 아예 커밋하지 않기가 그 자리다.

그 대조가 **공집합 위에서 참이 되어서는 안 된다**(SHALL NOT). 착지 지점을 고정하는
`revision: current` 증거가 하나도 없으면 그 기록은 유효하지 않다(SHALL). "모두 일치"는
번들이 0 이면 자동으로 참이고, 그러면 저자가 구간의 바닥을 골라 요구 집합을 비운 뒤
면제 표식으로 통과할 수 있다. 증거를 다른 change 에서 빌리는 change 는 착지 지점을 가질 수
없다(아래).

빌린 증거를 쓰는 change 는 비교 base 를 빌려주는 change 와 공유해야 하지만(SHALL), 비교 창을
**좁혀서는 안 된다**(SHALL NOT). 그 change 의 착지 지점 기록은 받지 않으며(SHALL NOT) 5단계와
기록 명령은 같은 사유를 이름으로 말해야 하고(SHALL), 빌려주는 change 의 기록을 그 change 의
비교 대상에 써서는 안 된다(SHALL NOT). 빌려주는 change 의 착지 지점은 빌려주는 쪽 증거의 가장
낮은 값이라, 빌리는 change 가 그 값을 쓰면 빌리는 쪽이 그 **뒤에** 한 Go 작업이 창 밖으로 나가
어떤 판정도 보지 못한다. 빌리는 change 는 자기 번들이 없으므로 자기 작업이 어디 착지했는지 말할
증거를 소유하지 않는다.

`revision: current` 증거의 source hash 대조 대상은 비교 대상 쪽 끝과 같아야 한다
(SHALL). 착지 지점이 기록된 change 의 증거를 워킹트리와 대조해서는 안 된다(SHALL NOT).

명시적 legacy 실행 기준선 이관 경로의 비교 대상 쪽 끝은 그 이관 기록이 감사한
`source_commit` 이어야 한다(SHALL). 그 경로에서 착지 지점 기록을 받아들여서는 안 된다
(SHALL NOT) — 이관 예외의 정당성은 판정에 들어가는 입력을 빠짐없이 열거하고 digest 로
묶은 데 있는데, 착지 지점 기록은 그 열거 어디에도 없으면서 비교 대상을 고르기 때문이다.
비교 대상을 이관 기록이 아닌 워킹트리로 삼아서도 안 된다(SHALL NOT).

완료 게이트는 아카이브된 change 의 함수 분석도 그 id 로 재검사할 수 있어야 한다(SHALL).
실행 기준선 이관 예외도 같다. 이관의 신원은 디렉터리 이름이 아니라 완료 게이트가 요청받은
id 로 판정해야 하고(SHALL), 이관 기록이 적은 옮기기 전 증거 경로는 지금 자리에서 같은
digest 로 읽어야 한다(SHALL). 다른 id 의 디렉터리에 복사된 이관 기록이 그 id 를 이관 대상으로
만들어서는 안 된다(SHALL NOT). 해독할 수 없는 착지 지점 기록도 기록이며, 그것 때문에 완료
게이트가 판정 없이 중단되어서는 안 된다(SHALL NOT).

아래 명시적 legacy 실행 기준선 이관의 모든 조건을 만족하는 a063만 고정된 실행
기준 E로 판정할 수 있으며, 원래 구현 전 증거 규칙을 소급 충족했다고 보고해서는
안 된다(SHALL NOT). 원래 계획 기준 P는 계속 보존해야 한다(SHALL).

#### Scenario: 보조 문맥과 현재 HEAD 충돌
- **WHEN** CodeGraphContext 또는 기억 결과가 현재 HEAD와 다르면
- **THEN** 현재 HEAD와 CodeGraph를 다시 확인·동기화한 뒤 그 결과를 구현 근거로 사용한다

#### Scenario: 기존 함수 내부 변경
- **WHEN** 기존 함수의 분기·early return·mutation·side effect를 변경하면
- **THEN** 구현 전에 Function Logic Map과 Branch Test Map을 만들고 변경 후 source hash·함수·분기와 묶인 증거로 최신화한다

#### Scenario: Function Logic Map 면제 시도
- **WHEN** `not-applicable` 면제를 기록했지만 비교 기준 대비 기존 Go 함수가 수정되었다
- **THEN** gate는 수정 함수를 diff에서 계산하고 해당 함수의 완전한 증거 묶음이 없으면 실패한다

#### Scenario: diff 본문 줄이 파일 헤더와 글자가 같을 때
- **WHEN** 바뀐 Go 소스 줄이 `-- ` 또는 `++ ` 로 시작해서 통합 diff 본문에 파일 헤더(`--- `/`+++ `)와
  같은 모양으로 나타나면
- **THEN** gate 는 그것을 본문으로 읽고(파일 이름은 그 파일의 첫 훅 **앞에서만** 정해진다) 그 파일의
  수정 함수 요구를 그대로 유지한다

#### Scenario: git 이 바뀐 Go 파일을 이진으로 다룰 때
- **WHEN** 바뀐 `*.go` 파일을 git 이 이진으로 다루면 — `.gitattributes` 의 `binary`·`-diff` 속성이나
  실제 이진 내용 때문이며, 그 `.gitattributes` 는 추적되지 않아도 된다
- **THEN** gate 는 그 change 를 **거절한다**(요구가 조용히 0 건이 되게 두지 않는다). 판별은
  `--numstat` 의 `-`/`-` 이며, 본문이 바뀌지 않은 정상적인 mode-only 변경(`0`/`0`)은 거절하지 않는다

#### Scenario: 바뀐 Go 파일의 diff 본문이 사라질 때
- **WHEN** git 이 내용이 바뀌었다고 세면서도(`--numstat` 이 `0`/`0` 도 `-`/`-` 도 아니다) 판정이 읽는
  diff 에 그 파일의 본문이 하나도 없으면 — 외부 diff 명령이나 `textconv` 필터가 지운 것이다
- **THEN** gate 는 그 change 를 **거절한다**. 즉 gate 는 앞단 가드가 센 파일 집합과 판정이 본문을 읽은
  파일 집합을 **대조해야** 하며, 한쪽만 믿어서는 안 된다

#### Scenario: git 이 워킹트리를 비교 전에 다시 쓰거나 안 읽을 때
- **WHEN** 판정 대상이 워킹트리이고, git 설정(필터 드라이버의 clean·process, fsmonitor, stat 캐시 검사)이나
  stat 캐시 자체, 인덱스 플래그(assume-unchanged, skip-worktree), 내용을 바꾸는 속성(`ident`,
  `working-tree-encoding`)이 바뀐 Go 파일의 편집을 git 의 워킹트리 비교에서 감추면
- **THEN** gate 는 git 의 워킹트리 비교를 쓰지 않고, 추적된 `*.go` 의 디스크 바이트를 직접 읽어 base 와
  대조해야 한다(SHALL) — 그 편집이 요구하는 Function Logic Map 은 설정이 없을 때와 같다
- **AND** 그 대조를 위해 gate 는 실제 인덱스와 객체 저장소에 아무것도 쓰지 않는다
- **AND** skip-worktree 인 파일이 디스크에 없으면 sparse-checkout 이 꺼내지 않은 것으로 보고 인덱스의 내용을
  쓴다. 충돌 중이거나 인덱스가 일반 파일이라 하지 않는 `*.go` 는 이름을 대고 거절한다
- **AND** gate 가 띄우는 git 은 교체 참조(`refs/replace`)를 따르지 않는다

#### Scenario: 저장소가 객체 이름(oid)에 다른 바이트를 내놓을 때
- **WHEN** 저장소가 교체 참조, 위조한 loose·pack 객체, 다시 쓴 트리나 커밋으로 어떤 oid 에 그 oid 로 해시되지 않는
  바이트를 담고 있으면 — 판정 대상이 워킹트리든 커밋이든, 편집 쪽이든 base 쪽이든, 이름을 바꾼 파일이든
- **THEN** gate 가 판정하는 바이트는 디스크에서 직접 읽은 것이거나 gate 가 해시를 검증한 객체뿐이어야 하고(SHALL),
  git 의 diff 는 그 바이트만 담은 격리한 저장소에서 돈다 — 워킹트리 쪽 위조는 판정에 닿지 않고, 판정에 필요한
  객체의 해시가 맞지 않으면 gate 는 그 객체를 이름 대고 거절한다
- **AND** 목록과 diff 는 pathspec 을 받지 않으므로 pathspec 환경 변수(`GIT_*_PATHSPECS`)가 판정 대상을 바꾸지 않는다

#### Scenario: 해시는 맞지만 git 이 쓰지 않는 모양의 트리
- **WHEN** base 나 커밋 대상의 트리가 해시는 맞지만 git 이 쓰지 않는 모양이면 — 0 으로 채운 모드(`040000`), 기본 `git fsck`
  가 받지 않는 모드, 같은 이름의 항목 둘(붙어 있지 않아도), 비었거나 `/` 가 들었거나 `.` · `..` 인 이름, git 의 정렬을 어긴 순서
- **THEN** gate 는 그 트리를 git 과 다르게 읽을 수 있으므로 이름 대고 거절해야 한다(SHALL) — 그 아래 파일의 요구가
  조용히 사라지면 안 된다
- **AND** git 이 쓰는 트리(실행 파일 · 심링크 · gitlink · 하위 트리, 하위 트리를 이름 뒤 `/` 로 견주는 정렬)와 초기 git 이 쓴
  `100664` 모드는 거절하지 않는다

#### Scenario: 판정 diff 의 형식을 바꾸는 사용자 설정
- **WHEN** 사용자나 저장소 설정이 diff 의 접두사(`diff.noprefix`), 색(`color.ui` · `color.diff`), 훅 합치기
  (`diff.interHunkContext`)를 바꾸거나 환경에 `GIT_DIFF_OPTS` 가 있으면
- **THEN** gate 의 판정은 그 설정이 없을 때와 같아야 한다(SHALL) — 요구가 다른 경로로 옮겨 가거나, 헛거절하거나,
  편집하지 않은 함수까지 요구하면 안 된다

#### Scenario: 바뀐 Go 파일의 경로에 공백이나 유니코드 줄 구분자가 있을 때
- **WHEN** 바뀐 `*.go` 의 경로에 공백이나 U+2028 · U+2029 · U+0085 가 들어 있으면
- **THEN** gate 는 그 파일의 수정 함수 요구를 그대로 세고, diff 의 줄 경계는 `\n` 하나로만 읽는다

#### Scenario: 실행 기준선 이관 기록이 없는 변경
- **WHEN** 변경에 execution-baseline 이관 기록이 없으면
- **THEN** 기존 불변 기준과 전체 수정 함수 분석 규칙을 그대로 적용한다

#### Scenario: 병합 뒤에 닫히는 배포 후 실측 태스크

- **WHEN** 어떤 change 의 작업이 착지하고 그 지점이 기록된 뒤, 다른 change 들이 트리에
  들어오고, 그 change 의 배포 후 실측 태스크를 닫아 완료 게이트를 실행하면
- **THEN** 5단계는 그 change 가 실제로 고친 함수에 대해서만 증거를 요구하고, 사이에
  들어온 다른 change 의 함수를 요구하지 않는다

#### Scenario: 착지 지점이 없는 작업 중 change

- **WHEN** 착지 지점이 기록되지 않은 change 로 완료 게이트를 실행하면
- **THEN** 5단계는 지금과 같이 `base-commit.txt` 와 워킹트리를 비교한다

착지 지점을 기록한 change 의 요구 집합이 비었는데 function-logic 번들이 존재하면,
5단계는 통과하더라도 요구 집합이 비었음을 출력해야 한다(SHALL). 조용한 통과는 증거로
통과한 것과 구분되지 않는다.

5단계는 **통과하든 실패하든** 비교의 두 끝(base 와 대상)과 요구된 함수 수를 출력해야
한다(SHALL). 실패 출력이 요구된 함수 이름만 담아서는 안 된다(SHALL NOT) — 어떤 창이
그 이름들을 만들었는지가 없으면 남의 change 의 함수와 자기 함수를 가를 수 없다.
대상이 워킹트리이면 그 창에 base 뒤로 착지한 커밋이 몇 개 들어 있는지도 출력해야
한다(SHALL).
5단계가 착지 지점 기록 명령을 권할 때는 그 명령이 후보를 걷기 전에 거절할 사유가 없어야
하며(SHALL), 사유가 있으면 명령을 권하지 않고 그 명령이 낼 거절 사유를 출력해야 한다(SHALL).
그 판단은 기록 명령이 거절에 쓰는 것과 같은 판정이어야 한다(SHALL) — 따로 두면 명령에 거절이
늘 때 조언만 옛 조건으로 남는다. 후보를 걸어야만 아는 거절은 예측하지 않되, 권하는 문장이
기록을 약속해서는 안 된다(SHALL NOT). 대상을 말하는 줄은 기록이 디스크에만 있을 때도 참이어야
한다(SHALL) — 게이트는 기록을 커밋에서 읽는다.

#### Scenario: 착지 지점이 base 뒤에 있어 요구 집합이 비는 change

- **WHEN** 착지 지점이 기록된 change 의 `base-commit.txt` 가 이미 그 change 의 Go
  작업을 담고 있어 base 와 착지 지점 사이에 바뀐 함수가 없고, 그 change 에
  function-logic 번들이 있으면
- **THEN** 5단계는 번들 자체의 유효성은 계속 검사하고, 요구된 함수가 0개라는 사실을
  출력한다

#### Scenario: 착지 지점이 증거와 맞지 않는다

- **WHEN** 기록된 착지 지점에서 그 change 의 `revision: current` 증거 묶음 중 하나라도
  source hash 가 맞지 않으면
- **THEN** 5단계는 통과하지 않고 어느 묶음이 맞지 않는지 이름으로 말한다

#### Scenario: 아카이브된 change 의 재검사

- **WHEN** 아카이브된 change 의 id 로 완료 게이트 함수 분석을 실행하면
- **THEN** `archive/<YYYY-MM-DD>-<id>/` 의 기록으로 판정하고 `missing base-commit.txt`
  로 실패하지 않는다

#### Scenario: 위조된 착지 지점

- **WHEN** 착지 지점 기록이 유효 조건을 만족하지 않는 값을 담고 있으면
- **THEN** 5단계는 통과하지 않고 그 사유를 이름으로 말한다

#### Scenario: 5단계가 실패했을 때의 출력

- **WHEN** 5단계가 증거 부족으로 실패하면
- **THEN** 요구된 함수 이름과 함께 비교의 base, 비교 대상, 요구된 함수 수를 출력하고,
  대상이 워킹트리이면 base 뒤에 착지한 커밋 수를 함께 출력한다

#### Scenario: base 의 소스를 적은 번들 옆에서의 조언

- **WHEN** 비교 대상이 워킹트리이고 그 change 의 `revision: current` 번들 중 하나가
  **base 의 소스**를 적고 있으면
- **THEN** 5단계는 착지 지점을 기록하라고 권하지 않고, 어느 번들이 base 를 적고 있는지
  이름으로 말하며 그 번들을 현재 소스로 갱신하라고 말한다
- **AND** 그런 번들이 하나도 없고 기록 명령이 걷기 전에 거절할 사유도 없으면, 착지 지점을
  기록하면 창이 좁아진다는 사실과 기록할 커밋이 없으면 명령이 그렇게 말한다는 것을 출력한다

#### Scenario: 기록 명령이 거절할 change 옆에서의 조언

- **WHEN** 비교 대상이 워킹트리이고, 착지 지점 기록이 커밋되지 않은 채 디스크에만 있거나
  (아카이브 이동이 staged 인 경우 포함), 증거를 빌리거나, 고정 번들이 없거나, 추적 파일이
  수정됐거나, 고정 번들이 역사에 없거나, 기록 자리가 심링크이면
- **THEN** 5단계는 착지 지점을 기록하라고 권하지 않고, 그 기록 명령이 낼 거절 사유를 출력한다
- **AND** 비교 대상을 말하는 줄은 기록 파일이 없다고 말하지 않고 HEAD 에 기록이 없다고 말한다
- **AND** 기록이 디스크에만 있으면 기록 명령은 그 경로와 함께 그것이 커밋되지 않았다고 말한다

#### Scenario: 자기 증거보다 앞선 착지 지점

- **WHEN** 기록된 착지 지점이 그 change 의 고정 번들이 역사에 들어온 커밋보다 앞서면
- **THEN** 5단계는 통과하지 않고, 그 기록이 자기 증거보다 앞선다는 것과 그 증거가
  들어온 지점을 이름으로 말한다

#### Scenario: 역사에 없는 증거로 고정한 착지 지점

- **WHEN** 착지 지점을 기록한 change 의 `revision: current` 번들이 역사에 하나도 없으면
- **THEN** 5단계는 통과하지 않고, 그 기록을 고정할 증거가 역사에 들어온 적이 없다고
  이름으로 말한다

#### Scenario: 아카이브가 증거를 옮긴 뒤의 기록

- **WHEN** 유효한 착지 기록을 가진 change 를 아카이브하고 그 id 로 다시 판정하면
- **THEN** 5단계는 아카이브의 이동을 증거가 새로 들어온 것으로 보지 않고 아카이브
  전과 같은 판정을 낸다

#### Scenario: 아카이브가 증거를 옮기면서 고친 경우

- **WHEN** 아카이브의 이동과 **같은 커밋**에서 증거 묶음의 내용이 바뀌면
- **THEN** 5단계는 그 커밋을 증거가 새로 들어온 지점으로 보고, 그보다 앞선 기록을
  통과시키지 않는다

#### Scenario: 판정이 읽은 증거를 들고 있지 않은 착지 지점

- **WHEN** 기록된 착지 지점 커밋에 그 change 의 `ast.json` 이 없거나 그 내용이 판정이
  읽은 것과 다르면
- **THEN** 5단계는 통과하지 않고 어느 묶음인지 이름으로 말한다
- **AND** 도구는 그런 커밋을 착지 지점으로 계산해 기록하지 않는다

#### Scenario: 계산값이 아닌 수락값을 적은 기록

- **WHEN** 기록된 착지 지점이 유효 조건을 전부 통과하지만 도구가 그 change 의 증거로
  계산하는 값이 아니면
- **THEN** 5단계는 통과하지 않고, 기록된 값과 계산된 값을 **둘 다** 이름으로 말한다

#### Scenario: 하한이 사람의 git 설정에 흔들리지 않는다

- **WHEN** 저장소와 기록이 같고 실행하는 사람의 git 설정만 다르면
- **THEN** 5단계의 판정은 같다

#### Scenario: 게이트가 착지 지점을 계산해 기록한다

- **WHEN** 착지 기록이 없는 change 에 대해 기록을 요청하면
- **THEN** 도구는 고정 번들이 역사에 들어온 지점 이후이면서 고정을 통과하는 가장 낮은
  커밋을 기록하고, 그런 커밋이 없으면 아무것도 쓰지 않고 사유를 이름으로 말한다

#### Scenario: 자기를 담은 커밋에서 이미 틀린 증거

- **WHEN** `revision: current` 증거 묶음이 그것을 역사에 들인 커밋에서 이미 소스와 맞지
  않고, 그 뒤 어느 커밋도 모든 묶음과 맞지 않으면
- **THEN** 도구는 아무것도 기록하지 않고, 그 커밋에서 맞지 않는 소스를 이름으로 말한다
- **AND** 그 사유는 증거가 이 역사의 어느 리비전도 기술하지 않는다고 말하지 않는다 —
  계산은 그 커밋 아래를 걷지 않는다
- **AND** 기록이 없는 동안 5단계는 그 묶음의 source hash 가 낡았다고 이름으로 말하며
  통과하지 않는다

#### Scenario: 편집 전 소스를 적은 증거로 좁힌 착지 지점

- **WHEN** FLM 을 먼저 커밋하고 Go 를 고친 뒤 번들을 갱신하지 않은 change 에 착지 지점 기록을
  요청하거나, 그 편집 전 커밋을 착지 지점으로 기록하면
- **THEN** 도구는 기록하지 않고 5단계는 그 기록을 받지 않으며, 둘 다 고정 소스가 base 이후
  바뀌지 않았다고 이름으로 말한다
- **AND** 기록이 없는 동안 5단계는 그 묶음의 source hash 가 낡았다고 이름으로 말하며 통과하지
  않는다

#### Scenario: 작업 이전의 곁가지에서 병합된 증거

- **WHEN** Go 작업 뒤에, 작업 이전에서 갈라진 곁가지에 base 상태를 적은 증거를 두고 병합하면
- **THEN** 도구는 그 곁가지 커밋을 착지 지점으로 기록하지 않는다

#### Scenario: 병합 커밋 안에서 처음 커밋한 증거

- **WHEN** 병합을 마치는 커밋에서 그 change 의 `revision: current` 번들을 처음 커밋하고 착지
  지점 기록을 요청하면
- **THEN** 도구는 기록하지 않고, 병합이 아닌 커밋이 그 증거를 들이지 않았다고 말한다
- **AND** 기록이 없는 동안 5단계는 워킹트리를 대상으로 판정하며 그 이유로 새 오류를 더하지 않는다

#### Scenario: 착지 기록 뒤에 그 change 가 고친 Go

- **WHEN** 착지 지점을 기록한 뒤 그 change 의 디렉터리를 만지면서 Go 파일도 고친 비병합 커밋이
  서고 번들을 갱신하지 않으면
- **THEN** 5단계는 그 기록을 받지 않고, 그 change 자신의 나중 작업이 착지 뒤에 있다고 가장 오래된
  그 커밋을 이름으로 말한다
- **AND** 같은 문장이 기록을 옮기는 경로를 말한다
- **AND** 기록이 아직 없으면 도구는 그 앞 커밋을 착지 지점으로 계산해 기록하지 않는다

#### Scenario: 이웃이 나중에 같은 파일을 고친 change

- **WHEN** 착지 지점을 기록한 뒤 그 change 의 디렉터리를 만지지 않는 커밋이 같은 Go 파일을 고치면
- **THEN** 5단계는 그 기록을 계속 받는다

#### Scenario: 수리와 그 문서를 쪼갠 커밋

- **WHEN** 착지 기록 뒤의 Go 수리와 그 change 의 문서 편집을 서로 다른 커밋으로 나누면
- **THEN** 5단계는 그 기록을 계속 받는다 — 알려진 한계다(놓치는 쪽이므로 이 규칙 이전 상태로
  남는다)

#### Scenario: 곁가지에서 한 수리를 감추는 병합

- **WHEN** 그 change 의 디렉터리와 Go 를 같이 고친 커밋이 곁가지에 있고, 병합이 그 디렉터리에
  대해 줄기와 같아져 경로 조회의 기본 단순화가 그 커밋을 버리면
- **THEN** 5단계는 그 커밋을 그대로 읽어 기록을 받지 않는다

#### Scenario: 신호를 잴 수 없는 경우

- **WHEN** 이 신호를 재는 이력 조회가 실패하면
- **THEN** 5단계는 통과시키지 않고 무엇을 못 읽었는지 이름으로 말한다

#### Scenario: 증거를 갱신한 뒤의 재기록

- **WHEN** 착지 기록이 거절된 change 가 번들을 갱신하고 기록을 지우는 커밋을 한 뒤 다시 기록을
  요청하면
- **THEN** 도구는 갱신된 증거가 세우는 하한 이후의 값을 새로 계산해 기록한다
- **AND** 기록이 남아 있는 동안에는 덮어쓰지 않고 그 경로를 이름으로 말한다

#### Scenario: base 가 작업 뒤로 옮겨진 change

- **WHEN** 재기준화로 base 가 그 change 의 Go 작업 뒤에 놓여, 증거가 base 의 소스를 정확히
  적고 착지 후보에서 고정 소스가 base 와 같으면
- **THEN** 그 change 는 착지 지점을 얻지 못하고, 5단계는 워킹트리를 대상으로 판정한다

#### Scenario: 고정할 증거가 없는 착지 지점

- **WHEN** function-logic 번들이 하나도 없거나 전부 `revision: base` 인 change 가 착지
  지점을 기록하면
- **THEN** 5단계는 통과하지 않고, 그 기록을 고정할 증거가 없다고 이름으로 말한다

#### Scenario: 빌린 증거를 쓰는 change 의 착지 지점 기록

- **WHEN** `function-logic-reference.txt` 로 다른 change 의 증거를 빌리는 change 에 착지 지점
  기록이 있으면 — 빌려주는 change 의 기록과 같은 값이거나 읽을 수 없는 내용이어도
- **THEN** 5단계는 통과하지 않고, 빌린 창은 좁혀지지 않는다고 이름으로 말한다
- **AND** 그 change 에 기록을 요청하면 도구는 쓰지 않고 같은 사유를 말한다

#### Scenario: 빌려주는 change 만 착지 지점을 기록한 빌림

- **WHEN** 빌려주는 change 에만 착지 지점 기록이 있고 빌리는 change 의 Go 작업이 그 지점 뒤에
  있으면
- **THEN** 빌리는 change 의 5단계는 워킹트리를 대상으로 판정하고 그 작업을 요구한다

#### Scenario: 양쪽 다 착지 지점을 기록하지 않은 빌림

- **WHEN** 증거를 빌리는 change 와 빌려주는 change 가 둘 다 착지 지점을 기록하지 않으면
- **THEN** 5단계는 그 이유로 새 오류를 더하지 않고 워킹트리를 대상으로 판정한다

#### Scenario: 실행 기준선 이관 경로의 비교 대상

- **WHEN** 실행 기준선 이관 예외로 5단계를 실행하면
- **THEN** 비교 대상은 워킹트리가 아니라 이관 기록이 감사한 `source_commit` 이고,
  출력이 그 값과 그것을 고정한 것이 이관 감사임을 이름으로 말한다

#### Scenario: 이관 경로에 있는 착지 지점 기록

- **WHEN** 실행 기준선 이관 예외를 쓰는 change 에 착지 지점 기록이 있으면
- **THEN** 5단계는 통과하지 않고, 이관 경로가 그 기록을 받지 않는다는 것과 창의 끝이
  감사된 source commit 이라는 것을 이름으로 말한다

#### Scenario: 이관 경로의 해독할 수 없는 착지 지점 기록

- **WHEN** 실행 기준선 이관 예외를 쓰는 change 에 UTF-8 로 읽을 수 없는 착지 지점 기록이
  있으면
- **THEN** 5단계는 도구 오류로 중단되지 않고, 이관 경로가 그 기록을 받지 않는다는 것을
  이름으로 말한다

#### Scenario: 아카이브된 이관 change 의 재검사

- **WHEN** 실행 기준선 이관 예외를 쓰는 change 를 아카이브하고 그 id 로 다시 판정하면
- **THEN** 이관 예외로 판정하고, 비교 대상 쪽 끝은 아카이브 전과 같은 감사된 source
  commit 이다

#### Scenario: 다른 id 로 복사한 이관 기록

- **WHEN** 이관 예외를 쓰는 change 의 디렉터리를 다른 id 로 통째로 복사하고 그 id 로
  판정하면
- **THEN** 5단계는 그 id 에 이관 예외를 적용하지 않고 그 사유를 이름으로 말한다

### Requirement: SDD 도구의 실재와 동기 검증
에이전트 규칙에 등재된 SDD 도구와 경로는 저장소의 `make sdd-check`로 검증 가능해야 한다(SHALL).
상세 개발 절차의 단일 정본은 `docs/WORKFLOW.md`여야 하며(SHALL), Claude·Codex 진입 파일은
동일한 최소 안전 부트스트랩과 `docs/WORKFLOW.md` 포인터를 포함해야 한다(SHALL).
설치되지 않은 필수 CLI, 존재하지 않는 산출물 경로, 최소 부트스트랩 drift 또는 정본 포인터
누락은 게이트를 실패시켜야 한다(SHALL).

#### Scenario: ast-grep 누락
- **WHEN** 개발 환경에서 `make sdd-check`를 실행했으나 ast-grep CLI가 없으면
- **THEN** 설치 명령을 포함한 오류로 실패한다

#### Scenario: Codex 최소 안전 부트스트랩 drift
- **WHEN** `.codex/agents.md`의 최소 안전 부트스트랩이 `.claude/CLAUDE.md`와 다르면
- **THEN** 설정 동기 검사가 실패하고 재생성 명령을 안내한다

#### Scenario: 상세 워크플로 정본 포인터 누락
- **WHEN** Claude 또는 Codex 진입 파일에서 `docs/WORKFLOW.md` 필수 참조가 누락되면
- **THEN** 설정 동기 검사가 실패한다

#### Scenario: CodeGraph hard-evidence index drift
- **WHEN** 마지막 `make sdd-sync` 이후 tracked 또는 untracked 소스가 변경되었다
- **THEN** `make sdd-check`는 stale CodeGraph fingerprint로 실패하고 재동기화를 요구한다

### Requirement: 두 계층 기억
파일 기반 episodic memory를 primary 정본으로 사용하고 GBrain을 의미 검색 보조 계층으로 사용해야 한다(SHALL). 기억은 검증 전 자동으로 canonical 승격되어서는 안 되며(SHALL NOT), 시크릿·개인정보·검증되지 않은 실거래 수익 결론을 저장해서는 안 된다(SHALL NOT).

#### Scenario: 검증되지 않은 작업 학습 저장
- **WHEN** 완료 전 학습을 retain하면
- **THEN** episodic 상태로만 기록되고 canonical 승격은 별도 검증 근거를 요구한다

#### Scenario: 동시 retain과 canonical ID 재사용
- **WHEN** 여러 에이전트가 동시에 memory를 retain하거나 기존 canonical ID를 episodic으로 다시 retain한다
- **THEN** ledger transaction은 직렬화되고 canonical ID의 교체·강등은 거절된다

### Requirement: 비차단 SDD 관측
에이전트 저장과 git commit은 마스킹된 로컬 이벤트로 관측되어야 하며(SHALL), TypeDB SDD Control Graph와 Neo4j Create Context Graph로 best-effort 동기화될 수 있다. 관측 서비스·CLI 오류는 저장, 커밋, 테스트를 차단해서는 안 된다(SHALL NOT). StockOS와 서비스를 공유할 때 database와 source namespace는 TossOS 전용으로 분리되어야 한다(SHALL).

#### Scenario: TypeDB 중단 상태의 커밋
- **WHEN** TypeDB에 접속할 수 없는 상태에서 post-commit hook이 실행되면
- **THEN** 로컬 마스킹 이벤트는 보존되고 hook은 성공 종료하여 커밋을 방해하지 않는다

#### Scenario: 공유 Neo4j 사용
- **WHEN** TossOS 이벤트를 공유 Neo4j에 ingest하면
- **THEN** TossOS source 이름으로 기록되고 StockOS source를 덮어쓰지 않는다

### Requirement: PM 계층과 OpenSpec 역추적
활성 OpenSpec change는 PM story와 1:1로 연결되거나 명시적 bootstrap 예외를 가져야 하며(SHALL), generator는 initiative→epic→feature→story→change의 양방향 참조와 고아 항목을 검사해야 한다(SHALL). StockOS의 PM 데이터·기억·인덱스를 TossOS 사실로 복제해서는 안 된다(SHALL NOT).

#### Scenario: 고아 OpenSpec change
- **WHEN** 활성 change가 어떤 TossOS story에도 연결되지 않고 bootstrap 예외에도 없으면
- **THEN** PM check가 실패하고 누락 change id를 출력한다

### Requirement: 프로젝트 GBrain 단일 프로세스 소유권
The TossOS project-local GBrain wrapper SHALL serialize MCP and CLI processes
entering the same `GBRAIN_HOME` with a kernel-lifetime singleton lock. 살아 있는 소유자가
있을 때 후발 프로세스는 PGLite 내부 timeout을 기다리거나 잠금 파일을 삭제해서는 안 되며,
소유자 진단을 포함한 temporary-failure로 즉시 종료해야 한다.

#### Scenario: 두 에이전트가 동시에 GBrain MCP를 시작한다
- **WHEN** 첫 번째 `gbrain serve`가 project singleton lock을 보유한 동안 두 번째 세션이 같은 wrapper로 `serve`를 시작하면
- **THEN** 두 번째 실행은 실제 GBrain/PGLite를 시작하지 않고 exit 75와 busy 진단을 반환한다

#### Scenario: 소유 프로세스가 비정상 종료한다
- **WHEN** singleton lock 소유 프로세스가 정상 cleanup 없이 종료하면
- **THEN** 커널은 소유권을 자동 회수하고 다음 wrapper 실행은 stale 파일 삭제 없이 GBrain을 시작할 수 있다

#### Scenario: 변경 전 GBrain 소유자가 남아 있다
- **WHEN** singleton flock을 사용하지 않는 기존 프로세스가 같은 홈의 PGLite lock에 살아 있는 PID와 steal grace 안의 heartbeat로 기록되어 있으면
- **THEN** 새 wrapper는 그 lock을 삭제하거나 기다리지 않고 legacy 소유자를 busy로 보고한다

#### Scenario: legacy PID는 살아 있지만 heartbeat가 stale이다
- **WHEN** PGLite lock의 PID가 존재하더라도 `refreshed_at`이 GBrain steal grace보다 오래되었으면
- **THEN** wrapper는 lock을 삭제하지 않고 실제 GBrain 실행에 넘겨 upstream stale-lock recovery가 소유권을 판정하게 한다

### Requirement: GBrain contention의 advisory 격리
SDD synchronization SHALL execute GBrain commands through the project wrapper and
treat verified active-owner contention only as a GBrain freshness warning. 해당 contention은
CodeGraph hard-evidence 동기화와 성공 상태 기록을 실패시키거나 지연시켜서는 안 된다.
contention 이외의 GBrain 오류는 기존 incomplete 진단을 유지해야 한다.

#### Scenario: 활성 MCP 중 make sdd-sync를 실행한다
- **WHEN** project GBrain MCP가 singleton을 보유한 상태에서 SDD sync의 source probe가 실행되면
- **THEN** probe는 빠르게 busy로 분류되고 CodeGraph hard-evidence 동기화 결과는 정상 기록된다

#### Scenario: GBrain 자체 오류가 발생한다
- **WHEN** singleton contention이 아닌 GBrain init, source registration 또는 sync 오류가 발생하면
- **THEN** SDD sync는 해당 GBrain 오류를 incomplete failure로 보고하고 GBrain freshness를 갱신하지 않는다

### Requirement: GBrain 중복 복구의 소유자 보존
Operational recovery SHALL cross-check command line, `GBRAIN_HOME`, and the PGLite
lock owner PID.
자동 복구는 동일 홈의 비소유 중복 `gbrain serve`에만 정상 종료 신호를 보낼 수 있으며,
활성 잠금 소유자·에이전트 부모 프로세스·잠금 데이터 디렉터리를 종료 또는 삭제해서는 안 된다.

#### Scenario: 비소유 중복 프로세스를 복구한다
- **WHEN** 동일 홈의 두 `gbrain serve` 중 한 PID만 PGLite lock owner로 확인되면
- **THEN** 복구 절차는 비소유 GBrain 자식에만 SIGTERM을 보내고 소유자와 부모 에이전트를 유지한다

### Requirement: 신규 Story와 OpenSpec은 같은 번호를 사용한다
TossOS는 `a040` 이후 신규 번호형 change를 `aNNN-kebab-intent`로 명명하고 정확히 하나의 `STORY-TOS-aNNN`과 연결해야 한다 (SHALL). 기존 비번호형 change와 `STORY-TOS-001~039`는 historical legacy로 허용해야 한다 (SHALL).

#### Scenario: 정상 신규 change
- **WHEN** `a041-complete-exit-line-contract`와 `STORY-TOS-a041`이 서로를 가리킨다
- **THEN** PM 검증은 번호·slug·1:1 mapping을 승인한다

#### Scenario: 번호 불일치
- **WHEN** `STORY-TOS-a041`이 `a042-*` change를 가리킨다
- **THEN** PM 검증은 번호 불일치로 실패한다

#### Scenario: legacy 보존
- **WHEN** 기존 `STORY-TOS-039`가 기존 무번호 archived change를 가리킨다
- **THEN** PM 검증은 migration을 강요하지 않고 기존 mapping을 승인한다

### Requirement: 번호와 intent 형식은 기계적으로 검증된다
PM 검증기는 신규 change의 3자리 번호 중복, 빈 intent, 대문자·underscore·비-kebab slug를 거부해야 한다 (MUST).

#### Scenario: 중복 번호
- **WHEN** 서로 다른 두 active change가 같은 `a047` 번호를 사용한다
- **THEN** 검증은 두 경로를 모두 지목하며 실패한다

### Requirement: 고정된 legacy 실행 기준선 예외
The checker SHALL permit execution-baseline adoption only for
`a063-align-attestation-renewal-profile` with planning base
`da80ce31b6a1ab5d443016768f970a82bab102db` and execution base
`e65e394bf84b3c6e4559a219e816af96d341d75d`. It SHALL preserve the planning
base file, verify its regular committed bytes contain the full fixed P, and reject another change, base pair, malformed record or moving
snapshot reference. Invalid adoption SHALL fail closed rather than fall back.
The result SHALL be labeled `execution-baseline adoption exception` with
`retrospective-exception` provenance. The environment SHALL NOT select a base:
`SDD_BASE_REF` SHALL match only the valid record-derived effective base.

#### Scenario: 나중 기준으로 구현 변경을 숨기려는 시도
- **WHEN** 이관 기록이 허용된 E보다 뒤의 커밋이나 다른 변경 ID를 지정하면
- **THEN** 검사는 실패하고 원래 P를 덮어쓰거나 환경 변수로 우회하지 않는다

#### Scenario: 유효한 예외의 CI 기준
- **WHEN** 유효한 이관 기록의 E와 다른 커밋으로 SDD_BASE_REF를 설정하면
- **THEN** 검사는 기준 불일치로 실패한다

### Requirement: 이관 이력의 전수 회계와 독립 검토
An adoption SHALL retain a strict versioned ledger containing the complete
topologically ordered P..E reachable commit range, each commit's full parent
list and path/status differences against every parent (including merge parents
and root changes), and the net P-to-E modified-existing Go function
inventory, and the complete E..S Go path/function inventory. The validator SHALL
recompute these from immutable Git objects and reject missing, duplicate or
altered entries, ambiguous JSON, invalid schemas or digest mismatches. The
inherited range SHALL be identified as committed historical work with any
missing original analysis still outstanding; it SHALL NOT be labeled completed
FLM or a waiver. A committed adoption record SHALL bind the ledger and two
distinct, actually performed adversarial and subsequent gstack review documents
by repository-contained regular paths and SHA-256. The draft generator SHALL NOT
create approval claims, overwrite output or change a baseline or checkout.

#### Scenario: 과거 함수 한 개가 원장에서 빠짐
- **WHEN** 기록의 요약 건수와 해시가 있더라도 재계산한 P..E 함수가 원장에 없으면
- **THEN** 검사는 이력 누락으로 실패한다

#### Scenario: 소급 준수 주장
- **WHEN** 과거 누락 분석을 완료 또는 면제로 표시하거나 구현 전 작성된 증거라고 주장하면
- **THEN** 해당 이관 기록은 유효한 예외 증거로 인정되지 않는다

#### Scenario: 검토 자료 교체
- **WHEN** 원장이나 검토 문서가 지정 해시 또는 커밋된 내용과 다르거나 심볼릭 링크이면
- **THEN** 검사는 자료 교체로 실패한다

### Requirement: 커밋된 소스 스냅샷과 전체 함수 의무
Adoption acceptance SHALL run at clean detached HEAD H with full-commit source
snapshot S and verified P <= E <= S < H ancestry. H SHALL differ from S only
within `openspec/` and `docs/pm/`; staged or unstaged tracked changes SHALL fail.
All tracked Go paths, modes and bytes SHALL match S, and additional untracked or
ignored Go files and symlink substitutions SHALL fail. Other untracked/ignored
files SHALL also fail except fixed generated SDD/index/cache data locations
specified by the reviewed design. Such exceptions SHALL be regular,
non-executable, non-source files and SHALL NOT be actual Go/test/embed inputs
reported by successful untagged and `tossos_testseams` Go package enumeration;
enumeration failure SHALL block adoption. E..S Go changes SHALL be
limited to the a063 CLI soak, soak attestation/renewal diagnostic and console
files enumerated in the reviewed design. Every modified-existing E..H function
SHALL still satisfy the ordinary full bundle/hash/revision/branch/call/test
checks, with a recomputed inventory matching E..S. This exception SHALL NOT
waive operational evidence, other change tasks, full tests, or the final gate.

#### Scenario: 검토 이후 소스 추가 또는 수정
- **WHEN** H의 소스가 S와 다르거나 추적되지 않은 또는 ignored Go 파일이 존재하면
- **THEN** 이관 검사는 실패하고 검토 스냅샷의 테스트를 현재 소스의 성공으로 사용하지 않는다

#### Scenario: 실행 기준 이후 삭제된 기존 함수
- **WHEN** E에 있던 Go 함수가 S에서 삭제되고 해당 base-revision 번들이 없으면
- **THEN** 일반 함수 분석 검사가 실패한다

#### Scenario: 소스 확장자가 아닌 빌드 입력을 숨김
- **WHEN** 추적되지 않은 C/assembly/header 파일이나 embedded asset이 생기거나 허용 메타데이터 파일이 실제 Go 테스트 입력으로 사용되면
- **THEN** 확장자나 ignore 규칙으로 면제하지 않고 이관 검사를 실패시킨다

#### Scenario: 정상 SDD 인덱스가 존재함
- **WHEN** 고정된 허용 위치에 실행 불가능한 일반 인덱스 데이터만 있고 실제 빌드 입력과 겹치지 않으면
- **THEN** 그 생성 데이터 자체는 소스 오염으로 간주하지 않으며 다른 이관 조건은 모두 계속 검사한다

#### Scenario: 이관 증거만 뒤에 기록
- **WHEN** 소스 S 이후 원장과 검토 문서를 커밋한 깨끗한 detached H에서 모든 이관 조건과 함수 번들이 검증되면
- **THEN** E를 비교 기준으로 사용할 수 있지만 a063의 남은 운영 및 최종 수용 조건은 계속 검사한다

### Requirement: 이관 소스 밖의 명시적 SDD 인터프리터
The SDD doctor SHALL support an optional `SDD_PYTHON` absolute external
interpreter while preserving its existing local-venv behavior when the variable
is absent. An explicit interpreter SHALL be outside the repository both by its
path and resolved target, executable, and able to probe the repository-pinned
TypeDB driver version. Invalid explicit values or failed/mismatched dependency
probes SHALL fail without falling back to a local environment. An external
interpreter symlink to another external executable MAY be used. This option
SHALL NOT relax adoption source validation, choose an execution baseline or
create a repository-local environment. The doctor SHALL report the selected
interpreter and dependency result.

#### Scenario: 이관 작업 공간에서 외부 도구 환경 사용
- **WHEN** 소스 스냅샷 밖의 유효한 SDD_PYTHON으로 고정된 드라이버를 검사하면
- **THEN** 저장소 안에 가상환경을 만들지 않고 의존성을 확인하며 이관 소스 검사를 그대로 적용한다

#### Scenario: 잘못된 외부 인터프리터 지정
- **WHEN** SDD_PYTHON이 비어 있거나 상대 경로, 저장소 내부 경로, 실행 불가능한 파일 또는 잘못된 드라이버 환경을 가리키면
- **THEN** doctor는 실패하고 기존 로컬 가상환경으로 조용히 대체하지 않는다

#### Scenario: 기존 일반 작업 공간
- **WHEN** SDD_PYTHON을 지정하지 않으면

