# Branch Test Map: `TestA135StageBoundaries`

- 시험 함수 자신. 「덮는 시험」은 자기 자신이고, 갈래 B2~B4 는 실패 보고 경로라 통과 실행에서는 타지 않는다(정상).
- base 실행: `-run 'TestA135'` rc 0(`../../implementation/base-a135-run.txt`) — 이 함수 PASS.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 12행 순회 | 자기 자신 | 해당 없음(기존) | yes — base PASS |
| B2 | stage 불일치 보고 | 자기 자신(실패 경로) | 2단계: 「D-8」「2일 1초」 행이 여기서 FAIL 예상 | 미측정(편집 전) |
| B3 | days 불일치 보고 | 자기 자신(실패 경로) | — | 미측정 |
| B4 | critical 불일치 보고 | 자기 자신(실패 경로) | 「2일 1초」 행 | 미측정 |
