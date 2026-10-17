# alert_outbox 행 중 계좌 정보가 실린 행 수만 셈 — 값은 출력하지 않음(불변식 8).
import json, re, sqlite3, sys, os
home = os.path.expanduser("~/.config/tossctl")
acct = json.load(open(os.path.join(home, "capability-attestation.json"))).get("account_ref", "")
digits = re.sub(r"[^0-9]", "", acct)
mask = re.compile(r"\*{3,}[0-9](?:-?[0-9])*")
derived = re.compile(r"(^|[^A-Za-z0-9])(?:(pos|rec|mode|adj)-[0-9a-f]{24}|(corr)-[0-9a-f]{32})([^0-9a-f]|$)")
run = re.compile(r"[0-9]+(?:-[0-9]+)*")
def has_acct(s):
    if not s: return False
    if acct and acct in s: return True
    if len(digits) >= 8 and any(re.sub(r"-", "", m) == digits for m in run.findall(s)): return True
    return False
con = sqlite3.connect(f"file:{home}/journal.db?mode=ro", uri=True)
rows = con.execute("SELECT state, event_type, event_key, title, body, payload, last_error FROM alert_outbox").fetchall()
print("registered_ref_present", bool(acct), "digits_len", len(digits))
print("total_rows", len(rows))
from collections import Counter
st = Counter(r[0] for r in rows); print("by_state", dict(st))
cols = ["event_key","title","body","payload","last_error"]
for i, c in enumerate(cols, start=2):
    raw = Counter(r[0] for r in rows if has_acct(r[i]))
    msk = Counter(r[0] for r in rows if r[i] and mask.search(r[i]))
    der = Counter(r[0] for r in rows if r[i] and derived.search(r[i]))
    print(c, "raw", dict(raw), "mask", dict(msk), "derived", dict(der))
acctkey = Counter(r[0] for r in rows if r[5] and any(k.replace("_","").lower() in ("account","accountref","accountno","maskedaccount","accountseq") for k in (json.loads(r[5]) if r[5].startswith("{") else {})))
print("payload_account_key", dict(acctkey))
ext = Counter((r[0]) for r in rows if has_acct(r[3]) or has_acct(r[4]) or (r[3] and (mask.search(r[3]) or derived.search(r[3]))) or (r[4] and (mask.search(r[4]) or derived.search(r[4]))))
print("title_or_body_any(would_be_sent_if_pending)", dict(ext))
types = Counter(r[1] for r in rows if r[0]=="PENDING"); print("pending_types_count", len(types), sum(types.values()))
