# a140 원장(커밋된 archive 파일)에서 호가 관문 가정(design D3)의 근거 수치를 다시 잼 — 읽기 전용, 네트워크 0.
# 실행: python3 -I openspec/changes/a203-breakout-kr-is-replayed-on-past-sessions/analysis/a140-ledger-quote-stats.py
# 출력에는 종목 코드 · 금액을 넣지 않음(틱 · 분 · 건수 · 비율만).
import collections
import datetime
import hashlib
import json
import pathlib
import statistics

ROOT = pathlib.Path(__file__).resolve().parents[4]
LEDGER = ROOT / "openspec/changes/archive/2026-10-22-a140-breakout-signals-are-measured-in-shadow/analysis/shadow-ledger/ledger.jsonl"


def ms(text):
    return int(datetime.datetime.fromisoformat(text.replace("Z", "+00:00")).timestamp() * 1000)


raw = LEDGER.read_bytes()
rows = [json.loads(line) for line in raw.splitlines() if line.strip()]
print(f"ledger sha256 {hashlib.sha256(raw).hexdigest()} · rows {len(rows)}")

attempts = [r for r in rows if r["kind"] == "signal" and r.get("attempt")]
print("ARMED 시도 행(단계, 거절):", sorted(collections.Counter((r["phase"], r.get("refusal") or "-") for r in attempts).items()))

setups = collections.defaultdict(list)
for r in attempts:
    setups[r["setup_id"]].append(r)
proposed = [r for r in rows if r["kind"] == "signal" and r.get("phase") == "PROPOSED"]
print(f"ARMED setup {len(setups)} · 그중 PROPOSED {len(proposed)}")
print("PROPOSED 의 시도 번호:", sorted(collections.Counter(r["attempt"] for r in proposed).items()))

# 호가 ask 와 진입 제안가(ARMED 봉 종가)의 차이 — 틱 단위
print("ask − 진입가(틱):", sorted((r["ask"] - r["entry"]) // r["tick"] if r["ask"] >= r["entry"] else -((r["entry"] - r["ask"]) // r["tick"]) for r in proposed))
print("스프레드(틱):", sorted(collections.Counter((r["ask"] - r["bid"]) // r["tick"] for r in proposed).items()))
print("worst == 진입가:", sum(1 for r in proposed if r["worst"] == r["entry"]), "/", len(proposed))
gap = [(r["worst"] - r["entry"]) / r["entry"] * 100 for r in proposed]
print(f"(worst − 진입가)/진입가 %: 중앙 {statistics.median(gap):.3f} · 평균 {statistics.mean(gap):.3f} · 최대 {max(gap):.3f}")

# 시각: ARMED 봉 여는 시각 → 제안 시각, → 라벨 시작 봉(분, 내림)
to_prop, to_label = collections.Counter(), collections.Counter()
for r in proposed:
    open_ms = ms(r["session"].split(":")[1] + "T00:00:00Z")  # KRX 정규장 개장 09:00 KST = 00:00Z
    armed_open = open_ms + r["armed_index"] * 60_000
    to_prop[(ms(r["at"]) - armed_open) // 60_000] += 1
    to_label[(r["label_from_ms"] - armed_open) // 60_000] += 1
print("ARMED 봉 여는 시각 → 제안(분):", sorted(to_prop.items()))
print("ARMED 봉 여는 시각 → 라벨 시작 봉(분):", sorted(to_label.items()))
