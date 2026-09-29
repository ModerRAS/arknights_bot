import { spawn } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
const ROOT = 'C:/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes';
let spec = null;
for (const fn of ['static-specs.ndjson', 'dynamic-render-specs.ndjson', 'complex_render_specs.ndjson']) {
  for (const l of readFileSync(ROOT + '/specs/' + fn, 'utf8').split('\n')) {
    if (!l.includes('"box"')) continue;
    const d = JSON.parse(l);
    if (d.id === 'box') { spec = d; break; }
  }
  if (spec) break;
}
function render() {
  return new Promise((res, rej) => {
    const p = spawn('node', ['runner.mjs', '--ndjson'], { cwd: ROOT + '/renderer', env: { ...process.env, SATORI_ASSET_MANIFEST: ROOT + '/src/utils/media/testdata/visual/baseline/resource-manifest.json' } });
    let out = '', err = '';
    p.stdout.on('data', d => out += d); p.stderr.on('data', d => err += d);
    p.on('close', () => {
      for (const line of out.trim().split('\n')) {
        const d = JSON.parse(line);
        if (d.ok && d.id === 'box') return res(Buffer.from(d.dataBase64, 'base64'));
      }
      rej(new Error(err.slice(-300)));
    });
    p.stdin.write(JSON.stringify({ id: spec.id, component: spec.component, width: spec.width, height: spec.height, scale: spec.scale, props: spec.props }) + '\n');
    p.stdin.end();
  });
}
async function score() {
  const png = await render();
  const { default: sharp } = await import('file:///' + ROOT + '/renderer/node_modules/sharp/dist/index.mjs');
  const a = await sharp(ROOT + '/src/utils/media/testdata/visual/final/old/box.jpg').removeAlpha().raw().toBuffer({ resolveWithObject: true });
  const b = await sharp(png).removeAlpha().raw().toBuffer({ resolveWithObject: true });
  let d = 0;
  const A = a.data, B = b.data, n = Math.min(A.length, B.length);
  for (let i = 0; i < n; i++) d += Math.abs(A[i] - B[i]);
  return 1 - d / (a.info.width * a.info.height * 4 * 255);
}
const variants = JSON.parse(readFileSync(process.argv[2], 'utf8'));
for (const v of variants) {
  writeFileSync(ROOT + '/renderer/components/box.mjs', readFileSync(ROOT + '/.iter/box.bak.mjs', 'utf8'));
  let s = readFileSync(ROOT + '/renderer/components/box.mjs', 'utf8');
  for (const [f, t] of v.repl) {
    if (f === '__NONE__') continue;
    if (!s.includes(f)) { console.log(v.name, 'PATCH MISS:', f.slice(0, 50)); process.exit(1); }
    s = s.replace(f, t);
  }
  writeFileSync(ROOT + '/renderer/components/box.mjs', s);
  try { console.log(v.name, '->', (await score()).toFixed(5)); } catch (e) { console.log(v.name, 'FAIL', String(e).slice(-100)); }
}
writeFileSync(ROOT + '/renderer/components/box.mjs', readFileSync(ROOT + '/.iter/box.bak.mjs', 'utf8'));
console.log('restored');
