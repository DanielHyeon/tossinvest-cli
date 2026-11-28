# Branch Test Map: `engineFillDetector` (cmd/tossctl)

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | hints 있음 | `TestTheHintConsumerIsValidatedAtAssembly` | 기존 | yes(`../../green/green-2-neighbours.txt`) |
| B2 | hints 검증 실패 | 같음 | 기존 | yes |
| return@771 | 경보 통로 배선(nil/비nil) | `TestA191TheDetectorIsWiredToTheEngineNotifier`, `TestA191FillAlertsFilterTheNilNotifier` | **관측 안 함** — 시험을 구현 뒤에 씀. 대신 변이 m28(배선 줄 삭제)·m29(nil 거름 삭제) CAUGHT(`../../mutation/mutation-run-3-wiring.txt`) | yes |
