# Manager 검증 — 공용 helper·MCP·engine reconcile

고정 대상 `194d66548e6341d94e519e875d798df73855ecff`, 기준 `c412c0e55f2948e16e1f33f1e1e8fadaf2092c56`.

Manager는 생산 변경 세 파일을 손읽기했다. root의 성공 factory를 빌리고 required 오류를 반환하며, MCP의 기존 eligibility/WTS 분기와 engine reconcile의 계좌 선택·최초 Accounts 이전 결속을 유지한다. 외부 transport option의 authority 제한은 기존 공식 constructor가 유지한다.

독립 Terra는 정확한 다섯 시험 `TestA166RemainingOfficialFactory`, `TestA166MCPRequiredFailureDoesNotFallback`, `TestA166MCPWTSDoesNotOpenBudget`, `TestA166EngineReconcileOptionsGuard`, `TestA166EngineReconcileLocalTLS`를 Go 1.25 normal·tagged race로 실행했다. 원문 `/tmp/a166-cli-first-independent-normal.log`(0.621초), `/tmp/a166-cli-first-independent-tagged-race.log`(4.887초), `/tmp/a166-cli-first-independent-netns.log`를 Manager가 확인했다. 별도 network namespace의 loopback만 활성화되고 외부 route가 없는 경계 안에서 실행했다. fixture는 local TLS URL/transport를 명시하지만 별도 loopback 전용 DialContext는 없으므로 egress 차단 근거는 namespace다. 관련 변이 검출·원문 AST/FLM 및 scoped 수동 gstack 기록을 대조했다. 독립 P0/P1 없음.

앞선 verify 시험은 DefaultTransport 교체가 official 자체 transport를 막지 못해 가짜 자격으로 외부 401이 발생했을 가능성이 있다. 실제 endpoint/wire 횟수는 trace가 없어 확정하지 않는다. 주문·실자격 사용은 없었다는 fixture 기록을 확인했다. `test-boundary-incident.md`와 실패 원문을 보존하고 PASS 또는 의도된 RED에 포함하지 않는다. 이후 네트워크 검증은 격리 namespace만 허용한다.

판정: 이 작은 로트의 비강제 병합·푸시 승인. MCP 정상 required token/Accounts의 실제 호출 통합, verify-reconcile/OpenAPI/console/HTTP API 등 나머지 배선, 전체 tests/gate 및 운영은 미검증이다. a166 전체 완료나 archive 승인이 아니다. scoped OpenSpec·PM·diff 검사 후 통합하며 a128 기존 변경을 보존한다. 원문 공백 검사 실패 뒤 커밋이 진행된 이력은 후속 수리와 gzip 원본 보존 기록에 남기며 실패를 통과로 재분류하지 않는다.
