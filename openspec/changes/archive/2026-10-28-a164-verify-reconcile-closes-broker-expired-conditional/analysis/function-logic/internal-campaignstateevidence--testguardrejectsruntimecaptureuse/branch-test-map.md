# Branch Test Map: `TestGuardRejectsRuntimeCaptureUse`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 두 import 표기 반복 | 이 시험(본문 불변) | n/a(기존) | yes(green.txt) |
| B2 | 이름 있는 lane import 의 캡처 참조가 거절됨 | 이 시험 · 변이 M4 가 다른 시험으로 잡힘 | n/a(기존) | yes |
| B3 | 점 import 캡처 참조가 거절됨 | 이 시험 | n/a(기존) | yes |
