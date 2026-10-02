// A 路：legacy 截图 PNG -> Chromium 编码 -> (bytes 落盘，解码留给 Python/PIL)
// B 路在 Go 侧做。两路共用同一组 quality，逐 q 比较。
const fs = require('fs');
const PORT = 9223;
const BASE = 'http://127.0.0.1:38126';
const OUT = 'C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/t15';
const ROUTES = JSON.parse(fs.readFileSync(OUT + '/routes.json', 'utf8'));
const QS = JSON.parse(fs.readFileSync(OUT + '/qualities.json', 'utf8'));

const sleep = ms => new Promise(r => setTimeout(r, ms));

(async () => {
  fs.mkdirSync(OUT + '/png', {recursive: true});
  fs.mkdirSync(OUT + '/chrom', {recursive: true});
  const ver = await (await fetch(`http://127.0.0.1:${PORT}/json/version`)).json();
  const ws = new WebSocket(ver.webSocketDebuggerUrl);
  await new Promise((res, rej) => { ws.onopen = res; ws.onerror = rej; });
  let id = 0; const pending = new Map(); const events = [];
  ws.onmessage = ev => {
    const m = JSON.parse(ev.data);
    if (m.id && pending.has(m.id)) { pending.get(m.id)(m); pending.delete(m.id); }
    else if (m.method) events.push(m);
  };
  const cmd = (method, params = {}, sessionId) => new Promise((res, rej) => {
    const mid = ++id;
    pending.set(mid, m => (m.error ? rej(new Error(method + ':' + JSON.stringify(m.error))) : res(m.result)));
    ws.send(JSON.stringify({id: mid, method, params, sessionId}));
  });
  const {targetId} = await cmd('Target.createTarget', {url: 'about:blank'});
  const {sessionId} = await cmd('Target.attachToTarget', {targetId, flatten: true});
  await cmd('Page.enable', {}, sessionId);
  await cmd('Runtime.enable', {}, sessionId);

  for (const [scene, route] of Object.entries(ROUTES)) {
    await cmd('Page.navigate', {url: BASE + route}, sessionId);
    const t0 = Date.now();
    while (Date.now() - t0 < 15000) {
      if (events.some(e => e.method === 'Page.loadEventFired')) { events.length = 0; break; }
      await sleep(100);
    }
    await sleep(800);
    const shot = await cmd('Page.captureScreenshot', {format: 'png'}, sessionId);
    const b64 = shot.data;
    fs.writeFileSync(`${OUT}/png/${scene}.png`, Buffer.from(b64, 'base64'));
    // 用 Chromium 自己的编码器把该 PNG 重新编码为 JPEG（canvas.toDataURL）
    const expr = `(async () => {
      const img = new Image();
      img.src = "data:image/png;base64,${b64}";
      await img.decode();
      const c = document.createElement('canvas');
      c.width = img.naturalWidth; c.height = img.naturalHeight;
      const g = c.getContext('2d');
      g.drawImage(img, 0, 0);
      const out = ${JSON.stringify(QS)}.map(q => [q, c.toDataURL('image/jpeg', q/100)]);
      return JSON.stringify(out);
    })()`;
    const r = await cmd('Runtime.evaluate',
      {expression: expr, awaitPromise: true, returnByValue: true}, sessionId);
    const pairs = JSON.parse(r.result.value);
    for (const [q, dataurl] of pairs) {
      const b64j = dataurl.split(',')[1];
      fs.writeFileSync(`${OUT}/chrom/${scene}_q${q}.jpg`, Buffer.from(b64j, 'base64'));
    }
    process.stderr.write(`${scene} -> png + ${pairs.length} chromium jpegs\n`);
  }
  ws.close();
  process.exit(0);
})().catch(e => { process.stderr.write('FATAL ' + e.message + '\n'); process.exit(2); });