// 通过 CDP 读 legacy 页面的 getBoundingClientRect —— 只读，不截图。
// 零 npm 依赖：Node 24 自带 fetch + 全局 WebSocket。
const fs = require('fs');

const CDP_PORT = process.env.CDP_PORT || 9222;
const BASE = 'http://127.0.0.1:38127';
const OUT = process.env.OUT || 'C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/resolved/rects_t2.json';
const ROUTES = JSON.parse(fs.readFileSync(process.env.ROUTES_FILE, 'utf8'));

const COLLECT = `(() => {
  const out = [];
  for (const el of document.querySelectorAll('*')) {
    const id = el.id || '';
    const cls = (el.getAttribute('class') || '');
    if (!id && !cls) continue;
    const r = el.getBoundingClientRect();
    if (r.width <= 0 || r.height <= 0) continue;
    out.push({tag: el.tagName.toLowerCase(), id, cls,
              x: +r.x.toFixed(2), y: +r.y.toFixed(2),
              w: +r.width.toFixed(2), h: +r.height.toFixed(2)});
  }
  return JSON.stringify(out);
})()`;

const sleep = ms => new Promise(r => setTimeout(r, ms));

(async () => {
  const ver = await (await fetch(`http://127.0.0.1:${CDP_PORT}/json/version`)).json();
  const ws = new WebSocket(ver.webSocketDebuggerUrl);
  await new Promise((res, rej) => { ws.onopen = res; ws.onerror = rej; });
  let id = 0;
  const pending = new Map();
  const events = [];
  ws.onmessage = ev => {
    const m = JSON.parse(ev.data);
    if (m.id && pending.has(m.id)) { pending.get(m.id)(m); pending.delete(m.id); }
    else if (m.method) events.push(m);
  };
  const cmd = (method, params = {}, sessionId) => new Promise((res, rej) => {
    const mid = ++id;
    pending.set(mid, m => (m.error ? rej(new Error(method + ': ' + JSON.stringify(m.error))) : res(m.result)));
    ws.send(JSON.stringify({id: mid, method, params, sessionId}));
  });

  const {targetId} = await cmd('Target.createTarget', {url: 'about:blank'});
  const {sessionId} = await cmd('Target.attachToTarget', {targetId, flatten: true});
  await cmd('Page.enable', {}, sessionId);
  await cmd('Runtime.enable', {}, sessionId);

  const result = {};
  for (const [scene, route] of Object.entries(ROUTES)) {
    await cmd('Page.navigate', {url: BASE + route}, sessionId);
    // 等 load 事件
    const t0 = Date.now();
    while (Date.now() - t0 < 15000) {
      if (events.some(e => e.method === 'Page.loadEventFired')) { events.length = 0; break; }
      await sleep(100);
    }
    await sleep(600);                       // 等字体/图片落位
    const r = await cmd('Runtime.evaluate',
      {expression: COLLECT, returnByValue: true}, sessionId);
    try {
      result[scene] = {route, rects: JSON.parse(r.result.value)};
    } catch (e) {
      result[scene] = {route, error: String(e)};
    }
    process.stderr.write(`${scene} -> ${result[scene].rects ? result[scene].rects.length : 'ERR'} elems\n`);
  }
  fs.writeFileSync(OUT, JSON.stringify(result, null, 1), 'utf8');
  process.stderr.write('written ' + OUT + '\n');
  ws.close();
  process.exit(0);
})().catch(e => { process.stderr.write('FATAL ' + e.message + '\n'); process.exit(2); });