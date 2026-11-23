# Branch Test Map: `judgeattestation`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파일 없음·깨짐 | `TestA135UnusableAttestationsAreCritical` | 해당 없음(기존) | yes — 99.20-101.3 count 1 |
| B2 | Verify 거절 | 아래 | 해당 없음 | yes — 104.62-105.10 |
| B3 | 오류 switch | 아래 | 해당 없음 | yes — 104.62-105.10 |
| B4 | 만료 | `TestA135StageBoundaries` | 해당 없음 | yes — 106.42-107.26 |
| B5 | 범위 상실 | `TestA135ScopeLossIsCritical` | 해당 없음 | yes — 108.54-110.46 |
| B6 | 계좌 불일치 | `TestA135UnusableAttestationsAreCritical` | 해당 없음 | yes — 111.50-112.34 |
| B7 | 필수 칸 빔 | 같은 시험 | 해당 없음 | yes — 113.45-114.29 |
| B8 | 알 수 없는 거절 | 없음 | 해당 없음 | no — 115.11-117.29 count 0(도달 불가) |
| B9 | 기한 지남·파일 통과 → proof_overdue | `TestA196DeadlineBoundaries` · `TestA196OverdueIsCriticalOncePerDay` | yes — red.txt(「증거 나이 정확히 V」 등 FAIL) | yes — 125.20-130.3 count 1 |
| B10 | 날수 switch | 아래 | 해당 없음 | yes — 132.2-133.9 |
| B11 | healthy | `TestA196ProofDeadlineGovernsTheDayCount`(7d1s) · StageBoundaries「D-8」 | 해당 없음(기존) | yes — 134.42-135.25 |
| B12 | notice(감독 증거 기한 D-7) | `TestA196ProofDeadlineGovernsTheDayCount` | yes — red.txt(만료 기준이라 healthy 로 FAIL) | yes — 136.44-137.24 |
| B13 | imminent | `TestA196DeadlineBoundaries`「정확히 2일」 | yes — red.txt | yes — 138.10-139.26 |
