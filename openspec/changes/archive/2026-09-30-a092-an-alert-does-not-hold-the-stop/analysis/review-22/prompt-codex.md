# a092 22라운드 적대 리뷰 — 공통 브리프

읽기 전용 리뷰다. 파일을 만들거나 고치지 말고, git 상태를 바꾸는 명령(commit·checkout·stash·reset·add·init·config)을 실행하지 마라.
LIVE 주문·운영 토글·엔진 기동·`mutating: true` 명령 실행 금지. 운영 원장(`~/.config/tossctl/`)과 자격 증명 파일은 열지 마라.

## 대상

change `openspec/changes/a092-an-alert-does-not-hold-the-stop/` 의 **22판**(커밋 `7cf80832`). 21라운드는 세 보이스 모두 BLOCK 이었고
(공통 P0: exit goroutine 의 운영 모드 전이 통지 입구가 동기 전송으로 남음), Manager 판정 뒤 22판이 처분을 반영했다.

읽을 것:
- `specs/engine-safety/spec.md`(MODIFIED 둘 · ADDED 하나 — archive 되면 정본 「등급화된 알림」과 「배달 실행자의 정지가 다른 루프를 내려서는 안 된다」를
  **통째로 교체**한다), `specs/exit-policy/spec.md`(MODIFIED 하나)
- `design.md` 의 **D0.3g**(22판 — 21라운드 처분 반영). D0.3e · D0.3f 는 21판 절이고, 그 위 절들은 옛 판 기록이다(⛔ 표지, 규칙상 본문을 안 고친다).
- `proposal.md` 머리의 22판 · 21판 표와 「열린 질문」, `tasks.md` 의 「22. 22판 작업」「21. 21판 작업」
- `analysis/head-ast-21/`(HEAD AST 추출 39개 · `extract.py` · `MANIFEST.txt`)
- 정본: `openspec/specs/engine-safety/spec.md`, `openspec/specs/risk-management/spec.md`, `openspec/specs/exit-policy/spec.md`
- 착수 조건 a124 아카이브: `openspec/changes/archive/2026-09-27-a124-a-deliverer-that-keeps-failing-blocks-entry/`
- 코드: `internal/obs/{notifier.go,mode.go}`, `internal/app/engine/{alertdelivery.go,exitloop.go,exitwiring.go,runtime.go,gateway.go,engine.go,alert_control*.go,alertops.go,reconcileloop.go}`,
  `internal/journal/{operating_mode.go,outbox.go,alert_claim.go,core_domain.go}`, `internal/execgw/{modegate.go,retry.go,replay.go}`, `cmd/tossctl/{engine.go,engine_alerts.go}`
- **`review.md` 와 `analysis/review-21/` · `analysis/review-22/` 의 다른 보이스 원문은 열지 마라**(이 파일만 예외). 이전 판정에 끌려가지 않기 위함이다.

## 판정 기준

- 모든 주장은 파일:줄 증거와 함께. 함수 내부 분기를 근거로 한 주장은 AST 추출물이나 코드를 직접 인용.
- 등급: **P0** = 이대로 구현/archive 하면 안전 불변식 위반·정본 모순·목표 미달성 · **P1** = 구현 전 반드시 고칠 결함 · **P2** = 고칠 것 · 문서만의 결함은 (T).
  추측은 추측이라고 적어라.
- 안전 불변식(최우선): 손절·비상 청산의 즉시성을 약화하지 않는다 · 토글 OFF = upstream · 사람 승인 없는 LIVE side effect 금지 ·
  `mutating: true` 자동 실행 금지 · 손절/익절/사이징 변경은 보수 방향만.

## 반드시 판정할 것 (각각 한 줄 판정 + 근거)

1. **제목 달성**: 22판 설계대로면 exit 관측 goroutine 에서 원격 전송을 동기로 기다리는 경로가 **0** 인가. 호출 자리가 아니라 **도달 경로**로 전수하라
   (`Notify` 입구 · `AnnounceOperatingMode` 입구 · 그 밖에 있으면 그것까지). D0.3g 1번의 측정(공유 Retrier 는 설정값만 가진 구조체이고 소비자별로 포인터를 대입받는다 →
   exit 쪽 주입 지점 셋만 기록 전용)이 참인가.
2. **기록 경로 안 나**(D0.3g 2번 — 원장 공개 래퍼 `RecordAlert(ctx, a, remindAfter)` = `recordAlertTx` 한 트랜잭션, 임차 없음): 재무장 · 중복 제거 ·
   `Acknowledge` 셈~해제 배제(`n.mu`)가 유지되는가. 새 원장 표면의 위험.
3. **「모든 발송자」 원칙 E**(D0.3g 3번): 범위 밖 동기 발송자의 세 래치 자리에 해제 세대 + `BlockUnlessClearedSince` 를 두는 설계가 승인 뒤 재잠금을 닫는가,
   그리고 게이트를 잘못 **여는** 경로를 새로 만드는가.
4. **커밋 순서 하나**(D0.3g 5번 — rowid): 방향 판정 · 복원 · 투영 울타리가 같은 순서를 쓰게 되는가. rowid 가 커밋 순서라는 근거(DELETE/UPDATE 0 · BEGIN IMMEDIATE)가 참인가.
   AC2 문구(적용 세대 역행 금지 · 보장 시점 = 전이 호출 반환)가 정본 risk-management 「journal 영속과 동시에 EntryGate 계좌 latch로 투영」과 모순인가.
5. **archive 정합성**: 두 델타(MODIFIED 셋 · ADDED 하나)가 archive 되면 정본에 모순·지킬 수 없는 SHALL·사본·거짓이 될 HEAD 사실이 들어가는가.
   새 MODIFIED 「배달 실행자의 정지가 …」 는 정본과 근거 ① 만 다른가.
6. **Q4 구조 핀**(D0.3g 4번의 금지 형태 셋)과 **C8 일반 등급 이관**(별도 보조 실행자 · 유계 버퍼)의 설계가 서는가.
7. **a094 와의 교차(Manager 요청)**: a094 6라운드가 알림 **에피소드 신원** 문제를 찾았다 — `EnqueueAlert` 는 `remindAfter` 0 이라(`outbox.go:142-146`,
   `claimOwed` `remindAfter <= 0` → owed 아님) 전달·승인된 같은 key 행을 PENDING 으로 되돌리지 않아, 재시작·새 연속이 옛 행에 **조용히 흡수**된다. a094 의 처분 방향은
   「에피소드-키(키에 에피소드 신원을 넣음), outbox 계약 무변경」이다. a092 22판의 `RecordAlert` 래퍼 · 재알림(재무장) 절 · Q4 핀(`parkAlert` → `EnqueueAlert`) ·
   C8 이관이 같은 지반을 밟는가 — 충돌·중복·서로의 전제를 깨는 곳이 있는가.
8. 문서 좌표·AST 인용의 정확성(가능한 한 전수).

## 출력

맨 위에 **판정 한 줄: APPROVE / BLOCK**. 그 아래 발견 표(`# | 등급 | 주장 | 증거(파일:줄) | 권고`), 그 아래 위 1~8 의 판정.
마지막 줄에 `Recommendation: <action> because <reason>`.

## 이 보이스의 렌즈 (codex — 교차 모델)

전 범위 적대 리뷰. 특히 21라운드 P0(모드 전이 통지 입구)가 정말 닫혔는지 도달 경로로 다시 세고, 22판이 새로 넣은 것(RecordAlert 래퍼 · 동기 경로의 BlockUnlessClearedSince · rowid 순서 · 일반 등급 보조 실행자 · 정본 MODIFIED)이 새 결함을 만드는지 보라. 참고: 이 트리는 git 저장소가 아니다(git archive).
