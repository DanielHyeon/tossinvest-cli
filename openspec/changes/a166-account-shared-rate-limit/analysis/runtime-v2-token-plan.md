# runtime v2와 token 관측 결속 — 후속 로트

기준: 게시 main `2938622a8bb8ef43e0f11e410e7affaf784bcab1`. observed 후보 `aee3b8df6`의 독립·Manager 검증과 게시를 확인했다. 독립 설계 검토는 P0–P2 없음, 기준 SHA 표기 P3를 이 문단에서 고쳤다. Manager는 이 제한 로트 구현을 승인한다. 기존 `observation-transport-plan.md`의 전송·정리 계약을 변경하지 않는다. 이 로트는 required factory와 token 종류만 연결하며 Retrier poll 소비와 전체 참여자 완료를 선언하지 않는다.

## 조사와 결정

Manager와 Terra가 현재 runtime_linux.go의 Open/readManifest/rejectDuplicates/NewClient/Close/Budget, observed_linux.go의 OpenObserved, observation_linux.go의 OpenObservation와 token.go의 exchange를 손으로 읽었다. 기존 stdlib context, strict JSON parser, trusted path/identity 검증과 관측 store를 확장한다. 새 의존성·네트워크 서비스·lock 구현은 필요 없다. 저장소 source와 Go 1.25 API로 충분해 외부 라이브러리 검색·설치를 확대하지 않았다. 신규 외부 라이선스 영향 없음. 사전 AST/FLM/BTM은 구현자가 최신 통합 기준으로 고정한다.

manifest는 `version: 2`와 최상위 `event_lock: {device, inode}`를 필수로 갖는다. 기존 `identity.root/queue/wire` 구조는 보존한다. 모든 nested field의 중복·unknown·잘못된 타입·trailing JSON을 기존 parser로 거절하며 event pin의 0도 거절한다. v1을 호환 허용하지 않는다. required v1·missing/copy/replaced event lock·dirty/invalid state는 client 노출과 wire 전에 오류다. OFF는 기존 nil budget과 official.New를 유지한다.

Open은 외부 manifest를 기존 규칙으로 읽고 OpenObserved에 동일 root/scope/instance, 신뢰 clock.System 및 clock.SystemBootSource, 명시 event pin을 넘긴다. 실제 타입 생성 형태는 기존 clock API를 재사용한다. 이어 기존 VerifyIdentity(root/queue/wire)가 성공한 같은 Budget만 Factory에 넣는다. 모든 초기화 실패에서 원 오류와 Budget.Close 오류를 보존한다. OpenObserved 내부 실패가 이미 닫은 객체를 이중 Close하지 않는다. store는 별도 영속 fd를 보유하지 않으므로 Factory.Close 소유권은 Budget 하나로 유지한다. Close active/retryable 실패 및 재사용 규칙을 바꾸지 않는다.

자동 ProvisionObservation, v1 migration, 임의 event pin 자기 승인, 운영 manifest 수정은 없다. 테스트에서만 임시 root를 provision하고 그 결과를 외부 manifest에 명시한다. runtime가 초기화 중 기존 state를 읽는 것은 검증이며 새 상태를 생성하는 provisioning과 구분한다. parser fixture와 현재 required factory를 쓰는 테스트 입력은 v2로 이행하되 deprecated v1 거부 fixture를 남긴다. 상수 치환으로 검증을 생략하지 않는다.

## token 종류

accountbudget 내부의 전용 `WithTokenObservation(ctx)` helper는 기존 private scope의 collector를 유지하고 kind만 Token, key는 zero로 바꾸는 파생 context를 반환한다. priority, deadline, cancellation 및 다른 context 값은 부모를 계승한다. 외부에서 scope의 private collector를 읽거나 setter로 공유 객체를 바꾸지 않는다.

official token.exchange의 accountbudget.Transport 가지에서만 HTTP request 생성 전에 위 helper를 적용한다. 캐시 hit에는 새 wire나 event가 생기지 않는다. OFF의 cache·OAuth body·status/error·retry·선택 계좌 동작은 변경하지 않는다. required의 token 429는 같은 store에 Token event로 남고 부모 Query collector도 유지하지만 poll hint를 쓰지 않는다. token HTTP response를 받은 뒤 기록 실패가 OAuth HTTP 결과를 NOT_SENT로 바꾸지 않는다. request parent cancellation과 deadline을 연장하지 않는다.

## 수용 증거

1. v1·event pin 누락/0/복사·nested duplicate/unknown·wrong type·dirty state·root/queue/wire mismatch를 wire0으로 거절. 실패 후 fd/owner 정리. OFF 선택 golden 유지.
2. 같은 v2 factory의 복수 client가 동일 Budget/store를 사용. fake local TLS에서 실제 official.Client token + GET, cached token, 401 후 token 교환 및 재전송마다 Begin/Finish. 순서·wire 개수·Token/Other 종류·공유 429 event를 확인한다.
3. 명시 Poll scope의 실제 official.Client가 token POST에서 poll key를 제거하고 collector/critical/취소·deadline을 보존. token 429는 부모 collector에 남고 해당 poll slot hint는 0. GET 429와 body 실패도 event를 유지한다.
4. 기존 a112 clone·budgeted replay·probe fixture와 두 process factory event 조회에서 같은 binding을 증명한다. 별도 참여자 runtime가 아직 미배선이면 실제 범위를 기록하고 전체 completion을 거절한다. 순수 leaf fixture를 production wiring 증거로 대신하지 않는다.
5. runtime Close active/error와 OFF/unsupported 플랫폼 계약을 선택 시험으로 확인한다. 전 source census에서 raw required constructor 잔여를 분류한다. 이번 로트 밖 소비자 공백은 없어진 것으로 표시하지 않는다.

모든 네트워크 fixture와 mutant는 fake 입력·로컬 TLS·고정 DialContext 및 loopback-only/no-route namespace. 원본 default endpoint에 닿아야 하는 RED는 사용하지 않는다. 필요한 경우 nil 기본 동작을 유지하는 옵션 seam을 먼저 고정한다. 정확 선택 normal/tagged race, 변이, scope 수동 gstack와 독립 검증을 시행한다. 전체 tests/gate·브로커·운영·활성화는 실행하지 않는다.

## 다음 로트와 미해결 경계

Retrier 소비는 별도 문서로 확정한다. 현재 canonical design:73의 engine poll 역할은 검증된 engine Retrier 묶음이며 같은 RequiredQuery를 쓰는 보호·일반 read가 같은 hint에 보류될 수 있다. 보호 read를 자동 제외하는 새 정책은 도입하지 않는다. 주문 본체/token·flatten/별도 walk에 계좌 전역 hint를 적용하지 않는다. AccountRef/selected seq/동일 client 결속, slot 안정성, AuthFatal 우선 처리와 429 뒤 freshness·재시도 차단은 후속 로트 필수다.

grill-me/grilling: 계좌 전체 3건/초 우선·보호 대기 허용은 사용자 확정이며 재질문하지 않는다. v1 cutover는 개발 계약이고 실제 운영 전환 권한은 아니다. 실패 경계와 수용 조건을 위에 고정했다. search-first: 기존 구현 확장 결정. impeccable: N/A — backend runtime/token-only. gstack/AST/선택 시험은 구현 이후 증거로만 기록한다.
