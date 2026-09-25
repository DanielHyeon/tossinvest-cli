import sys
def sub(path, old, new):
    s=open(path,encoding='utf-8').read()
    assert s.count(old)==1, (path, old[:60], s.count(old))
    open(path,'w',encoding='utf-8').write(s.replace(old,new))
