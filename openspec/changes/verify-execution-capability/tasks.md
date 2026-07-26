# Tasks: verify-execution-capability

> 선행: harden-execution-base 완료(됨). 사용자 협조 필수. 이 change의 측정 결과가 2c `add-protection-orders` 스펙 작성의 입력이다 — **2c는 2.5~2.9 완료 전에 작성하지 않는다.**

## 1. Soak·실측 [T]

- [x] 1.1 조회 전용 soak 도구 작성 (mutation transport 컴파일 제외, 실 hostname POST 불가 가드 상속)
- [ ] 1.2 자격증명 무인 갱신 soak 연속 3일+ 실행·기록 (사용자 환경)
- [ ] 1.3 rate limit 실측 → retry matrix·폴링 SLO 수치 확정 반영
- [ ] 1.4 attestation 파일 생성 (만료·계좌 식별·성공 endpoint 집합·속성 결과) + 엔진 기동 인터록 연동 확인 — endpoint 집합은 `engine.RequiredEndpoints()`와 drift 가드 테스트로 동기화한다(2a가 목록을 확장하면 자동 강제; 조건주문 endpoint는 2c가 추가)

- [x] 1.5 [T] **안내형 실계좌 검증 도구** `tossctl verify run`: 2.1·2.2·2.5·2.7·2.8의 측정을 단계 목록으로 구동 — 승인은 **run당 1회 배치 typed-confirmation**(시작 시 계획된 전 mutation의 전체 요약 목록 + 만료 nonce — 사용자 결정 2026-07-26; 1회성 측정이므로 §0.1 취지 내 완화)이 기본이고 `--confirm-each`로 단계별 확인 옵트인, 자동화 플래그 금지·TTY 전용은 유지, 최소 수량·즉시 취소, 단계별 증거 JSONL 기록(→ attestation 속성 입력), 중단·재개 가능(조건주문 존속 확인은 재실행 2회 구조), 비용은 검증 주문 자체의 execution.commission에서 수집. 자동 테스트가 아니라 운영자 도구다(테스트는 httptest 전용, testenv 가드 상속). 유효 창 경계(2.7)는 의도적 이중 주문 절차임을 단계 안내문에 명시하고 기본 생략(--include-ttl-edge 옵트인)

- [x] 1.6 [T] **로컬 운영 콘솔** `tossctl console` (사용자 결정 2026-07-26 — 웹 화면으로 검증 수행): 127.0.0.1 전용 바인딩(비루프백 거부), 기동 시 1회성 세션 토큰 URL 출력. 화면 — 대시보드(soak 진행·attestation 상태·verify 진행·evidence 요약, read-only), verify 실행(단계 목록 → 배치 요약 표시 → **nonce 타이핑 폼**으로 승인 → 진행 로그 → resume 안내), report 뷰. 승인 등가성: 세션 토큰+CSRF+화면 표시 nonce 타이핑 3중 — TTY 타이핑과 동일한 "사람의 의도적 승인"이며 runner의 모든 레일(계획 인가·상한·취소·ErrOutsidePlan)은 무변경. CLI 단독으로 비대화 승인이 가능해지는 플래그·경로를 만들지 않는다(콘솔 내부 배선만). 조건주문 존속 측정은 콘솔 재시작 안내로 프로세스 경계 유지. 게이트 ON은 이 콘솔의 범위가 아니다(2c 후). 테스트: httptest·testenv 가드·루프백 전용·오승인(틀린/만료 nonce·토큰 부재) 전수

- [x] 1.7 [T] **재측정 경로·실행 강건성** (1차 실행 2026-07-26의 실측 갭 — measurements.md): ① 콘솔에 **재측정** 시작 모드 — 마지막 verdict가 `fail`·`skipped`인 단계를 `Runner.Redo`로 재실행(`deferred`·`pass` 제외, 대상 단계 수를 시작 화면에 표기), 반드시 새 배치 승인(신규 nonce) 경유 — 비대화 승인 경로·게이트 접근 신설 금지, ttl-edge 등 설계상 skip은 preflight가 다시 걸러 무해. ② 시작 화면에 **장시간 advisory**: KST 평일 09:00–15:30 밖이면 "mutation 단계는 order-hours-closed(422 실측)로 실패할 수 있음" 경고 — advisory만, 하드 차단 금지(주문 접수 창은 [미측정]). ③ **429 강건성**: account-seq 해석을 run당 1회로 캐시, 읽기 전용 단계는 ErrRateLimited에 한해 한도 내 백오프 재시도(mutation 자동 재시도 금지), verify 실행 중 soak cycle 일시정지(XDG lockfile advisory — stale lock은 mtime으로 무시). 테스트: httptest·testenv 가드, 재측정이 pass 단계를 건드리지 않음·승인 없이 redo 불가 전수

- [x] 1.8 [T] **콘솔 운영 자동화·전면 한글화** (사용자 결정 2026-07-26 — 터미널 수동 조작 제거): ① **웹 콘솔 재시작** — [콘솔 재시작] 버튼(세션+CSRF POST): 동일 바이너리 경로 re-exec로 새 프로세스 인스턴스 시작(one-run-per-process 상한 초기화·조건주문 존속 경계 유지 — 존속 판정이 record의 `process.instance_id` 기준인지 확인하고 테스트로 고정; pid 기준이면 instance_id로 교정), 브라우저 연속성은 **1회성 핸드오프 토큰**(0600 파일, 즉시 소모, 재사용 거부)으로만 — 이미 인증된 세션에서 시작된 재시작에 한함, 새 세션 URL은 종전대로 터미널에도 출력. ② **soak 무인화** — 대시보드 [soak 재시작] 버튼(SIGINT 후 autostart 방식 재기동, 자격증명 무접촉), soak 자신은 cycle 경계에서 설치 바이너리 변경 감지 시 self re-exec(record·streak 무영향), 대시보드에 실행 중/설치 바이너리 불일치 경고(콘솔·soak 각각). ③ **전면 한글화** — 화면·verify runner 진행 출력을 한글로(step ID·HTTP/에러 코드·endpoint 경로·JSON 다운로드 내용은 원문 유지), evidence record의 `title`은 영어 유지(기존 레코드 비교성)하고 렌더 계층에서 StepID→한글 라벨 매핑(전 단계 커버 drift 테스트). 레일 불변: 게이트 라우트 부재, 승인 등가성(새 nonce 타이핑) 무변경, 재시작은 주문 능력과 무관한 read-only 도구 재기동(§0.7 운영 토글 비해당 — issues.md 판정), 비대화 승인 경로 신설 금지. 테스트: httptest·testenv 가드, 핸드오프 재사용·CSRF 부재·만료 거부 전수, re-exec·프로세스 기동은 seam으로 fake

## 2. 실계좌 검증 [M+사용자]

- [ ] 2.1 주문 status enum 실측 fixture 수집 → 상태 파생 표 보강. **CANCEL_REJECTED/REPLACE_REJECTED "별도 주문 레코드"의 실제 형태**(목록 조회 노출 여부·원주문 링크 유무) 관측 포함 — 2a 브로커 상태 파생과 2c 귀속 규칙의 입력
- [ ] 2.2 실계좌 주문 경로 1회성 검증 절차 실행(사용자): 최소 수량·limit-only·즉시 취소, 매도 경계(부분/전량/보유초과), KR cancel/amend
- [ ] 2.3 flatten-all `--dry-run` 리허설 1회 실행·기록
- [ ] 2.4 약관·자동화 허용 범위 검토 기록 + 계정 정지 시 포지션 방침 문서화
- [ ] 2.5 **조건주문 능력 검증**(사용자): 시장·유형별 등록·조회·취소·정정(신규 ID 발급·기존 ID 무효화 — openapi 문서 확인), 프로세스 종료 후 존속, 발동 관측과 `triggeredOrderId` 노출 지연, 정규장 밖 동작, 만료, OCO sibling 취소 시점, 부분체결 잔량, **SINGLE+MARKET 손절의 실동작**(OCO/OTO는 LIMIT 전용 — openapi), 조건주문 예약이 매도가능수량에 반영되는지, 조건주문과 일반 매도 동시 제출의 거부 의미 — 최소 수량으로 수행, 결과를 ProtectiveCapability 속성(시장·수량 종류·조건 유형·발동 주문 유형·세션·modify 의미·triggeredOrderId 노출·검증 시각·증거 digest)으로 기록
- [ ] 2.6 2.5 결과로 보호 유형 확정(기본 가설: SINGLE+MARKET 손절 단독, 익절은 로컬 청산 — "한 심볼에 브로커측 매도 청구권 1개" 불변식) + 미검증 시장·유형의 자동 진입 금지 목록 산출
- [ ] 2.7 **멱등키 실동작 검증**(사용자): 일반 주문·조건주문 각각 동일 `clientOrderId`·동일 본문 재요청이 이전 결과를 재반환하는지, 유효 창(문서상 10분)과 **안전 마진 산정을 위한 왕복 지연 관측**, 키의 계좌 스코프, 본문 상이 시 `idempotency-key-conflict` — 최소 수량·즉시 취소. **주의**: 유효 창 경계 확인은 의도적으로 두 번째 라이브 주문을 만드는 절차다 — 양쪽 즉시 취소 범위로 한정하고, 사용자가 거부하면 창 경계는 미검증으로 남기고 재생은 보수 마진(TTL/2)으로만 허용하거나 비활성 유지
- [ ] 2.8 매도가능수량 의미 실측: 담보·미결제·미체결 매도·조건주문 예약이 `sellableQuantity`에 반영되는 방식 — 2c 청산 예약 공식의 입력
- [ ] 2.9 실측 비용표(수수료·거래세 bps) 수집 → 2d 비용 모델의 Toss 검증값 입력

## 3. 완료 게이트 [M]

- [ ] 3.1 attestation 유효 확인 + 검증 기록 docs/ 반영, 미검증 속성의 명시적 목록(= 자동 경로 금지 목록) 산출
- [ ] 3.2 `make gate CHANGE=verify-execution-capability`
