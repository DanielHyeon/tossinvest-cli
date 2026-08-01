## Context

engine runtime에는 entry loop가 없고 Tracer는 production에서 호출되지 않는다. a046 verdict를 a045 보호 전제 아래 공식 주문 경로에 연결하는 최소 lane architecture가 필요하다.

## Goals / Non-Goals

**Goals:** 독립 lane, durable decision provenance, Guardian-only LIVE 진입을 구현한다.

**Non-Goals:** 다수 전략 동시 이식, scheduler, 자동 최적화, paper/shadow/canary 경로.

## Decisions

1. lane는 순수 `Evaluate(ApprovedCandidate) EntryDecision`을 반환하고 broker/journal을 직접 받지 않는다.
2. orchestrator만 RiskIntent→Guardian→durable attempt→official gateway를 수행한다.
3. 첫 lane는 StockOS `parker_vwap_trend_v1`의 KRX conservative profile을 `krx_parker_vwap_conservative_v1`로 순수 이식한다. source는 StockOS commit `d75113d3c338148606d86c8aedbbeb7ed446c0b8`, 관련 source-set SHA-256은 `09260ac29e50ed4d2a43d0e274f9a17465e00ee36fb61d759127f158985c23bd`다. StockOS가 KOSPI/KOSDAQ tick rule에 calibration됐다는 source 주석에 따라 KRX regular-session, closed 5-minute bars만 허용하고 US/장전·장후는 지원하지 않는다.
4. lane desired state와 effective state를 분리한다. OFF는 신규 entry만 차단하고 exit loop에 전달되지 않는다.
5. 자동 재시작은 운영자가 승인한 desired state와 유효 gate/protection을 모두 재확인한다.
6. UI 소유 카테고리는 `strategy-runtime`이다. `전략 파라미터`, `lane 상태`, `자동 기동`, `LIVE 주문 승인`을 별도 section과 별도 save/confirm action으로 분리하며 `한 번에 모두 활성화` control을 금지한다.
7. lane desired 기본값과 auto-start 기본값은 OFF다. 첫 lane source/market/constants가 동결되기 전 descriptor는 `not_configured`, read-only이며 숫자 0이나 임의 StockOS 값을 UI default로 넣지 않는다.
8. 모든 설정 descriptor는 label/help/default/desired/effective/unit/range/provenance/apply timing과 effective refusal reason을 포함한다.
9. 신규 진입 권한은 immutable activation manifest 하나로만 표현한다. manifest는 account/profile,
   build/commit digest, lane ID/version/source/market/constants digest, threshold/evidence digest,
   settings digest, capability-attestation digest/expiry, Guardian version/limits, reconciliation watermark,
   protection profile, operating policy, scheduler/calendar scope, lane/scheduler/autostart/gate/LIVE의 개별
   승인, actor/issued-at/expires-at/generation/audit ID를 묶는다.
10. orchestrator는 durable dispatch 직전에 manifest 전 필드를 다시 검증하고 decision/attempt에 digest를
    기록한다. mismatch, expiry, kill switch, reconcile degradation 또는 high-risk config 변경은 entry만
    OFF로 만들며 reduce-only exit/reconcile에는 전달하지 않는다.
11. 첫 lane의 immutable constants는 `min_vwap_slope_pct=0.08`, `ema_touch_tolerance_pct=0.25`,
    `min_forward_space_pct=1.2`, `min_expected_rr=1.5`, `tangled_band_pct=0.35`,
    `max_band_expansion_rate=1.8`, `hard_stop_pct=0.7`, `partial_take_profit_at_r=3.0`,
    `skip_open_minutes=10`, `no_entry_after_buffer_minutes=45`, `max_signal_age_seconds=15`, `max_entry_price_drift_pct=0.20`,
    `symbol_state_stale_seconds=30`이다. KRX open은 source와 동일하게 `09:00 KST`로 봉인하고 임의 이동을 거부한다.
    session 상수는 source와 동일한 시가 동시호가 `08:30..09:00`(`open-30m..open`),
    종가 동시호가 `close-10m..close`, 시간외 `15:40 KST`(`open+400m`)이며, session refusal 순서는
    non-trading → open auction → close auction → after-hours → opening skip → cutoff다. gate 순서는 profile/state → session → closed-bar integrity →
    symbol state → no existing position → nonzero volume → indicator completeness → VWAP above/slope →
    EMA9 bullish pullback → LVN forward space → untangled/band expansion → RR → HVN ceiling → age/drift다.
12. source parity는 StockOS fixture를 Go fixture로 번역해 regular/early-close와 auction/after-hours refusal reason,
    accept, entry, 0.7% stop,
    3R target과 expected RR이 일치하도록 검증한다. source digest나 constants 변경은 새 lane version이며
    기존 manifest를 무효화한다. UI는 이 값을 직접 입력받지 않고 fixed server preset/provenance만 보여준다.
13. official candle endpoint의 `1m` decimal strings를 exact decimal로 보존한 뒤 KST 정규장 경계에서
    연속된 다섯 개의 closed minute만 하나의 5m bar로 집계한다. 미완성 bucket, 누락 minute, 장외 minute,
    timezone-naive timestamp와 호출 시점에 아직 닫히지 않은 bar는 typed refusal이며 float64를 lane 입력
    정본으로 사용하지 않는다.
14. HALT/LIMIT/MANAGED와 freshness를 판정하는 authoritative symbol-state source는 별도 typed interface다.
    quote 존재나 일일 price-limit 값으로 상태를 추측하지 않는다. authority가 없거나 30초보다 stale이면
    lane은 `not_configured`/effective OFF다.
15. a045 ProtectionReady가 WIRED가 아니거나 a046 candidate provenance, a048 scheduler/calendar claim,
    reproducible StockOS source manifest 중 하나라도 없으면 구현은 dormant evaluation/descriptor까지만
    제공한다. runtime loop, auto-start와 LIVE activation은 생성하지 않으며 exit/reconcile은 계속된다.
16. a047의 console 표면은 fixed lane·constants·provenance와 각 blocker를 보여주는 read-only 상태 card다.
    실제 lane/autostart/gate/LIVE 제어는 a050 canonical control plane에서 server preset과 별도 action으로
    조립한다. text/number/textarea/contenteditable, 임의 symbol/reason, typed confirmation과 `enable all`은 없다.
17. frozen source set은 sorted relative path, blob SHA-256과 canonical source-set digest를 담은 manifest로
    재현되어야 한다. 고정된 `09260ac…`를 manifest로 재계산하지 못하면 provenance는 미검증이고 effective
    state는 OFF이며 새 digest를 조용히 대입하지 않는다.

## Risks / Trade-offs

- [전략이 gateway를 우회] → package dependency/static test로 lane의 broker import를 금지한다.
- [재시작 중 중복 진입] → deterministic decision/client order identity와 duplicate Guardian check를 사용한다.
- [직접 LIVE 위험] → 기본 OFF, 전체 gate, 단일 운영 승인과 kill switch를 유지한다.
- [서로 다른 시점의 승인이 조합됨] → 원자적 activation manifest와 submit-time TOCTOU 재검증을 사용한다.
- [전략값 저장과 엔진/LIVE 활성화 혼동] → 독립 section·버전·확인 흐름을 사용하고 설정 저장이 운영 토글을 바꾸지 않음을 preview에 고정한다.

## Migration Plan

lane interface와 dormant evaluation/runtime seam을 배포한다. a045/a046/a048와 source manifest가 모두
검증된 별도 activation change 전에는 LIVE manifest를 발급하지 않는다. canary 경로는 없다. rollback은
entry desired state OFF 후 exit/reconcile을 계속 실행한다.

## Open Questions

고정 source-set digest를 재현하는 sorted file/blob manifest가 구현 gate에서 확정되어야 한다. 재현 실패는
새 값을 선택하는 근거가 아니라 dormant `not_configured` 상태를 유지하는 근거다. 첫 lane의 수익성 승인을
의미하지 않는다.
