#!/usr/bin/env python3
"""a140 4.1 보고서 보조 집계 — 원장만 읽음(쓰기 0 · API 0). 종목 · 가격 · setup ID 는 출력하지 않음.

도구 보고서(`go run ./tools/a140-breakout-signal-shadow report`)가 내지 않는 셋을 셈:
세션 유효성(측정 규칙 대조) · 호출 예산 실측(D11 대조) · 결과 분포(판정 아님, n<30).

    python3 -I openspec/changes/a140-breakout-signals-are-measured-in-shadow/analysis/report_supplement.py
"""

import collections
import hashlib
import json
import pathlib
import statistics
import sys

LEDGER = pathlib.Path(__file__).resolve().parent / "shadow-ledger" / "ledger.jsonl"


def main() -> int:
    raw = LEDGER.read_bytes()
    rows = [json.loads(line) for line in raw.decode().splitlines()]
    print(f"ledger sha256 {hashlib.sha256(raw).hexdigest()} · rows {len(rows)} · bytes {len(raw)}")
    sessions = sorted({r["session"] for r in rows})
    lots = {r["lot"] for r in rows}
    print(f"sessions {len(sessions)} · lots {len(lots)}")

    # 세션 유효성 — design §2.1 · §5 의 규칙에서 뽑은 조건(달력 출처 · 한 판 · 종목 30 · 세션 끝 요약 · 제안 전부 라벨 · MINUTE_GAP 0)
    print("\n## 세션 유효성")
    for s in sessions:
        rs = [r for r in rows if r["session"] == s]
        kinds = collections.Counter(r["kind"] for r in rs)
        uni = [r for r in rs if r["kind"] == "universe"]
        proposed = {r["setup_id"] for r in rs if r["kind"] == "signal" and r.get("phase") == "PROPOSED"}
        labeled = {r["setup_id"] for r in rs if r["kind"] == "label"}
        minute_gap = sum(1 for r in rs if r["kind"] == "gap" and "MINUTE_GAP" in r.get("reason", ""))
        checks = {
            "calendar": all(u.get("session_source") == "calendar" for u in uni) and len(uni) == 1,
            "one_lot": len({r["lot"] for r in rs}) == 1,
            "universe30": len(uni) == 1 and len(uni[0].get("symbols", [])) == 30,
            "session_summary": kinds["session"] == 1,
            "all_proposals_labeled": proposed == labeled,
            "no_minute_gap": minute_gap == 0,
        }
        print(f"- {s}: {'VALID' if all(checks.values()) else 'INVALID'} · " + " · ".join(f"{k}={v}" for k, v in checks.items())
              + f" · first {rs[0]['at']} · last {rs[-1]['at']} · proposed {len(proposed)} · labeled {len(labeled)}")

    # 호출 예산 실측 — budget 행(10 분 창)의 그룹별 GET 합과 수집 시간
    print("\n## 호출 예산 실측")
    for s in sessions:
        b = [r for r in rows if r["session"] == s and r["kind"] == "budget"]
        cs = [r.get("collect_seconds", 0.0) for r in b]
        chart = sum(r.get("reads", {}).get("MARKET_DATA_CHART", 0) for r in b)
        md = sum(r.get("reads", {}).get("MARKET_DATA", 0) for r in b)
        print(f"- {s}: 10분 창 {len(b)} · CHART GET {chart} · MARKET_DATA GET {md} · 수집 시간 중앙 {statistics.median(cs):.1f}s"
              f" · 최대 {max(cs):.1f}s · 300s(창의 절반) 초과 창 {sum(1 for c in cs if c > 300)}")

    # 결과 분포 — 판정 아님(spec: n<30 판정 보류)
    print("\n## 결과 분포 (판정 아님)")
    labels = [r for r in rows if r["kind"] == "label"]
    print(f"- 라벨 n={len(labels)} · " + " · ".join(f"{k} {v}" for k, v in sorted(collections.Counter(r['outcome'] for r in labels).items())))
    buckets = collections.Counter()
    for r in labels:
        x = r.get("r_ppm", 0) / 1e6
        buckets["R<=-0.5" if x <= -0.5 else "-0.5<R<0" if x < 0 else "0<=R<0.5" if x < 0.5 else "R>=0.5"] += 1
    print("- R 구간: " + " · ".join(f"{k} {buckets[k]}" for k in ("R<=-0.5", "-0.5<R<0", "0<=R<0.5", "R>=0.5")))

    # RVOL 반사실 — 종목-세션마다 「어느 행에서든」 깃발이 선 적 있음(도구 보고서는 처음 깃발이 선 행 기준이라 1 씩 다를 수 있음)
    rank = {"RANGE_LOCKED": 1, "BREAKOUT_CLOSED": 2, "RETEST_WAIT": 3, "ARMED": 4, "PROPOSED": 5}
    furthest, flags = {}, collections.defaultdict(lambda: [False, False, False])
    for r in rows:
        if r["kind"] != "signal":
            continue
        key = (r["session"], r["symbol"])
        if rank.get(r["phase"], 0) > rank.get(furthest.get(key), 0):
            furthest[key] = r["phase"]
        for i, name in enumerate(("rvol_1_2x", "rvol_2_0x", "rvol_2_5x")):
            flags[key][i] = flags[key][i] or bool(r.get(name))
    print("\n## RVOL 반사실(종목-세션, 어느 행에서든)")
    breakout = [k for k, ph in furthest.items() if rank[ph] >= 2]
    proposed_keys = [k for k, ph in furthest.items() if ph == "PROPOSED"]
    near_miss = sum(1 for k, ph in furthest.items() if rank[ph] < 2 and flags[k][0])
    print(f"- 종목-세션 {len(furthest)} · 돌파 {len(breakout)} · 돌파 못 한 것 중 1.2× 근접 {near_miss}")
    for i, label in enumerate(("1.2x", "2.0x", "2.5x")):
        print(f"- {label}: 돌파 {sum(flags[k][i] for k in breakout)} · 그 중 PROPOSED {sum(flags[k][i] for k in proposed_keys)}")

    # 거절 분포 — ARMED 시도 행과 그 밖(행 기준)
    print("\n## ARMED 시도 결과(시도 행 기준)")
    att = collections.Counter((r.get("refusal") or r.get("phase")) for r in rows if r["kind"] == "signal" and r.get("attempt"))
    print("- " + " · ".join(f"{k} {v}" for k, v in sorted(att.items())))
    gaps = collections.Counter((r["stage"], r["reason"].split(":")[1].strip()) for r in rows if r["kind"] == "gap")
    print("\n## 결손 행(단계 · 코드)")
    for (stage, code), n in sorted(gaps.items()):
        print(f"- {stage} {code} {n}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
