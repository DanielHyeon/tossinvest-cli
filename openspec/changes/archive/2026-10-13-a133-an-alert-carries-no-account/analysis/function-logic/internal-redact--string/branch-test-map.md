# Branch Test Map: `String`

- Source: `internal/redact/redact.go` (79-109) — 편집 뒤. B1~B5 는 편집 전과 같은 ID · 좌표, B6 · B7 신설(마지막 반환 자리).

| Branch | Anchor | Scenario | Test | a133 |
|---|---|---|---|---|
| B1 | if at 80:2 | 빈 문자열은 그대로 | `TestA133AdjacentDerivedIDsAreAllRedacted`(`String("")`) | 무편집 |
| B2 | range at 84:2 | 등록 원문 순회 | `TestA131StringReplacesEveryAccountShape` | 무편집 |
| B3 | if at 85:3 | 등록 원문 치환 | `TestA131StringReplacesEveryAccountShape` · `TestA133CriticalFieldsLeaveNoAccount` | 무편집 |
| B4 | if at 89:2 | 숫자열 규칙 | `TestA131StringReplacesEveryAccountShape` · `TestA131StringLeavesShortNumbersAlone` | 무편집 |
| B5 | if at 91:4 | 숫자열 일치 / 불일치 | `TestA131StringLeavesShortNumbersAlone` · `TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags` | 무편집 |
| B6 | for at 102:2 | 유도 id 치환을 되풀이 — 구분자 한 글자로 붙은 둘째 id | `TestA133AdjacentDerivedIDsAreAllRedacted` | 신설 |
| B7 | if at 104:3 | 고정점에서 반환(id 없음 → 첫 판 반환) | `TestA133AdjacentDerivedIDsAreAllRedacted` · `TestA131StringCoversSeparatedMasksAndEveryDerivedID` | 신설 |
