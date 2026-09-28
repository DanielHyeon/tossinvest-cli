# a125 design — 특례를 지우고 a063 을 일반 경로로 옮기는 순서

## D1 — 지우는 것과 남기는 것

| 자리 | 처분 |
|---|---|
| `execution_baseline.py` (validate · draft CLI · 원장 계산) | 삭제 — 부르는 곳은 `check_analysis.resolve_base` 하나, draft CLI 는 README 가 a063 에만 권한다 |
| `test_execution_baseline.py` | 삭제 — 지운 모듈의 단위 시험. 반전된 행동 시험은 `test_check_analysis.py` 에 둔다(D3) |
| `check_analysis.py` 이관 분기 | 삭제. 함수 시그니처에서 `adopted`/`audited` 인자를 뺀다(호출자 전부 이 파일) |
| `ADOPTION_REFUSES_A_LANDING` 문장 | 삭제 — a063 은 착지 규칙을 일반 change 와 같이 받는다 |
| `SDD_BASE_REF` 검사 | 유지 — 대조 대상이 `effective`(이관이면 E)에서 persisted 로 단순화된다 |
| a063 `execution-baseline.json` | 삭제 커밋 |
| a063 원장 · 이관 리뷰 문서 · `historical-planning-base/` | 유지(역사 기록, 게이트 미참조) |
| `sdd_doctor` `SDD_PYTHON` | 유지(스펙은 RENAMED + MODIFIED 로 이관 서술만 뺀다) |

## D2 — 남은 기록은 읽지 않는다(무력) — fail-closed 가 아니다

두 선택지가 있다. (a) 기록을 발견하면 거절. (b) 읽지 않음. (b) 를 고른다.

- (a) 는 판정에 새 입력(파일 존재)을 더한다. 오늘 모집단에서 그 입력이 참인 곳은 a063 하나이고, 삭제 커밋이 그것을 없앤다.
- 아카이브된 과거 스냅샷을 재검사할 때 (a) 는 이유 없는 거절을 만든다.
- (b) 는 "기록이 창을 바꾸지 않는다" 를 시험으로 못 박는다(D3). 기록의 **모양**이 판정에 닿지 않으므로
  위조 · 복사 · 아카이브 변형을 따로 막을 필요도 없어진다.

## D3 — 시험 반전 규칙

특례를 못 박던 시험은 삭제하지 않고 **같은 픽스처로 반대 사실**을 못 박는다(대상 목록은 편집 전 AST 로 센다 — 0.3).

- 유효 a063 기록 + base 파일 → 비교 기준은 `base-commit.txt`(P), 문맥에 이관 키가 없다.
- a063 에 착지 기록 → 착지 규칙이 판정한다(`ADOPTION_REFUSES_A_LANDING` 이 안 나온다).
- `SDD_BASE_REF=E` → 거절(P 가 아니므로). `=P` → 수락.
- 아카이브·복사된 기록 → 일반 판정.
- 출력에 `audited source-commit` · `adoption exception` 라벨이 **없다**.
- import 수준: `check_analysis` 가 `execution_baseline` 을 import 하지 않고, 저장소에 그 모듈이 없다.

## D4 — a063 재고정(사람 절차) 조건

`docs/WORKFLOW.md` 「사람 승인 base 재고정」의 a071 변형이다.

1. 귀속 실측: base 뒤 자기 Go 비병합 커밋과, 그 커밋들이 고친 기존 함수 ↔ a063 번들 대조표. 옛 디렉터리명 없음(renumber 이력 확인).
2. 번들 신선도: current 6 이 워킹트리 소스와 해시가 맞는가. base 3(`soak_test` 둘 · `console_test` 하나)은 함수가 오늘
   존재하는가 — 존재하면 일반 규칙에서 `revision: base` 는 틀린 revision 이므로 current 로 재추출하는 근거를 적는다.
3. 새 base 는 재고정 커밋의 부모(그 시점 HEAD)다. 영수증은 a063 `review.md` 에, `base-commit.txt` 는 단독 커밋.
4. 프로브: `check_analysis.py --change a063-…` 결과(rc · required · 오류 줄)를 기록한다. 통과해도 a063 의 4.2~4.4 는 남는다.

## 검증 순서

1. 편집 전 Python FLM(대상 함수 AST 열거).
2. RED: 반전 시험(D3)을 먼저 쓰고 오늘 코드에서 빨간 것을 확인한다.
3. 창 영향 전수: 활성+아카이브 전 change 의 편집 전/후 판정(rc · required · 대상 라벨)이 a063 외에는 같다.
4. GREEN 삭제.
5. 변이: 사본 · 무변이 대조군 선행. 예: `resolve_base` 에 이관 분기를 되살림, 착지 거절 문장 복원 → 반전 시험 빨강.
6. 적대 보이스 1 + gstack.
