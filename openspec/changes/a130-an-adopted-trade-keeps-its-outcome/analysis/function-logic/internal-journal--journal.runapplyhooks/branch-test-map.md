# Branch Test Map: `runApplyHooks`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:268` 훅 셋 다 nil | 훅 없는 저널 시험 | no | no |
| B2 | `:275` Project | 전부 | no | no |
| B3 | `:276` Project 오류 | 프로젝션 실패 시험 | no | no |
| B4 | `:280` Campaign | 캠페인 시험 | no | no |
| B5 | `:281` Campaign 오류 | 있음 | no | no |
| B6 | `:285` 위험 소유자 바인딩 오류 | risk_bucket_owner_test | no | no |
| B7 | `:289` Exit | exit 시험 | no | no |
| B8 | `:290` Exit 오류 | 있음 | no | no |

## a130 시나리오(편집 계획에 묶인 시험)

- B1~B8 — 무편집 갈래 — 기존 시험(위 표)
- (리터럴) — 싱크가 핸들에 실림 — a130 거절 싱크 시험(싱크 nil 이면 0 회 · 있으면 1 회)
