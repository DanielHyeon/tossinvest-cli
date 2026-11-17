# Branch Test Map: `engineRuntime`

> 편집 전(pre-edit). 커버 카운트는 2단계 Pre-Edit(tasks 2.3)에서 `go test -coverprofile`(대상 패키지만)로 채운다 — 아래 「GREEN observed」 의 「미측정」 은 그 전까지의 값이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 감지기 생성 실패 | 표 시험에 행 없음(그 표는 B2·B3·B4) — 다른 시험은 2.3 커버로 확인 | 해당 없음(기존 분기) | 미측정(2.3) |
| B2 | 대사 드라이버 생성 실패 | 같은 시험 「B2 reconcile construction」 | 해당 없음(기존 분기) | 미측정(2.3) |
| B3 | exit 관측자 생성 실패 | 같은 시험 「B3 exit observer construction」 | 해당 없음(기존 분기) | 미측정(2.3) |
| B4 | 복구 생성 실패 | 같은 시험 「B4 recovery construction」 | 해당 없음(기존 분기) | 미측정(2.3) |
| B5 | 전략 진입 감독자 생성 실패 | 표 시험에 행 없음 — 2.3 커버로 확인 | 해당 없음(기존 분기) | 미측정(2.3) |
| B6 | 배달 실행자 생성 실패 | 표 시험에 행 없음 — 2.3 커버로 확인 | 해당 없음(기존 분기) | 미측정(2.3) |
| (성공) | 보조 실행자 집합에 `in-doubt-settle` 포함, 감독 루프 집합 불변 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(기대 목록 갱신) · `TestTheAlertDelivererIsNotASupervisedLoop`(무편집) · 2단계에 생길 a193 배선 시험(tasks 3.10 — 이름은 아직 없는 함수라 여기 적지 않음; check_analysis 가 존재하지 않는 시험 이름을 거절함) | 새 시험: 편집 전 FAIL 예정(이름 없음) | 2단계 |
