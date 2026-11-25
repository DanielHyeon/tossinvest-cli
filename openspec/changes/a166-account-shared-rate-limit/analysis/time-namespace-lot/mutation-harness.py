#!/usr/bin/env python3
"""time namespace 로트 변이 하네스 — 저장소 파일을 고치지 않고 `go test -overlay` 사본으로만 변이함.

사용: mutation-harness.py <repo> <scratch-dir> <expected-bootstamp-sha256> <expected-bootstamp_linux-sha256>
- 시작 때 두 원본의 sha256 을 단언(다른 판이 섞이지 않게), 무변이 대조군이 GREEN 이 아니면 멈춤.
- 한 번에 한 판, 각 판은 정확히 한 자리 치환(개수 1 단언). 결과는 stdout 원문.
"""
import hashlib, json, os, pathlib, re, subprocess, sys

repo = pathlib.Path(sys.argv[1]).resolve()
scratch = pathlib.Path(sys.argv[2]).resolve()
expect = {"internal/clock/bootstamp.go": sys.argv[3], "internal/clock/bootstamp_linux.go": sys.argv[4]}
src = {}
for rel, sha in expect.items():
    data = (repo / rel).read_bytes()
    got = hashlib.sha256(data).hexdigest()
    assert got == sha, f"{rel} sha {got} != {sha}"
    src[rel] = data.decode()
print("start-sha OK", {k: v[:12] for k, v in expect.items()})

B = "internal/clock/bootstamp.go"
BL = "internal/clock/bootstamp_linux.go"
MUTANTS = [
    ("M0", "무변이 대조군", B, None, None),
    ("M1", "B3 uptime 항 삭제", B, "now.BootID == saved.BootID && now.Uptime < saved.Uptime", "now.BootID == saved.BootID && false"),
    ("M2", "B4 uptime 항 삭제", B, "now.BootID == anchor.BootID && now.Uptime < anchor.Uptime", "now.BootID == anchor.BootID && false"),
    ("M3", "B2 uptime 항 삭제", B, "anchor.BootID == saved.BootID && anchor.Uptime < saved.Uptime", "anchor.BootID == saved.BootID && false"),
    ("M4", "B3 에 ns 항 복원", B, "now.BootID == saved.BootID && now.Uptime < saved.Uptime", "now.BootID == saved.BootID && (now.Namespace != saved.Namespace || now.Uptime < saved.Uptime)"),
    ("M4a", "B2 에 ns 항 복원", B, "anchor.BootID == saved.BootID && anchor.Uptime < saved.Uptime", "anchor.BootID == saved.BootID && (anchor.Namespace != saved.Namespace || anchor.Uptime < saved.Uptime)"),
    ("M4b", "B4 에 ns 항 복원", B, "now.BootID == anchor.BootID && now.Uptime < anchor.Uptime", "now.BootID == anchor.BootID && (now.Namespace != anchor.Namespace || now.Uptime < anchor.Uptime)"),
    ("M4c", "B8 에 ns 항 복원", B, "\tif now.Uptime < base.Uptime {", "\tif now.Namespace != base.Namespace || now.Uptime < base.Uptime {"),
    ("M5", "B7 다른 boot 를 같은 boot 처럼 차감", B, "if now.BootID != base.BootID {", "if false {"),
    ("M6", "B8 분기 삭제", B, "\tif now.Uptime < base.Uptime {", "\tif false {"),
    ("M8", "B5 duration 0 반환을 B2 앞으로", B, "\tif anchor != nil && anchor.BootID == saved.BootID", "\tif duration == 0 {\n\t\treturn HintRemaining{}, nil\n\t}\n\tif anchor != nil && anchor.BootID == saved.BootID"),
    ("M7", "Snapshot offset 0 검사 삭제(L152)", BL, "if e = parseZeroOffsets(offsets); e != nil {", "if e = parseZeroOffsets(offsets); false {"),
    ("M7a", "Snapshot current≠children 검사 삭제(B6 L145)", BL, "if e != nil || current != children {", "if e != nil {"),
    ("M7b", "Snapshot 표본 앞뒤 ns 비교 삭제(B12 L168)", BL, "if e != nil || currentAfter != current || childrenAfter != children {", "if e != nil || (false && (currentAfter != current || childrenAfter != children)) {"),
]
PKGS = [["./internal/clock"], ["-run", "CrossNamespace|HintMaxAndRestore|ExpiredSlots", "./internal/accountbudget"]]
env = dict(os.environ)
for mid, desc, rel, old, new in MUTANTS:
    overlay = {}
    if old is not None:
        text = src[rel]
        assert text.count(old) == 1, f"{mid}: {old!r} count {text.count(old)}"
        path = scratch / f"a166-tns-mut-{mid}-{pathlib.Path(rel).name}"
        path.write_text(text.replace(old, new))
        overlay[str(repo / rel)] = str(path)
    ov = scratch / f"a166-tns-mut-{mid}.json"
    ov.write_text(json.dumps({"Replace": overlay}))
    failed, rcs = [], []
    for args in PKGS:
        p = subprocess.run(["go", "test", "-count=1", "-v", "-overlay", str(ov), *args], cwd=repo, capture_output=True, text=True, env=env)
        rcs.append(p.returncode)
        failed += re.findall(r"^--- FAIL: (\S+)", p.stdout, re.M)
        if p.returncode != 0 and not failed:
            failed.append("BUILD/OTHER: " + (p.stdout + p.stderr).strip().splitlines()[-1])
    verdict = "GREEN" if all(rc == 0 for rc in rcs) else "RED"
    if mid == "M0":
        assert verdict == "GREEN", f"control not GREEN: {failed}"
    status = "CONTROL" if mid == "M0" else ("CAUGHT" if verdict == "RED" else "SURVIVED")
    print(f"{mid}\t{status}\t{desc}\trc={rcs}\tfailed={sorted(set(failed))}")
    for rel2, sha in expect.items():  # 판마다 원본 무변 단언
        assert hashlib.sha256((repo / rel2).read_bytes()).hexdigest() == sha, f"{rel2} changed during {mid}"
print("end-sha OK")
