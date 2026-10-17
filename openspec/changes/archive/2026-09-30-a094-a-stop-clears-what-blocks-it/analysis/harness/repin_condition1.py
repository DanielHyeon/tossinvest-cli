# a094 base 재고정 조건 ① 둘째 갈래 영수증 — 자기 Go 커밋이 바꾼 기존 함수 전수의 증거 대응 + 형제 몫 귀속.
# 실행: 저장소 루트(깨끗한 격리 워크트리)에서 python3 <이 파일> <게이트 ⑤ 로그>. 쓰기 0.
import sys,re
from pathlib import Path
sys.path.insert(0,'tools/logic-map')
import check_analysis as ca
root=Path('.').resolve()
BASE='1ffe2295994de4ca4530fc1086535c1f7dc37835'; LAND='99dfa5bc413081ad352ed8bebf81a1dddb748d2e'
own='766a8456 b74875e7 d485a45f ffd88707 05de14c3 41f3d9f6 bfa61fb4 6a90ebeb 48100446 cb36caf4 e5e7a67f f6a5bcd9'.split()
req=set(ca.changed_existing_functions(root,BASE,LAND).keys())
log=open(sys.argv[1]).read()
missing=set()
for m in re.finditer(r'missing evidence for modified function (\S+?):(\S+)',log):
    missing.add((m.group(1),m.group(2)))
reqs={f'{p}:{n}' for p,n in req}
miss={f'{p}:{n}' for p,n in missing}
print('required',len(reqs),'missing',len(miss),'missing⊆required',miss<=reqs)
touched=set()
for c in own:
    ks={f'{p}:{n}' for p,n in ca.changed_existing_functions(root,c+'^',c).keys()}
    touched|=ks
t_at_base=touched&reqs
print('own-touched (existing at parent)',len(touched),'∩ required(existing at base)',len(t_at_base))
bad=t_at_base&miss
print('own-touched base functions WITHOUT evidence:',sorted(bad) or 0)
print('sibling-only missing:',len(miss-touched))
for f in sorted(t_at_base): print('  own',f)
# 형제 몫 귀속 — 창 안 비-a094 Go 커밋마다 부모 대비 바뀐 기존 함수와 missing 의 교집합
sib='3260f4eb 36ade9b2 277a105a cc480a88 00e1b9bd 8d2f1e12 df3a6c69'.split()
left=set(miss)
for c in sib:
    ks={f'{p}:{n}' for p,n in ca.changed_existing_functions(root,c+'^',c).keys()}
    hit=sorted(ks&miss)
    left-=set(hit)
    print('sibling',c,len(hit),hit)
print('unattributed missing:',sorted(left) or 0)
