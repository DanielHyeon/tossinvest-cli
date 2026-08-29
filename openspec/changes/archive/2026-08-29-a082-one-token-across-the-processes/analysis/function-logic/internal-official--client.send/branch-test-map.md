# Branch Test Map: `Client.send`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 토큰 획득 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B2 | 최초 요청 생성 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B3 | 최초 전송 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B4 | 401 → 갱신 후 재시도. 채택한 토큰도 거부당하면 발급해 한 번 더 | `TestGetUnwrapsEnvelopeAndRetriesOn401`(기존), `TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne` | **yes** (M5: 상한을 1로 줄이면 후자가 깨진다) | yes |
| B5 | 갱신 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B6 | 재시도 요청 생성 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B7 | 재시도 전송 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B8 | 발급한 토큰이 거부당하면 더 돌지 않는다 | `TestGetUnwrapsEnvelopeAndRetriesOn401` (기존 — 갱신이 정확히 1회임을 `calls == 2`로 고정) | no (기존 통과) | yes |
| B9 | 최종 비-2xx 분류 | `TestRawReadsClassifyErrorsLikeEveryOtherRead` | **yes** (M3) | yes |
