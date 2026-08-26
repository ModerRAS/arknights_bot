import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const args = Object.fromEntries(process.argv.slice(2).reduce((pairs, value, index, values) => value.startsWith('--') ? [...pairs, [value.slice(2), values[index + 1]]] : pairs, []));
const root = path.resolve(args.root || process.env.VISUAL_ROOT || path.resolve(here, '../../..'));
const baseline = path.resolve(args.baseline || process.env.VISUAL_BASELINE || path.join(root, 'src/utils/media/testdata/visual/baseline'));
const clockErratumPath = path.resolve(args['clock-erratum'] || process.env.VISUAL_CLOCK_ERRATUM || path.join(root, 'src/utils/media/testdata/visual/capture-clock-erratum.json'));
const out = args.out || process.env.VISUAL_OUT;
const baseURL = (args['base-url'] || process.env.VISUAL_LEGACY_URL || 'http://127.0.0.1:38126').replace(/\/$/, '');
const playwrightModule = args.playwright || process.env.VISUAL_PLAYWRIGHT_MODULE || 'playwright';
const settleMS = Number(args['settle-ms'] || process.env.VISUAL_SETTLE_MS || 3000);
if (!out || !Number.isFinite(settleMS) || settleMS < 0) throw new Error('usage: node legacy-capture.mjs --out <final-artifact-dir> [--root <repo-root>] [--baseline <baseline-dir>] [--clock-erratum <json>] [--base-url <legacy-url>] [--playwright <module>] [--settle-ms <milliseconds>] [--id <page-id>]');
const require = createRequire(import.meta.url);
const { chromium } = require(playwrightModule);
const cache = path.join(baseline, 'cache');
const manifest = JSON.parse(fs.readFileSync(path.join(baseline, 'manifest.json'), 'utf8'));
const clockErratum = JSON.parse(fs.readFileSync(clockErratumPath, 'utf8'));
const entries = args.id ? manifest.entries.filter(entry => entry.id === args.id) : manifest.entries;
if (entries.length === 0) throw new Error(`unknown manifest id: ${args.id}`);
const clock = Object.fromEntries((clockErratum.immutableBaselineEvidence || []).map(entry => [entry.id, entry.captureClockUnixSeconds]));
const sha = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const cacheMap = {
  'https://fixture-cache.invalid/card/secretary-painting.png': 'card-secretary-1024.png',
  'https://fixture-cache.invalid/card/player-avatar.png': 'card-player-portrait-180x360.png',
  'https://fixture-cache.invalid/operator/amiya-painting.png': 'operator-painting-1024.png',
  'https://fixture-cache.invalid/operator/building-skill-icon.png': 'operator-building-36.png',
  'https://fixture-cache.invalid/operator/skill-icon.png': 'operator-skill-128.png',
  'https://fixture-cache.invalid/enemy/originium-slug.png': 'enemy-originium-slug-158.png',
  'https://fixture-cache.invalid/recruit-amiya.png': 'amiya-avatar.webp',
  'https://fixture-cache.invalid/depot-lmd.png': 'depot-lmd.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/portrait/char_002_amiya%231.png': 'card-player-portrait-180x360.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_002_amiya%231.png': 'avatar-amiya-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_103_angel%231.png': 'avatar-angel-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_102_texas%231.png': 'avatar-texas-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_112_siege%231.png': 'base-avatar-siege-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_202_demkni%231.png': 'base-avatar-saria-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_003_kalts%231.png': 'base-avatar-kalts-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_172_svrash%231.png': 'base-avatar-silverash-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_180_amgoat%231.png': 'base-avatar-eyjafjalla-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skill/skchr_amiya_3.png': 'base-skill-amiya-128.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_1001_amiya2%232.png': 'state-training-amiya-180.png',
  'https://web.hycdn.cn/arknights/game/assets/char_skill/skcom_magic_rage%5B3%5D.png': 'boxdetail-skill-amiya.png',
  'https://web.hycdn.cn/arknights/game/assets/uniequip/type/icon/original.png': 'boxdetail-equip-original.png',
  'https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90': 'gacha-avatar-amiya.webp',
  'https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90': 'gacha-avatar-exusiai.webp',
  'https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90': 'amiya-half.webp',
  'https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90': 'amiya-avatar.webp',
  'https://media.prts.wiki/3/3e/%E5%A4%B4%E5%83%8F_%E6%95%8C%E4%BA%BA_%E6%BA%90%E7%9F%B3%E8%99%AB.png': 'enemy-originium-slug-158.png'
};
const route = {'box-detail': 'boxDetail', 'box-summary': 'boxSummary', lottery: 'lotteryDetail'};
const target = path.join(out, 'old');
fs.mkdirSync(target, {recursive: true});

const browser = await chromium.launch({headless: true});
const results = [];
try {
  for (const entry of entries) {
    const start = Date.now();
    const id = entry.id;
    const context = await browser.newContext({viewport: {width: 1280, height: 720}, deviceScaleFactor: entry.scale, timezoneId: 'Asia/Shanghai'});
    if (clock[id]) await context.addInitScript(({ms, advance}) => {
      const RealDate = Date;
      const realStart = RealDate.now();
      const now = advance ? () => ms + (RealDate.now() - realStart) : () => ms;
      function FrozenDate(...values) { return values.length ? new RealDate(...values) : new RealDate(now()); }
      FrozenDate.now = now;
      FrozenDate.parse = RealDate.parse;
      FrozenDate.UTC = RealDate.UTC;
      FrozenDate.prototype = RealDate.prototype;
      globalThis.Date = FrozenDate;
    }, {ms: clock[id] * 1000, advance: id === 'gacha'});
    const page = await context.newPage();
    const failed = [], pending = new Set();
    page.on('request', request => pending.add(request.url()));
    page.on('requestfinished', request => pending.delete(request.url()));
    page.on('requestfailed', request => { pending.delete(request.url()); failed.push({url: request.url(), error: request.failure()?.errorText || ''}); });
    await page.route('**/*', request => {
      const file = cacheMap[request.request().url()];
      return file ? request.fulfill({body: fs.readFileSync(path.join(cache, file)), contentType: file.endsWith('.webp') ? 'image/webp' : 'image/png'}) : request.continue();
    });
    const result = {id, scale: entry.scale, format: 'jpeg', baselineSha256: entry.sha256, elapsedMs: 0, failed, pending: []};
    try {
      const response = await page.goto(`${baseURL}/${route[id] || id}`, {waitUntil: 'networkidle', timeout: 30000});
      await page.waitForFunction(() => document.fonts.ready.then(() => [...document.images].every(image => image.complete)), null, {timeout: 10000});
      if (id === 'gacha') await page.waitForTimeout(settleMS);
      const main = page.locator('#main');
      result.status = response?.status();
      result.visible = await main.isVisible();
      result.bbox = await main.boundingBox();
      const image = path.join(target, `${id}.jpg`);
      await main.screenshot({path: image, type: 'jpeg'});
      result.path = path.relative(out, image).replaceAll('\\', '/');
      result.sha256 = sha(image);
      result.shaMatchesBaseline = result.sha256 === entry.sha256;
      if (!result.shaMatchesBaseline) result.error = `legacy SHA mismatch: ${result.sha256}`;
    } catch (error) {
      result.error = String(error?.message || error);
    }
    result.pending = [...pending];
    result.elapsedMs = Date.now() - start;
    results.push(result);
    await context.close();
  }
} finally {
  await browser.close();
}
results.sort((a, b) => a.id.localeCompare(b.id));
const report = {engine: 'Playwright Chromium legacy-template service', fixtureServer: baseURL, manifestSha256: sha(path.join(baseline, 'manifest.json')), clockErratum: path.relative(out, clockErratumPath).replaceAll('\\', '/'), settleMS, results, rendered: results.filter(x => x.path).length, shaGatePassed: results.filter(x => x.shaMatchesBaseline).length, failed: results.filter(x => x.error || x.status !== 200 || !x.visible || x.failed.length || x.pending.length).length};
fs.writeFileSync(path.join(out, 'old-render-report.json'), JSON.stringify(report, null, 2) + '\n');
console.log(JSON.stringify({rendered: report.rendered, shaGatePassed: report.shaGatePassed, failed: report.failed}));
if (report.rendered !== entries.length || report.shaGatePassed !== entries.length || report.failed) process.exitCode = 1;
