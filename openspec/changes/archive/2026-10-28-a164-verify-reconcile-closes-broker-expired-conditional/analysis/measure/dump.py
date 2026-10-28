import json, sys
from pathlib import Path
root = Path(sys.argv[1]); fl = Path(sys.argv[2])
for d in sorted(fl.iterdir()):
    v = json.loads((d/'ast.json').read_text())
    lines = (root/v['file']).read_text().splitlines()
    print(f"### {d.name} {v['file']} {v['start']['line']}-{v['end']['line']}")
    for b in v.get('branches') or []:
        print(f"  {b['id']} {b['kind']} {b['at']['line']}: {lines[b['at']['line']-1].strip()}")
    for r in v.get('returns') or []:
        print(f"  RET {r['at']['line']}: {lines[r['at']['line']-1].strip()}")
    for c in v.get('defers') or []:
        print(f"  DEFER {c['at']['line']}: {c['text']}")
    print("  CALLS", len(v.get('calls') or []), ", ".join(sorted({c['text'] for c in v.get('calls') or []})))
