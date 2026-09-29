// Analyze harness-written heatmap (red channel = Go-computed peak delta).
// Usage: node scripts/analyze-diff.mjs <scene> [blockW blockH]
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
const require = createRequire('C:/WorkSpace/Golang/arknights_bot-satori/renderer/package.json');
const sharp = require('sharp');

const [scene, bw = 60, bh = 20] = process.argv.slice(2);
const hm = path.join('src/utils/media/testdata/visual/final/compare', `${scene}.heatmap.png`);
const { data, info } = await sharp(readFileSync(hm)).raw().toBuffer({ resolveWithObject: true });
const W = info.width, H = info.height, ch = info.channels;
const cols = Math.ceil(W / bw), rows = Math.ceil(H / bh);
const grid = Array.from({ length: rows }, () => new Array(cols).fill(0));
const cnt = Array.from({ length: rows }, () => new Array(cols).fill(0));
for (let y = 0; y < H; y++) {
  for (let x = 0; x < W; x++) {
    const d = data[(y * W + x) * ch]; // red channel = peak delta
    if (d >= 32) {
      grid[Math.floor(y / bh)][Math.floor(x / bw)] += d;
      cnt[Math.floor(y / bh)][Math.floor(x / bw)]++;
    }
  }
}
console.log(`heatmap ${W}x${H} ch=${ch} (red=peak delta, grid mean >=32)`);
for (let r = 0; r < rows; r++) {
  let line = '';
  for (let c = 0; c < cols; c++) {
    const mean = cnt[r][c] ? Math.round(grid[r][c] / cnt[r][c]) : 0;
    line += mean === 0 ? '.' : String.fromCharCode(48 + Math.min(9, Math.round(mean / 12)));
  }
  console.log(String(r).padStart(2) + ' ' + line);
}
// row bands
console.log('--- row bands (per 8px, mean peak delta) ---');
for (let y = 0; y < H; y += 8) {
  let s = 0, n = 0;
  for (let yy = y; yy < Math.min(y + 8, H); yy++) {
    for (let x = 0; x < W; x += 4) {
      const d = data[(yy * W + x) * ch];
      if (d >= 32) { s += d; n++; }
    }
  }
  if (n > 0) console.log(`  y=${String(y).padStart(4)} mean=${Math.round(s / n)} n=${n}`);
}
// column bands
console.log('--- col bands (per 16px, mean peak delta) ---');
for (let x = 0; x < W; x += 16) {
  let s = 0, n = 0;
  for (let xx = x; xx < Math.min(x + 16, W); xx++) {
    for (let y = 0; y < H; y += 4) {
      const d = data[(y * W + xx) * ch];
      if (d >= 32) { s += d; n++; }
    }
  }
  if (n > 0) console.log(`  x=${String(x).padStart(4)} mean=${Math.round(s / n)} n=${n}`);
}