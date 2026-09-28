#!/usr/bin/env python3
"""a125 4.2 — a063 의 soak_test base 번들 둘이 P 블롭과 **해시만** 다른지 잰다(design D4-2 의 근거).

    cd <저장소 루트> && python3 <이 파일> [--ref <번들을 읽을 커밋, 기본 76d0816a^>]

P(`da80ce31`)의 `cmd/tossctl/soak_test.go` 를 임시 파일로 꺼내 추출기로 두 함수를 뽑고, 번들(`--ref` 커밋의 것)과
분기 목록 · 시작/끝 좌표를 대조한다. 같고 `source_sha256` 만 다르면 "해시만 P 판본으로" 고친 것이 정당하다.
"""
import argparse
import hashlib
import json
import subprocess
import tempfile
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("--ref", default="76d0816a^")
args = parser.parse_args()
P = "da80ce31b6a1ab5d443016768f970a82bab102db"
A = "openspec/changes/a063-align-attestation-renewal-profile/analysis/function-logic/"
pairs = (("TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing", "cmd-tossctl--testsoakattestrefusesanunfinishedsoakandwritesnothing"),
         ("TestSoakAttestWritesAVerifiableAttestation", "cmd-tossctl--testsoakattestwritesaverifiableattestation"))
blob = subprocess.run(["git", "show", f"{P}:cmd/tossctl/soak_test.go"], capture_output=True, check=True).stdout
ok = True
with tempfile.TemporaryDirectory() as raw:
    source = Path(raw) / "soak_test.go"
    source.write_bytes(blob)
    for function, bundle in pairs:
        at_p = json.loads(subprocess.run(["go", "run", "./tools/logic-map", "--file", str(source), "--func", function],
                                         capture_output=True, text=True, check=True).stdout)
        recorded = json.loads(subprocess.run(["git", "show", f"{args.ref}:{A}{bundle}/ast.json"],
                                             capture_output=True, text=True, check=True).stdout)
        strip = lambda value: [{k: v for k, v in b.items() if k not in ("line", "column")} for b in (value.get("branches") or [])]
        row = {"function": function, "branches_same": strip(at_p) == strip(recorded),
               "start_same": at_p["start"] == recorded["start"], "end_same": at_p["end"] == recorded["end"],
               "recorded_sha": recorded["source_sha256"], "p_blob_sha": hashlib.sha256(blob).hexdigest()}
        ok &= row["branches_same"] and row["start_same"] and row["end_same"]
        print(json.dumps(row, ensure_ascii=False))
print("SAME EXCEPT HASH" if ok else "DIFFERENT")
raise SystemExit(0 if ok else 1)
