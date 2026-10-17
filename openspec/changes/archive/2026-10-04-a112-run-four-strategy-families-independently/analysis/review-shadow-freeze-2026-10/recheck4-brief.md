# SHADOW re-freeze — 재검 4라운드 표적 브리프 (2026-10-05)

Manager 판정: 재검 3라운드(codex FAIL · 보이스 1 FAIL · 보이스 2 PASS · 보이스 3 FAIL, P0 0)의 남은 항목만 표적으로 본다. 전면 재리뷰 아님. 보이스 2 는 PASS 유지(그 노트 2 「파도 스탬프는 시작 시점 ·
같은 조립에서」 는 v3.1 §5 에 반영 — 4라운드 대상 아님).

- 안전 규칙: `brief.md` 그대로(`~/.codex` 금지 · 읽기 전용 · `/tmp` 사본만 `git -C /mnt/D/Axipient/workspace/TossOS archive 4d22d726 | tar -x -C <사본>` · `set -euo pipefail` · `git -C` · toplevel 단언 ·
  네트워크/브로커/LIVE/토글 금지 · 시작/끝 `status --short` · `rev-parse HEAD` 동일 보고).
- **좌표: 커밋 `4d22d726`**(Go 변경 0 — 3라운드와 같은 코드). 입력: 브리프 v3.1 `design-brief-v3.1-under-review.md` — sha256 `c2ea7e821d87224b7bc28923a72017bad95886f5149178826e04b1823688671c`,
  v3(3라운드 입력)과의 차이 `brief-v3-to-v3.1.diff` · 3라운드 출력 `codex-recheck3-output.md` · `voice{1,2,3}-recheck3-output.md`.
- Manager 판정(2026-10-05, v3.1 의 근거):
  1. 묶음 출처 = `evaluate` 가 받아 `record` 가 파도 번호를 올리는 **같은 잠금 안** 에서 시장 칸 `{wave, batch}` 로 보관; shadow 단계는 cycle 클로저가 nil 반환 뒤 **동기로** 복사한 값만 씀;
     refresh · 조립 · 파도 합류 함수와 공유 캐시 필드는 shadow 단계 폐포에서 0; 교차 시장 비잠금 fault 핀; race 목록 등재.
     **핀 ① 재진술 (A):** `runProductionStrategyMarketCycle` 본문의 shadow 식별자는 `evaluate` 인자 한 자리뿐 · shadow 호출 0 · 마지막 문장 dispatch 유지, 그 자리는 §2 ② 허용 목록에 이름으로;
     + ④ evaluate/record 안 단일 보관 대입(호출 · 순회 0) + ⑤ 파도 증가와 같은 임계 구역. 기각 (B)(refresh 시점 칸 + 같은 시장 주기 비중첩 가정).
  2. 만료 = 투영 시점에 `runtime.clk.Now()` 로 `expiresAt` 검사(판정 함수 하나) + 핀 (e).
  3. 계보 충돌 닫힘 = 「관측 없음」(부재 값 — 빈 묶음 아님) + 수집은 안쪽 루프의 문장 수준 helper 하나 · 충돌 return 은 부재 값 리터럴.
  4. FLM 목록: `NewPairedStrategyEntryProductionAssembly` · `record` · `projection` 추가, `refreshPaired…` · `strategyLaneInputs` 는 무편집 사유 + 본문 digest 대조.
  5. codex N2(unsafe/reflect 직접 import, 모듈 패키지 한정) 종결.
- **표적(리뷰어별):**
  - codex: 3라운드 N2 가 v3.1 §2 로 닫히는가(가능하면 사본에서 모듈 패키지 직접 import 0 을 실측) + v3.1 이 새로 연 P0/P1.
  - 보이스 1: N2 PARTIAL(파도 정지 + 만료) → v3.1 §5.1 · N3(묶음 출처 · 금지 집합 · 교차 시장 fault · race) → v3.1 §2 ③ · §5. 3라운드 비등급 노트(충돌 부분 묶음 · 문장 수준 helper 핀) → §4.
  - 보이스 3: P1-1 PARTIAL(충돌 부분 묶음) → v3.1 §4 · FLM 목록 PARTIAL → v3.1 §4 · P1-C(묶음 출처 · 파도 정의 · refresh 호출 금지) → v3.1 §2 ③ · §5.
  - 공통: v3.1 이 새로 연 P0/P1 만 — 특히 핀 ① 재진술 (A) 가 주문 경로 앞에 shadow **일** 을 들이는 길(④ · ⑤ 로 막히는가), `{wave, batch}` 보관이 evaluate 의 오류 · panic 갈래와 어떻게 맞물리는가
    (record 전에 오류가 나면 칸이 이전 파도 그대로인지), 부재 값 대 빈 묶음 구분의 시험 가능성.
- 출력: 전체 판정(PASS = 자기 남은 항목 전부 CLOSED · 새 P0/P1 0 / FAIL) + 표(항목 · 판정 · v3.1 좌표 · 근거) + 새 P0/P1 + 저장소 무변경 확인.
