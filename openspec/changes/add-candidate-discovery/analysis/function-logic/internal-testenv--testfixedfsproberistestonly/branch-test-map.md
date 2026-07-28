# Branch Test Map: `TestFixedFSProberIsTestOnly`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 정의 allowlist 2건 순회 | 이 테스트 | yes (candidate 사본 추가 전에는 1건 — 사본이 use로 잡혀 FAIL) | yes |
| B2 | 정의 파일 읽기 실패 진단 | 동일 | no (진단) | yes |
| B3 | 선언하지 않는 파일을 allowlist에 넣으면 FAIL | 동일 | yes (가짜 항목 주입 시 FAIL) | yes |
| B4 | 프로덕션 파일 순회 | 동일 | no (구조) | yes |
| B5 | 파일 읽기 실패 진단 | 동일 | no (진단) | yes |
| B6 | 미언급 파일은 건너뛴다 | 동일 | no (구조) | yes |
| B7 | 정의 파일은 선언에 대해 면제 | 동일 | yes | yes |
| B8 | 정의 파일 안의 호출은 실패 | 동일 (`callsFixedFSProber` 경유) | yes (fsguard.go에 호출 주입 시 FAIL) | yes |
| B9 | doc 주석 언급은 use가 아니다 | 동일 | yes | yes |
