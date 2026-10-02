const fs=require('fs');
const CASES=[['case_noise','C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/t15/disc/case_noise.png'],
             ['case_base','C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/t15/disc/case_base.png']];
const QS=[70,90,100];
(async()=>{
  const ver=await (await fetch('http://127.0.0.1:9223/json/version')).json();
  const ws=new WebSocket(ver.webSocketDebuggerUrl);
  await new Promise((r,j)=>{ws.onopen=r;ws.onerror=j;});
  let id=0;const p=new Map();
  ws.onmessage=e=>{const m=JSON.parse(e.data);if(m.id&&p.has(m.id)){p.get(m.id)(m);p.delete(m.id);}};
  const cmd=(m,pr={},s)=>new Promise((res,rej)=>{const i=++id;p.set(i,x=>x.error?rej(new Error(JSON.stringify(x.error))):res(x.result));ws.send(JSON.stringify({id:i,method:m,params:pr,sessionId:s}));});
  const {targetId}=await cmd('Target.createTarget',{url:'about:blank'});
  const {sessionId}=await cmd('Target.attachToTarget',{targetId,flatten:true});
  await cmd('Page.enable',{},sessionId); await cmd('Runtime.enable',{},sessionId);
  for(const [name,path] of CASES){
    const b64=fs.readFileSync(path).toString('base64');
    const expr=`(async()=>{const img=new Image();img.src="data:image/png;base64,${b64}";await img.decode();
      const c=document.createElement('canvas');c.width=img.naturalWidth;c.height=img.naturalHeight;
      c.getContext('2d').drawImage(img,0,0);
      return JSON.stringify(${JSON.stringify(QS)}.map(q=>[q,c.toDataURL('image/jpeg',q/100)]));})()`;
    const r=await cmd('Runtime.evaluate',{expression:expr,awaitPromise:true,returnByValue:true},sessionId);
    for(const [q,du] of JSON.parse(r.result.value))
      fs.writeFileSync(`C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/t15/disc/${name}_chrom_q${q}.jpg`,Buffer.from(du.split(',')[1],'base64'));
    process.stderr.write(name+' chromium jpegs written\n');
  }
  ws.close();process.exit(0);
})().catch(e=>{process.stderr.write('FATAL '+e.message+'\n');process.exit(2);});
