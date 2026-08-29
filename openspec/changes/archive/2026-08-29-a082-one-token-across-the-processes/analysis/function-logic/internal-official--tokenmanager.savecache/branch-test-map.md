# Branch Test Map: `tokenManager.saveCache`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 디렉터리를 만들 수 없다 | 기존 커버리지 (손대지 않음) | no | yes |
| B2 | marshal 실패 — 구조체가 고정이라 도달 불가 | — | no | n/a |
| B3 | 임시 파일을 만들 수 없다 | 기존 커버리지 (같은 원인으로 B1과 함께 걸린다) | no | yes |
| B4 | `Chmod` 실패 → 임시 파일을 남기지 않는다 | 기존 커버리지 | no | yes |
| B5 | 쓰기 실패 → 임시 파일을 남기지 않는다 | 기존 커버리지 | no | yes |
| B6 | `Close` 실패 → 임시 파일을 남기지 않는다 | 기존 커버리지 | no | yes |
| (성공) | 갱신 중에 읽어도 온전한 토큰 하나를 본다 | `TestAReaderNeverSeesAHalfWrittenCacheFile` | **yes** — 변이 M6(`os.WriteFile`로 되돌림)에서 400회 중 **244회**가 빈 파일·파싱 실패 | yes (0회) |
