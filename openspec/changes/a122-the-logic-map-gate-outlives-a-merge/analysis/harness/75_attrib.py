"""task 7.5 MEASURE — spawn 이 **어느 함수**에서 나오는가 + 한 개 fetch 의 두 방식 비교.

편집 집합을 추측으로 고르지 않는다. 배치가 값을 내는 자리가 `_pinning_at` 하나뿐이면
나머지는 안 고친다(KISS). 한 개짜리 fetch 가 `git show` 보다 느리면 `_committed_bytes` 를
배치 위에 올리는 설계는 죽는다.
"""
import importlib.util
import statistics
import subprocess
import sys
import time
import traceback
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀌고, 체크아웃 이름이 바뀌면 절대경로는 고아가 된다
# ([[renamed-checkout-strands-absolute-path-state]]).
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
spec = importlib.util.spec_from_file_location("ca75a", ROOT / "tools" / "logic-map" / "check_analysis.py")
ca = importlib.util.module_from_spec(spec)
sys.modules["ca75a"] = ca
spec.loader.exec_module(ca)

REAL = subprocess.run
BY_FUNC: dict[str, int] = {}


def attributing(*args, **kwargs):
    # 부른 **저장소 함수**로 귀속한다 — 스택에서 check_analysis 의 가장 안쪽 프레임.
    for frame in traceback.extract_stack()[::-1]:
        if frame.filename.endswith("check_analysis.py"):
            if frame.name in ("_committed_bytes", "_is_ancestor"):
                continue        # 잎이 아니라 **호출자**로 귀속한다
            BY_FUNC[frame.name] = BY_FUNC.get(frame.name, 0) + 1
            break
    return REAL(*args, **kwargs)


ca.subprocess.run = attributing
changes = ROOT / "openspec" / "changes"

for name in sys.argv[1:]:
    d = changes / "archive" / name if (changes / "archive" / name).is_dir() else changes / name
    m = ca.ARCHIVED_CHANGE.fullmatch(name)
    cid = m.group("change") if m else name
    base = ca.resolve_base(d, ROOT, {}, change_id=cid)
    BY_FUNC.clear()
    start = time.monotonic()
    try:
        value, _ = ca.compute_landing(ROOT, base, d / "analysis" / "function-logic")
    except ca.GATE_FAULTS as exc:      # 7.5.1 부터 git 결함은 예외다 — 한 change 가 나머지를 멈추지 않게
        value = ""
        print(f"{name[:44]:46s} RAISED {type(exc).__name__}: {exc}")
    elapsed = time.monotonic() - start
    total = sum(BY_FUNC.values())
    print(f"{name[:44]:46s} {elapsed:7.2f}s · spawn {total:6d} · {value[:12] or 'none'}")
    for fn, n in sorted(BY_FUNC.items(), key=lambda kv: -kv[1]):
        print(f"{'':48s} {n:6d} ({n / total:5.1%}) {fn}")

# --- 한 개짜리 fetch: show vs cat-file ---
ca.subprocess.run = REAL
PATH = "tools/logic-map/README.md"
HEAD = REAL(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True,
            check=True).stdout.strip()


def one_show():
    return REAL(["git", "show", f"{HEAD}:{PATH}"], cwd=ROOT, capture_output=True,
                timeout=30, check=False).stdout


def one_batch():
    return REAL(["git", "cat-file", "--batch", "-z"], cwd=ROOT,
                input=f"{HEAD}:{PATH}\0".encode(), capture_output=True,
                timeout=60, check=False).stdout


def med(fn, n=25):
    out = []
    for _ in range(n):
        s = time.monotonic()
        fn()
        out.append(time.monotonic() - s)
    return statistics.median(out) * 1000


print(f"\n한 개 fetch: show {med(one_show):.2f} ms · cat-file --batch -z {med(one_batch):.2f} ms")
