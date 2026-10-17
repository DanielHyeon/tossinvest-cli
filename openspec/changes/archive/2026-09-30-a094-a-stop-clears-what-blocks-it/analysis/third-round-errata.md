# a094 3판 정오표 — 현재 코드와 어긋난 인용 (설계 변경 없음)

작성 2026-09-27, HEAD `d2f5d3f1`. **이 문서는 정오표만 담는다.** 3판 proposal·design·tasks·spec 본문은 고치지
않는다 — 3라운드 적대 리뷰(0.5d) 전에는 초안을 개정하지 않는다(Manager 지시). 리뷰어는 3판을 읽을 때
이 표로 줄을 옮겨 읽는다.

## 1. 방법과 범위

- 대상: `proposal.md` · `design.md` · `tasks.md` · `issues.md` · `specs/*/spec.md` 의 `파일.go:줄` 인용 전부.
  `review.md` 는 라운드별 역사 기록이라 대상이 아니다.
- 기준: 3판이 인용한 소스는 base `ec29dc72` 이다(그 base 의 번들이 3판의 근거였다). 각 인용의 base 줄
  텍스트를 HEAD 의 같은 줄과 대조했다(앞뒤 공백 무시). 다르면 HEAD 에서 같은 텍스트를 찾았고, 한 줄짜리가
  여러 곳에 맞으면 인용 범위 전체를 블록으로 대조해 확정했다. 파일 이름만 적힌 인용(`gateway.go` 처럼 둘 이상의
  파일에 해당)은 문맥으로 파일을 정했다.
- 스크립트: `/tmp/claude-1000/a094-lot/scan.py`·`scan2.py`(세션 스크래치). 결과를 아래 표로 옮겼다.
- 집계: 인용 **91건** — 그대로 맞음 **57** · 줄 이동 **32** · 내용 변경 **2**.
- 이 정오표가 보지 **않은** 것: 함수 이름·타입 이름이 아직 있는지(인용 줄 텍스트 대조로 간접 확인만),
  문장이 주장하는 동작의 참·거짓(그것은 3라운드 리뷰의 일).

## 2. 줄 이동·내용 변경 (34 행 — 같은 인용이 두 번 나오면 두 행)

| 문서:줄 | 인용(base `ec29dc72` 기준) | 파일 | HEAD 위치 | 판정 |
|---|---|---|---|---|
| `design.md:14` | `exitloop.go:1296-1300` | `internal/app/engine/exitloop.go` | `:1402-1406` | 줄 이동 |
| `design.md:26` | `exitloop.go:1082` | `internal/app/engine/exitloop.go` | `:1185` | 줄 이동 |
| `design.md:30` | `exitloop.go:1117` | `internal/app/engine/exitloop.go` | `:1223` | 줄 이동 |
| `design.md:55` | `apply_hook.go:846-849` | `internal/journal/apply_hook.go` | `:861-864` | 줄 이동(블록 일치로 확정) |
| `design.md:55` | `exitloop.go:1317` | `internal/app/engine/exitloop.go` | `:1423` | 줄 이동 |
| `design.md:57` | `apply_hook.go:704-706` | `internal/journal/apply_hook.go` | `:719-721` | 줄 이동 |
| `design.md:233` | `exitloop.go:1334-1392` | `internal/app/engine/exitloop.go` | `:1440-1498` | 줄 이동 |
| `design.md:265` | `exitloop.go:166-223` | `internal/app/engine/exitloop.go` | `:167-224` | 줄 이동 |
| `design.md:266` | `exitloop.go:263-282` | `internal/app/engine/exitloop.go` | `:264-283` | 줄 이동(블록 일치로 확정) |
| `design.md:268` | `cmd/tossctl/engine.go:346` | `cmd/tossctl/engine.go` | `:632` | 줄 이동 |
| `design.md:316` | `cmd/tossctl/engine.go:391-396` | `cmd/tossctl/engine.go` | `:692-697` | 줄 이동(블록 일치로 확정) |
| `design.md:319` | `engine.go:420` | `cmd/tossctl/engine.go` | `:722` | 줄 이동 |
| `design.md:324` | `exitloop.go:97` | `internal/app/engine/exitloop.go` | `:98` | 줄 이동 |
| `design.md:354` | `engine.go:349` | `cmd/tossctl/engine.go` | `:635` | 줄 이동 |
| `design.md:365` | `exitloop.go:1141-1144` | `internal/app/engine/exitloop.go` | `:1247-1250` | 줄 이동 |
| `design.md:368` | `exitloop.go:1113-1116` | `internal/app/engine/exitloop.go` | `:1219-1222` | 줄 이동 |
| `design.md:377` | `exitloop.go:1145-1148` | `internal/app/engine/exitloop.go` | `:1251-1254` | 줄 이동 |
| `design.md:421` | `exitloop.go:1673-1679` | `internal/app/engine/exitloop.go` | `:1779-1785` | 줄 이동 |
| `design.md:441` | `notifier.go:216-218` | `internal/obs/notifier.go` | `:381-383` | 줄 이동(블록 일치로 확정) |
| `design.md:593` | `gateway.go:1044` | `internal/execgw/gateway.go` | `:1049` | 줄 이동 |
| `design.md:596` | `exitloop.go:1281-1286` | `internal/app/engine/exitloop.go` | `:1387-1392` | 줄 이동 |
| `design.md:629` | `exitloop.go:166-223` | `internal/app/engine/exitloop.go` | `:167-224` | 줄 이동 |
| `proposal.md:109` | `exitloop.go:1237-1312` | `internal/app/engine/exitloop.go` | `:1343-1418` | 줄 이동 |
| `proposal.md:129` | `recovery.go:207-296` (`:253`) | `internal/reconcile/recovery.go` | `:238-329` (`:284`) | 줄 이동 |
| `proposal.md:131` | `cmd/tossctl/engine.go:374` | `cmd/tossctl/engine.go` | — | **내용 변경** — a102 `6cd643ca` 가 재시작 복구 호출을 `engineRecoverySequence`(`cmd/tossctl/engine.go:604-606`, `return r.Run`)로 옮겼다 — `_, rerr := recovery.Run(ctx)` 줄은 HEAD 에 없다. 호출 자리는 runtime 의 Recover 단계다 |
| `proposal.md:181` | `exitloop.go:1077-1197` | `internal/app/engine/exitloop.go` | `:1177-1297` | 줄 이동 |
| `proposal.md:188` | `exitloop.go:1334-1392` | `internal/app/engine/exitloop.go` | `:1440-1498` | 줄 이동 |
| `proposal.md:292` | `internal/app/engine/exitloop.go:166-223` | `internal/app/engine/exitloop.go` | `:167-224` | 줄 이동 |
| `proposal.md:342` | `exitloop.go:1082` | `internal/app/engine/exitloop.go` | `:1185` | 줄 이동 |
| `proposal.md:346` | `internal/reconcile/recovery.go:207` | `internal/reconcile/recovery.go` | `:238` | 줄 이동 |
| `tasks.md:115` | `cmd/tossctl/engine.go:346` | `cmd/tossctl/engine.go` | `:632` | 줄 이동 |
| `tasks.md:119` | `engine.go:332` | `cmd/tossctl/engine.go` | `:618` | 줄 이동 |
| `tasks.md:159` | `exitloop.go:1141-1144` | `internal/app/engine/exitloop.go` | `:1247-1250` | 줄 이동 |
| `tasks.md:186` | `cmd/tossctl/engine.go:374` | `cmd/tossctl/engine.go` | — | **내용 변경** — a102 `6cd643ca` 가 재시작 복구 호출을 `engineRecoverySequence`(`cmd/tossctl/engine.go:604-606`, `return r.Run`)로 옮겼다 — `_, rerr := recovery.Run(ctx)` 줄은 HEAD 에 없다. 호출 자리는 runtime 의 Recover 단계다 |

## 3. 분기 번호가 바뀐 함수 — `ExitObserver.record`

번들 refresh(3937e341)에서 `record` 는 a111 `882a0b49` 이 앞에 분기 둘(`!o.quoteUsable(quote)`,
`judgement.ObservationSource == ""`)을 더해 **14 → 16** 이 됐다. difflib 대응 B1..B14 → B3..B16.
3판이 `record` 의 분기 번호를 인용한 두 자리는 번호도 옮겨 읽어야 한다.

| 문서:줄 | 3판 표기 | HEAD 표기 |
|---|---|---|
| `proposal.md:181` | `engine.record`(`exitloop.go:1077-1197`, 분기 14) **B3** `:1117` | `exitloop.go:1177-1303`, 분기 16, **B5** `:1223` |
| `design.md:449` | `record` **B3** `:1117` | `record` **B5** `:1223` |

다른 여섯 refresh 함수(`clearTheSymbol`·`submit`·`checkSymbolFree`·`armExitProposalTx`·`ResolveExitProposal`·
`Recovery.Run`)는 분기 번호가 항등이라 번호 인용은 그대로 유효하다(줄은 §2).

## 4. 이웃이 바꾼 것 중 3판의 주장에 닿는 것 (판단은 3라운드 몫)

- **재시작 복구의 호출 자리**(§2 의 `cmd/tossctl/engine.go:374` 두 행). a102 `6cd643ca` 가 복구 호출을
  runtime 의 Recover 단계로 옮겼다(`engineRecoverySequence`, `cmd/tossctl/engine.go:604-606`). 3판 tasks 4.4
  「재시작 복구의 순회는 무변화(`cmd/tossctl/engine.go:374`)」 의 고정 대상 줄이 없어졌다.
- **`Recovery.Run` 본문**. a102 `1c76a580`(「복구가 rate limit에 죽지 않는다」)이 본문을 바꿨다 — 분기 번호는
  항등이다. 3판 R3/R1 소급이 이 함수의 순회를 전제로 하는지 리뷰어가 대조해야 한다.
- **`ExitObserver.record` 앞머리**. a111 의 새 B1 `if !o.quoteUsable(quote) { return nil }`(`exitloop.go:1180-1182`)이
  `record` 의 맨 앞에서 조기 반환한다. 새 B2(`ObservationSource == ""`, `:1199`)는 반환하지 않는 값 보정이다.
  3판 R2 가 `record` B3(현 B5)을 "이미 게이트" 로 쓰는 논증(`design.md:449`)은 그 앞에 새 조기 반환이 생겼다는
  사실과 함께 읽어야 한다.

## 5. a094 R1 과 a089 R2 의 겹침 — 문장 대조

두 change 는 **같은 브로커 본문의 같은 필드**를 읽고, 그 값으로 **동작을 바꾸는가**에서 정반대다.

| | a094 R1 (`specs/order-execution/spec.md`) | a089 R2 (`a089…/specs/engine-safety/spec.md`) |
|---|---|---|
| 무엇을 읽는가 | "이 판정은 응답 본문의 **`code` 필드 값**으로만 성립해야 한다(SHALL)" (:57) · "판정은 **최상위 `code`와 `error.code`를 모두 읽어야**" (:59) | design D4: "`official.APIError.Body`의 JSON에서 `error.code`와 `error.data.field`를 꺼내 별도 필드로 남긴다" |
| 그 값으로 무엇을 하는가 | "브로커 응답이 **요청 자체를 서술하는 code**를 실었으면 그 code로 분류해야 하며 … code가 확정 거절을 뜻하면 그 attempt는 IN_DOUBT로 가지 않고 **종결해야 한다(SHALL)**" (:53) · "`opposite-pending-order-exists`는 확정 거절이다(SHALL)" (:55) | "브로커가 주문을 거부하며 사유 코드와 문제 필드를 제공하면 그 값들을 **질의 가능한 별도 필드로 기록**해야 한다(SHALL)" · "이 기록에 따라 **재시도·억제·에스컬레이션 같은 동작을 분기해서는 안 된다(SHALL NOT)** — 사유별 옳은 대응을 정할 근거가 아직 측정되지 않았고" |
| 근거 | 프로덕션 409 `opposite-pending-order-exists` 3건(proposal R1) | 2026-08-05 실측, "사유별 옳은 대응 … 측정되지 않았고" |

- **겹침**: 파싱 대상이 같다 — `error.code`(a094 는 최상위 `code` 도). 두 change 가 각자 본문 파서를 더하면
  같은 필드의 해석이 두 곳에 산다.
- **충돌**: a089 R2 의 SHALL NOT(사유로 동작 분기 금지)과 a094 R1 의 SHALL(`code` 로 종결)은 둘 다 main spec 에
  들어가면 서로 부정한다. 스펙 파일이 달라(engine-safety / order-execution) `openspec validate` 는 잡지 못한다.
- **a094 의 자기 서술이 틀렸다**: `proposal.md:387-388` "a089·a091·a092와 겹치지 않는다. 각각 계측·알림 등급·
  알림 체류이고, 어느 것도 409 분류나 충돌 해소를 다루지 않는다" 와 `tasks.md:274` "a089 … 독립. 계측이고
  대응을 분기하지 않는다" — 분류는 다루지 않지만 **같은 필드를 읽고, 그 필드로 분기하는 것을 금지한다.**
- **참고**: a089 는 2026-09-27 조사에서 "불구현 아카이브 + a090 신설" 이 사용자 결정 큐에 올라가 있다(R2 는
  "채택 안 함 또는 a094 로 이관" 제안). 그 결정이 나면 이 충돌은 문서상으로만 남는다. 3라운드 리뷰어는 a089 의
  처분과 무관하게 a094 R1 이 main spec 의 기존 조항과 충돌하지 않는지 확인한다.
