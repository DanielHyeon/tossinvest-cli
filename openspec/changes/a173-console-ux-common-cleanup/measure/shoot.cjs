const { chromium } = require('/home/daniel/gstack/node_modules/playwright');
const fs=require('fs');const [,,base,tok,out]=process.argv;
const pages=['dashboard','verify-console','positions','orders','history','signals','report','verify','settings','settings/standing','settings/daily','settings/strategy','settings/tools','optimization','optimization?category=exit-protection','optimization?category=position-management','optimization?category=strategy-runtime','optimization?category=performance-history','optimization?category=candidate-filters','position-management','strategy-runtime','strategy-runtime/market-schedule','performance-history','openapi/login'];
(async()=>{const b=await chromium.launch();
for(const [tag,vp] of [['d',{width:1280,height:900}],['m',{width:390,height:844}]]){
const ctx=await b.newContext({viewport:vp,colorScheme:'light',reducedMotion:'reduce'});const p=await ctx.newPage();
await p.goto(base+'/?session='+tok);
for(const n of pages){const f=n.replace(/[\/?=&]/g,'_');const r=await p.goto(base+'/'+n);
 if(tag==='d') fs.writeFileSync(`${out}/${f}.html`,await r.text());
 await p.screenshot({path:`${out}/${tag}-${f}.png`,fullPage:true});
 const m=await p.evaluate(()=>({h:document.documentElement.scrollHeight,forms:document.forms.length,inputs:document.querySelectorAll('input,select,textarea').length,buttons:document.querySelectorAll('button,input[type=submit]').length,sections:document.querySelectorAll('section,fieldset,details,.card').length,words:document.body.innerText.split(/\s+/).length}));
 console.log(tag,n,r.status(),JSON.stringify(m));}
await ctx.close();}
await b.close();})();
