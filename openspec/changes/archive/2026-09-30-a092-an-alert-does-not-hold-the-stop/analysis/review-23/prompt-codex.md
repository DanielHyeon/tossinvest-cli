# a092 23라운드 적대 리뷰 — 공통 브리프

읽기 전용 리뷰다. 파일을 만들거나 고치지 말고, git 상태를 바꾸는 명령(commit·checkout·stash·reset·add·init·config)을 실행하지 마라.
LIVE 주문·운영 토글·엔진 기동·`mutating: true` 명령 실행 금지. 운영 원장(`~/.config/tossctl/`)과 자격 증명 파일은 열지 마라.

## 대상

change `openspec/changes/a092-an-alert-does-not-hold-the-stop/` 의 **23판**(커밋 `a88e7079`). 21·22라운드는 BLOCK 이었다. 22라운드의
공통 판정은 「exit goroutine 의 원격 동기 대기 0 은 설계상 섰다」였고, 새로 막은 것은 22판이 새로 연 자리들(모드 통지 신원 · 동기 경로 승격 실패 ·
핀 정의 · 교차 change 기록자)이었다. 23판은 Manager 판정대로 그것을 반영했다.

읽을 것:
- `specs/engine-safety/spec.md`(MODIFIED 둘 · ADDED 하나 — archive 되면 정본 「등급화된 알림」·「배달 실행자의 정지가 다른 루프를 내려서는 안 된다」를
  **통째로 교체**한다), `specs/exit-policy/spec.md`(MODIFIED 하나)
- `design.md` 의 **D0.3h**(23판) — 필요하면 D0.3g(22판) · D0.3f · D0.3e(21판). 그 위 절들은 옛 판 기록이다(⛔ 표지).
- `proposal.md` 머리의 23판 표와 「열린 질문」, `tasks.md` 의 「23.」「22.」「21.」 절
- `analysis/head-ast-21/`(HEAD AST 추출 39개 · `extract.py` · `MANIFEST.txt`)
- 정본: `openspec/specs/engine-safety/spec.md`, `openspec/specs/risk-management/spec.md`, `openspec/specs/exit-policy/spec.md`
- 착수 조건 a124 아카이브: `openspec/changes/archive/2026-09-27-a124-a-deliverer-that-keeps-failing-blocks-entry/`
- 코드: `internal/obs/{notifier.go,mode.go}`, `internal/app/engine/{alertdelivery.go,exitloop.go,exitwiring.go,runtime.go,auxiliary.go,gateway.go,engine.go,alert_control*.go,alertops.go,reconcileloop.go,risk_relaxation_command.go}`,
  `internal/journal/{operating_mode.go,outbox.go,alert_claim.go,core_domain.go,backup.go}`, `internal/execgw/{modegate.go,retry.go,replay.go,gateway.go}`, `cmd/tossctl/{engine.go,engine_alerts.go,flatten.go}`, `internal/flatten/`
- **`review.md` 와 `analysis/review-21/` · `review-22/` · `review-23/` 의 다른 원문은 열지 마라**(이 파일만 예외).

## 판정 기준

- 모든 주장은 파일:줄 증거와 함께. 함수 내부 분기를 근거로 한 주장은 AST 추출물이나 코드를 직접 인용. 추측은 추측이라고 적어라.
- 등급: **P0** = 이대로 구현/archive 하면 안전 불변식 위반·정본 모순·목표 미달성 · **P1** = 구현 전 반드시 고칠 결함 · **P2** = 고칠 것 · 문서만의 결함은 (T).
- 안전 불변식(최우선): 손절·비상 청산의 즉시성을 약화하지 않는다 · 토글 OFF = upstream · 사람 승인 없는 LIVE side effect 금지 ·
  `mutating: true` 자동 실행 금지 · 손절/익절/사이징 변경은 보수 방향만.

## 반드시 판정할 것 (각각 한 줄 판정 + 근거)

1. **K1 — 모드 전이 통지 신원 = 전이 rowid**: 「강화 → 완화 → 재알림 창 안 재강화」의 통지가 흡수되지 않는가. 「변화 없으면 통지 전 반환이라 폭주 없음」 근거
   (`TransitionOperatingMode` B15 `:410` · B16 `:417`)가 참인가. 키를 바꾸면 기존 중복 억제(정본 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만 전송한다」)와 모순되는가.
2. **K2 · K4 — 동기 경로 보수 조항**: 허용 창 + 승격 실패 무조건 차단이 a124 정본과 같은 규범인가, 그리고 게이트를 잘못 여는 경로가 남는가.
3. **K3 — 「동시에」 해석 문장 + 구조 핀**: 해석(같은 전이 호출 안에서 커밋 뒤 투영)과 「분리하면 위반」이 정본 risk-management 와 정합하고, 핀(Commit → Project 순서, 사이에 `go`·반환 없음)이 그 해석을 실제로 지키는가.
4. **K5 (i) — 조립 생성자 전수 핀**: `execgw.New(` 비시험 호출자가 정말 둘(엔진 `gateway.go:296` · flatten CLI `cmd/tossctl/flatten.go:233`)이고 둘 다 `Entry` 를 넘기는가.
   핀이 「존재 확인」이 아니라 「전수 + 역할」을 세는 형태로 적혔는가(호출자를 하나 더한 변이 · `Entry` 를 뺀 변이가 잡히는가).
5. **K6 · K7 — critical 기록의 단일 입구**: 입구 정의(알림기 배제 잠금 아래 임차 없는 기록, 기록자별 재알림 창 · 0 허용)와 입구 밖 기록자 전수(`parkAlert` · a066 `risk_relaxation_command.go:158`)가 맞는가.
   재무장 SHALL 을 exit 기록으로 한정한 것이 다른 기록자(a094 · a090 계획)와 양립하는가.
6. **새 잔여 — 프로세스 밖 기록자(flatten CLI)**(design D0.3h 4): flatten CLI 가 자기 프로세스의 게이트를 만든다(`flatten.go:232`). 그 경로가 실제로 `ReplayInDoubt` → `parkAlert`(`replay.go:358` → `:551`)에 닿는가(`internal/flatten/` 추적).
   닿는다면 엔진 프로세스의 `Acknowledge` 셈~해제 창과의 관계, 잔여로 수용해도 되는가, 아니면 P0/P1 인가.
7. **archive 정합성**: 두 델타가 archive 되면 정본에 모순·지킬 수 없는 SHALL·사본·거짓이 될 HEAD 사실이 들어가는가. MODIFIED 「배달 실행자의 정지가 …」 의 정정 표지·좌표 병기가 규범 문장을 바꾸지 않았는가.
8. 문서 좌표·AST 인용의 정확성(가능한 한 전수), 그리고 23판이 22라운드 발견을 **적은 대로** 반영했는지(빠뜨림·과잉).

## 출력

맨 위에 **판정 한 줄: APPROVE / BLOCK**. 그 아래 발견 표(`# | 등급 | 주장 | 증거(파일:줄) | 권고`), 그 아래 위 1~8 의 판정.
마지막 줄에 `Recommendation: <action> because <reason>`.

## 이 보이스의 렌즈 (codex — 교차 모델)

전 범위 적대 리뷰. 23판이 22라운드 발견을 닫았는지, 그리고 새로 넣은 문장(단일 입구 · 조립 생성자 전수 · 「동시에」 해석 · 통지 신원 · 무통지 예외 · 캡 한정어)이 새 결함을 만드는지 보라. 특히 6번(flatten 프로세스 밖 기록자)을 코드로 추적하라. 참고: 이 트리는 git 저장소가 아니다(git archive).
