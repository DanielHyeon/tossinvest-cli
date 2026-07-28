# Branch Test Map: `TestTheConsoleIsHandedOneCapabilityAndNotABroker`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | holdings seam이 두 번째 메서드를 갖게 되면 FAIL | 이 테스트 | yes (변이로 확인 가능) | yes |
| B2 | 메서드 이름 수집(진단 경로) | 동일 | no (진단) | yes |
| B3 | orders seam이 `Orders` 하나만 선언 | 이 테스트 (이 change 추가분) | yes (seam 부재 시 컴파일/형상 FAIL) | yes |
| B4 | 메서드 이름 수집(진단 경로) | 동일 | no (진단) | yes |
| B5 | `console.Options` 리터럴 파싱 실패 시 가드가 스스로 실패 | 동일 | no (자기 진단) | yes |
| B6 | `Holdings` 필드 누락 감지 | 동일 | yes | yes |
| B7 | `Orders` 필드 누락 감지 | 동일 (이 change 추가분) | yes (필드 추가 전 FAIL) | yes |
| B8 | 필드 순회 | 동일 | no (구조) | yes |
| B9 | 예외 필드는 이름 검사를 건너뛴다 | 동일 | yes (예외 없이는 `Orders`가 금지어에 걸려 FAIL) | yes |
| B10 | 이유 없는 예외는 실패 | 동일 | yes (빈 문자열 주입 시 FAIL) | yes |
| B11 | 금지 단어 순회 | 동일 | no (구조) | yes |
| B12 | `broker`/`client`/`place`/`cancel`을 이름에 담은 새 필드는 실패 | 동일 | yes | yes |
