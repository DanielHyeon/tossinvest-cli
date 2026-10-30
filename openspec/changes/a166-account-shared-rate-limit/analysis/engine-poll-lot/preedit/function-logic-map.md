# 사전 함수 논리 지도

기준 cf0f0919e19bd384995b8bfb7fb8ee44a42031a7. 인접 16개 AST의 source_sha256·분기 열거가 정본이며 기존 분기 전체 보존. UI N/A backend-only; 기존 store·Retrier·Go stdlib 재사용, 신규 의존성 없음. 승인 `engine-poll-observation-plan.md` 및 두 독립 검토 원문 참조.

- Transport.RoundTrip: 기존 사전 URL/proxy 검사→Acquire→Begin→poll recheck→wire→Finish→release 및 ErrNotSent/body 소유. Acquire 전 collector latch 추가 시 기존 429/local error와 body Close 한 번·ErrNotSent 보존. nil collector는 기존 그대로. 최초 HTTP 응답은 바꾸지 않음.
- Retrier.Query: 기존 policy/default clock·attempt 루프, OK freshness, AuthFatal gate/escalation, permanent/cancel return, rateLimitWait, transient backoff. required immutable binding이 있으면 AccountRef/client/seq/kind 검증→공유 hint→query-local collector를 먼저 적용하고 매 fn 반환 직후 결과 검사. AuthFatal 우선 부작용, 이후 관측 429/fault 무재시도·성공 기록0. OFF 모든 기존 분기 유지.
- NewContext: 기존 account resolve 성공 뒤 실 official/Budget/selected seq를 바인딩하고 실패 startup 거절. buildGateway/newRetrier는 기존 journal/gate/clock/authority 배선과 오류 정리를 유지하며 binding 전달. exitSideRetrier는 값 복사로 같은 immutable binding 참조.
- NewTracer, NewExitObserver, Context.ReconcileDriver, NewPositionPolicyCommandService: 기존 nil/설정/게이트 가드 유지, required에서 구체 *official.Client 일치 검증, nil·foreign·uncomparable 안전거부; OFF 기존 injection 허용. Context.ExitObserver/private reconcileFloor의 공식 client 배선을 손읽기·AST로 확인.
- Query closure 9개는 tracer 네 곳, exitloop 가격, reconcileFloor 두 holdings, adoption 가격, position policy 가격. 본문 자체 수정 없이 source/AST 증명. 각 보호/일반/재편입 동일 account/kind 동일 key. accountbudget에 새 key/remaining helper 및 execgw immutable constructor는 신규 함수로 기존 함수 사전 FLM 대상 아님.
