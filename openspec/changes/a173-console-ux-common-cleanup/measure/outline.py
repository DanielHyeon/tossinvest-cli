import sys,re,html,glob,os
from html.parser import HTMLParser
class P(HTMLParser):
    def __init__(s): super().__init__(); s.out=[]; s.cur=None; s.buf=''; s.depth=0
    def handle_starttag(s,t,a):
        a=dict(a)
        if t in('h1','h2','h3','h4','summary','legend','label','button','th'): s.cur=t; s.buf=''
        if t=='form': s.out.append(f"  [form {a.get('method','get')} {a.get('action','')}]")
        if t=='input' and a.get('type') not in ('hidden',): s.out.append(f"    <input {a.get('type','text')} name={a.get('name')}>")
        if t=='select': s.out.append(f"    <select name={a.get('name')}>")
    def handle_endtag(s,t):
        if s.cur==t:
            txt=' '.join(s.buf.split())[:90]
            if txt: s.out.append({'h1':'# ','h2':'## ','h3':'### ','h4':'#### ','summary':'  > ','legend':'  legend: ','label':'    label: ','button':'    [btn] ','th':'    th: '}[t]+txt)
            s.cur=None
    def handle_data(s,d):
        if s.cur: s.buf+=d
for f in sorted(glob.glob(sys.argv[1]+'/*.html')):
    p=P(); p.feed(open(f,encoding='utf-8').read())
    print('=====',os.path.basename(f)); print('\n'.join(p.out))
