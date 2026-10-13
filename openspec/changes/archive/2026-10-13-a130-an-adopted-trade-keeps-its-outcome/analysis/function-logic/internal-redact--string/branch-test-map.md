# Branch Test Map: `String`

- Source: `internal/redact/redact.go` (79-109) — `727a841e` 기준(a133 `2c2daae7` 편집 뒤, 그 뒤로 이 파일 무변경).
- 출처: a133 이 편집한 함수가 a130 창 안에 끼어서 요구됨 — 겹친 change. 시험은 a131 · a133 시험을 인용함.
- 측정(2026-10-13, HEAD `272dba00`): 시험마다 따로
  `go test -count=1 -run '^<Test>$' -covermode=count -coverpkg=./internal/obs,./internal/redact -coverprofile=<f> ./<pkg>`
  (전부 rc 0). 칸의 `a/b` = 분기 문장을 담은 블록의 실행 횟수 / 분기 본문(같은 줄, 조건 뒤에서 시작하는 블록)의 실행 횟수.
  `if` 는 `a > b` 이면 거짓 갈래도 지남.

| Branch | Anchor | Scenario | Test | 측정 |
|---|---|---|---|---|
| B1 | if at 80:2 | 빈 문자열은 그대로 | `TestA133AdjacentDerivedIDsAreAllRedacted`(`String("")`) | 7/1 |
| B2 | range at 84:2 | 등록 원문 순회 | `TestA131StringReplacesEveryAccountShape` · `TestA133AdjacentDerivedIDsAreAllRedacted` | 1/2 · 6/12 |
| B3 | if at 85:3 | 등록 원문 치환(참) · 불포함(거짓) | 참: `TestA131StringReplacesEveryAccountShape` · `TestA133CriticalFieldsLeaveNoAccount`(`internal/obs`), 거짓: `TestA131StringLeavesShortNumbersAlone` | 2/2 · 52/3 · 6/1 |
| B4 | if at 89:2 | 등록 숫자열 있음(참) · 없음(거짓) | 참: `TestA131StringReplacesEveryAccountShape` · `TestA131StringLeavesShortNumbersAlone`, 거짓: `TestA133TheNotificationTestCarriesOnlyItsText`(`cmd/tossctl`, 등록 없음) | 1/1 · 2/2 · 3/0 |
| B5 | if at 91:4 | 숫자열 일치 → 가림(참) · 불일치 → 원문(거짓) | 참: `TestA131StringReplacesEveryAccountShape`, 거짓: `TestA131StringLeavesShortNumbersAlone` · `TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags`(`internal/obs`) | 6/1 · 2/0 · 3/0 |
| B6 | for at 102:2 | 유도 id 치환 되풀이 — 구분자 한 글자로 붙은 둘째 id | `TestA133AdjacentDerivedIDsAreAllRedacted` | 6/16 |
| B7 | if at 104:3 | 고정점 → 반환(참) · 바뀜 → 다시(거짓) | `TestA133AdjacentDerivedIDsAreAllRedacted` · `TestA131StringCoversSeparatedMasksAndEveryDerivedID` | 16/6 · 2/1 — 두 갈래 다 |

- a133 번들과 다른 점(측정으로 정정): a133 은 B5 의 "일치 / 불일치" 를 `TestA131StringLeavesShortNumbersAlone` ·
  `TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags` 로 적었으나 둘 다 본문 0(불일치 갈래만)임.
  일치 갈래(`return Placeholder`, 92:5)를 지나는 인용 시험은 `TestA131StringReplacesEveryAccountShape`(6/1) 하나라 그것을 더함.
- 미덮음 분기: 없음.
