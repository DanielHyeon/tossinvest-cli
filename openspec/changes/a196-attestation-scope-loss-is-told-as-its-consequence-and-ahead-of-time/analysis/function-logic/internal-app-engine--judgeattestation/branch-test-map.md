# Branch Test Map: `judgeAttestation`

- 편집 전(pre-edit). 커버(cover mode set — 1 = 실행됨): `../../implementation/base-cover-attestation_watch.out`
  (`GOFLAGS=-trimpath go test -count=1 -run 'TestA135' -covermode=set ./internal/app/engine`, HEAD `064f59fd9`, rc 0 — `../../implementation/base-a135-run.txt`).
- RED/GREEN 열은 2단계에서 채운다. 아래 「GREEN observed」 는 base 에서 기존 시험이 그 갈래를 탔는지(커버 블록)다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파일 없음·깨짐 | `TestA135UnusableAttestationsAreCritical` | 해당 없음(기존 분기) | yes — 블록 88.20-90.3 count 1 |
| B2 | Verify 거절 | 아래 B4~B7 | 해당 없음 | yes — 93.62-94.10 count 1 |
| B3 | 오류 종류 switch | 아래 B4~B8 | 해당 없음 | yes — 93.62-94.10 count 1 |
| B4 | 만료 | `TestA135StageBoundaries`「만료 순간」「만료 1시간 뒤」 | 해당 없음 | yes — 95.42-96.26 count 1 |
| B5 | 범위 상실 | `TestA135ScopeLossIsCritical` | 해당 없음 | yes — 97.54-99.46 count 1 |
| B6 | 계좌 불일치 | `TestA135UnusableAttestationsAreCritical`「다른 계좌」 | 해당 없음 | yes — 100.50-101.34 count 1 |
| B7 | 필수 칸 빔 | `TestA135UnusableAttestationsAreCritical`「endpoint 없음」 | 해당 없음 | yes — 102.45-103.29 count 1 |
| B8 | 알 수 없는 거절 | 없음 | 해당 없음 | **no — 104.11-106.29 count 0**(도달 불가, a196 범위 밖) |
| B9 | 날수 단계 switch | 아래 B10~B12 | 해당 없음 | yes — 112.2-113.9 count 1 |
| B10 | healthy | `TestA135StageBoundaries`「D-8」「30일」 | 해당 없음 | yes — 114.42-115.25 count 1 |
| B11 | notice | `TestA135StageBoundaries` | 해당 없음 | yes — 116.44-117.24 count 1 |
| B12 | imminent | `TestA135StageBoundaries` | 해당 없음 | yes — 118.10-119.26 count 1 |
| (a196 새 갈래) | 감독 증거의 기한이 만료보다 이르다 / 기한 지남·파일 통과 / 감독 증거 없음 | design §5-1~5(새 파일 `a196_attestation_deadline_internal_test.go`) | 2단계 | 2단계 |

주의(실측, overlay 프로브): a135 픽스처 `a135Attestation` 은 감독 증거를 발급 1시간 전으로 찍는다 → D1 식의 기한 = 만료 − 1h. 편집 뒤
`TestA135StageBoundaries` 의 「D-8 은 아직 알리지 않음」(7d+1s → 기한 167h0m1s → D-7 notice)과 「2일 1초 = D-3」(→ 47h0m1s → D-2 imminent) 두 행이 뒤집힌다.
「1초 = D-1」 은 기한 −59m59s 로 `daysLeftCeil` 이 1 을 돌려줘 우연히 통과한다(구현이 0 이하를 다르게 다루면 깨짐).
