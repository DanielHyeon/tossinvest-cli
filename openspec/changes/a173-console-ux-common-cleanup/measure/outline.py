"""사용: python3 outline.py <html-폴더> [outline|count]
outline: 제목·폼·입력 개요 / count: 사유 문장 중복·식별자 노출·상자 수를 JSON 으로 출력(수치만)."""
import sys,re,glob,os,json
from html.parser import HTMLParser
class P(HTMLParser):
    def __init__(s): super().__init__(); s.out=[]; s.cur=None; s.buf=''
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
def strip_tech(h):  # details.tech / 스크립트·스타일 제외한 본문
    h=re.sub(r'<details class="tech".*?</details>','',h,flags=re.S)
    h=re.sub(r'<style.*?</style>','',h,flags=re.S)
    return h
def text(h): return ' '.join(re.sub(r'<[^>]+>',' ',h).split())
def counts(h):
    body=text(strip_tech(h))
    full=text(h)
    return {
     'notice_danger':len(re.findall(r'class="[^"]*\b(?:notice|danger)\b',h)),
     'change_id_body':len(re.findall(r'\ba0\d{2,3}\b',body)),
     'abs_path_body':len(re.findall(r'(?<![\w.])/(?:home|mnt|tmp|var|etc|root)/[\w./-]+',body)),
     'hex64_body':len(re.findall(r'\b[0-9a-f]{64}\b',body)),
     'not_observed':full.count('not_observed'),
     'unmeasured':body.count('미측정'),
    }
mode=sys.argv[2] if len(sys.argv)>2 else 'outline'
files=sorted(glob.glob(sys.argv[1]+'/*.html'))
if mode=='count':
    print(json.dumps({os.path.basename(f):counts(open(f,encoding='utf-8').read()) for f in files},ensure_ascii=False,indent=1))
else:
    for f in files:
        p=P(); p.feed(open(f,encoding='utf-8').read())
        print('=====',os.path.basename(f)); print('\n'.join(p.out))
