// Minimal matrix: which combo gives 24px line box at 16px font?
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
const { chromium } = require('C:/Users/ModerRAS/AppData/Roaming/npm/node_modules/clawdbot/node_modules/playwright-core');

const html = `<!doctype html><html><head><style>
@font-face { font-family: 'NotoSansHans'; src: url('/assets/font/NotoSansHans-Regular.ttf'); }
body { margin: 0; font-family: 'NotoSansHans', serif; }
table { border-spacing: 0; }
td { padding: 1px; }
</style></head><body>
<table><tr><td id="a">1742</td></tr></table>
<div id="b" style="display:inline-block">1742</div>
<div id="c">1742</div>
<span id="d">1742</span>
</body></html>`;

const browser = await chromium.launch({ headless: true, executablePath: 'C:/Users/ModerRAS/AppData/Local/ms-playwright/chromium-1223/chrome-win64/chrome.exe' });
const page = await browser.newPage();
await page.route('**/assets/font/NotoSansHans-Regular.ttf', route => route.fulfill({ path: 'C:/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes/assets/font/NotoSansHans-Regular.ttf' }));
await page.setContent(html, { waitUntil: 'networkidle' });
await page.evaluate(() => document.fonts.ready);
const out = await page.evaluate(() => {
  const h = (sel) => { const el = document.querySelector(sel); const r = el.getBoundingClientRect(); const range = document.createRange(); range.selectNodeContents(el); const tr = range.getBoundingClientRect(); const cs = getComputedStyle(el); return { boxH: r.height, textH: +tr.height.toFixed(1), font: cs.fontFamily, fs: cs.fontSize }; };
  return { td: h('#a'), inlineBlock: h('#b'), div: h('#c'), span: h('#d'), fonts: [...document.fonts].map(f => f.family + ':' + f.status) };
});
console.log(JSON.stringify(out, null, 1));

// also: measure with font blocked (fallback serif) for comparison
await page.route('**/NotoSansHans-Regular.ttf', route => route.abort());
await page.reload({ waitUntil: 'networkidle' });
const out2 = await page.evaluate(() => {
  const el = document.querySelector('#a'); const range = document.createRange(); range.selectNodeContents(el);
  return { tdTextH: +range.getBoundingClientRect().height.toFixed(1) };
});
console.log('fallback serif td textH:', JSON.stringify(out2));
await browser.close();
