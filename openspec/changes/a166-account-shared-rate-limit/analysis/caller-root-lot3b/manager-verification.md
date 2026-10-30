# Manager 검증 — 첫 root caller 배선

대상 `06066149be49ac09da2f0f891a70a89baad80d98` 및 monitor/Close 원인 보존 수정 `a8b7afd2966ad0f079be7df9c0c611e5db5bd18c`.

Manager는 root owner와 env 행렬, app의 lazy provider, engine decorator 이후 factory 재결속, flatten의 bootstrap 전 factory 주입, main 종료 및 monitor delta를 손읽기했다. OFF/WTS 선택, 원래 authority seal, caller context와 명시적 Close 수명 계약에 부합한다. monitor와 Close가 함께 실패할 때 원인을 버리던 구현은 수정했다. private wrapper는 두 오류를 `errors.Is`로 보존하고 추가 monitor sentinel 줄을 출력하지 않는다.

독립 Terra의 정확 선택 Go 1.25 normal·tagged race 원문 `/tmp/a166-root-independent-exact-go125-normal.log`, `/tmp/a166-root-independent-exact-go125-tagged-race.log`를 확인했다. cmd/app/engine 각각 1.890/0.007/0.480초 및 3.223/1.034/1.519초 PASS다. 최초 넓은 선택 실행은 중단됐고 PASS로 계산하지 않는다. delta의 세 시험 원문 `/tmp/a166-root-delta-independent-normal.log`, `/tmp/a166-root-delta-independent-tagged-race.log`는 0.795/1.908초 PASS다. 독립 AST source hash 대조와 다섯 배선 변이·원인 유실 변이 검출 기록을 확인했다. scoped 수동 gstack 검토이며 전체 자동 리뷰 파이프라인을 실행했다고 주장하지 않는다.

판정: 이 로트의 비강제 병합·푸시 승인. clock leaf가 착지한 최신 main 위에 통합하고 scoped OpenSpec·PM·diff를 확인한다. 원래 a128 ledger 변경을 보존한다. 전체 tests/gate, 실제 hard-exit kill, 다른 CLI·console·도구 배선, 보호 priority/replay 출처, 429 이벤트·hint 및 실제 운영은 이 검증에 포함되지 않는다. a087 admission nil을 유지하며 a166 전체 완료 또는 archive는 승인하지 않는다.
