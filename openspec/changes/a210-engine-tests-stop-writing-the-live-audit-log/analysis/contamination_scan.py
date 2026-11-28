# a210 D5 — 운영 audit 읽기 전용 스캔. 파일을 열기만 하고 쓰지 않음.
# 출력: 묶음마다 첫 시각 · 행 수 · 설정 이름 · 불리언 토글의 old→new 만(한도 · 금액 · 토픽 · URL 값은 출력하지 않음).
# 규칙(design F6): subject 가 호스트 사용자인 행이 1 초 안에 이어 붙은 묶음 중
#   (가) 기동 설정 이름(engine.*)이 2 개 이상이고 (나) 불리언 토글 하나 이상이 true → false 인 것.
import json, sys
from datetime import datetime

BOOL = {"true", "false", ""}
path, host_subject = sys.argv[1], sys.argv[2]
rows = []
with open(path, "rb") as f:
    for line in f:
        r = json.loads(line)
        r["_t"] = datetime.fromisoformat(r["at"].replace("Z", "+00:00"))
        rows.append(r)

bursts, cur = [], []
for r in rows:
    if r.get("subject") != host_subject:
        if cur: bursts.append(cur); cur = []
        continue
    if cur and (r["_t"] - cur[-1]["_t"]).total_seconds() > 1.0:
        bursts.append(cur); cur = []
    cur.append(r)
if cur: bursts.append(cur)

# 다른 subject(운영 컨테이너)가 처음 쓴 시각 — 그 전 묶음은 호스트 엔진이 진짜로 기동했을 수 있어 판정 불가로 가름
first_other = next((r["_t"] for r in rows if r.get("subject") != host_subject), None)

def restored(b):
    """묶음에서 true→false 된 토글을 다음 다른-subject 행이 false→true 로 되돌렸는지."""
    end = b[-1]["_t"]
    flips = {r["setting"] for r in b if r["old"] == "true" and r["new"] == "false"}
    seen = {}
    for r in rows:
        if r["_t"] <= end or r.get("subject") == host_subject or r["setting"] not in flips or r["setting"] in seen:
            continue
        seen[r["setting"]] = (r["old"] == "false" and r["new"] == "true")
    return f'{sum(seen.values())}/{len(flips)}'

hits = 0
for b in bursts:
    names = sorted({r["setting"] for r in b if r["setting"].startswith("engine.")})
    off = [r for r in b if r["old"] == "true" and r["new"] == "false"]
    if len(names) < 2 or not off:
        continue
    hits += 1
    toggles = sorted({f'{r["setting"]} {r["old"] or "∅"}→{r["new"] or "∅"}' for r in b
                      if r["old"] in BOOL and r["new"] in BOOL})
    era = "container" if first_other and b[0]["_t"] > first_other else "pre-container"
    print(f'{b[0]["at"]}\tera={era}\trestored={restored(b)}\trows={len(b)}\tsettings={",".join(names)}\ttoggles={"; ".join(toggles)}')
print(f"# host-subject bursts={len(bursts)} matching={hits} total_rows={len(rows)}")
