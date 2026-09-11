# Branch Test Map: `validate` — task 6.2 가 바꾸는 갈래만

전수 표가 아니다. 편집이 조건·경로를 바꾸는 갈래만 적고, 나머지 38 갈래는
**이 change 가 재지 않았다**(`test_execution_baseline.py` 의 몫이며 커버리지를 주장하지 않는다).

| 갈래 | 편집 뒤 무엇을 묻나 | 시험 | 편집 전 | 편집 뒤(기대) |
|---|---|---|---|---|
| B4·B5 | 요청받은 id 가 a063 인가 | `test_an_archived_adoption_is_rechecked_by_its_id` | RED(이름) | GREEN |
| B4·B5 | 복사본은 a063 이 아니다 | `test_a_copied_adoption_record_does_not_make_another_change_a063` | GREEN(이름) | GREEN(신원) |
| B14 | 적힌 경로가 a063 분석 안인가 | `test_an_archived_adoption_is_rechecked_by_its_id` | — (B5 에서 멈춤) | GREEN |
| L452 | 원장을 지금 자리에서 읽는다 | 같음 | — | GREEN |
| B32 | 리뷰 둘을 지금 자리에서 읽는다 | 같음 | — | GREEN |
| 활성 경로 | 편집 전과 같다 | `test_valid_adoption_uses_e_and_requires_complete_current_bundle` 외 이관 시험 전부 | GREEN | GREEN |

각 행은 편집 뒤 **변이로** 확인한다(넷을 하나씩 되돌려 아카이브 시험이 빨개지는지,
신원 판정을 지워 복사 시험이 빨개지는지).

## 편집 뒤 실측

| 갈래 | 시험 | 편집 전 | 편집 뒤 | 변이 |
|---|---|---|---|---|
| B4·B5 (아카이브) | `test_an_archived_adoption_is_rechecked_by_its_id` | **FAIL** | OK | M1 · M6 CAUGHT |
| B4·B5 (복사) | `test_a_copied_adoption_record_does_not_make_another_change_a063` | OK(이름이 막음) | OK(신원이 막음) | M2 CAUGHT, 신원을 지우면 복사본이 **통과**함을 직접 잼 |
| B14 | 아카이브 시험 | — | OK | M3 CAUGHT |
| L452 | 아카이브 시험 | — | OK | M4 CAUGHT |
| B32 | 아카이브 시험 | — | OK | M5 CAUGHT |
| 활성 경로 | 기존 이관 시험 전부 | OK | OK | — |
