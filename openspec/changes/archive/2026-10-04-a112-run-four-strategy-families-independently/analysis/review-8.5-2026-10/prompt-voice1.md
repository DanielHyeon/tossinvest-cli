# 보이스 ① — 적대 fail-open / 노출 (a112 8.5)

공통 브리프: 같은 디렉터리의 `brief.md` 를 **먼저 전부** 읽고 그 안전 규칙을 지켜라. 네 배역은 아래 하나다 — 다른 축은 다른 보이스가 본다.

**배역: 노출이 새는 길을 찾는다.**
1. B1 — `ErrProductionFamilyActivationUndeclared` 변장: 미선언이 **아닌** 실패(파일 없음 · 핀 불일치 · 결속 불일치 · 만료 · 폐기 · 해석 실패 · ctx)가 `errors.Is(err, Undeclared)` 를
   참으로 만드는 경로가 있는가 — 이중 `%w`(`fmt.Errorf("%w: manifest file %s: %w", ErrUnavailable, name, err)`)의 `err` 가 Undeclared 를 감쌀 수 있는지, decode/validate 가 돌려주는
   오류 사슬, 엔진 `familyGateFor`(internal/app/engine/strategy_family_activation.go:132 부근)의 판별까지 따라가라. 참이면 「선언됐는데 쓸 수 없는」 시장이 기존 경로(관문 없음)를
   타는 노출 결함이다(P0 후보). 반대로 미선언이 다른 sentinel 로 바뀌는 길(→ 기존 경로 닫힘, 안전 방향)도 적어라.
2. B2 — 닫힌 시장이 실어 보내는 **검증된 활성화**가 그 주기에 닿는 곳을 전수로 따라가라: 레인 관측 승격(`strategyLaneRuntime.evaluate` 의 promotion) · lease 상한
   (`FamilyActivation.LeaseCeiling`) · dispatch 의 ProtectionReady 하한 · 투영 · 승격/잠금 원장 기록. 닫힌 시장(항목 0)에서 주문 · lease · 승격 기록 · 잠금 복구 세대가 생기는
   경로가 있는가? 앞 판(0b441270^)에서는 영값이던 자리들이다.
3. A — 2.3 B7 이 결정 · 봉인을 바꾸는 입력(노출 쪽: 제안이 생기거나 수량이 달라지는 것).
실험은 커밋 좌표 사본에서(브리프 「좌표 · 입력」). 출력 형식은 브리프 「출력」.
