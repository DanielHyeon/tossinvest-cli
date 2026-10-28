# Branch Test Map: `Gateway.checkSymbolFree`

- Source: `internal/execgw/gateway.go` (798-829)
- AST branches 7 — 실행 칸 미측정(이번 라운드 시험 실행 금지, 2.1 전 coverprofile 로 채움)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 801:2` | unsettledFor 오류 → 전송 없음 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B2 | `if at 804:2` | 같은 종목 미종결 attempt → SymbolInFlight(모든 변경) | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B3 | `if at 810:2` | 취소·청산 매도(`raisesExposure=false`) → latch 전에 반환 | 미측정 | 계획(D11 RED-b): R3 latch 가 선 종목의 엔진 재취소·청산 매도·flatten 취소가 `checkSymbolFree` 를 통과(latch 질의 호출 0) | GREEN 재추출 때 |
| B4 | `if at 814:2` | UnresolvedAttempts 오류 → 전송 없음 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B5 | `range at 817:2` | 미해소 attempt 순회 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B6 | `if at 819:3` | attemptTargets 오류 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B7 | `if at 822:3` | 같은 종목 UNRESOLVED_IN_DOUBT → 진입 거절 | 미측정 | 계획: 기존 시험이 latch 추가 뒤에도 같은 사유(UnresolvedInDoubt)로 거절됨을 유지(가드 위치 고정) | GREEN 재추출 때 |
| (신규) | B7 뒤 · 828 앞 | `cancelRefusalLatch` 거절/오류 | 해당 없음(편집 전 없음) | 계획(D11 RED-a·c·d·e): R3 뒤 같은 종목 진입 거절 / 해제 조건별 통과 / 재시작 뒤 유지 / 다른 종목 영향 0 / 질의 오류 → 진입 전송 없음 | GREEN 재추출 때 |
