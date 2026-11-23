# Branch Test Map: `AllReasonCodes`

- 편집 전. 커버: `../../implementation/base-cover-execgw.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestReasonCodeEnumIsStable` · `TestTheSenderDownReasonIsRegisteredInTheEnumeration`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy path 254-311 | 선택 집합(머리) | 해당 없음 | yes — 254.36,309.40 count 1 |
