# Branch Test Map: `NewContext`

**편집 전 좌표**(AST 17 분기, base `b93ed133a`, `engine.go` `451`–`633`). 실행 0 — 시험 이름은 rg 로 고른 것만, 대부분 미특정(이번 회차에 시험을 찾거나 돌리지 않았다). 편집이 닿는 것은 B5 뒤 한 줄이며 그 분기의 시험은 구현 단계 RED 에서 더한다(design §4-18).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 주문 경로 실패 | 미특정 | no | no |
| B2 | audit log 열기 실패 | 미특정 | no | no |
| B3 | 시계 기본값 | 미특정 | no | no |
| B4 | 발행기 주입 | 미특정 | no | no |
| B5 | 설정 audit 실패 → 기동 거절 | 미특정(함수 단위: `TestNotificationSettingsAreAudited` 등은 `recordGateSettings` 를 직접 부름) | no | no |
| B6 | 계좌 해석 실패 | 미특정 | no | no |
| B7 | 원장 열기 실패 | 미특정 | no | no |
| B8 | apply hook 실패 | 미특정 | no | no |
| B9 | 예산 factory 있음 | a166 시험(미특정) | no | no |
| B10 | 예산 오류 | 미특정 | no | no |
| B11 | 보호 manifest 오류 | 미특정 | no | no |
| B12 | 생산 Guardian 생성 | 미특정 | no | no |
| B13 | 기본 factory | 미특정 | no | no |
| B14 | Guardian 생성 오류 | 미특정 | no | no |
| B15 | interlock 오류 | 미특정 | no | no |
| B16 | 청산 전용 아님 → Guardian 비움 | a199 시험(미특정) | no | no |
| B17 | 전략 projection 오류 | 미특정 | no | no |
