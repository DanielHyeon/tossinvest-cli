#!/usr/bin/env python3
"""a199 census 갱신 영수증 — shapes.json 키별 비교(HEAD 대 작업 트리)와 sites.json 동일성. 저장소 루트에서 실행."""

import json
import subprocess

D = "internal/accountbudgetruntime/testdata/a166-wiring/"


def head(path: str) -> str:
    return subprocess.run(["git", "show", f"HEAD:{path}"], check=True, capture_output=True, text=True).stdout


old = json.loads(head(D + "shapes.json"))
new = json.load(open(D + "shapes.json", encoding="utf-8"))
changed = sorted(k for k in old.keys() & new.keys() if old[k] != new[k])
print(f"shapes.json: keys HEAD {len(old)} now {len(new)}, added {sorted(new.keys() - old.keys())}, "
      f"removed {sorted(old.keys() - new.keys())}, changed {changed}, unchanged {len(old) - len(changed)}")
sites_same = json.loads(head(D + "sites.json")) == json.load(open(D + "sites.json", encoding="utf-8"))
print(f"sites.json identical {sites_same}")
