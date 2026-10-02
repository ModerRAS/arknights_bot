const fs=require('fs');
(async()=>{
  const b64=fs.readFileSync('C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/t15/ctrl/noise.png').toString('base64');
  const ver=await (await fetch('http://127.0.0.1:9223/json/version')).json();
  const ws=new WebSocket(ver.webSocketDebuggerUrl);
  await new Promise((r,j)=>{ws.onopen=r;ws.onerror=j;});
  let id=0;const p=new Map();
  ws.onmessage=e=>{const m=JSON.parse(e.data);if(m.id&&p.has(m.id)){p.get(m.id)(m);p.delete(m.id);}};
  const cmd=(method,params={},sessionId)=>new Promise((res,rej)=>{const i=++id;p.set(i,m=>m.error?rej(new Error(JSON.stringify(m.error))):res(m.result));ws.send(JSON.stringify({id:i,method,params,sessionId}));});
  const {targetId}=await cmd('Target.createTarget',{url:'about:blank'});
  const {sessionId}=await cmd('Target.attachToTarget',{targetId,flatten:true});
  await cmd('Page.enable',{},sessionId); await cmd('Runtime.enable',{},sessionId);
  const expr=`(async()=>{const img=new Image();img.src="data:image/png;base64,${b64}";await img.decode();
    const c=document.createElement('canvas');c.width=img.naturalWidth;c.height=img.naturalHeight;
    c.getContext('2d').drawImage(img,0,0);return c.toDataURL('image/jpeg',0.70);})()`;
  const r=await cmd('Runtime.evaluate',{expression:expr,awaitPromise:true,returnByValue:true},sessionId);
  fs.writeFileSync('C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/t15/ctrl/noise_chrom_q70.jpg',Buffer.from(r.result.value.split(',')[1],'base64'));
  process.stderr.write('chromium noise jpeg written\n');ws.close();process.exit(0);
})().catch(e=>{process.stderr.write('FATAL '+e.message+'\n');process.exit(2);});
