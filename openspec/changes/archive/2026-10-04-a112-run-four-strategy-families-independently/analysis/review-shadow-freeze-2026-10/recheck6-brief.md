# SHADOW re-freeze — 재검 6라운드(마지막) 반영 검증 브리프 (2026-10-05)

Manager 판정(2026-10-05): v3.3 을 **종결판** 으로 선언. 6라운드 = 마지막. 범위 = 아래 접기 다섯의 **충실 반영 검증만** — 항목별 이진 판정(반영됨 / 안 됨). **새 사냥 축 금지.**
예외는 P0(실주문 · 노출 · 안전 불변식)뿐 — 발견하면 즉시 보고. 새 P1 급 행동 추정은 판정에 넣지 말고 「구현 단계 RED 후보」 로 따로 적어라(설계 루프 밖에서 실코드 RED · 변이 · 게이트로 잰다).

- 안전 규칙: `brief.md` 그대로(`~/.codex` 금지 · 읽기 전용 · `/tmp` 사본만 · **모든 셸 첫 줄 `set -euo pipefail`** · `git -C` · toplevel 단언 · 네트워크/브로커/LIVE/토글 금지 ·
  시작/끝 `status --short` · `rev-parse HEAD` 동일 보고).
- 좌표: 커밋 `4d22d726`(Go 변경 0). 입력: `design-brief-v3.3-under-review.md` — sha256 `df2d07bcd7b3efdf4ce9d0ce33ad62bdd0576461559cdb60e1841c0f454418ff`, v3.2 대비 `brief-v3.2-to-v3.3.diff` · 5라운드 출력
  `codex-recheck5-output.md`(+ `_codex/recheck5/report.md`) · `voice1-recheck5-output.md` · `voice3-recheck5-output.md`.
- **접기 다섯(각각 반영됨 / 안 됨 + v3.3 좌표):**
  1. P1-E(codex): §5.1 연속성 주장 좁힘 — 「같은 파도 관측은 나이 조건만으로 거부되지 않음」 + record → 다음 게시 사이 UNOBSERVED 간극은 신선도 규칙의 의도된 결과로 **별도 시험**; 이전 파도 수용안 기각.
  2. N5(보이스 1): `strategyLanesMu` 를 잡는 nil-안전 접근자(잠금 순서 strategyLanesMu → runtime.mu) · 생성자에서 epoch 맵 · `invalidateShadow` nil 가드 · **성공 플래그** `if !returnedNil`
     (recover() 아님) · 핀 (v) · central-integrity 신원 보존 시험. recoverMarketLanes/persistMarketLatches 잠금 창 전환은 기각(기록만 확인).
  3. P2 잔여(보이스 3, P1 동급): 경과 ≥ `MaximumStrategyCycleLimit` 이면 nil 반환이라도 실패 취급(폐기 · shadow 미기동) + 핀 (iv) 「마감 + δ 뒤 늦은 nil」(경계 한도 / 한도 − 1ns).
  4. 미등급(보이스 1): 수집 helper 수신자 = 주소 지정 가능한 지역 값 또는 새 slice 를 돌려주는 순수 함수 — shape 핀에 포함.
  5. 사실 정정(보이스 3 비등급): 「panic 은 시장을 잠근다」 는 effective worker 만, refreshOnly 는 삼키고 계속 돈다 — 폐기는 두 경우 모두 작동.
- 출력: 전체(전원 반영됨 = PASS / 하나라도 안 됨 = FAIL) + 다섯 줄 표(접기 · 반영됨/안 됨 · v3.3 좌표 · 한 줄 근거) + P0(있으면) + 「구현 단계 RED 후보」(있으면) + 저장소 무변경 확인.
