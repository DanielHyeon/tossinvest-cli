// 사용: node shoot.cjs <playwright-모듈-경로> <base-url> <session-token> <출력-폴더> <metrics-json-경로>
// 22+ 화면 x 1280/390 x light/dark 촬영. HTML 은 light 데스크톱만 덤프. GET 이동만 하며 버튼은 누르지 않는다.
const [,,pwPath,base,tok,out,metricsPath]=process.argv;
if(!metricsPath){console.error('인자 5개 필요: playwright base token outdir metrics.json');process.exit(2);}
const { chromium } = require(pwPath);
const fs=require('fs');
fs.mkdirSync(out,{recursive:true});
const pages=['dashboard','verify-console','positions','orders','history','signals','report','verify','settings','settings/standing','settings/daily','settings/strategy','settings/tools','optimization','optimization?category=exit-protection','optimization?category=position-management','optimization?category=strategy-runtime','optimization?category=performance-history','optimization?category=candidate-filters','position-management','strategy-runtime','strategy-runtime/market-schedule','performance-history','openapi/login'];
const sample=()=>{
  const px=e=>e?Math.round(e.getBoundingClientRect().height):0;
  let maxP=0;
  document.querySelectorAll('p').forEach(e=>{const r=e.getBoundingClientRect();if(r.width>maxP&&e.innerText.length>40)maxP=Math.round(r.width);});
  const floating=document.querySelectorAll('.notice,.danger').length;
  const strip=document.querySelector('.console-status');
  return {h:document.documentElement.scrollHeight,forms:document.forms.length,
    inputs:document.querySelectorAll('input,select,textarea').length,
    buttons:document.querySelectorAll('button,input[type=submit]').length,
    notices:floating,stripH:px(strip),maxP,
    words:document.body.innerText.split(/\s+/).length};
};
(async()=>{const b=await chromium.launch();const all={};
for(const scheme of ['light','dark'])
for(const [tag,vp] of [['d',{width:1280,height:900}],['m',{width:390,height:844}]]){
  const ctx=await b.newContext({viewport:vp,colorScheme:scheme,reducedMotion:'reduce'});const p=await ctx.newPage();
  await p.goto(base+'/?session='+tok);
  for(const n of pages){const f=n.replace(/[\/?=&]/g,'_');const r=await p.goto(base+'/'+n);
    if(tag==='d'&&scheme==='light') fs.writeFileSync(`${out}/${f}.html`,await r.text());
    try{await p.screenshot({path:`${out}/${scheme}-${tag}-${f}.png`,fullPage:true});}
    catch(e){console.error('스크린샷 실패',scheme,tag,n,e.message.split('\n')[0]);}
    const m=await p.evaluate(sample);m.status=r.status();
    all[`${scheme}/${tag}/${n}`]=m;}
  await ctx.close();}
await b.close();
fs.writeFileSync(metricsPath,JSON.stringify(all,null,1));
console.log('화면수',Object.keys(all).length,'->',metricsPath);})();
