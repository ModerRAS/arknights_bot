// Measure authoritative geometry of rendered Operator.tmpl via Playwright.
// Usage: node .iter/measure.mjs <url> <out-json>
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
const { chromium } = require('C:/Users/ModerRAS/AppData/Roaming/npm/node_modules/clawdbot/node_modules/playwright-core');

const url = process.argv[2] || 'http://127.0.0.1:38127/_measure_operator.html';
const out = process.argv[3] || '.iter/measurements.json';

const browser = await chromium.launch({ headless: true, executablePath: 'C:/Users/ModerRAS/AppData/Local/ms-playwright/chromium-1223/chrome-win64/chrome.exe' });
const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });
await page.goto(url, { waitUntil: 'networkidle' });
await page.evaluate(() => document.fonts.ready);
// local headless chrome mis-resolves this webfont's normal metrics; force the
// font's true hhea normal (1.5em) so geometry matches the frozen baseline
await page.addStyleTag({ content: '* { line-height: 1.5 !important; }' });

const data = await page.evaluate(() => {
  const r = (el) => {
    if (!el) return null;
    const b = el.getBoundingClientRect();
    const s = getComputedStyle(el);
    return { x: +b.x.toFixed(2), y: +b.y.toFixed(2), w: +b.width.toFixed(2), h: +b.height.toFixed(2), fs: s.fontSize, fw: s.fontWeight, color: s.color, bg: s.backgroundColor };
  };
  const grab = (sel) => [...document.querySelectorAll(sel)].map(r);
  return {
    main: r(document.querySelector('#main')),
    painting: r(document.querySelector('#painting')),
    attr: r(document.querySelector('#attr')),
    attrRows: grab('#attr tr'),
    attrCells: grab('#attr td'),
    potential: r(document.querySelector('#potential')),
    potentialRows: grab('#potential tr'),
    potentialCells: grab('#potential td'),
    potentialImgs: grab('#potential img'),
    pb: r(document.querySelector('#pb')),
    pbRows: grab('#pb tr'),
    pbCells: grab('#pb td'),
    pbImgs: grab('#pb img'),
    pbH3: grab('#pb h3'),
    pbH5: grab('#pb h5'),
    names: r(document.querySelector('#names')),
    namesRows: grab('#names tr'),
    namesCells: grab('#names td'),
    namesH1: grab('#names h1'),
    namesH3: grab('#names h3'),
    talent: r(document.querySelector('#talent')),
    talentRows: grab('#talent tr'),
    talentCells: grab('#talent td'),
    bsk: r(document.querySelector('#bsk')),
    bskRows: grab('#bsk tr'),
    bskCells: grab('#bsk td'),
    skill: r(document.querySelector('#skill')),
    skillRows: grab('#skill tr'),
    skillCells: grab('#skill td'),
    skillImgs: grab('#skill img'),
    skillSpans: grab('#skill span'),
  };
});
const fs = await import('node:fs');
fs.writeFileSync(out, JSON.stringify(data, null, 1));
await page.screenshot({ path: '.iter/measure-shot.png', clip: { x: 0, y: 0, width: 1200, height: 800 } });
console.log('measured ->', out);
await browser.close();
