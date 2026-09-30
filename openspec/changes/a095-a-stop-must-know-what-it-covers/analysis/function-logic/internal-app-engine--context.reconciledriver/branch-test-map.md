# Branch Test Map: `Context.ReconcileDriver`

a095의 편집은 분기 없는 무조건 복사 한 줄이다. 그 줄의 효과는 생산 조립을 거치는 시험(tasks 2.5 · 2.5a · 2.5b — 알림 꺼짐 · 켜짐+topic 없음 ·
거부된 알림 블록)이 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil Context | 기존 — 무변화 | n/a | n/a |
| B2 | 자동화 게이트 미검증 | 기존 — 무변화 | n/a | n/a |
| B3 | 시세 읽기 미지정 | 기존 — 무변화 | n/a | n/a |
| B4 | 알림기 미지정 ∧ Context 알림기 있음 | 기존 — 무변화 | n/a | n/a |
| B5 | 로그 미지정 | 기존 — 무변화 | n/a | n/a |
