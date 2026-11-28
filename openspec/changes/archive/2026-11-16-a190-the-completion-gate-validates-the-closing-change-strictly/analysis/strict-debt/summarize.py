"""strict-all.json(실측 원문)에서 실패 항목·경고 종류 목록을 만든다. 판정에 쓰지 않는 기록용."""
import collections, json, re, sys

KINDS = (
    ("요구 문장 500자 초과", re.compile(r"Requirement text is very long")),
    ("Purpose 자리표시", re.compile(r"Purpose section is still a placeholder")),
    ("Purpose 너무 짧음", re.compile(r"Purpose section is too brief")),
    ("첫 ## 위 본문(메인 스펙에 안 옮겨짐)", re.compile(r'Text above the first "## " section')),
)

def kind(message):
    for name, pattern in KINDS:
        if pattern.search(message):
            return name
    return "기타: " + message[:80]

data = json.load(open(sys.argv[1], encoding="utf-8"))
totals = data["summary"]["totals"]
print(f"# 원문: {sys.argv[1]} · totals {totals}")
levels = collections.Counter()
kinds = collections.Counter()
failed = [item for item in data["items"] if not item["valid"]]
print(f"# 실패 항목 {len(failed)}")
print("| # | type | id | 경고 종류(건수) |")
print("|---|---|---|---|")
for n, item in enumerate(sorted(failed, key=lambda i: (i["type"], i["id"])), 1):
    per = collections.Counter()
    for issue in item["issues"]:
        levels[issue["level"]] += 1
        per[kind(issue["message"])] += 1
        kinds[kind(issue["message"])] += 1
    print(f"| {n} | {item['type']} | {item['id']} | " + ", ".join(f"{k}({v})" for k, v in sorted(per.items())) + " |")
print()
print(f"# issue level 합계: {dict(levels)}")
print(f"# 경고 종류 합계: {dict(kinds)}")
