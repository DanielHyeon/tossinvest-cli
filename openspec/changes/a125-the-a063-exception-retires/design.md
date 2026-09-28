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
- **침묵 건너뛰기 원칙과의 관계**(Manager 조건, 2026-09-28): 기록을 읽는 코드(`execution_baseline.py`)가 삭제되므로
  이 파일은 판정으로 들어가는 문이 아니라 데이터다. 그 모양을 거절하는 fail-closed 가드를 두면, 읽는 코드가 없는
  모양을 지키는 죽은 가드가 된다. "조용히 건너뛰는 갈래가 공격 경로가 된다" 는 원칙은 **판정이 읽는 입력**에만
  적용되고, 여기서는 그런 갈래 자체가 없다.

## D3 — 시험 반전 규칙

특례를 못 박던 시험은 삭제하지 않고 **같은 모양의 픽스처로 반대 사실**을 못 박는다(대상 목록은 편집 전 AST 로 센다 — 0.3).
옛 픽스처(`_adoption_with_complete_bundle`)는 지울 모듈의 `draft` 를 부르므로, 반전 시험은 모듈 없이 세운
**정적 기록** 픽스처(`_a063_fixture` — 키 집합은 a120 초안기와 같고 `planning_base`·`execution_base`·`source_commit` 은
실제 픽스처 커밋)를 쓰고, id 는 리터럴 `a063-align-attestation-renewal-profile` 이다(id 로 우회하는 변이를 잡는다 —
freeze F8). 기록이 FIFO · 디렉터리 · 심링크 · 해독 불가여도 일반 판정이 나온다(F9 — "안 읽음" 의 핀).

- 유효 a063 기록 + base 파일 → 비교 기준은 `base-commit.txt`(P), 문맥에 이관 키가 없다.
- a063 에 착지 기록 → 착지 규칙이 판정한다(`ADOPTION_REFUSES_A_LANDING` 이 안 나온다).
- `SDD_BASE_REF=E` → 거절(P 가 아니므로). `=P` → 수락.
- 아카이브·복사된 기록 → 일반 판정.
- 출력에 `audited source-commit` · `adoption exception` 라벨이 **없다**.
- import 수준: `check_analysis` 가 `execution_baseline` 을 import 하지 않고, 저장소에 그 모듈이 없다.

## D4 — a063 재고정(사람 절차) 조건

`docs/WORKFLOW.md` 「사람 승인 base 재고정」의 a071 변형이다.

1. 귀속 실측: base 뒤 자기 Go 비병합 커밋과, 그 커밋들이 고친 기존 함수 ↔ a063 번들 대조표. 옛 디렉터리명 없음(현재
   디렉터리는 `47a7f90a` 에서 태어났다 — `0c563c6c`·`b8f31f27` 의 "[a063]" 은 renumber 전 다른 change, 현 a069). `c727ad12`(S)는
   HEAD 조상이 아니고 `676bd4b4` 가 main 사본이다(freeze F7).
2. 번들 신선도(freeze F4): base 번들 3 은 E(=`676bd4b4^`)의 파일 해시를 적는데 세 함수는 오늘 존재한다 — 일반 규칙에서는
   current 여야 하므로 **current 로 재추출**한다. 그 뒤 특례 없는 도구로 **옛 base P 에서** `check_analysis` 를 돌려, 출력에서
   9 함수 이름의 missing · stale · revision 오류가 0 임을 필터 영수증으로 남긴다 — 재고정 뒤에는 어떤 게이트도 ① 을 다시
   재지 않으므로 이 영수증이 유일한 기계 증거다.
3. **재고정은 a063 게이트 직전에 한 번**(Manager 결정 (나), 2026-09-29). 재고정 값은 게이트 시점의 사실이라 사람 운영(a063
   4.2~4.4)이 끝난 뒤에만 의미가 있다 — 중간 재고정은 낡을 값을 커밋하는 것이다. a125 는 ① 영수증까지이고, 게이트 직전 재고정은
   WORKFLOW 「사람 승인 base 재고정」을 그대로 밟는다: P → 그때 HEAD 귀속 재실측(오늘의 '9' 는 오늘의 측정값) · 번들 재검증 ·
   분리 커밋 사람 귀속 확인(codex F2).
4. 영수증의 판정은 **양성 단언**이다(codex F1): 판정 완료 · 실제 P/HEAD/창 · 귀속 함수별 `(소스, 함수, 번들)` 대응 · 그 번들의
   전체 오류 0. 이름 필터는 번들 디렉터리명으로 나오는 오류와 조기 실패를 놓친다.

## D5 — 불변 주장의 범위(codex F4)

"a063 외 판정 불변" 은 **현재 모집단**의 주장이다. 비교 계약은 판정 줄 · 문맥(제거 승인된 키 `execution_baseline_adoption` 만
정규화) · rc · `main` 출력이다. 증거는 셋이다 — 모집단(기록 보유 a063 하나) · 옛 판본 논증 핀(`ab_absent_record.py`, guard 단언) ·
옛/새 A/B(`ab_old_new.py`: 워킹트리 창 · 착지 기록 · 빌림 세 모양, 차이 0). 기록을 가진 비-a063 change 가 생기면 옛 도구는 거절,
새 도구는 무시로 갈린다 — D2 가 승인한 차이다.

## 검증 순서

1. 편집 전 Python FLM(대상 함수 AST 열거).
2. RED: 반전 시험(D3)을 먼저 쓰고 오늘 코드에서 빨간 것을 확인한다.
3. 창 영향 전수: 활성+아카이브 전 change 의 편집 전/후 판정(rc · required · 대상 라벨)이 a063 외에는 같다.
4. GREEN 삭제.
5. 변이: 사본 · 무변이 대조군 선행. 예: `resolve_base` 에 이관 분기를 되살림, 착지 거절 문장 복원 → 반전 시험 빨강.
6. 적대 보이스 1 + gstack.
