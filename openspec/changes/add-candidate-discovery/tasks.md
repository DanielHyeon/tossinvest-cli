# Tasks: add-candidate-discovery

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 이 change는 읽기 전용이므로 실계좌
mutating 단계가 없다 — 검증은 fixture와 주입 clock으로 한다.

## 1. 기반 — 관측과 영속

- [ ] 1.1 [T] `Observation`·`Candidate` 타입 + 파생값 표시. 원천이 준 값과 계산값을 구분.
- [ ] 1.2 [T] 후보 저장소(SQLite, 원장과 **별개 파일·별개 락**). ext4 강제·fuseblk 거부는
  원장 규칙을 따른다. RED: 재시작 후 이전 프로세스의 관측을 읽지 못함 → GREEN.
- [ ] 1.3 [T] 저장소가 엔진 원장 락을 잡지 않음을 테스트로 고정 — 스캔이 도는 동안
  `tossctl engine run`의 flock 획득이 막히지 않는다.
- [ ] 1.4 [T] 후보 수명(활성/냉각/만료)과 **재진입 시 `first_seen_at` 유지**.
  RED: 냉각 중 재진입이 `first_seen_at`을 갱신 → GREEN(유지). 만료 후는 새 후보.

## 2. 원천

- [ ] 2.1 [T] 공식 Open API 어댑터 — `Rankings`(`MARKET_TRADING_AMOUNT`,
  `MARKET_TRADING_VOLUME`, `TOP_GAINERS`), `MarketInvestorTrading`, 지표 캔들.
  **공식 원천만으로 후보가 산출됨**을 테스트로 고정.
- [ ] 2.2 [T] WTS 어댑터 — 인기 순위, 투자자 순매수 순위, 테마, AI 시그널, 스크리너.
  가산 원천이며 없어도 산출이 멈추지 않는다.
- [ ] 2.3 [T] 원천 강등 — 일부 실패 시 남은 원천으로 계속하고 `sources`·`completeness`·
  `degraded`를 후보와 스캔 결과에 기록. RED: WTS 전부 실패 시 후보 0건 → GREEN.
- [ ] 2.4 [T] 원천 병합 — 같은 `(market, symbol)`을 올린 여러 원천이 후보 **하나**가 되고
  각 원천이 그 하나를 지지한 것으로 기록된다.
- [ ] 2.5 [T] hybrid 우회를 강등으로 기록(D14). RED: 공식이 429로 WTS 우회된 후보가 공식이
  응답한 후보와 구분되지 않는다 → GREEN(우회 사실이 `sources`에 남는다).
- [ ] 2.6 [T] rate 예산 잔량 헤더(`X-RateLimit-Limit`/`Remaining`/`Reset`) 기록(D13 결정 2).
  지금은 어디서도 읽지 않아 429가 나야 예산을 안다. RED: 200 응답에서 잔량이 유실 → GREEN.
- [ ] 2.7 [T] 원천별 간격 — 시장이 아니라 원천에 붙고, 느린 원천이 빠른 원천을 묶지 않는다.
  기본값 WTS 5초(하한 3초) / 공식 랭킹 15초(하한 5초).

## 3. 시간축 지표

- [ ] 3.1 [T] 주입 clock 배선. 테스트가 시간 경과를 만든다.
- [ ] 3.2 [T] 거래대금·거래량 변화율 — **경과 시간으로 정규화**. RED: 관측당 차분을 쓰면
  1분 간격과 10분 간격이 같은 값 → GREEN(1분 쪽이 크다).
- [ ] 3.3 [T] 가속도를 **구간 속도의 비**로 산출(D9). 두 구간의 실제 초를 함께 저장.
  RED: 합의 비를 쓰면 backoff로 3배 늘어난 창이 가속도 3배를 만든다 → GREEN(1배 부근).
- [ ] 3.4 [T] `WARMING_UP` — 직전 구간이 없으면 가속도를 산출하지 않는다. RED: 첫 관측이
  무한/과대 가속도로 임계를 통과 → GREEN(미산출, 임계 통과 아님).
- [ ] 3.5 [T] shadow 임계 5종(1.3/1.5/1.8/2.0/2.5) 통과 여부를 전부 기록. 판정은 하지 않는다.
- [ ] 3.6 [T] 최초 발견가 대비 확장률.
- [ ] 3.7 [T] 일중 range position(고점 근접도) — 캔들 기반, 캔들 부재 시 **미측정**으로 표시.
- [ ] 3.8 [T] 순위 백분위 변화 + `newly_listed`를 **별개 사실**로 기록(D8). RED: 신규 진입이
  직전 최하위로 채워져 순위 속도를 자동 통과 → GREEN(합성하지 않는다).

## 4. 추격 위험 veto

- [ ] 4.1 [T] `seen_late` — 최초 관측 시점에 이미 임계를 크게 초과.
- [ ] 4.2 [T] `extended` — 최초 발견가 대비 확장률 상한 초과.
- [ ] 4.3 [T] `near_high` — 일중 고점 근접도 상한 초과. 경계 고정: 1.99% true / 2.00% false.
- [ ] 4.4 [T] **3-상태 veto**(D10) — `true`/`false`/`unmeasured`. RED: 캔들을 못 받은 후보의
  `near_high`가 `false`로 저장되어 "veto 통과"로 집계 → GREEN(`unmeasured`는 통과가 아니다).
  이 change에서 **가장 중요한 테스트**다 — 상세조회는 초당 5회·종목당 1회이므로 미측정이
  예외가 아니라 평시다.
- [ ] 4.5 [T] veto가 점수와 **분리**됨을 고정. RED: 가속도 최상위가 `near_high`를 상쇄해
  통과 → GREEN(거부권 유지).
- [ ] 4.6 [T] veto 후보도 저장·보고된다. 늦게 본 비율을 셀 수 있어야 한다.

## 5. 사람이 보는 표면

- [ ] 5.1 [T] `tossctl candidate scan` — 1회 스캔. `mutating: false`. 출력에 임계별 통과 수,
  강등·백오프 표시.
- [ ] 5.2 [T] `tossctl candidate watch` — 반복 스캔. 간격 하한 강제.
- [ ] 5.3 [T] 실계좌 검증 runlock이 살아 있으면 `watch` 시작 거부.
- [ ] 5.4 [T] 429 백오프 + 백오프 사실을 스캔 결과에 기록.
- [ ] 5.5 [T] 콘솔 `/candidates` — 후보 목록, 최초 발견 시각, 지표, veto 사유, 원천·완전성.
  **읽기 전용 화면**. 확인 문자열 타이핑 같은 마찰을 넣지 않는다.

## 6. 격리와 게이트

- [ ] 6.1 [T] 의존성 격리 테스트 — `internal/candidate`의 import 그래프에 주문 실행 경로가
  없음을 고정. 후보에서 주문으로 가는 코드가 컴파일되지 않아야 한다.
- [ ] 6.2 Function Logic Map + `check_analysis.py`
- [ ] 6.3 PM registry allowlist + fixture 등록
- [ ] 6.4 `docs/ROADMAP.md` Phase 3 갱신 — T3.1을 이 change로 분리하고 착수 조건을 고친다
- [ ] 6.5 `make sdd-sync && make sdd-check && make gate CHANGE=add-candidate-discovery`
- [ ] 6.6 독립 리뷰(gstack) — 특히 격리 테스트가 실제로 컴파일 차단을 하는지

## 착수 조건 — 충족 (2026-07-28)

세 가지 모두 확정됐다. 값은 design.md "결정된 계약값", 근거는 D8~D14.

1. **임계 기본값** — 공통 순위 하드 컷 없음(D8), 가속도는 구간 속도의 비(D9),
   `near_high` 2.0% 3-상태(D10). 레인별 임계는 기록만 하고 판정은 T3.2가 한다.
2. **스캔 간격** — 시장이 아니라 원천별(D13). WTS 5초, 공식 랭킹 15초 + 실측 기록.
3. **KR·US** — 처음부터 동시 ON, `policy_version`을 시장별로 분리.
