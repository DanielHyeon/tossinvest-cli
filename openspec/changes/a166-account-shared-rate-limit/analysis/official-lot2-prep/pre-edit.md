# a166 공식 경로 연결 준비 — 미착수

기준 `da67f3e1e9282e181160467fb2a476345a08627a`. 읽기 전용 분석이며 제품 구현·RED 실행·승인 아님. lot1 API가 확정되면 기준 SHA와 각 source hash를 다시 확인하고 Pre-Edit 및 기존 함수별 FLM/BTM을 확정한다. 이 문서는 생산 배선이나 활성화를 승인하지 않는다.

## 기존 AST freshness

`analysis/evidence/ast/`의 `official--New`, `official--Client.authorityOriginLocked`, `official--Client.hcTransport`, `official--Client.doRequest`, `official--Client.send`, `official--tokenManager.exchange`, `official--a112MBUSReadAt`, `execgw--HTTPReplay.ReplayPlace` JSON 각각 `source_sha256`을 기준 파일 SHA-256과 비교해 일치했다. 이는 AST 자료의 파일 신선도 확인이다. 아래 분기와 제안은 소스 손 읽기이며 새 AST 분석이나 테스트 PASS가 아니다. a165 cancel 및 journal tracker는 아래 좌표 손 읽기다.

## 영향 지도와 함수 로직 / Branch Test Map 초안

| 자리 | 현재 흐름·불변식 | 연결 전 RED 계획 |
|---|---|---|
| `internal/official/client.go:99` New | defaults → options → configMu 잠금 → tokenManager에 같은 `c.hc` 전달 → seal. private authority는 constructor의 base·원래 transport 정체성을 확인(:157–168). | OFF에서 요청 bytes·순서·timeout·redirect·authority 원형 유지. enforcing은 신뢰된 단일 wrapper만 원래 private transport로 연결; 임의 wrapper·`WithHTTPClient`·base override·옵션 재실행은 권한 승격 금지. token manager와 공식 client가 같은 budget instance 사용. |
| `client.go:191` doRequest | `hc.Do` 실패를 `fmt.Errorf("%w: %s", ErrTransport, err)`로 변환해 원 오류 identity를 잃음; body 읽기 실패도 전송 오류. | 무전송 admission/첫 Write 전 종료를 typed evidence로 보존, 이미 Write한 오류는 ambiguous; 문자열만 검사하지 않음. HTTP 응답 상태·body·기존 trace 및 OFF 반환 보존. |
| `client.go:320` send | token → 첫 doRequest → 401일 때 refresh와 최대 두 추가 doRequest, adopted 아니면 break. | 실제 official.Client와 로컬 fake에서 token 교환·401 재전송 각각 별 admission. 첫 mutation Write 뒤 후속 prewrite 실패가 원 attempt의 불명 결과를 지우지 않음. |
| `internal/official/token.go:130` exchange | 독립 `m.hc.Do`; 요청 생성·전송·body 오류를 ErrTransport로 문자열화. | 보호 ctx 등급 상속, token Write와 주문 본체 Write 분리. token 실패가 주문 미전송 판단을 오염시키지 않으며 429 이벤트는 body 성공과 독립. |
| `internal/official/a112_mbus_read.go:103` | sealed authority 검증 후 원래 transport clone·proxy nil·직접 client Do(:121–153). | enforcing에서 clone도 같은 budget 적용; a112 전용 header/redirect/cache/timeout과 권한 검사 보존. nil/off는 기존 경로 그대로. |
| `internal/execgw/replay_transport.go:68` | 자체 client `Do`, replay body·headers 보존. | 명시 budget instance의 매 wire admission, 기존 idempotency/replay 상태와 body 불변. OFF golden. |
| `tools/a087-market-sell-probe/send.go:54,85` | 별도 send client·Do. | 별도 수동 승인 도구는 budget 누락 시 시작 거부; 기능 시험 fake만, 실주문 금지. |
| `internal/execgw/cancel_refusal.go:43–104` | 첫 409 ambiguous 후 재전송 NOT_SENT 또는 context 종료면 첫 결과 보존(:91–100). | 첫 admission 거절은 해당 attempt NOT_SENT; 첫 409 후 재전송 admission 거절은 최초 ambiguous 보존, `DispatchVerified` 종단 검증. |
| `internal/journal/dispatch.go:250–305` | `httptrace` tracker는 header callback으로 partial 표지. | 토큰과 주문 본체의 tracker 분리 및 prewrite 확실성 증명; tracker 미관측만으로 Write가 없었다는 추정 금지. 불확실하면 ambiguous. |

## lot1에 필요한 최소 계약 — 확정 전 협의

- 동일 계좌 instance를 official/token/a112/replay/probe에 명시 주입하고 OFF(nil)와 required 누락 오류를 구분하는 opaque budget/transport 계약. 임의 wrapper·client의 private authority 승격 불가.
- caller context의 critical/normal/aux 등급과 취소를 유지하며 각 실제 요청 시작에 독립 소유권을 적용한다. 반환 body 수명·late dial·drain을 포함해야 한다. 429 event 기록과 기록 실패의 fail-closed 계약은 후속 lot3 범위이며 여기서 구현·검증하지 않는다.
- `errors.Is` 또는 별도 typed result로 **현재 주문 본체의 첫 Write 이전임이 증명된 경우**만 전달하고, 이미 시작했거나 구별이 불가능하면 ambiguous. `doRequest`·token의 `%s` wrapping과 journal tracker의 공유 context 경계를 해결해야 한다.
- 현재 lot1 초안 `Config{Root,Scope,Instance,Clock}`, `Provision(root,scope,instance)`, `Open`, `Acquire(ctx, Priority)`, `Lease.Release`, `ErrUnavailable`, `ErrNotSent`은 개발 중이다. 이름·의미·transport 경계·error 증거는 동결되지 않았고 연결 코드를 작성하지 않는다.

## 범위·skill·검증

기존 연구 `analysis/manager-review-v4.md`의 search-first(Go 1.25/1.26.5, x/sys v0.46.0, stdlib/flock·라이선스), grill-me/grilling, impeccable N/A 기록을 재사용하되 lot1 API 및 코드 기준 변경 시 재확인한다. scoped test·전체 gate·운영·실계좌 호출·비밀 접근은 수행하지 않았다. 실행 전 실제 선택 경로마다 FLM/BTM, 영향 caller, RED·GREEN·race·변이와 unrun 범위를 새 기준에 고정한다.

## prewrite 증거 전파 추가 검토 — 손 읽기

`official/client.go:191–196`은 `observeAttempt`에 원 `hc.Do` 오류를 넘긴 뒤 `fmt.Errorf("%w: %s", ErrTransport, err)`로 반환한다. `official/trace.go`의 observer는 호출자가 설치할 때만 동작하고 request role·첫 Write 여부 필드가 없다. gateway mutation `internal/execgw/gateway.go:711–729`은 이를 설치하지 않고 `journal.WithSendTracker`의 단일 tracker를 사용한다. `token.go:130–146`은 같은 ctx의 별도 `hc.Do`이며 AttemptObserver를 호출하지 않는다. tracker의 httptrace는 token과 주문·401 후 요청을 같은 단조 상태로 합칠 수 있어 token만 보낸 뒤 주문을 안 보낸 경우도 보수적으로 ambiguous가 될 수 있다. 반대로 앞선 401 주문 시도의 tracker 상태가 올라갔다면 뒤 prewrite 실패가 그 상태를 되돌리지 않으므로 tracker 자체가 그 이유로 NOT_SENT로 낮아지지는 않는다. trace callbacks가 실제 첫 plaintext Write와 동치라는 증명은 별개다.

최소 후보: enforcing에서만 검증된 `ErrNotSent` identity를 `fmt.Errorf("%w: %w", ErrTransport, err)` 같은 표준 wrapping으로 보존한다(문자열은 기존 형식과 일치 가능). OFF는 기존 코드·오류 문자열·`errors.Is`를 그대로 둔다. 다만 오류 identity 하나만으로 거래 전체를 NOT_SENT로 바꾸면 안 된다. 관문이 보증한 **해당 주문 본체**의 prewrite와 이전 주문 본체 시도가 없음을 함께 판정하고, 증거 미확인·선행 시도·write 진행은 ambiguous를 유지한다. token 요청과 주문 요청의 분리된 attempt 증거가 필요한데 현 `AttemptTrace`만으로는 이를 증명하지 못한다. 별도 타입/상태가 정말 필요한지는 lot1 transport API 확정과 실제 gateway 경로 조사 후 최소로 선택한다. 기존 public `AttemptTrace` 표면 시험도 영향 범위다. a165 최초 409 뒤 재전송 NOT_SENT의 최초 ambiguous 보존은 `cancel_refusal.go:91–100`이다.

RED 후보: OFF 오류 문자열 및 `errors.Is(ErrTransport)` 동일; enforcing prewrite 거부 identity 전파; token만 전송 후 주문 미전송(잘못된 NOT_SENT/ambiguous 판정 방지); 첫 주문 Write/401 이후 후속 prewrite 거부는 이전 시도 보존; Write 진입 후 취소·연결 오류는 ambiguous; trace callback 선점/누락 시 안전 실패; 최초 409 후 재전송 admission 거절은 첫 ambiguous 보존; 사용자 observer의 조합과 기존 strict-read trace 표면 회귀. 이 절은 소스 손 읽기이며 실행·AST 분기 증명 아님.
