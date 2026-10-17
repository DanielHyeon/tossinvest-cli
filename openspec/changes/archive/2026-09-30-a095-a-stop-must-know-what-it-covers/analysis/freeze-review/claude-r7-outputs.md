# a095 7판 — Claude 독립 보이스 결과 두 건 (2026-09-29)

프롬프트: `claude-r7-prompt.md` (sha256 `f633fda9…375da`, 실행 사본 일치). 트리: HEAD `90ce43b1` export + a095 오버레이(diff 0),
스크래치패드 `r7-tree` — 두 실행 모두 끝까지 트리가 있었다. 두 보이스는 서로의 결과와 `analysis/freeze-review/`를 보지 않았다.

- **r7a** — 첫 실행(백그라운드). 보고가 늦게 도착해 Manager 지시에 따라 r7b를 다시 돌렸고, 그 뒤 r7a 보고가 도착했다. 30 도구 호출, 약 17분.
  스스로 적은 범위: 번들 8개 해시 대조(나머지 21개는 해시 미대조).
- **r7b** — 재실행(전경). 34 도구 호출, 약 44분. 스스로 적은 범위: 번들 29개 해시 전부 일치, 인용 좌표 약 45개 ast.json 대조.

아래는 두 보고의 요지를 원문 표 그대로 옮긴 것이다(서론 문장 일부 생략, 판정 · 발견 표 · VERDICT 줄은 원문).

---

## r7a — VERDICT: APPROVE

> **VERDICT: APPROVE** — no P0/P1: every revision-7 SHALL is true of the code and jointly satisfiable by one implementation;
> the two P2 items (R5-1's lock-hold rate on the exit loop, and which reports the delta's rules cover) should be recorded
> before implementation.

처분 정합: R4-1~R4-6 · R5-1~R5-4 · R6-1 CONSISTENT, **R6-2 PARTIAL**(→ F3).

| id | sev | finding | evidence | suggested fix |
|---|---|---|---|---|
| F1 | P2 | R5-1이 Q7 **첫째 면**의 노출도 키우는데 문서는 둘째 면으로만 기록 — 래치가 있을 때는 포지션당 약 1회였던 대사 쪽 critical 배달이 PENDING 동안 사이클마다 `n.mu`를 배달 예산만큼 쥘 수 있고, 그동안 exit goroutine의 기존 critical 발신이 같은 뮤텍스를 기다린다. a092 21판 이관이 있어 P1은 아님 | design.md:89-93; proposal.md:204-205; notifier.go:254-255, :421-428; exitwiring.go:341-342, reconcileloop.go:367-368; exitloop.go:831, :1633, :1657, :1687 | D1 귀결과 이관 기록에 첫째 면 빈도 상승을 적고, a092 21판의 모든-보유자 SHALL이 덮는다는 것과 a095 선착지 창을 명시하거나 tasks 7.3에 순서 메모. Manager가 결정 (1)의 이관이 이 빈도를 덮지 않는다고 보면 P1 |
| F2 | P2 | 델타가 어떤 보고를 「무관리 보고」로 보는지 명시하지 않음 — `checkExternalIncrease`도 같은 종류(`EventExitPositionUnmanaged`)를 내므로, 종류로 읽으면 사실 식별 · normal 래치 규칙이 수량 증가 보고에도 걸려 Q4(수량을 key에)와 475150 재생(3.4)의 모양을 조용히 좁힌다 | adoption.go:465-467; spec.md:41-45; design D1 조건 목록(:101-102)에 grown 없음 | 델타가 다스리는 발신 자리(두 `alertUnmanaged`)를 명명하고 수량 증가 보고의 식별 · 억제는 Q4에 남긴다고 적기 |
| F3 | P3 | R6-2가 1개 번들에만 — 「이 함수는 그것을 되던진다」가 28개 맵에 남음, 결과값 없는 함수 5개(reconcile `alertUnmanaged` · `ExitObserver.alertUnmanaged` · `checkExternalIncrease` · `publishBestEffort` · `judgeHoldings`) 포함 | render_bundles.py:34-35, :337; 각 ast.json `results=0` | 번들별 `error_contract` 또는 중립 기본값 + grep 영수증 |
| F4 | P3 | 거부된 알림 블록은 「꺼짐」으로 셈 — 안전하지만 명시 안 됨 | config/notifications.go:54, :106; notifications.go:77 | 델타나 design 「판정 근거」에 명명 |
| F5 | P3 | tasks 6.4 「OFF 동작이 이전과 같다」와 exit-policy 시나리오 「기존 값을 유지한다」가 코드보다 강함 | tasks.md:139-141; exit-policy spec.md:28; issues.md:14 | 6.4를 「등급과 진입 차단 결과」로, 시나리오를 「평단 하락 때문에 낮아지지 않는다」로 |

(a) 공동 충족 가능 — 한 구현(키 `unmanaged|<site>|<condition>|<posID>`, 사실별 normal 전용 래치, critical 래치 무통과, 설정
`enabled`)이 전부 만족. (c) exit-policy 델타 참 — `AvgPrice`/`avg_price`가 exit 경로에 없음. (b) F1 외 소유되지 않은 손절 지연 ·
진입 차단 경로 없음.

---

## r7b — VERDICT: REJECT

> **VERDICT: REJECT** — two P1 wording defects in the engine-safety delta: F1 decides the open Q2(a) through the
> "operator-chosen" definition, and F2's scenario contradicts the notifications-off SHALL NOT and decides Q2(b). Both are
> one-line fixes; F3–F8 may be recorded.

처분 정합: R4-1~R4-6 · R5-1~R5-4 · R6-1 CONSISTENT(일부 문언 결함은 아래), **R6-2 PARTIAL/INCONSISTENT**(→ F7).

| id | sev | finding | evidence | suggested fix |
|---|---|---|---|---|
| **F1** | **P1** | 델타의 「운영자가 고른 상태」 정의가 **거부된 편입 블록**도 포함해 열린 Q2(a)(설정 거부 등급)를 「non-critical」로 정해 버림 — 거부된 블록은 0으로 만들어져 `Enabled=false` · include 없음이 되어 정의에 정확히 맞고, 정본상 거부는 `enabled` 참이거나 include가 있을 때만 일어나므로 거부된 엔진은 운영자가 보호를 **요청한** 엔진이다. Q2(a)=critical이면 이 SHALL NOT은 충족 불가 | spec :47-50(시나리오 :81-83) vs :7 · design :53 · tasks 2.6; `internal/config/adoption_test.go:95-99`; 정본 `exit-policy/spec.md:85`; `adoption.go:135` B12 도달, `:405` B3 `Rejected` 사유 선택 | 정의와 시나리오에 「설정이 거부되지 않았고(`adoption.Rejected == ""`)」를 넣거나 거부를 Q2(a)로 명시해 빼기 |
| **F2** | **P1** | 시나리오 「앞선 normal 보고 뒤의 시도 실패」에 `notifications.enabled ∧ adoption.enabled` 전제가 없는데 THEN이 「critical로 기록된다」 — 알림 off에서 SHALL NOT :52 · 시나리오 :105-107과 모순, `adoption.enabled=false` + include면 Q2(b)를 critical로 정함 | spec :89-91 vs :52, :105-107, :7; 비교: 시나리오 :74는 두 전제를 적음 | :74의 전제를 WHEN에 옮기고 「enabled 시도 실패」를 명시 |
| F3 | P2 | 거부된 알림 블록 — 0으로 만들어져 로드된 `enabled`가 거짓. 운영자가 `enabled: true`에 잘못된 `base_url`을 적어도 B5가 non-critical이 됨. 문서는 B2를 나열만 하고 등급을 안 정함(안전하지만 R4-2 논거와 반대) | `config/notifications.go:53-55`, `:109-111`; `notifications.go:20-24`, `:77`; design :135-139; tasks 2.5/2.5a | Q1에 넣거나 「거부 = 꺼짐」을 명시하고 시험 |
| F4 | P2 | SHALL :38-39 「기록이 실패한 critical 보고는 다음 관측에서 다시 기록을 시도해야 한다」가 무조건 — 문자 그대로면 이미 해소된 사실도 다시 기록하는 메모리 재시도 큐가 필요해 Q8을 악화. 시나리오 :97-99와 design :85는 「같은 실패가 다시 관측되면」 뜻 | spec :38-39 vs :97-99; design :85 | 「같은 사실이 다시 관측되면 그 관측에서 기록을 시도해야 한다」로 |
| F5 | P2 | R5-1 귀결을 Q7 둘째 면으로만 기록 — 첫째 면 빈도도 크게 늘어남(r7a F1과 같은 발견) | design :89-93, :175-185; notifier.go:254, :428-431; exitloop.go:831, :1633, :1687 | design 「남는 경합」 · proposal 「이관 기록」 · 7.3 공시에 증폭 기록, 재결정 불요 |
| F6 | P2 | 시나리오 THEN 둘이 SHALL보다 강함 — (i) exit-policy 「자동 경로의 유효 손절가는 기존 값을 유지한다」(refresh B23이 스칼라 · 스냅샷 불일치 시 낮출 수 있음), (ii) engine-safety :107 「진입 게이트 래치도 … 일어나지 않는다」 vs :56-57(원장에 남은 PENDING 행의 차단은 유지). exit-policy SHALL 자체는 참(자동 경로가 평단을 안 읽음, `ActionReadopt`는 운영자 명령 · 콘솔에서만) | exit-policy delta :27-28 vs design :273 / I1 :14; engine-safety :105-107 vs :56-57; `position_policy_command.go:177,277` | THEN을 「평단 하락으로 인해」 · 「그 보고로 인해」로 한정 |
| F7 | P3 | R6-2가 29개 중 1개 번들에만(r7a F3과 같은 발견) — `checkExternalIncrease`(tasks 3.1과도 모순) · `adoptOne` · `judgeHoldings` · `alertUnmanaged` · `publishBestEffort` | 예: `…checkexternalincrease/function-logic-map.md:33` | 번들별 `error_contract`, 값 단위로 전수 |
| F8 | P3 | 사실 식별 목록 모호(「연기」가 조건 하나인지 셋인지, include 전용 연기 후보의 조건 없음 — 오늘은 B6 「tried」 사유) · 시험 픽스처에 같은 사이클 `adopt` B7 조기 반환 포함 권고 · 머리말 낡음(design :3 「번들 21개」, spec 머리말 「6판」) | design :101-102; `adoption.go:201-207`, `:412` | 편집 |

기타: exit goroutine에 새 critical 없음(`n.mu`는 :254 · :734 · :851에서만). F1 · F3 외 소유되지 않은 진입 차단 경로 없음.
생산에서 critical 기록 가능(`cmd/tossctl/engine.go:668` `Alerts: ectx.Notifier`, 알림기는 항상 원장을 가짐 `exitwiring.go:73`).
