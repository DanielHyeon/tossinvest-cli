> design 1판(2026-10-29). 착수 승인(사용자 2026-10-29). 아래는 design §5~§10 의 순서이며 아직 아무 구현도 시작하지 않았다.
> High-risk(retry matrix·rate limit) — FLM/BTM·Pre-Edit·변이 원장·다중 적대 리뷰 면제 없음. 실계좌 부하 시험 금지.

## 0. 착수 준비

- [ ] 0.1 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a166-account-shared-rate-limit` 로 base 고정(a165 착지 `5e2facaee` 이후).
- [ ] 0.2 2026-10-22 「엔진 + 수동 조회 동시 429」 관측의 로그 영수증 확보(로컬 기록만 — 계좌 API 재호출 금지). 없으면 「영수증 없음」으로 기록.
- [x] 0.3 전송 자리 전수 — design §2.1(S1~S5 + 대상 아님 4), 프로세스 지도 §2.2, 파일시스템 §2.3. CodeGraph 1.6.0 `callers`·`affected --filter '*_test.go'` 영수증 `analysis/code-context/*-1st.txt`.
- [ ] 0.4 search-first(외부 레지스트리·MCP) — 이 1판은 저장소만 봤다(선례 `strategyevidence.SharedRateBudget` 은 프로세스 안·생산 미배선, `enginelock` flock, `x/time/rate` 없음). 외부 채널은 **안 봤음**. 들이는 것은 사람 승인 + `dependency-auditor`.
- [ ] 0.5 FLM/BTM 번들(gate ⑤ 형식, `scaffold_analysis.py`) — design §10 목록. 1판 AST(`analysis/evidence/ast/`, HEAD `5e2facaee`)와 `source_sha256` 대조.
- [ ] 0.6 evidence reconciliation — `analysis/code-context/codegraph-baseline.md`·`codegraphcontext-context.md`·`evidence-reconciliation.md`(CodeGraph 가 `tokenManager.exchange` 호출자를 못 찾은 것 포함).

## 1. 계약 (design freeze)

- [x] 1.1 design.md 1판 — 선택지 비교·권장 ①, 키, 우선순위, 429 패널티, 이벤트, 닫기 경계, 실패 처분, 배포·rollback, 소비자 계약 C1~C8.
- [x] 1.2 spec delta — 「Retry Matrix 산출물」 MODIFIED + 「계좌 공유 요청 admission」 ADDED.
- [ ] 1.3 사람 결정 H1~H10(design §8) 회신 → design 2판.
- [ ] 1.4 proposal-freeze 리뷰(적대적 Eng 보이스 포함, 작성자와 다른 계열) → `review.md`.

## 2. RED

- [ ] 2.1 Pre-Edit 선언(`review.md`) — 대상 심볼 design §10.
- [ ] 2.2 RED `internal/accountbudget`: T-SPACE-1 · T-PRIO-1 · T-429-1 · T-CLK-1 · T-SYNC-1(프로세스 안).
- [ ] 2.3 RED `internal/official` 통합: T-INT-1~3(a087 이 인용할 이름 확정) · T-EVT-1 일부 · T-OFF-1 · T-STR-1.
- [ ] 2.4 RED 엔진·호출자: T-RET-1 · T-MUT-1 · T-STR-2(보호 GET 표지 전수) · T-STR-3(재생 결속) · T-WALK-1.
- [ ] 2.5 RED 다중 프로세스: T-MP-1 · T-SYNC-1 의 `StopOnShared429`(자식 프로세스) — `-race` 아래.

## 3. GREEN · REFACTOR · VERIFY

- [ ] 3.1 GREEN `internal/accountbudget`(새 패키지) + `official` 배선(`WithAccountBudget`·`doRequest`·`exchange`·`ErrAdmissionRefused`).
- [ ] 3.2 GREEN 엔진(`NewOrderPath` Budget·marker `account_budget_instance`·`ReasonAccountBudgetUnshared`·보호 GET 표지·`Retrier.Query` 429 즉시 반환) + 사람 결정 H9 이면 `RetryAfterTransport` 일가 삭제.
- [ ] 3.3 GREEN 그 밖의 Client 생성 자리(콘솔·httpapi·soak·verify·CLI·도구 S3·S4) + `compose.yaml` env. **정본 retry matrix §0.3·§1·§2·§3·§4 와 코드 기본값을 같은 커밋에서**.
- [ ] 3.4 VERIFY 영향 범위 시험: `internal/accountbudget`·`internal/official`·`internal/execgw`·`internal/app/engine`·`cmd/tossctl` 관련 + `-race`(`make test-race` 목록에 새 패키지 추가) — 선택 이유·미실행 범위 기록.
- [ ] 3.5 변이 원장(design §9 목록, 양성 대조군 포함) — 사본 하네스는 무변이 대조군 GREEN 확인 뒤.
- [ ] 3.6 FLM/BTM 최신화(편집 뒤 AST 재추출), `check_analysis.py --change a166-account-shared-rate-limit`.
- [ ] 3.7 **사람 승인**: 컨테이너↔호스트 flock 교차 배제 1회 확인(예산 파일만, 브로커 호출 0 — design §8 H10).

## 4. 리뷰

- [ ] 4.1 다중 적대 리뷰(적대적 Eng 포함) — 반례 목표: 손절·flatten 지연, 잠금 쥔 채 정지, 시계 역행, 독점, 우회 전송 자리 누락, 보호 GET 표지 누락, admission 거절의 IN_DOUBT 오분류.
- [ ] 4.2 gstack review + cso.
- [ ] 4.3 Manager 독립 검증.

## 5. 배포·닫기

- [ ] 5.1 **사람**: 이미지 재빌드 + 두 서비스 재시작(KST 05:00~09:00). 배포 뒤 `wire_429`·`admission_refused` 이벤트를 로컬 기록으로만 관찰(계좌 API 추가 호출 0).
- [ ] 5.2 (선택, **사람 승인 시에만**) 합산 한도 실측 — 이 change 는 요구하지 않는다.
- [ ] 5.3 완료 게이트(사용자 명시 요청 시) → archive → Story 경로·PM sync. a087 5.5·a169 재개에 착지 sha 와 T-INT·T-EVT 시험 이름 전달.
- [ ] 5.4 검증된 교훈만 Hindsight retain.
