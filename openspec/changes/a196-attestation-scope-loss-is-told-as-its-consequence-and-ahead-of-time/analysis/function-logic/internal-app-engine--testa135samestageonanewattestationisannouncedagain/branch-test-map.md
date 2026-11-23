# Branch Test Map: `TestA135SameStageOnANewAttestationIsAnnouncedAgain`

- 시험 함수 자신. base 실행: `-run 'TestA135'` rc 0(`../../implementation/base-a135-run.txt`) — 이 함수 PASS.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전반 사건 수 ≠ 2 보고 | 자기 자신(실패 경로) | 2단계: D5 편집 뒤 FAIL 예상 | 미측정 |
| B2 | 전반 Key 같음 보고 | 자기 자신(실패 경로) | — | 미측정 |
| B3 | 후반 신원 보고 | 자기 자신(실패 경로) | D1 식 때문에 통과 예상(통과 이유가 바뀜) | 미측정 |
