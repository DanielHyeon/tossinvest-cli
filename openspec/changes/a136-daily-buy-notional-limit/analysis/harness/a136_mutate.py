#!/usr/bin/env python3
"""a136 변이 하네스 — 공유 워킹트리를 건드리지 않고 사본(pid 접미사)에서 변이를 하나씩 적용해 시험이 실패하는지 잼.

규칙(기억: 「뮤테이션은 대상에 닿아야 한다」·「사본 하네스는 무변이 대조군이 GREEN 이 아니면 멈출 것」):
  1. 사본을 만들고 무변이 대조군이 GREEN 인지 먼저 확인 — 아니면 중단.
  2. 각 변이는 원문 문자열이 정확히 한 번 나와야 적용(닿음 단언).
  3. 변이 뒤 지정 시험이 실패하면 CAUGHT, 통과하면 SURVIVED. 매 변이 뒤 원본 파일을 되돌림.
사용: python3 a136_mutate.py <repo-root> [out.log]
"""
import os, shutil, subprocess, sys, tempfile

ROOT = os.path.abspath(sys.argv[1])
LOG = sys.argv[2] if len(sys.argv) > 2 else None
COPY = tempfile.mkdtemp(prefix=f"a136-mut-{os.getpid()}-")
for item in ("go.mod", "go.sum", "internal", "cmd", "schemas"):  # a137 의 미커밋 internal/dailyloss 는 아래에서 지움
    src = os.path.join(ROOT, item)
    (shutil.copytree if os.path.isdir(src) else shutil.copy2)(src, os.path.join(COPY, item))

shutil.rmtree(os.path.join(COPY, "internal", "dailyloss"), ignore_errors=True)  # 다른 change 의 미커밋 패키지 — 이 측정 밖
J = ("./internal/journal/", "TestA136")
E = ("./internal/execgw/", "TestA136")
A = ("./internal/app/engine/", "TestA136")
C = ("./internal/config/", "TestA136")
K = ("./internal/clock/", "TestA136")
R = ("./internal/risk/", "TestA136")
M = ("./internal/manualbuy/", "TestA136")
CON = ("./internal/console/", "TestA136")
CMD = ("./cmd/tossctl/", "TestA136")

MUTANTS = [
    ("M1 ≥ → >", "internal/journal/daily_buy.go", "if cmp >= 0 {", "if cmp > 0 {", [J]),
    ("M2 released rows never counted", "internal/journal/daily_buy.go", "if day != today {", "if day != today || true {", [J]),
    ("M3 Z3 ignores release reason/attempt shape", "internal/journal/daily_buy.go",
     "if releaseReason != ReleaseReasonBrokerTerminal ||", "if false && releaseReason != ReleaseReasonBrokerTerminal ||", [J]),
    ("M4 Z3 drops trading-day scope", "internal/journal/daily_buy.go",
     "AND f.trading_day = TRIM(i.trading_day) AND f.symbol", "AND f.symbol", [J]),
    ("M5 Z3 ignores lineage/events", "internal/journal/daily_buy.go", "return others == 0, nil", "return true, nil", [J]),
    ("M6 rollover offset 0", "internal/clock/account_day.go", "const AccountDayRollover = 8 * time.Hour",
     "const AccountDayRollover = 0 * time.Hour", [K, J]),
    ("M7 issue-day check removed", "internal/journal/daily_buy.go", "if issuedDay != today {", "if false {", [J]),
    ("M8 missing ceiling accepted", "internal/journal/daily_buy.go", "if req.DailyBuyLimit == nil {",
     "if req.DailyBuyLimit == nil {\n\t\treturn nil\n\t}\n\tif false {", [J]),
    ("M9 hook removed from reserveRows", "internal/journal/reservations.go",
     "if err := checkDailyBuyNotional(ctx, tx, req, plan, now); err != nil {", "if err := error(nil); err != nil {", [J]),
    ("M10 unreadable amount read as zero", "internal/journal/daily_buy.go",
     "amount, err := riskcalc.CanonicalDecimal(c.amount)\n", "amount, err := riskcalc.CanonicalDecimal(c.amount)\n\t\tif err != nil {\n\t\t\tamount, err = \"0\", nil\n\t\t}\n", [J]),
    ("M11 doors carry a huge ceiling", "internal/execgw/daily_buy.go",
     "Amount:    g.policy.MaxDailyBuyNotional.Amount,", "Amount:    \"999999999999\",", [E]),
    ("M12 interlock daily-buy check skipped", "internal/app/engine/interlock.go",
     "if err := checkGuardianDailyBuyLimit(facts.guardian, *status); err != nil {", "if err := error(nil); err != nil {", [A]),
    ("M13 default applied to every currency", "internal/config/daily_buy.go",
     "if strings.EqualFold(strings.TrimSpace(g.LimitCurrency), \"KRW\") {", "if strings.EqualFold(strings.TrimSpace(g.LimitCurrency), \"KRW\") || true {", [C, A]),
    ("M14 currency change keeps the old figure", "internal/config/limits_io.go",
     "} else if prev != next.Currency && next.DailyBuy == nil {", "} else if prev != next.Currency && next.DailyBuy == nil && false {", [C]),
    ("M15 digest ignores the ceiling", "internal/execgw/riskguardian.go",
     "sum := sha256.Sum256([]byte(g.limitsJSON + \"\\ndaily_buy_notional=\" +\n\t\tg.policy.MaxDailyBuyNotional.Amount + \" \" + g.policy.MaxDailyBuyNotional.Currency))",
     "sum := sha256.Sum256([]byte(g.limitsJSON))", [E]),
    ("M16 policy validation skips the ceiling", "internal/risk/input.go",
     "\t\t{\"maximum daily buy notional\", p.MaxDailyBuyNotional},\n", "", [R, A]),
    ("M17 Z1 accepts any release", "internal/journal/daily_buy.go",
     "return releaseReason == ReleaseReasonExpiredUnconsumed, nil", "return true, nil", [J]),
    ("M18 Z2 includes CONFIRMED", "internal/journal/daily_buy.go",
     "case StateNotDispatched, StateFailedConfirmed:", "case StateNotDispatched, StateFailedConfirmed, StateConfirmed:", [J]),
    # --- 1.5 수동 매수(D9=b) ---
    ("M19 manual total not added", "internal/journal/daily_buy.go",
     "if spent, err = riskcalc.AddDecimal(spent, manual); err != nil {", "if _, err = riskcalc.AddDecimal(spent, manual); err != nil {", [J, E]),
    ("M20 manual staleness ignored", "internal/journal/daily_buy.go", "if age > ManualBuyStaleness {", "if false {", [J]),
    ("M21 missing manual total accepted", "internal/journal/daily_buy.go", "if manual == nil {",
     "if manual == nil {\n\t\treturn \"0\", nil\n\t}\n\tif false {", [J]),
    ("M22 manual day ignored", "internal/journal/daily_buy.go", "if strings.TrimSpace(manual.Day) != today {", "if false {", [J]),
    ("M23 foreign manual buys skipped without authority", "internal/execgw/manual_buy.go",
     "authority, ok := fx[market]\n\t\tif !ok {", "authority, ok := fx[market]\n\t\tif !ok {\n\t\t\tcontinue\n\t\t}\n\t\tif false {", [E]),
    ("M24 engine orders counted as manual", "internal/manualbuy/summary.go", "if engineOrders[id] {", "if false && engineOrders[id] {", [M, A]),
    ("M25 pending remainder dropped", "internal/manualbuy/summary.go", "if remaining.Sign() > 0 {", "if false && remaining.Sign() > 0 {", [M, A]),
    ("M26 failed read published as known", "internal/app/engine/manual_buy_refresh.go", "snapshot.Err = err\n", "snapshot.Err = nil\n", [A]),
    ("M27 no board accepted", "internal/execgw/manual_buy.go", "if g.manualBuys == nil {", "if g.manualBuys == nil {\n\t\treturn &journal.ManualBuyTotal{Amount: \"0\", AsOf: now, Day: \"\"}, nil\n\t}\n\tif false {", [E]),
    # --- 1.3 콘솔 칸 ---
    ("M28 every save writes the daily key", "internal/config/limits_io.go", "\tif l.DailyBuy != nil {\n\t\t// a136",
     "\tif l.DailyBuy == nil {\n\t\tzero := 0.0\n\t\tl.DailyBuy = &zero\n\t}\n\tif l.DailyBuy != nil {\n\t\t// a136", [C, CON, CMD]),
    ("M29 negative daily ceiling accepted", "internal/config/limits.go", "*l.DailyBuy < 0)", "*l.DailyBuy < -1e18)", [C]),
    ("M30 USD unset says nothing", "internal/console/settings_limits.go",
     "\" 계좌에는 기본값이 없어 엔진이 기동을 거부한다. 값을 넣어야 한다\"", "\" 계좌\"", [CON]),
    ("M31 form drops the daily field", "internal/console/settings_limits.go", "next.DailyBuy = &dailyBuy", "next.DailyBuy = nil", [CON]),
    ("M32 explicit new-currency value cleared", "internal/config/limits_io.go",
     "} else if prev != next.Currency && next.DailyBuy == nil {", "} else if prev != next.Currency {", [C]),
]

def run(targets, tags):
    for pkg, pattern in targets:
        cmd = ["go", "test", "-count=1", "-run", pattern, pkg]
        if tags:
            cmd[2:2] = ["-tags", "tossos_testseams"]
        env = dict(os.environ, GOFLAGS="-trimpath")
        r = subprocess.run(cmd, cwd=COPY, capture_output=True, text=True, env=env, timeout=900)
        if r.returncode != 0:
            text = (r.stdout + r.stderr)
            if "[build failed]" in text or "[setup failed]" in text:
                # 컴파일 실패는 시험이 잡은 것이 아님 — 별도 표지(round1 에서 M2·M13·M14 가 이렇게 「CAUGHT」 로 찍혔음).
                return None, pkg + ": BUILD-FAILED"
            fails = [l for l in text.splitlines() if l.startswith("--- FAIL")]
            return False, pkg + ": " + (", ".join(fails[:3]) if fails else "FAIL")
    return True, ""

out = []
def say(line):
    print(line, flush=True)
    out.append(line)

ok, why = run([J, E, A, C, K, R, M, CON, CMD], True)
say(f"CONTROL (무변이) {'GREEN' if ok else 'RED — 중단: ' + why}")
if ok is not True:
    sys.exit(1)
ONLY = os.environ.get("A136_ONLY", "")  # 쉼표로 고른 변이만(예: "M14") — 재측정용
for name, rel, old, new, targets in MUTANTS:
    if ONLY and name.split()[0] not in ONLY.split(","):
        continue
    path = os.path.join(COPY, rel)
    original = open(path).read()
    if original.count(old) != 1:
        say(f"{name}: NOT-REACHED (원문 {original.count(old)}회)")
        continue
    open(path, "w").write(original.replace(old, new))
    try:
        passed, why = run(targets, True)
    finally:
        open(path, "w").write(original)
    verdict = "SURVIVED" if passed else ("NOT-COMPILED — " + why if passed is None else "CAUGHT — " + why)
    say(f"{name}: {verdict}")
shutil.rmtree(COPY)
if LOG:
    open(LOG, "w").write("\n".join(out) + "\n")
