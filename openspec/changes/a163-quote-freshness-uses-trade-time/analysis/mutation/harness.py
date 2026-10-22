#!/usr/bin/env python3
"""a163 tasks 4.1 — 변이 원장 하네스.

저장소의 **추적 파일 사본**(scratch, pid 접미사)에서만 변이를 걸고, 무변이 대조군이 GREEN 인지 먼저 잰 뒤 변이마다
  ① 원문이 정확히 한 번 있는지 단언(닿지 않는 변이 금지) → ② 치환 → ③ 지정 시험 실행 → ④ 원복 + sha 대조
를 함. 원본 worktree 는 읽기만 함(git ls-files 로 목록, 바이트 복사). 결과는 같은 디렉터리의 ledger.json · ledger.md.

사용(저장소 루트에서):
    python3 -I openspec/changes/a163-quote-freshness-uses-trade-time/analysis/mutation/harness.py <scratch-parent>
"""

from __future__ import annotations

import hashlib
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]

ENGINE = "./internal/app/engine/"
OFFICIAL = "./internal/official/"
CONFIG = "./internal/config/"
OUTPUT = "./internal/output/"
CMD = "./cmd/tossctl/"

EXIT = "internal/app/engine/exitloop.go"
AGE = "internal/app/engine/exit_source_age.go"
MR = "internal/official/market_reads.go"

# (id, 설명, 파일, 원문, 치환, [(패키지, -run 정규식)]) — 원문은 파일에서 정확히 한 번이어야 함.
MUTATIONS = [
    ("①", "보고 사유(시각 없음)가 있으면 판정을 건너뜀", EXIT,
     "\t\tcycle.Judged++\n",
     "\t\tif quote.SourceTimestamp == \"\" {\n\t\t\to.noteUnobservedCause(&cycle, state.position.ID, unobservedNoQuote)\n\t\t\tcontinue\n\t\t}\n\t\tcycle.Judged++\n",
     [(ENGINE, "TestA163AMissingOrMalformedBrokerTimeJudgesTheSame|TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows")]),
    ("②", "보고를 판정 순회 안(판정 직후)으로 이동", EXIT,
     "\t\t\tcycle.Err = err\n\t\t}\n\t}\n\t// a163: 체결 나이 보고도 모든 판정 · 제출 뒤 한 번 — a090 집계를 비우는 settleUnobserved 앞이어야 판정된 포지션을 읽음.\n\t// 토글 판단 · 패닉 격리는 그 함수 안(exit_source_age.go) — 이 줄은 토글과 무관하게 무조건임.\n\to.reportSourceAge(ctx, &cycle, states, quotes)\n",
     "\t\t\tcycle.Err = err\n\t\t}\n\t\to.reportSourceAge(ctx, &cycle, states, quotes)\n\t}\n",
     [(ENGINE, "TestA163TheReportComesAfterEveryJudgementAndBeforeTheUnobservedSettlement")]),
    ("③", "어댑터가 빈 시각을 받은 시각 문자열로 메움", MR,
     "\t\t\tSourceTimestamp: p.Timestamp,\n",
     "\t\t\tSourceTimestamp: func() string {\n\t\t\t\tif p.Timestamp == \"\" {\n\t\t\t\t\treturn time.Now().UTC().Format(time.RFC3339Nano)\n\t\t\t\t}\n\t\t\t\treturn p.Timestamp\n\t\t\t}(),\n",
     [(OFFICIAL, "TestA163AdaptPricesKeepsTheBrokerTimeVerbatim|TestA163PricesCarriesTheBrokerTimeInOneRequest")]),
    ("④a", "어댑터가 FetchedAt 을 브로커 시각으로", MR,
     "\t\t\tFetchedAt: time.Now().UTC(),\n\t\t\t// a163:",
     "\t\t\tFetchedAt: func() time.Time {\n\t\t\t\tif t, err := time.Parse(time.RFC3339Nano, p.Timestamp); err == nil {\n\t\t\t\t\treturn t.UTC()\n\t\t\t\t}\n\t\t\t\treturn time.Now().UTC()\n\t\t\t}(),\n\t\t\t// a163:",
     [(OFFICIAL, "TestA163AdaptPricesKeepsTheBrokerTimeVerbatim")]),
    ("④b", "observe 리터럴의 FetchedAt 에 브로커 시각 대입", EXIT,
     "next[symbol] = observedQuote{Price: decimalOf(q.Last), FetchedAt: effectiveAt,",
     "next[symbol] = observedQuote{Price: decimalOf(q.Last), FetchedAt: func() time.Time {\n\t\t\t\tif t, err := time.Parse(time.RFC3339Nano, q.SourceTimestamp); err == nil {\n\t\t\t\t\treturn t.UTC()\n\t\t\t\t}\n\t\t\t\treturn effectiveAt\n\t\t\t}(),",
     [(ENGINE, "TestA163IdentityAndObservedAtComeFromTheReceiptNotTheBrokerTime|TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows")]),
    ("⑤", "래치 제거(이미 알린 포지션도 매 주기 알림)", AGE,
     "\t\tif rec.alerted || alerts >= sourceAgeAlertsPerCycle {\n",
     "\t\tif alerts >= sourceAgeAlertsPerCycle {\n",
     [(ENGINE, "TestA163AMissingOrUnparseableTimeIsAlertedOnceAndRearmsWhenATimeArrives")]),
    ("⑥", "보고 등급을 critical 로(등급표 등재)", "internal/obs/event.go",
     "\tEventExitStopSoldNothing: true,\n}",
     "\tEventExitStopSoldNothing: true,\n\tEventExitSourceMissing:   true,\n}",
     [(ENGINE, "TestA163TheReportNeverBlocksEntryOrEscalates|TestA163AMissingOrUnparseableTimeIsAlertedOnceAndRearmsWhenATimeArrives")]),
    ("⑦", "어댑터가 빈 · 깨진 timestamp 행을 버림", MR,
     "\tfor _, p := range raw {\n\t\tout = append(out, domain.Quote{\n\t\t\tSymbol:    p.Symbol,\n\t\t\tLast:      parseDecimal(p.LastPrice),",
     "\tfor _, p := range raw {\n\t\tif _, err := time.Parse(time.RFC3339Nano, p.Timestamp); err != nil {\n\t\t\tcontinue\n\t\t}\n\t\tout = append(out, domain.Quote{\n\t\t\tSymbol:    p.Symbol,\n\t\t\tLast:      parseDecimal(p.LastPrice),",
     [(ENGINE, "TestA163AMissingOrMalformedBrokerTimeJudgesTheSame")]),
    ("⑧", "어댑터가 빈 · 깨진 timestamp 에 오류 반환(Prices)", MR,
     "\treturn adaptPrices(raw), nil\n",
     "\tfor _, p := range raw {\n\t\tif _, err := time.Parse(time.RFC3339Nano, p.Timestamp); err != nil {\n\t\t\treturn nil, err\n\t\t}\n\t}\n\treturn adaptPrices(raw), nil\n",
     [(ENGINE, "TestA163AMissingOrMalformedBrokerTimeJudgesTheSame")]),
    ("⑨", "observe B7 필터의 effectiveAt 을 브로커 시각 기준으로", EXIT,
     "\t\t\teffectiveAt, source := q.FetchedAt.UTC(), \"quote_fetched_at\"\n",
     "\t\t\teffectiveAt, source := q.FetchedAt.UTC(), \"quote_fetched_at\"\n\t\t\tif t, err := time.Parse(time.RFC3339Nano, q.SourceTimestamp); err == nil {\n\t\t\t\teffectiveAt = t.UTC()\n\t\t\t}\n",
     [(ENGINE, "TestA163AStaleBrokerTimeStillLiquidatesInTheSameCycle|TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows")]),
    ("⑩", "snapshotContext 가 stableObservationID 에 브로커 시각을 넘김", EXIT,
     "\tobservationID, err := stableObservationID(o.opts.AccountRef, m, quote, fallback)\n",
     "\tif t, perr := time.Parse(time.RFC3339Nano, quote.SourceTimestamp); perr == nil {\n\t\tquote.FetchedAt = t\n\t}\n\tobservationID, err := stableObservationID(o.opts.AccountRef, m, quote, fallback)\n",
     [(ENGINE, "TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange|TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows")]),
    ("⑪", "보고기 recover 경계 제거", AGE,
     "\tdefer func() {\n\t\tif recover() != nil {\n\t\t\to.sourceAgeRecovered()\n\t\t\treturn\n\t\t}\n\t\to.sourceAgePanicLogged = false // 패닉 없이 끝난 보고 — 다음 패닉 연속의 첫 줄을 다시 쓰게 함\n\t}()\n\tif !o.opts.SourceAgeReport {",
     "\tif !o.opts.SourceAgeReport {",
     [(ENGINE, "TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops")]),
    ("⑫", "engineRuntime 의 SourceAgeLog 배선 삭제", "cmd/tossctl/engine.go",
     "\t\tSourceAgeLog: logger,\n",
     "",
     [(CMD, "TestA163EngineRuntimeWiresTheSourceAgeLogger")]),
    ("⑬", "JSON 태그에서 omitempty 삭제", "internal/domain/models.go",
     "`json:\"source_timestamp,omitempty\"`",
     "`json:\"source_timestamp\"`",
     [(OUTPUT, "TestA163TheWTSQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime"),
      (OFFICIAL, "TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime")]),
    ("⑭", "전이 대신 매 주기 줄", AGE,
     "\t\tif rec.bucket != bucket {\n",
     "\t\tif true {\n",
     [(ENGINE, "TestA163AnAgeLineIsWrittenOnlyWhenTheBucketChanges")]),
    ("⑮", "held 기준 정리 삭제", AGE,
     "\tfor id := range o.sourceAge {\n\t\tif _, ok := tally.held[id]; !ok {\n\t\t\tdelete(o.sourceAge, id)\n\t\t}\n\t}\n",
     "",
     [(ENGINE, "TestA163AClosedPositionLeavesTheLatch")]),
    ("⑯", "주기 상한 삭제", AGE,
     "\t\tif rec.alerted || alerts >= sourceAgeAlertsPerCycle {\n",
     "\t\tif rec.alerted {\n",
     [(ENGINE, "TestA163AtMostEightSourceAlertsLeaveInOneCycle")]),
    ("⑰", "토글 무시(꺼짐이면 반환 삭제)", AGE,
     "\tif !o.opts.SourceAgeReport {\n\t\treturn // 꺼짐 = a163 이전 동작(줄 · 알림 · 래치 0)\n\t}\n",
     "",
     [(ENGINE, "TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn")]),
    ("⑱a", "기본값 ON — mergeEngine 이 키 없음을 true 로", "internal/config/engine.go",
     "\tif raw.ExitSourceAgeReport != nil {\n",
     "\tcfg.ExitSourceAgeReport = true\n\tif raw.ExitSourceAgeReport != nil {\n",
     [(CONFIG, "TestA163TheSourceAgeReportToggleDefaultsOff")]),
    ("⑱b", "기본값 ON — 진입 함수 판단을 뒤집음", AGE,
     "\tif !o.opts.SourceAgeReport {\n\t\treturn",
     "\tif o.opts.SourceAgeReport {\n\t\treturn",
     [(ENGINE, "TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn")]),
    ("⑲", "merge 를 automation_gate 조기 반환 뒤로 이동", "internal/config/engine.go",
     "\tif raw.ExitSourceAgeReport != nil {\n\t\tcfg.ExitSourceAgeReport = *raw.ExitSourceAgeReport\n\t}\n\tif raw.AutomationGate == nil {\n\t\treturn\n\t}\n",
     "\tif raw.AutomationGate == nil {\n\t\treturn\n\t}\n\tif raw.ExitSourceAgeReport != nil {\n\t\tcfg.ExitSourceAgeReport = *raw.ExitSourceAgeReport\n\t}\n",
     [(CONFIG, "TestA163TheSourceAgeReportToggleDefaultsOff")]),
    ("⑳", "exitwiring 덮어쓰기 삭제", "internal/app/engine/exitwiring.go",
     "\topts.SourceAgeReport = c.Config.Engine.ExitSourceAgeReport\n",
     "",
     [(ENGINE, "TestA163TheProductionAssemblyReadsTheLoadedToggle")]),
    ("㉑a", "audit 행 삭제", "internal/app/engine/interlock.go",
     "\t\t{audit.ActionExitSourceAgeReport, \"engine.exit_source_age_report\", strconv.FormatBool(sourceAgeReport), \"\"},\n",
     "",
     [(ENGINE, "TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten|TestA163AStartupAuditsTheLoadedReportToggle")]),
    ("㉑b", "audit 행 값 고정(false)", "internal/app/engine/interlock.go",
     "strconv.FormatBool(sourceAgeReport), \"\"},\n",
     "strconv.FormatBool(false), \"\"},\n",
     [(ENGINE, "TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten|TestA163AStartupAuditsTheLoadedReportToggle")]),
    ("㉑c", "NewContext 가 로드된 값 대신 false 를 넘김", "internal/app/engine/engine.go",
     "status.AttestationFile, cfg.Engine.ExitSourceAgeReport); err != nil {",
     "status.AttestationFile, false); err != nil {",
     [(ENGINE, "TestA163AStartupAuditsTheLoadedReportToggle")]),
    ("㉒", "토글 판단을 ObserveOnce 판정 순회 앞으로 — 꺼짐이면 판정 건너뜀", EXIT,
     "\tfor _, state := range states {\n\t\tquote, ok := quotes[",
     "\tif !o.opts.SourceAgeReport {\n\t\to.settleUnobserved(ctx, &cycle)\n\t\treturn cycle\n\t}\n\tfor _, state := range states {\n\t\tquote, ok := quotes[",
     [(ENGINE, "TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn|TestA163TheReportToggleIsReadOnlyWhereTheDesignAllows|TestA163AStaleBrokerTimeStillLiquidatesInTheSameCycle")]),
    # --- gstack 리뷰(2026-10-22) 처분으로 더한 변이 ---
    ("㉓", "거둔 패닉의 고정 문구 줄 삭제(흔적 없는 패닉)", AGE,
     "\t\t\to.sourceAgeRecovered()\n",
     "",
     [(ENGINE, "TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops")]),
    ("㉔", "고정 문구 줄 래치 삭제(매 주기 같은 줄)", AGE,
     "\tif o.sourceAgePanicLogged {\n\t\treturn\n\t}\n",
     "",
     [(ENGINE, "TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops")]),
    ("㉕", "미래(ahead) 구간도 알림 사유로", AGE,
     "\treturn bucket == sourceAgeMissing || bucket == sourceAgeUnparseable\n",
     "\treturn bucket == sourceAgeMissing || bucket == sourceAgeUnparseable || bucket == sourceAgeAhead\n",
     [(ENGINE, "TestA163EveryBucketIsClassifiedFromTheReceiptTime")]),
    ("㉖", "원문 정화 삭제(바이트 자르기 · 제어문자 통과)", AGE,
     "\treturn strings.Map(func(r rune) rune {",
     "\tif len(raw) >= 0 {\n\t\treturn raw\n\t}\n\treturn strings.Map(func(r rune) rune {",
     [(ENGINE, "TestA163TheRawBrokerTimeIsBoundedPrintableAndValidUTF8")]),
    ("㉗", "비상 청산(flatten) 경로가 브로커 시각을 읽음", "internal/flatten/liquidate.go",
     "\t\t\tif !strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) || q.Last <= 0 {\n",
     "\t\t\tif !strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) || q.Last <= 0 || q.SourceTimestamp == \"\" {\n",
     [(ENGINE, "TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows")]),
    ("㉘", "거대 원문 해독 상한 삭제(매 주기 끝까지 훑음)", AGE,
     "\tif len(raw) > sourceAgeRawLimit {\n\t\treturn sourceAgeUnparseable, 0, false\n\t}\n",
     "",
     [(ENGINE, "TestA163TheRawBrokerTimeIsBoundedPrintableAndValidUTF8")]),
]


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run_tests(copy: Path, pkg: str, pattern: str) -> tuple[int, str]:
    env = dict(os.environ, GOFLAGS="-trimpath")
    proc = subprocess.run(["go", "test", "-count=1", "-timeout", "4m", "-run", pattern, pkg],
                          cwd=copy, env=env, capture_output=True, text=True)
    out = proc.stdout + proc.stderr
    failed = sorted({line.split()[2] for line in out.splitlines() if line.lstrip().startswith("--- FAIL:")})
    panics = [line for line in out.splitlines() if line.startswith("panic:")]
    builds = [line for line in out.splitlines() if "[build failed]" in line or "[setup failed]" in line]
    if builds:
        summary = "BUILD FAILED: " + builds[0]  # 컴파일 실패는 시험이 잡은 것이 아님 — 판정에서 뺌
    elif failed:
        summary = ",".join(failed)
    elif panics:
        # go test 는 패닉한 시험 이름을 "--- FAIL" 없이 traceback 에 남김 — 실행 중이던 시험은 -run 묶음 안의 것
        summary = "PANIC in -run " + pattern + ": " + panics[0]
    else:
        summary = ("rc!=0: " + out.strip().splitlines()[-1]) if proc.returncode else ""
    return proc.returncode, summary


def main() -> int:
    parent = Path(sys.argv[1]).resolve()
    copy = parent / f"a163-mutation-{os.getpid()}"
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()
    dirty = subprocess.run(["git", "-C", str(ROOT), "status", "--porcelain", "--", "cmd", "internal", "schemas"],
                           capture_output=True, text=True, check=True).stdout
    if dirty.strip():
        print("refusing: tracked Go sources are dirty — the copy must equal HEAD", file=sys.stderr)
        return 2
    files = subprocess.run(["git", "-C", str(ROOT), "ls-files", "-z"], capture_output=True, check=True).stdout.split(b"\0")
    for name in filter(None, files):
        src = ROOT / name.decode()
        if src.is_file():
            dst = copy / name.decode()
            dst.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(src, dst)
    results = {"head": head, "copy": str(copy), "control": [], "mutations": []}

    # 무변이 대조군: 변이가 쓰는 시험 집합 전부가 사본에서 GREEN 이어야 함(사본 · 계측기 눈멂 차단).
    pairs = sorted({(pkg, pat) for mid, *_, tests in MUTATIONS for pkg, pat in tests
                    if not sys.argv[2:] or mid in sys.argv[2:]})
    for pkg, pat in pairs:
        rc, summary = run_tests(copy, pkg, pat)
        results["control"].append({"pkg": pkg, "run": pat, "rc": rc, "failed": summary})
        print(f"control {pkg} {pat}: rc={rc} {summary}", flush=True)
        if rc != 0:
            print("control is not GREEN — stopping", file=sys.stderr)
            json.dump(results, open(HERE / "ledger.json", "w"), ensure_ascii=False, indent=1)
            return 1

    only = set(sys.argv[2:])
    for mid, why, rel, old, new, tests in MUTATIONS:
        if only and mid not in only:
            continue
        path = copy / rel
        before = path.read_bytes()
        text = before.decode()
        count = text.count(old)
        entry = {"id": mid, "why": why, "file": rel, "reached": count == 1, "runs": []}
        if count != 1:
            entry["verdict"] = f"NOT REACHED (anchor count {count})"
            results["mutations"].append(entry)
            print(f"{mid}: NOT REACHED ({count})", flush=True)
            continue
        path.write_text(text.replace(old, new))
        caught = False
        for pkg, pat in tests:
            rc, summary = run_tests(copy, pkg, pat)
            entry["runs"].append({"pkg": pkg, "run": pat, "rc": rc, "failed": summary})
            caught = caught or (rc != 0 and not summary.startswith("BUILD FAILED"))
        path.write_bytes(before)
        if hashlib.sha256(path.read_bytes()).hexdigest() != hashlib.sha256(before).hexdigest():
            print(f"{mid}: revert failed", file=sys.stderr)
            return 3
        entry["verdict"] = "CAUGHT" if caught else "SURVIVED"
        results["mutations"].append(entry)
        print(f"{mid}: {entry['verdict']} {[r['failed'] for r in entry['runs']]}", flush=True)

    name = "ledger" if not sys.argv[2:] else "ledger-partial"
    json.dump(results, open(HERE / f"{name}.json", "w"), ensure_ascii=False, indent=1)
    caught = sum(1 for m in results["mutations"] if m["verdict"] == "CAUGHT")
    lines = [
        "# a163 변이 원장 (tasks 4.1)", "",
        f"- HEAD `{head[:9]}` 의 추적 파일 사본(`{copy.name}`)에서 실행. 원본 worktree 무편집. 생성: `harness.py` (손편집 금지 — 다시 돌릴 것)",
        f"- 무변이 대조군: 시험 묶음 {len(results['control'])} 개 전부 rc=0",
        f"- 결과: **{caught} CAUGHT / {len(results['mutations'])}**", "",
        "| # | 변이 | 파일 | 판정 | 실패한 시험 |", "| --- | --- | --- | --- | --- |",
    ]
    for m in results["mutations"]:
        failed = " · ".join(r["failed"] for r in m["runs"] if r["failed"]) or "—"
        lines.append(f"| {m['id']} | {m['why']} | `{m['file']}` | {m['verdict']} | {failed} |")
    (HERE / f"{name}.md").write_text("\n".join(lines) + "\n")
    shutil.rmtree(copy)
    return 0 if caught == len(results["mutations"]) else 4


if __name__ == "__main__":
    raise SystemExit(main())
