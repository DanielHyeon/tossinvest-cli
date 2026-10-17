# a092 21라운드 적대 리뷰 — 공통 브리프

읽기 전용 리뷰다. 파일을 만들거나 고치지 말고, git 상태를 바꾸는 명령(commit·checkout·stash·reset·add·init·config)을 실행하지 마라.
LIVE 주문·운영 토글·엔진 기동·`mutating: true` 명령 실행 금지. 운영 원장(`~/.config/tossctl/`)과 자격 증명 파일은 열지 마라.

## 대상

change `openspec/changes/a092-an-alert-does-not-hold-the-stop/` 의 **21판**(커밋 `8b38de8e` 까지). 20라운드(BLOCK, P0 2 · P1 7)
뒤 사용자 결정으로 범위가 줄고(20-1 ⓒ: *"a092 21판은 「exit goroutine 에서 동기 deliver 제거」로 좁히고 a124 를 착수 조건으로 인용한다"*),
그 뒤 Q3 사용자 결정으로 투영 배선 + AC2 + 사람의 완화 경로가 범위에 들어왔다(*"투영기 배선(SetModeProjector 생산 배선 + 기동
RestoreOperatingModeProjection + AC2 수리)은 21판 범위 포함, 모드 완화 경로는 승인 원칙(자동은 조이기만 / 완화는 OPERATOR + 승인 참조 +
commit 전 audit / journal API + tossctl mutating 명령 / 콘솔 없음)으로 설계한다"*).

읽을 것:
- `specs/engine-safety/spec.md`, `specs/exit-policy/spec.md` (21판 델타 — archive 되면 정본 「등급화된 알림」·「관측 경로와 fail-safe」를 **통째로 교체**하고 ADDED 요구 하나를 더한다)
- `design.md` 의 **D0.3e · D0.3f** (21판 절. 그 위 절들은 옛 판 기록이고 규칙상 본문을 안 고친다 — ⛔ 표지)
- `proposal.md` 머리의 21판 표와 「열린 질문」(Manager 판정 표 포함)
- `tasks.md` 머리의 21판 블록과 「21. 21판 작업」
- `analysis/head-ast-21/` (HEAD AST 추출물 20개 · `extract.py` · `MANIFEST.txt`), `analysis/head-ast-20/`
- 정본: `openspec/specs/engine-safety/spec.md`, `openspec/specs/risk-management/spec.md`, `openspec/specs/exit-policy/spec.md`
- 착수 조건 a124 아카이브: `openspec/changes/archive/2026-09-27-a124-a-deliverer-that-keeps-failing-blocks-entry/`(특히 design D6 · D7 · D10)
- 코드: `internal/obs/notifier.go`, `internal/app/engine/{alertdelivery.go,exitloop.go,runtime.go,gateway.go,engine.go,alert_control*.go,alertops.go}`,
  `internal/journal/{operating_mode.go,outbox.go,alert_claim.go}`, `internal/execgw/{modegate.go,retry.go,replay.go}`, `cmd/tossctl/engine_alerts.go`
- **`review.md` 는 열지 마라** (이전 판정에 끌려가지 않기 위함). a092 밖 다른 change 의 문서는 필요할 때만.

## 판정 기준

- 모든 주장은 파일:줄 증거와 함께. 함수 내부 분기를 근거로 한 주장은 AST 추출물(`analysis/head-ast-21/*.json`)이나 코드를 직접 인용.
- 등급: **P0** = 이대로 구현/archive 하면 안전 불변식 위반·정본 모순·목표 미달성 · **P1** = 구현 전 반드시 고칠 결함 · **P2** = 고칠 것 ·
  문서만의 결함은 (T) 를 붙인다. 추측은 추측이라고 적어라.
- 안전 불변식(최우선): 손절·비상 청산의 즉시성을 약화하지 않는다 · 토글 OFF = upstream · 사람 승인 없는 LIVE side effect 금지 ·
  `mutating: true` 자동 실행 금지 · 손절/익절/사이징 변경은 보수 방향만.

## 반드시 판정할 것 (각각 한 줄 판정 + 근거)

1. 20라운드 P0 **A-1 = B-1**(전달 실패의 진입 차단·승격 주인 없음)이 a124 착지로 정말 닫혔는가 — design D0.3e 3번 책임 대조 표가 빠뜨린 책임은 없는가.
2. 20라운드 P0 **A-2**(제목 경로 = 기록 경로)의 저자 선택 「안 가」(claim → `ReleaseAlertClaim` → 반환, exit goroutine 전용 기록 어댑터)가
   제목(*exit goroutine 은 원격 전송을 기다리지 않는다*)을 달성하는가. 남는 동기 호출자의 `claimAndDeliver` 잠금 범위 변경(D0.3e 5번)의 위험.
3. 21판 델타가 archive 되면 정본에 **모순되거나 지킬 수 없는 SHALL** 이 들어가는가(두 델타 사이, 델타와 다른 정본 요구 사이).
4. Q3 설계(D0.3f): 원장 API 판단(기존 `TransitionOperatingMode` 로 충분), 완화 명령 모양·엔진 프로세스 경유, 배선 순서(a124 AD3 (ii)),
   AC2 원자 교체 + 세대 울타리 안, 전개 절(운영 원장의 기존 `ENTRY_BLOCKED` 행 — 배포 첫 기동에서 진입 차단, 처분은 사람 몫).
5. 열린 질문 Q4(셈~해제를 덮지 않는 기록자 `parkAlert`→`EnqueueAlert`)와 Q6(배달 실행자가 실패마다 임차를 놓는 것) — 각 선택지에 대한 판정 권고.
6. 문서 좌표·AST 인용의 정확성(표본이 아니라 가능한 한 전수).

## 출력

맨 위에 **판정 한 줄: APPROVE / BLOCK**. 그 아래 발견 표(`# | 등급 | 주장 | 증거(파일:줄) | 권고`), 그 아래 위 1~6의 판정.
마지막 줄에 `Recommendation: <action> because <reason>`.

## 이 보이스의 렌즈 (codex — 교차 모델)

전 범위 적대 리뷰. 특히 Claude 저자가 못 볼 것: 동시성 순서(잠금 범위 변경 뒤 deliver 갈래가 Acknowledge 셈~해제와 겹칠 때), 세대 울타리 안의 기동 복원·재묶기, 완화 명령의 엔진 프로세스 경유가 실제 제어 소켓 코드와 맞는지, 델타의 SHALL 이 HEAD 코드로 달성 가능한지.
