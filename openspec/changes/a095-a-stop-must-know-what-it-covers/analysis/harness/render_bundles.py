#!/usr/bin/env python3
"""a095 3판 FLM 번들 생성기 — AST · 커버리지 · 산문을 네 파일로 씀.

분기 표는 `branch_rows.py` 와 같은 규칙으로 기계가 채움. 산문(역할 · 입력 · 호출 · 변이 · 안전 결론 ·
시험 요구)만 아래 `BUNDLES` 에 손으로 적음 — 분기 **주장**은 표의 행을 가리키는 방식으로만 씀.

사용 (저장소 루트에서):
    python3 openspec/changes/a095-a-stop-must-know-what-it-covers/analysis/harness/render_bundles.py \
        <ast 디렉터리> <coverprofile>...

`<ast 디렉터리>` 에는 `go run ./tools/logic-map --file F --func Q > <stem>.json` 산출물이 있어야 함
(stem 은 아래 `BUNDLES` 의 첫 칸). 위험 패턴 보고는 `tools/logic-map/risk_pattern_report.py` 로 만듦.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from branch_rows import ROOT, cell, load_blocks  # noqa: E402

CHANGE = Path(__file__).resolve().parents[2]
OUT = CHANGE / "analysis" / "function-logic"
COVERAGE_COMMAND = (
    "`go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · "
    "app/engine · reconcile 네 패키지에 각각"
)
DEFAULT_TEST = "기존 — a095는 이 함수를 바꾸지 않는다"

# stem, 번들 디렉터리, 역할, 입력 행들, 호출 문단, 변이 문단, 안전 경계, High-risk, {분기: 시험 요구}
BUNDLES: list[dict] = [
    {
        "stem": "event--SeverityOf",
        "dir": "internal-obs--severityof",
        "role": "이벤트 **종류** 하나를 받아 등급을 답한다. `criticalEvents` map만 본다 — 분기가 하나다.",
        "inputs": [
            ("`t`", "`EventType`", "호출자", "없음 — 순수"),
            ("`criticalEvents`", "등급 표(18종)", "소스의 리터럴 map", "미등재는 `SeverityNormal`"),
        ],
        "calls": "없다.",
        "mutations": "없다.",
        "boundary": "**3판은 이 함수의 입력이 종류뿐이라는 사실을 설계 제약으로 받는다.** 결정 (2)는 등급을 "
                    "「이벤트 종류가 아니라 사실로」 가르라고 한다. B1은 종류만 보므로 같은 종류의 두 발신 자리"
                    "(exit 관측 · reconcile 대사)에 다른 등급을 줄 수 없다 — 그것을 어떻게 싣는지가 Q1이다.",
        "high_risk": "yes — 이 답이 알림의 durable 여부와 진입 차단 도달 여부를 정한다.",
        "tests": {"B1": "**a095 2.1 · 2.2** — exit 관측 자리의 사실은 normal, reconcile 자리의 비선택 사실은 critical. "
                        "싣는 방식은 [비움 — Q1]"},
    },
    {
        "stem": "notifier--Notifier.Notify",
        "dir": "internal-obs--notifier.notify",
        "role": "이벤트를 등급 매기고 배달한다. B1이 best-effort 경로와 durable 경로를 가른다.",
        "inputs": [
            ("`e`", "이벤트", "호출자", "`SeverityOf(e.Type)`가 등급을 정한다"),
            ("`severity`", "`SeverityOf`의 답", "위 함수", "critical이 아니면 B1 창의 `n.publishBestEffort`"),
        ],
        "calls": "`SeverityOf` · `n.logEvent`(등급 판정 **앞**, 두 경로 공통) · `n.publishBestEffort`(B1 창) · "
                 "`n.notifyCritical`(B1 뒤).",
        "mutations": "없다 — 두 경로가 각자 부작용을 갖는다. `n.mu`는 이 함수가 잡지 않는다"
                     "(`claimAndDeliver` 번들 참조).",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** 결정 (1)의 요구 「exit goroutine 에 critical Notify 를 "
                    "새로 두지 않는다」는 exit 관측 자리의 사실이 B1 창(`publishBestEffort`)으로 가는 것으로 "
                    "성립한다 — 그 경로에는 `n.mu`도 outbox도 재시도 대기도 없다.",
        "high_risk": "yes — 알림이 원장에 남는지, 진입 차단에 닿는지가 여기서 갈린다.",
        "tests": {"B1": "**a095 2.1** — exit 관측 자리의 사실은 B1 창으로 간다 · **2.2** — reconcile 자리의 "
                        "비선택 사실은 B1을 지나 `notifyCritical`로 간다"},
    },
    {
        "stem": "notifier--Notifier.publishBestEffort",
        "dir": "internal-obs--notifier.publishbesteffort",
        "role": "일반 등급 알림을 보내고 잊는다. outbox 행도 재시도도 없다.",
        "inputs": [
            ("`n.Publisher`", "전송기", "`Notifier` 구성", "nil이면 B1 창의 return"),
            ("`n.Log`", "구조 로그", "`Notifier` 구성", "B2 조건의 `&& n.Log != nil`"),
        ],
        "calls": "`n.Publisher.Publish`(B2 조건 안) · `notificationFor` · `n.Log.Warn`(B2 창).",
        "mutations": "외부 전송뿐. 원장에 쓰지 않고 뮤텍스를 잡지 않는다.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** 2판은 B1을 「로그도 없다」로 적었으나 구조 로그는 "
                    "`Notify`가 등급 판정 전에 `logEvent`로 이미 남긴다(보이스 B B-P2-13). 이 경로의 부재는 "
                    "「durable outbox와 재시도가 없다」이다.",
        "high_risk": "no — 이 함수 자체는 설계대로다.",
        "tests": {},
    },
    {
        "stem": "notifier--Notifier.notifyCritical",
        "dir": "internal-obs--notifier.notifycritical",
        "role": "critical 등급의 durable 경로. outbox 기록을 먼저 하고 배달하며, 배달이 빚졌는데 못 보냈으면 "
                "운영 모드를 승격한다.",
        "inputs": [
            ("`n.Journal`", "원장", "`Notifier` 구성", "nil이면 B1 창에서 best-effort로 격하"),
            ("`sent, owed, err`", "`claimAndDeliver`의 답", "아래 번들", "B3 오류 → `n.escalate` · B4 `owed && !sent` → `n.escalate`"),
        ],
        "calls": "`n.claimAndDeliver`(뮤텍스 안) · `n.escalate`(B3 창 · B4 창, 뮤텍스 밖) · `n.publishBestEffort`(B1 창).",
        "mutations": "outbox 행(아래 번들) · 승격 기록(`escalate`).",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** 결정 (2)는 운영자가 고른 상태에서 a095의 사실이 B3·B4의 "
                    "`n.escalate`에 닿지 않을 것을 요구한다 — 그 사실이 이 함수에 **들어오지 않게** 하는 것이 "
                    "해법이고, 들어온 뒤 거르는 분기를 이 함수에 더하지 않는다.",
        "high_risk": "yes — 진입 차단 승격의 발화 자리다.",
        "tests": {"B4": "**a095 2.5** — 알림 off 엔진에서 a095의 사실이 이 창에 도달하지 않는다"},
    },
    {
        "stem": "notifier--Notifier.claimAndDeliver",
        "dir": "internal-obs--notifier.claimanddeliver",
        "role": "outbox에 전송 의무를 묻고, 빚졌으면 배달한다 — 둘 다 `n.mu` 아래에서.",
        "inputs": [
            ("`record.EventKey`", "dedupe 키", "`n.eventKey(e)` — `e.Key`가 있으면 그것", "같은 키의 전달된 행은 창 안에서 B5"),
            ("`n.remindAfter()`", "재알림 창", "`DefaultRemindAfter`", "창 안이면 `ClaimSettled`"),
        ],
        "calls": "`n.mu.Lock`/`n.mu.Unlock`(defer) · `n.Journal.ClaimAlertForDelivery` · `n.Gate.Block`(B3 창) · "
                 "`n.logClaimHeld`(B6) · `n.deliver`(switch 뒤, 뮤텍스 안).",
        "mutations": "outbox 행 claim · 실패 시 진입 게이트 래치(B3 창).",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** 두 사실이 3판의 제약이 된다. ① `n.deliver`가 `n.mu`를 "
                    "잡은 채 불린다 — reconcile 자리의 critical 배달이 그동안 같은 `Notifier`의 다른 critical "
                    "호출자를 기다리게 한다(Q7). ② B5(`ClaimSettled`)가 같은 키의 재전송을 창 안에서 삼킨다 — "
                    "수량 증가 재알림이 critical이라면 키 설계가 이 창과 대조되어야 한다(Q4).",
        "high_risk": "yes — 배달 직렬화와 재알림 억제의 자리다.",
        "tests": {"B5": "**a095 3.3** — [비움 — Q4] 수량 증가 사실의 키와 재알림 창"},
    },
    {
        "stem": "notifier--Notifier.deliver",
        "dir": "internal-obs--notifier.deliver",
        "role": "빚진 critical 알림을 재시도 예산 안에서 보내고, 끝내 못 보내면 진입 게이트를 래치한다.",
        "inputs": [
            ("`n.Publisher`", "전송기", "`Notifier` 구성", "nil이면 B3 창에서 `lastErr`를 세우고 루프를 벗어난다"),
            ("`attempts`", "재시도 횟수", "`n.Attempts` 또는 기본값(B1)", "B19 · B20 사이 대기"),
        ],
        "calls": "`n.Publisher.Publish` · `n.Journal.MarkAlertDelivered` · `n.Journal.MarkAlertAttemptFailed` · "
                 "`n.wait`(B20) · `n.Journal.ReleaseAlertClaim` · `n.Gate.Block`(B12 · B18 · B27 창).",
        "mutations": "outbox 행 상태 · 진입 게이트 래치.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** B3(`n.Publisher == nil`)은 **미진입**이다 — 알림을 끈 "
                    "엔진에서 critical이 어디로 가는지를 밟는 시험이 이 함수 단위로는 없다. 결정 (2)의 근거가 "
                    "이 경로이므로 a095는 그 사실이 여기 오지 않음을 호출자 쪽 시험(2.5)으로 고정한다.",
        "high_risk": "yes — 진입 차단 래치의 자리다.",
        "tests": {"B3": "**a095 2.5** — 알림 off에서 a095의 사실이 이 창에 오지 않는다(호출자 쪽에서 고정)"},
    },
    {
        "stem": "adoption--ReconcileDriver.judgeHoldings",
        "dir": "internal-app-engine--reconciledriver.judgeholdings",
        "role": "안정 스냅샷의 보유를 게이트에 통과시키고, 편입하거나 무관리로 모은다. "
                "**`checkExternalIncrease`와 reconcile 쪽 `alertUnmanaged`의 유일한 호출자다.**",
        "inputs": [
            ("`p.ExitEligible()`", "진입 결정 또는 편입 기록이 있나", "원장 `positions`", "B7 창 — 참이면 `continue`"),
            ("`p.Adopted()`", "편입 기록이 있나", "원장 `positions.adoption_id`", "B8 창 — 참일 때만 `checkExternalIncrease`"),
            ("`d.blocked` · `fresh`", "전이 상태", "RECONCILE 추적기 · 스냅샷 나이", "B9 · B10 — 무알림 `continue`"),
            ("`d.opts.Adoption`", "편입 설정", "런타임 config", "B11(exclude) · B12(off∧미지정) → `unmanaged`"),
        ],
        "calls": "`d.opts.Journal.CurrentPosition` · `d.checkExternalIncrease`(B8 창) · `d.blocked`(B9) · "
                 "`d.opts.Adoption.Excludes`(B11) · `d.opts.Adoption.Included`(B12) · `d.adopt` · "
                 "`d.alertUnmanaged`(B15 창).",
        "mutations": "`cycle.Unmanaged` 계수 · 편입(`d.adopt` 경유) · 알림(`alertUnmanaged` 경유).",
        "boundary": "**3판의 사실 셋이 이 함수에서 나온다.** ① B7 창은 `ExitEligible`이면 `continue`하고 B8만 "
                    "`checkExternalIncrease`를 부른다 — **엔진이 직접 연 포지션(편입 기록 없음)의 수량 증가는 "
                    "어디서도 검사되지 않는다**(결정 (3)(i)의 대상). ② 편입 기록이 없는 보유는 B8에 오지 않으므로 "
                    "`checkExternalIncrease` B2가 받는 입력이 아니다 — 2판 R2-B2의 전제가 거짓이었다. "
                    "③ 무관리 보유는 B11(exclude) · B12(off∧미지정) · B14(편입 실패)로 모여 B15에서 알려진다 — "
                    "운영자가 고른 상태(B11 · B12)와 고르지 않은 상태(B14)가 **다른 분기**로 이미 갈린다.",
        "high_risk": "yes — 편입과 무관리 보고의 입구다.",
        "tests": {
            "B7": "**a095 3.2** — [비움 — Q3] 엔진 개설 포지션의 수량 증가 검사",
            "B8": "**a095 3.1** — 편입된 포지션만 `checkExternalIncrease`에 온다(R2-B2 삭제의 근거)",
            "B9": "**a095 2.8** — 전이 상태 무알림 유지",
            "B10": "**a095 2.8** — 전이 상태 무알림 유지",
            "B11": "**a095 2.3** — exclude는 normal, 진입 차단 없음",
            "B12": "**a095 2.4** — off∧미지정은 normal(정본 exit-policy `adoption.enabled` false 동등)",
            "B14": "**a095 2.2** — 편입 시도 실패는 운영자가 고른 상태가 아니다 → critical",
            "B15": "**a095 2.2 · 2.3 · 2.4** — 모인 사유별 등급",
        },
    },
    {
        "stem": "adoption--ReconcileDriver.adopt",
        "dir": "internal-app-engine--reconciledriver.adopt",
        "role": "후보를 한 번의 묶음 시세 읽기로 값 매기고 편입할 수 있는 것을 편입한다. 편입된 id 집합을 돌려준다.",
        "inputs": [
            ("`d.observeCandidates`의 답", "후보 시세", "브로커 시세 경로", "B2 — 오류면 빈 집합을 돌려준다"),
            ("`quotes[key]`", "종목별 관측", "위 읽기", "B6 — 없으면 `cycle.Deferred`, 그 후보는 편입되지 않는다"),
            ("시세 나이", "`PriceStaleness`", "config · 기본값(B4)", "B7 — 넘으면 남은 후보 전부를 편입하지 않고 return"),
        ],
        "calls": "`d.observeCandidates`(B1 뒤) · `adoptionQuoteKey` · `d.logDeferred`(B7 창) · `d.adoptOne`(B8).",
        "mutations": "편입(`d.adoptOne` 경유) · `cycle.Deferred` · `cycle.Adopted` 계수.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** 이 함수가 돌려준 집합에 없는 후보는 호출자 `judgeHoldings` B14가 "
                    "무관리로 모은다. 따라서 B2(시세 읽기 오류) · B6(관측 없음) · B7(관측 묵음)로 **연기된** 후보도 "
                    "「enabled 시도 실패」 사유(`alertUnmanaged` B5)로 알려진다. 그 사유를 critical로 올리면 일시적 "
                    "시세 실패가 critical 알림이 된다 — 결정이 덮지 않는 귀결이다(Q2(b)).",
        "high_risk": "yes — 편입의 유일한 입구다.",
        "tests": {"B2": "**a095 2.6** — [비움 — Q2(b)] 연기된 후보의 등급",
                  "B6": "**a095 2.6** — [비움 — Q2(b)] 연기된 후보의 등급",
                  "B7": "**a095 2.6** — [비움 — Q2(b)] 연기된 후보의 등급"},
    },
    {
        "stem": "adoption--ReconcileDriver.alertUnmanaged",
        "dir": "internal-app-engine--reconciledriver.alertunmanaged",
        "role": "엔진이 관리하지 않는 보유를 알린다. why-matrix(B2 switch)가 사유를 고른다.",
        "inputs": [
            ("`d.unmanaged[p.ID]`", "이미 알렸나", "프로세스 메모리 map", "B1 — 프로세스당 1회"),
            ("`d.opts.Adoption`", "편입 설정", "런타임 config", "B3~B6이 사유 문구를 고른다"),
        ],
        "calls": "`d.opts.Adoption.Excludes`(B4) · `d.opts.Adoption.Included`(B6) · `d.alert` · `d.label`.",
        "mutations": "`d.unmanaged[p.ID] = true`(메모리) · 알림 1건. 키는 `exit.position_unmanaged|<posID>` — "
                     "exit 관측 자리와 **같은 철자**다.",
        "boundary": "**사유 분기는 이미 사실별로 갈려 있다** — B3(설정 거부) · B4(exclude) · B5(enabled 시도 실패) · "
                    "B6(include 지정 시도 실패) · 기본(off∧미지정). 결정 (2)는 B4와 기본을 normal로 두라 하고, "
                    "B5는 운영자가 고른 상태가 아니다. B3 · B6의 분류는 결정이 덮지 않는다(Q2). 이 함수의 키는 "
                    "결정 (3)(iii)에 따라 exit 관측 자리와 갈라야 한다.",
        "high_risk": "yes — reconcile 쪽 무보호 보고의 자리다.",
        "tests": {
            "B1": "기존 — 프로세스당 1회 래치는 유지",
            "B3": "**a095 2.6** — [비움 — Q2] 설정 거부의 등급",
            "B4": "**a095 2.3** — exclude는 normal",
            "B5": "**a095 2.2** — enabled 시도 실패는 critical",
            "B6": "**a095 2.6** — [비움 — Q2] include 지정 시도 실패의 등급",
        },
    },
    {
        "stem": "adoption--ReconcileDriver.checkExternalIncrease",
        "dir": "internal-app-engine--reconciledriver.checkexternalincrease",
        "role": "편입된 포지션의 수량이 편입 기록보다 늘었는지 보고 알린다. 주석이 t0 동결을 의도적 설계(A8)로 "
                "선언한다.",
        "inputs": [
            ("`d.grown[p.ID]`", "이미 알렸나", "프로세스 메모리 map", "B1 창의 return"),
            ("`AdoptionOf(p.ID)`", "편입 기록", "원장", "B2 창의 return. 호출자 가드(`judgeHoldings` B8)로 편입된 "
             "포지션만 오고 `positions.adoption_id`는 `position_adoptions(id)`를 참조하므로, 여기 오는 입력은 조회 "
             "오류다"),
            ("`p.Quantity` 대 `adoption.Quantity`", "현재 수량 대 편입 수량", "스냅샷 대 원장", "B3 — 늘지 않았으면 return"),
        ],
        "calls": "`d.opts.Journal.AdoptionOf`(B1 뒤) · `riskcalc.CompareDecimal`(B2 뒤) · `d.alert` · `d.label`.",
        "mutations": "`d.grown[p.ID] = true`(메모리) · 알림 1건(키 `…|grown|<posID>` — 수량이 없다). "
                     "원장의 exit state는 건드리지 않는다.",
        "boundary": "**2판 FLM의 「B2 — 엔진이 직접 연 포지션과 미편입 보유가 여기로 온다」는 거짓이었다** — "
                    "`judgeHoldings` B8 창이 편입된 포지션만 부른다. 결정 (3)이 R2-B2를 삭제했으므로 3판은 B2를 "
                    "바꾸지 않는다. 알림 본문 스스로 *\"늘어난 수량은 원래 수량 기준으로 산정된 손절의 보호를 "
                    "받는다\"*라고 쓴다 — 이 사실은 「무보호」가 아니며 그 종류·등급은 Q4다.",
        "high_risk": "yes — 편입 후 수량 증가를 알리는 유일한 자리다.",
        "tests": {
            "B1": "**a095 3.3** — [비움 — Q4] 수량 기준 재알림 여부",
            "B2": "**a095 3.1** — 무변화(R2-B2 삭제)",
            "B3": "**a095 3.3** — [비움 — Q4] 증가 사실의 종류·등급",
        },
    },
    {
        "stem": "exitloop--ExitObserver.ObserveOnce",
        "dir": "internal-app-engine--exitobserver.observeonce",
        "role": "exit 관측 한 사이클. 작업 집합을 만들고 시세를 읽은 뒤 포지션마다 판정한다.",
        "inputs": [
            ("`o.workingSet`의 답", "판정할 포지션", "원장", "B2 창의 return(오류)"),
            ("`o.observe`의 답", "시세", "브로커", "B4 창 — `o.checkOutage`"),
        ],
        "calls": "호출 순서(`ast.json` 좌표): `o.workingSet` `:426` → `o.observe` `:441` → `o.judge` `:465`.",
        "mutations": "사이클 계수 · 판정(`o.judge` 경유).",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** `o.workingSet`이 `o.observe` · `o.judge`보다 **먼저** 불린다 "
                    "— 작업 집합 안의 동기 알림은 이 사이클의 모든 손절 판정 앞에 선다. 결정 (1)이 그 자리에 "
                    "critical을 두지 않게 한다.",
        "high_risk": "yes — 손절 판정 루프다.",
        "tests": {"B1": "**a095 2.1 · 6.2** — 작업 집합 안의 발신은 normal이라 판정 앞 체류가 늘지 않는다"},
    },
    {
        "stem": "exitloop--ExitObserver.workingSet",
        "dir": "internal-app-engine--exitobserver.workingset",
        "role": "exit 관측이 판정할 포지션 집합을 만든다. 적격하지 않은 보유는 알리고 건너뛴다.",
        "inputs": [
            ("`p.ExitEligible()`", "진입 결정 또는 편입 기록", "원장", "B6 — 거짓이면 `o.alertUnmanaged` 후 `continue`"),
        ],
        "calls": "`o.opts.Journal.Positions` · `o.opts.Journal.OpenExitStateResults` · `o.alertUnmanaged`(B6 창) · "
                 "`o.openState`(B7) · `o.opts.Journal.QuarantineExitSnapshot`.",
        "mutations": "`cycle.Unmanaged` 계수 · 알림(B6) · 격리 기록.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** B6 창의 `o.alertUnmanaged`는 exit goroutine 안의 동기 "
                    "호출이며, 결정 (1)이 그 자리를 normal로 둔다. B6에는 전이 상태 판정이 없다(reconcile 쪽 "
                    "`judgeHoldings` B9 · B10과 다르다) — normal이므로 진입 차단에 닿지 않는다.",
        "high_risk": "yes — 손절 판정 앞의 작업 집합이다.",
        "tests": {"B6": "**a095 2.1** — 이 자리의 사실은 normal · `TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce`가 "
                        "이미 normal을 고정한다(`exitloop_test.go:508`)"},
    },
    {
        "stem": "exitloop--ExitObserver.alertUnmanaged",
        "dir": "internal-app-engine--exitobserver.alertunmanaged",
        "role": "exit 관측이 본 무관리 보유를 알린다.",
        "inputs": [("`o.unmanaged[p.ID]`", "이미 알렸나", "프로세스 메모리 map", "B1 창의 return")],
        "calls": "`o.alert` · `o.label` · `string`.",
        "mutations": "`o.unmanaged[p.ID] = true`(메모리) · 알림 1건. 키 `exit.position_unmanaged|<posID>` — "
                     "reconcile 쪽 `alertUnmanaged`와 **같은 철자**다.",
        "boundary": "**본문은 바꾸지 않는다.** 결정 (1)에 따라 이 자리의 사실은 normal로 남고, 결정 (3)(iii)에 따라 "
                    "키가 reconcile 자리와 달라야 한다.",
        "high_risk": "yes — exit goroutine 안의 발신이다.",
        "tests": {"B1": "**a095 2.1 · 2.7** — normal 유지 · reconcile 자리와 다른 키"},
    },
    {
        "stem": "exitwiring--notifierAlerter.ExternalPositionFound",
        "dir": "internal-app-engine--notifieralerter.externalpositionfound",
        "role": "reconcile 패키지의 fold 알림을 `Notifier`로 옮긴다. 주석이 등급을 normal로 선언한다.",
        "inputs": [("`a.notifier`", "알림기", "배선", "nil이면 B1 창의 return")],
        "calls": "`a.notifier.Notify` · `a.names.Label` · `strings.TrimSpace`.",
        "mutations": "알림 1건. 본문은 `exit_eligible`과 무관하게 「손절·익절이 자동으로 걸려 있지 않다」를 쓴다.",
        "boundary": "**a095는 이 함수와 그 등급을 바꾸지 않는다.** 생산 배선에서 이 함수의 유일한 호출 자리"
                    "(`Ingestor.IngestExternalPositions` B13)는 B12(`in.Alert == nil`)에 막힌다 — "
                    "`ReconcileDriver`가 `d.ingest.Alert = nil`로 복사하기 때문이다(`reconcileloop.go:338`). "
                    "2판 tasks 6.2a의 「오류가 대사를 실패시킨다」는 이 자리의 등급이 바뀔 때만 성립하므로 "
                    "3판에서는 성립하지 않는다.",
        "high_risk": "no — 생산에서 도달하지 않고 3판은 등급을 바꾸지 않는다.",
        "tests": {"B1": "**a095 2.9** — 등급 normal 유지"},
    },
    {
        "stem": "external--Ingestor.IngestExternalPositions",
        "dir": "internal-reconcile--ingestor.ingestexternalpositions",
        "role": "로컬 인스턴스가 없는 보유를 원장에 접어 넣고 알린다.",
        "inputs": [("`in.Alert`", "알림 어댑터", "배선", "B12 — nil이면 알림 없이 `continue`")],
        "calls": "`in.Journal.FillWatermark` · `in.Journal.ApplyPositionAdjustment` · `in.Alert.ExternalPositionFound`(B13).",
        "mutations": "`position_adjustments` · `positions`(조정 경유) · 알림.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** B12가 `in.Alert == nil`이면 알림을 건너뛴다 — "
                    "생산의 `ReconcileDriver`는 이 필드를 nil로 둔다. 따라서 네 번째 발신 자리는 생산에서 "
                    "도달하지 않는다(보이스 B B-P1-7의 주장을 이 분기로 확인).",
        "high_risk": "no — 3판에서 무변화.",
        "tests": {"B12": "**a095 2.9** — 생산 배선에서 알림 어댑터가 nil임을 고정"},
    },
    {
        "stem": "exit_state--Journal.recordExitJudgementTx",
        "dir": "internal-journal--journal.recordexitjudgementtx",
        "role": "exit 판정을 기록한다. `baseline_price`를 **값이 바뀌게** 쓰는 정상 경로다.",
        "inputs": [
            ("`recomputed`", "재계산 스냅샷", "호출자", "B23 — nil이면 옛 스칼라 단조 검사(B24 · B25)"),
            ("`current.Baseline`", "저장된 기준선", "원장", "B25 — `notBelow(\"baseline\", …)`"),
        ],
        "calls": "`notBelow`(B24 · B25) · `exitpolicy.SelectRecoverySnapshot`(B29 창) · `tx.ExecContext`(B39 — UPDATE).",
        "mutations": "`exit_states.baseline_price` 외 판정 열 · 제안 무장 · `exit_events`.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** 이 함수가 2판의 「유효 손절 쓰기 경로는 재편입 하나」를 "
                    "반증한다: 판정마다 `baseline_price`를 UPDATE하며(B39), 옛 경로는 B25의 `notBelow`가 하향을 "
                    "거부하고 스냅샷 경로는 B29 창의 `SelectRecoverySnapshot`이 고른다. 475150의 57,900은 이 "
                    "경로의 산물이다.",
        "high_risk": "yes — 손절선의 정상 갱신 자리다.",
        "tests": {"B25": "**a095 5.3** — 하향 거부가 이미 있다는 사실을 issues I1에 인용",
                  "B29": "**a095 5.3** — 스냅샷 경로의 선택을 issues I1에 인용"},
    },
    {
        "stem": "exit_observation_refresh--Journal.RefreshExitObservation",
        "dir": "internal-journal--journal.refreshexitobservation",
        "role": "판정 없이 관측 증거만 새로 고친다. `baseline_price`를 UPDATE 문에 포함한다.",
        "inputs": [("`request.Snapshot`", "새 관측 스냅샷", "호출자", "B23 — 운영 선(보호가 포함)이 다르면 거절")],
        "calls": "`compareObservationEvidence` · `sameExitOperationalLine`(B23 조건) · `tx.ExecContext`(UPDATE).",
        "mutations": "`exit_states` 관측 열 — `baseline_price`는 B23을 통과한 스냅샷의 `CurrentProtection`.",
        "boundary": "**a095는 이 함수를 바꾸지 않는다.** B23이 `sameExitOperationalLine`(보호가 · 워터마크 · 레벨 포함)이 "
                    "거짓이면 거절하므로, 이 함수가 `baseline_price`에 쓰는 값은 **저장된 effective 스냅샷의 "
                    "보호가와 같다**(B12가 그 스냅샷의 존재를 요구한다). 2판 리뷰가 이것을 「writer」로 셌다 — "
                    "UPDATE 문으로는 맞고 값의 이동으로는 아니다.",
        "high_risk": "yes — 손절선 열을 쓰는 자리다.",
        "tests": {"B23": "**a095 5.3** — issues I1에 「값 무변화 재기록」으로 인용"},
    },
    {
        "stem": "apply_hook--resetExitStateForReadoptTx",
        "dir": "internal-journal--resetexitstateforreadopttx",
        "role": "재편입 시 보호 기준 전체를 새 관측으로 다시 세운다. 주석: *\"the only reset writer for the four "
                "guarded execution-time columns\"* — 네 열은 손절 열이 아니다(보이스 A A-4).",
        "inputs": [
            ("`observation`", "재편입 관측가 · 합성 손절", "`positionpolicy.ActionReadopt`", "B1 — 무효면 거절"),
            ("`positionID`", "대상", "호출자", "B6 — 정확히 1행"),
        ],
        "calls": "`exitpolicy.OpenRatchetState` · `tx.QueryRowContext`(B2) · `tx.ExecContext`(UPDATE) · `appendExitEventTx`.",
        "mutations": "`exit_states`의 `entry_price` · `initial_stop` · `initial_risk` · `baseline_price` · `high_water` 등을 "
                     "새 값으로 덮어쓴다.",
        "boundary": "**a095는 이 함수를 바꾸지 않고 부르지도 않는다.** 분기 여섯은 전부 오류·행 수 검사이며 **이전 "
                    "`baseline_price`와 비교하는 분기가 없다** — 이 경로는 기준선을 낮출 수 있다(운영자 행동에서만 "
                    "불린다). 2판의 「기준을 다시 세우는 유일한 쓰기 자리」는 「reset 경로가 하나」로만 참이다.",
        "high_risk": "yes — 손절선을 덮어쓴다.",
        "tests": {"B6": "**a095 5.3** — issues I1에 「비교 없는 reset」으로 인용"},
    },
]


def render(bundle: dict, ast_dir: Path, profiles: list[str]) -> None:
    source_ast = ast_dir / f"{bundle['stem']}.json"
    value = json.loads(source_ast.read_text(encoding="utf-8"))
    relative = value["file"]
    qualified = (f"{value['receiver']}.{value['function']}" if value.get("receiver") else value["function"])
    lines = (ROOT / relative).read_text(encoding="utf-8").splitlines()
    blocks = load_blocks(profiles, relative)
    branches = value.get("branches") or []
    calls = value.get("calls") or []
    returns = value.get("returns") or []
    end = value["end"]["line"] + 1
    target = OUT / bundle["dir"]
    target.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source_ast, target / "ast.json")

    rows, test_rows, unentered, blockless = [], [], [], []
    for index, branch in enumerate(branches):
        line = branch["at"]["line"]
        upper = branches[index + 1]["at"]["line"] if index + 1 < len(branches) else end
        limit = max(upper, line + 1)
        window_calls = sorted({c.get("text") or "(unnamed)" for c in calls if line <= c["at"]["line"] < limit})
        window_returns = [f":{r['at']['line']}" for r in returns if line <= r["at"]["line"] < limit]
        if line not in blocks:
            entered = "—"
            blockless.append(branch["id"])
        elif blocks[line] > 0:
            entered = "예"
        else:
            entered = "아니오"
            unentered.append(branch["id"])
        condition = f"`:{line}` `{cell(lines[line - 1].strip())}`"
        rows.append(f"| {branch['id']} | {branch['kind']} | {condition} | "
                    f"{', '.join(f'`{cell(c)}`' for c in window_calls) or '—'} | "
                    f"{', '.join(window_returns) or '—'} | {entered} |")
        test = bundle["tests"].get(branch["id"], DEFAULT_TEST)
        test_rows.append(f"| {branch['id']} | {condition} | {entered} | {test} | no | no |")
    if not branches:
        test_rows.append("| B1 | branchless happy path | — | " + DEFAULT_TEST + " | no | no |")

    origin = ("> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 "
              "`[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 "
              f"{COVERAGE_COMMAND} 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — "
              "자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.")
    inputs = "\n".join(f"| {a} | {b} | {c} | {d} |" for a, b, c, d in bundle["inputs"])
    logic = f"""# Function Logic Map: `{qualified}`

- Source: `{relative}` (`{value['start']['line']}`–`{value['end']['line']}`)
- Qualified: `{qualified}`
- AST evidence: `ast.json` (`source_sha256` {value['source_sha256'][:16]}…)
- Risk scan: `risk-pattern-report.md`
- 분기 {len(branches)} · return {len(returns)} · 호출 {len(calls)}

**역할.** {bundle['role']}

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
{inputs}

## Branches and early returns

{origin}

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
{chr(10).join(rows) if rows else '| B1 | — | 분기 없음 | — | — | — |'}

## Calls and live bindings

{bundle['calls']}

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

{bundle['mutations']}

## Safety conclusion

- **Safe edit boundary**: {bundle['boundary']}
- **High-risk impact**: {bundle['high_risk']}
"""
    tests = f"""# Branch Test Map: `{qualified}`

- Source: `{relative}`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
{chr(10).join(test_rows)}

**미진입 분기 {len(unentered)}개**: {', '.join(unentered) or '없음'}
**자체 블록 없는 분기 {len(blockless)}개**: {', '.join(blockless) or '없음'} — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {{` 등)이며 미커버와 다르다.
"""
    (target / "function-logic-map.md").write_text(logic, encoding="utf-8")
    (target / "branch-test-map.md").write_text(tests, encoding="utf-8")
    subprocess.run([sys.executable, str(ROOT / "tools/logic-map/risk_pattern_report.py"), relative,
                    "--output", str(target / "risk-pattern-report.md")], cwd=ROOT, check=True)


def main() -> int:
    ast_dir, profiles = Path(sys.argv[1]), sys.argv[2:]
    for bundle in BUNDLES:
        render(bundle, ast_dir, profiles)
        print(f"rendered {bundle['dir']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
