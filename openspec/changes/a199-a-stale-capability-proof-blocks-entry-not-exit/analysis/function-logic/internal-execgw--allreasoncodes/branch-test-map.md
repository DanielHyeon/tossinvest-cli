# Branch Test Map: `AllReasonCodes`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-execgw.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestReasonCodeEnumIsStable` · `TestTheSenderDownReasonIsRegisteredInTheEnumeration`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy path 254-314 | 선택 집합(머리) | 해당 없음 | yes — 254.36,312.40 count 1 |
