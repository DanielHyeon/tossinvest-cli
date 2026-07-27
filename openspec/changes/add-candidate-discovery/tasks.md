# Tasks: add-candidate-discovery

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 이 change는 읽기 전용이므로 실계좌
mutating 단계가 없다 — 검증은 fixture와 주입 clock으로 한다.

## 1. 기반 — 관측과 영속

- [x] 1.1 [T] `Observation`·`Candidate` 타입 + 파생값 표시. 원천이 준 값과 계산값을 구분.
  → `candidate.go`. `Reported`가 원천의 말, 파생은 별도. 모든 decimal은 문자열이고 빈 값은
  0이 아니라 미측정 — `DayHigh`가 0으로 저장되면 `near_high`가 전부 통과한다.
- [x] 1.2 [T] 후보 저장소(SQLite, 원장과 **별개 파일·별개 락**). ext4 강제·fuseblk 거부는
  원장 규칙을 따른다. RED: 재시작 후 이전 프로세스의 관측을 읽지 못함 → GREEN.
  → `store.go`(`candidates.db`), `fsguard.go`. 원장 규칙은 import가 아니라 **복사 + 표류
  테스트**다(`fsguard_drift_test.go`가 `internal/journal/fsguard.go`를 소스로 읽어 대조).
  import하면 6.1 격리가 깨진다.
- [x] 1.3 [T] 저장소가 엔진 원장 락을 잡지 않음을 테스트로 고정 — 스캔이 도는 동안
  `tossctl engine run`의 flock 획득이 막히지 않는다.
  → `TestAScanningStoreDoesNotBlockTheEngineFromStarting`.
- [x] 1.4 [T] 후보 수명(활성/냉각/만료)과 **재진입 시 `first_seen_at` 유지**.
  RED: 냉각 중 재진입이 `first_seen_at`을 갱신 → GREEN(유지). 만료 후는 새 후보.
  → `Promote`/`Cool`. 만료는 sweeper가 아니라 **읽는 시점의 시각**으로 판정한다 — sweeper가
  돌아야 진행하는 상태는 방금 올라온 프로세스에서 거짓말을 하고, 그때가 바로 묻는 때다.
- [x] 1.5 [T] 리뷰(2026-07-28)가 찾은 §1 결함 수정. 전부 "아무것도 실패하지 않는" 종류였다.
  - **저장 시각이 순서 키가 아니었다(P0)**. `RFC3339Nano`는 소수부 뒤 0을 떼므로 TEXT 비교에서
    `…00Z` > `…00.5Z`(`Z`=0x5A > `.`=0x2E)다. `PruneObservations`가 **컷오프보다 새 행을
    지웠고**, `Observations(since)`가 요청한 창 안의 행을 빠뜨렸으며, "오래된 것 먼저"가 아니었다.
    §3·§4가 바로 이 두 질의 위에서 가속도와 veto를 계산한다. → 고정폭 stamp +
    `TestTheStoredInstantOrdersLikeTime`(기존 테스트는 전부 정각이라 이 분기에 닿지 못했다).
  - **아무도 냉각시키지 않은 후보가 영원히 활성**이었다. 냉각은 기록하는 쪽이 살아 있어야 하고
    스캔은 죽는다. → `stateAt`가 `last_seen_at` 기준 암묵 냉각(`DefaultStalenessTTL` 10분,
    최장 백오프 300초의 2배).
  - **과거 시각으로 살아 있는 후보를 만료시킬 수 있었다** — D1 우회의 뒷문. → `Promote`·`Cool`이
    마지막 관측보다 이른 시각을 거부.
  - **만료 후 새 후보가 죽은 후보의 `sources`·완전성·`degraded`를 상속**했다. → 만료 시 초기화.
  - **표류 테스트가 상수만 대조**하고 허용 **집합**은 대조하지 않았다 — 원장이 파일시스템을
    거부 목록으로 옮겨도 상수는 그대로다(테스트 헤더가 막겠다고 쓴 바로 그 방향).
    → 원장 소스에서 `allowedMagics`/`allowedNames`/`AllowedFilesystems`를 파싱해 양방향 대조.
    변이 테스트로 확인: `MagicExt` 제거 시 FAIL.
  - `_txlock=immediate` 누락(두 프로세스에서 `SQLITE_BUSY_SNAPSHOT`), DSN 경로 미이스케이프
    (`#` 포함 디렉터리가 다른 파일을 연다), `nearestExistingDir("")`가 **CWD를 승인**,
    `NoteSources`가 대상 없이 조용히 성공(D4가 감추지 않겠다는 바로 그 사실을 감춘다) +
    원천을 병합하지 않고 덮어씀, `rank>0`인데 `rank_total==0`(D8 백분위가 `+Inf`),
    `nullable`이 공백을 다듬지 않음.
  - `internal/testenv`의 `FixedFSProber` 허용을 **파일 단위 → 선언 단위**로 좁힘 — 정의 파일
    안에서의 호출도 잡는다.

## 2. 원천

- [x] 2.1 [T] 공식 Open API 어댑터 — `Rankings`(`MARKET_TRADING_AMOUNT`,
  `MARKET_TRADING_VOLUME`, `TOP_GAINERS`), `MarketInvestorTrading`, 지표 캔들.
  **공식 원천만으로 후보가 산출됨**을 테스트로 고정.
- [x] 2.2 [T] WTS 어댑터 — 인기 순위, 투자자 순매수 순위, 테마, AI 시그널, 스크리너.
  가산 원천이며 없어도 산출이 멈추지 않는다.
- [x] 2.3 [T] 원천 강등 — 일부 실패 시 남은 원천으로 계속하고 `sources`·`completeness`·
  `degraded`를 후보와 스캔 결과에 기록. RED: WTS 전부 실패 시 후보 0건 → GREEN.
- [x] 2.4 [T] 원천 병합 — 같은 `(market, symbol)`을 올린 여러 원천이 후보 **하나**가 되고
  각 원천이 그 하나를 지지한 것으로 기록된다.
- [x] 2.5 [T] hybrid 우회를 강등으로 기록(D14) — **적용 범위는 `official_prices`뿐**.
  랭킹·캔들·투자자는 hybrid에서 **official-only이고 우회가 없다**([client.go:185-220]) —
  최초 D14는 이것을 확인하지 않고 썼다. RED: 우회로 채운 `prices` 후보가 공식이 응답한 후보와
  구분되지 않는다 → GREEN. 랭킹 관측에 `ViaFallback`이 찍히면 그것은 우회가 아니라 버그다.
- [x] 2.5b [T] 랭킹 429는 **강등이 아니라 결손**으로 기록(D14 결정 2). 우회라는 완충재가 없으므로
  429 한 번에 원천이 통째로 사라진다. RED: 랭킹 실패가 `completeness`에 남지 않는다 → GREEN.
- [x] 2.6 [T] rate 예산 잔량 헤더(`X-RateLimit-Limit`/`Remaining`/`Reset`) 기록(D13 결정 2).
  지금은 어디서도 읽지 않아 429가 나야 예산을 안다. RED: 200 응답에서 잔량이 유실 → GREEN.
  **D15 결정**: `internal/official`에 읽기 전용 헤더 노출을 **가산**한다. proposal Impact의
  "수정 없음"을 그에 맞게 고친다 — 헤더는 429에서 sentinel로 바뀌며 버려지고
  ([execgw/retry.go:183]), 유일한 대안인 RoundTripper는 `internal/execgw`라 §6.1 격리가 막는다.
- [x] 2.7 [T] 원천별 간격 — 시장이 아니라 원천에 붙고, 느린 원천이 빠른 원천을 묶지 않는다.
  기본값 WTS 5초(하한 3초) / 공식 랭킹 15초(하한 5초). → `Schedule`(정책과 due 판정만;
  루프는 §5가 갖는다). `YieldToEngine`이 spec R7의 최상위 우선순위에 처음으로 동사를 준다.
- [x] 2.8 [T] **냉각은 침묵이 아니라 증거로 한다**(§2 구현 중 발견). 후보를 냉각하려면
  **그 후보를 올렸던 원천이 전부 이번 스캔에 응답**해야 한다. 느슨하게 하면 랭킹 429 한 번이
  전 watchlist를 냉각시키고 냉각 시계가 그것을 만료시켜, **일시적 rate limit이 우리가 일찍
  발견했다는 기록 전체를 지운다.** 아무것도 실패하지 않으면서.
  → `coolAbsent` + `TestAScanDoesNotCoolASymbolItDidNotLookFor`,
  `TestOneSurvivingSupporterIsNotEnoughToCool`.
- [x] 2.9 [T] 시장별 panel 구성 — 그 시장을 볼 수 없는 원천은 **panel에서 빠진다**.
  "응답했는데 목록에 없다"가 냉각의 근거이므로, 구조적으로 못 보는 원천이 panel에 있으면
  매 스캔 그 시장 후보를 전부 냉각시킨다. → `candidatesrc.Panel`.

## 3. 시간축 지표

- [ ] 3.1 [T] 주입 clock 배선. 테스트가 시간 경과를 만든다.
- [ ] 3.2 [T] 거래대금·거래량 변화율 — **경과 시간으로 정규화**. RED: 관측당 차분을 쓰면
  1분 간격과 10분 간격이 같은 값 → GREEN(1분 쪽이 크다).
- [ ] 3.3 [T] 가속도를 **구간 속도의 비**로 산출(D9). 두 구간의 실제 초를 함께 저장.
  RED: 합의 비를 쓰면 backoff로 3배 늘어난 창이 가속도 3배를 만든다 → GREEN(1배 부근).
- [ ] 3.4 [T] `WARMING_UP` — 직전 구간이 없으면 가속도를 산출하지 않는다. RED: 첫 관측이
  무한/과대 가속도로 임계를 통과 → GREEN(미산출, 임계 통과 아님).
- [ ] 3.4b [T] **못 쓰는 분모도 미산출**(D9 정정). `prior_rate == 0`(거래 정지·저유동성),
  직전 구간 실제 초 == 0(한 스캔에서 여러 원천이 같은 순간 보고), 누적값 델타가 음수(세션
  경계·원천 재기동). RED: 셋 다 `+Inf`/부호 반전으로 임계를 통과 → GREEN(미산출).
  미산출 사유를 `WARMING_UP`과 **구분해 기록** — 전자가 많으면 원천을, 후자면 아무것도 안 고친다.
- [ ] 3.4c [T] rate·가속도 계열은 `(market, symbol, **source**)`별(D9 정정). RED: 원천 섞어
  차분하면 컴파일도 되고 그럴듯한 숫자도 나온다 → GREEN(원천 필터 읽기 + 원천 간 차분 금지).
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
- [ ] 4.7 [T] **입력 나이 상한** — 오래된 캔들로 만든 `near_high=false`는 `unmeasured`다.
  캔들이 초당 5회라 한 번 얻은 고가는 재사용된다. 3분 된 캔들로 만든 "위험하지 않음"은 측정이
  아니라 기억이고, 그 구분이 사라지면 4.4가 막은 실패가 한 단계 아래에서 재현된다.

## 5. 사람이 보는 표면

- [ ] 5.1 [T] `tossctl candidate scan` — 1회 스캔. `mutating: false`. 출력에 **shadow 임계별
  교차 수**(veto 통과가 아니다 — 말을 갈라 쓴다), **미측정 veto를 가진 후보 수**, 강등·백오프.
- [ ] 5.2 [T] `tossctl candidate watch` — 반복 스캔. 간격 하한 강제. 보존 정리와 WAL 회수를
  스캔 주기 안에서 집행(D16) — 정리는 옵션이 아니다.
- [ ] 5.3 [T] 실계좌 검증 runlock이 살아 있으면 `watch` 시작 거부.
- [ ] 5.3b [T] **엔진이 돌면 간격을 늘린다**(spec R7 정정). 우선순위 표에서 최상위인 엔진에만
  지키는 동사가 없었다. RED: 엔진 실행 중에도 간격이 그대로 → GREEN(물러나고 그 사실을 기록).
- [ ] 5.3c [T] **여유 공간 하한**(D16). 아래로 내려가면 발굴이 먼저 쓰기를 멈추고 말한다.
  RED: 공간이 차면 원장 쓰기가 ENOSPC → GREEN(발굴이 먼저 멈춘다). 원장과 같은 파일시스템이다.
- [ ] 5.4 [T] 429 백오프 + 백오프 사실을 스캔 결과에 기록.
- [ ] 5.5 [T] 콘솔 **`/signals`** — 후보 목록, 최초 발견 시각, 지표, veto 사유, 원천·완전성.
  **읽기 전용 화면**. 확인 문자열 타이핑 같은 마찰을 넣지 않는다.
  경로 이름은 StockOS의 `/signals`에 맞춘다(사용자 지시 2026-07-28).
- [ ] 5.6 [T] **미측정이 통과처럼 보이지 않는다**(새 spec Requirement). veto 사유가 하나도
  없는 행은 "모두 측정했고 안전"과 "한 번도 확인 안 함" 둘 다일 수 있고, 후자가 평시다.
  RED: 세 veto가 모두 `unmeasured`인 후보가 통과 후보와 같게 보인다 → GREEN.
- [ ] 5.7 [T] 강등을 **빠진 원천 이름**으로 말한다. 사유 없는 불리언은 대응할 수 없는 표시다.

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
