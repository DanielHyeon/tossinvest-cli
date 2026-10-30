> **재설계 NOT READY (2026-10-31).** 사용자는 엄격 3건/초·보호 대기 정책과 남은 개발 진행을 승인했다. 기술 차단 해소·설계 검토·READY 뒤 Terra가 제품 구현한다. 활성화·실계좌·배포 승인은 별도다. 아래 2판 구현·시험 계획은 역사 기록이며 착수 지시가 아니다.

## 현재 진행 — Manager 설계 / Terra 구현·시험

- [x] OpenSpec scoped 검증 도구 복구: 임시 exact pin 1.4.1, `e32606d4b`의 세 change strict valid. 기존 전역 링크·저장소 의존성 불변. 형식 검증이며 READY 판정 아님.
- [ ] 실제 TLS 연구: 응답 body 보존, 취소·조기 응답·지연 dial·자동 401/token·redirect 및 내부 재시도의 요청별 중재 검증. fakeConn/수동 재요청 결과만으로 닫지 않음.
- [ ] 다중 프로세스 연구: 동일 ext4 공유 잠금, 소유자 종료·재기동, 실제 요청 시작 계수와 로컬 측정 경계 명시.
- [ ] 위 증거를 근거로 최신 proposal/design/spec delta와 현재 작업표를 하나의 계약으로 정렬하고 독립 리뷰. 2판 역사 기록을 구현 근거로 사용하지 않음.
- [ ] READY 후 제품 구현 → 영향 범위 RED/GREEN·race·변이 → gstack 리뷰 → Manager 검증 → Terra 종결. 실행하지 않은 전체 검증은 별도 미실행 기록.

## 3판에서 이어받은 설계 차단

- [ ] 3판 차단 해소: 실제 wire 재시도 중재 증명, 우선 중재 및 crash/reboot·inode·body 수명 검증 설계.
- [ ] 사람 결정: 보호 의존 사슬 범위·장애 결과·지연 수용 기준·429 정책.
- [ ] 3판 재설계 후 독립 적대 리뷰 및 scoped gstack review; 실패 시 구현 금지. a087 admission nil, a169 보류.
- [ ] research-first 결과의 공식 출처·버전·라이선스와 채널 미확인 사항 정리. UI/UX 없음: impeccable N/A — backend/documentation-only.

## 2판 작업 목록 — 역사 기록, 실행 금지

## 0. 착수 준비

- [ ] 0.1 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a166-account-shared-rate-limit` 로 base 고정(a165 착지 `5e2facaee` 이후).
- [ ] 0.2 2026-10-22 「엔진 + 수동 조회 동시 429」 관측의 로그 영수증 확보(로컬 기록만 — 계좌 API 재호출 금지). 없으면 「영수증 없음」으로 기록.
- [x] 0.3 전송 자리 전수 — design §2.1(S1~S5 + 대상 아님), §2.2 프로세스, §2.3 파일시스템. CodeGraph 1.6.0 영수증 `analysis/code-context/codegraph-{callers,affected}-1st.txt`·`codegraph-callers-2nd.txt`(주 체크아웃 인덱스, 같은 코드 트리 `5e2facaee` — 이 worktree 인덱스 미초기화).
- [ ] 0.4 search-first(외부 레지스트리·MCP) — 저장소만 봤다. 외부 채널은 **안 봤음**. 들이는 것은 사람 승인 + `dependency-auditor`.
- [ ] 0.5 FLM/BTM 번들(gate ⑤ 형식) — design §10 목록(삭제 대상은 revision base). `analysis/evidence/ast/`(1판 15 + 2판 8)와 `source_sha256` 대조.
- [ ] 0.6 evidence reconciliation — `analysis/code-context/codegraph-baseline.md`·`codegraphcontext-context.md`·`evidence-reconciliation.md`(CodeGraph 가 `tokenManager.exchange` 호출자를 못 찾은 것, worktree 인덱스 미초기화 포함).

## 1. 계약 (design freeze)

- [x] 1.1 design 1판.
- [x] 1.2 spec delta 1판.
- [x] 1.3 1판 독립 적대 리뷰(GPT 보이스) → `review.md` — FAIL(P0 4·P1 4·P2 2).
- [x] 1.4 design 2판 — 단일 RoundTripper 경계(D2), 명시 예산 주입·instance(D1), wire 시점 계약·ε 재admission(D3), critical 범위·안 A/B(D4), Close drain(D7), 실패·OFF(D9), wall clock 단일 판정(D10), 불변식 4 표(§6), 처분 대조(§11). spec delta 2판.
- [ ] 1.5 사람 결정: **H2(안 A/B)** · **H11(exit `/prices`·보호 매도 접수 확인 GET 등급)** · **H12(시계 역행 1회당 초과 수락)** · **H13(350ms·ε 40ms)** · H4 값 확인 → design 3판(필요 시).
- [ ] 1.6 2판 재리뷰(적대적 Eng 포함, 작성자와 다른 계열) → `review.md`.

## 2. RED (design §9 이름 — a087·a169 가 인용)

- [ ] 2.1 Pre-Edit 선언(`review.md`) — 대상 design §10.
- [ ] 2.2 RED `internal/accountbudget` 단위: T-WIRE-1 · T-SYNC-1(프로세스 안) · T-PRIO-A 또는 T-PRIO-B(H2 결과) · T-FS-1 · 패널티(H4 값) · 시계(단일 프로세스 fake).
- [ ] 2.3 RED `internal/official` 통합: T-INT-1~3 · T-OFF-1 · T-AUTH-1 · T-429-S 의 S1·S2·S3.
- [ ] 2.4 RED 엔진·호출자·도구: T-RET-1 · T-MUT-1~3(`DispatchVerified` 종단, a165 두 경우) · T-429-S 의 S4·S5 · T-STR-2 · T-WALK-1.
- [ ] 2.5 RED 구조: T-STR-1(`go/types` 전수 — 분류 안 된 http.Client 자리 0, 수집 자리 수·허용 목록 수 단언).
- [ ] 2.6 RED 다중 프로세스(`-race`): T-MP-1(지연 주입, `base.RoundTrip` 시작 시각 기준) · T-MP-CLK(offset ±1s·역행 1h) · T-SYNC-1 의 `StopOnShared429` 자식 판.

## 3. GREEN · REFACTOR · VERIFY

- [ ] 3.1 GREEN `internal/accountbudget`(새 패키지: Budget·Transport·instance·상태 파일·이벤트).
- [ ] 3.2 GREEN `official`: `WithAccountBudget`·`AccountBudget()`·`New` 안 감싸기·`hcTransport` 벗기기·`ErrAdmissionRefused`·`a112MBUSReadAt` clone 감싸기. `doRequest`·`send`·`exchange` 본문 무변경 확인.
- [ ] 3.3 GREEN 엔진: `NewOrderPath` Budget·marker `account_budget_instance`·`ReasonAccountBudgetUnavailable`·`WithCritical`(보호 dispatch)·precheck `WithTryOnce`·`buildGateway` 재생 client·`Retrier.Query` 즉시 반환·H9 삭제.
- [ ] 3.4 GREEN 그 밖: 루트 플래그 `--account-budget-dir`·env, flatten ctx 뿌리 `WithCritical`, a087 probe(`newSendClient` 감싸기·주입 없으면 시작 거부), `compose.yaml` env. **정본 retry matrix §0.3·§1·§2·§3·§4 와 코드 기본값을 같은 커밋에서**.
- [ ] 3.5 VERIFY 영향 범위 시험 + `-race`(`make test-race` 목록에 새 패키지) — 선택 이유·미실행 범위 기록.
- [ ] 3.6 변이 원장(design §9 목록, S1~S5 `Wrap` 삭제 양성 대조군 포함) — 사본 하네스는 무변이 대조군 GREEN 확인 뒤.
- [ ] 3.7 FLM/BTM 최신화, `check_analysis.py --change a166-account-shared-rate-limit`.

## 4. 리뷰

- [ ] 4.1 다중 적대 리뷰(적대적 Eng 포함) — 반례 목표: 손절·flatten 지연(§6 표 수치 대조), 잠금 쥔 채 정지, 시계 역행·offset, 독점, 래퍼 없는 http.Client, 보호 ctx 경로 누락, admission 거절의 IN_DOUBT 오분류, Close 뒤 시작.
- [ ] 4.2 gstack review + cso.
- [ ] 4.3 Manager 독립 검증.

## 5. 배포·닫기

- [ ] 5.1 **사람 gate H10**(배포 전, 브로커 호출 0): 실제 bind path 에서 컨테이너·호스트 두 프로세스가 같은 `instance`·같은 `dev/ino` 를 로그에 남기는지, flock 교차 배제, 다른 uid 열기 실패, fuseblk 거부.
- [ ] 5.2 **사람**: 이미지 재빌드 + 두 서비스 재시작(KST 05:00~09:00). **배포 체크**: 두 서비스 env `TOSSOS_ACCOUNT_BUDGET_DIR` 존재 + 엔진 기동 로그 `account_budget_open` 의 instance 가 marker 와 같음. 배포 뒤 `wire_429`·`admission_refused` 를 로컬 기록으로만 관찰.
- [ ] 5.3 (선택, **사람 승인 시에만**) 합산 한도 실측 — 이 change 는 요구하지 않는다.
- [ ] 5.4 완료 게이트(사용자 명시 요청 시) → archive → Story 경로·PM sync. a087 5.5·a169 재개에 착지 sha 와 T-INT·T-429-S·T-SYNC·T-MP 시험 이름 전달.
- [ ] 5.5 검증된 교훈만 Hindsight retain.
