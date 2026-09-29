// bd-sweep.mjs: parametric sweep for box-detail against official formula.
import { spawn } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';

const ROOT = 'C:/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes';
const specLine = readFileSync(ROOT + '/specs/static-specs.ndjson', 'utf8').split('\n').find(l => l.includes('"box-detail"'));
const spec = JSON.parse(specLine);

function render() {
  return new Promise((res, rej) => {
    const p = spawn('node', ['runner.mjs', '--ndjson'], { cwd: ROOT + '/renderer', env: { ...process.env, SATORI_ASSET_MANIFEST: ROOT + '/src/utils/media/testdata/visual/baseline/resource-manifest.json' } });
    let out = '', err = '';
    p.stdout.on('data', d => out += d); p.stderr.on('data', d => err += d);
    p.on('close', () => {
      for (const line of out.trim().split('\n')) {
        const d = JSON.parse(line);
        if (d.ok && d.id === 'box-detail') return res(Buffer.from(d.dataBase64, 'base64'));
      }
      rej(new Error(err.slice(-200)));
    });
    p.stdin.write(JSON.stringify({ id: spec.id, component: spec.component, width: spec.width, height: spec.height, scale: spec.scale, props: spec.props }) + '\n');
    p.stdin.end();
  });
}

async function score() {
  const png = await render();
  const { default: sharp } = await import('file:///' + ROOT + '/renderer/node_modules/sharp/dist/index.mjs');
  const a = await sharp(ROOT + '/src/utils/media/testdata/visual/final/old/box-detail.jpg').removeAlpha().raw().toBuffer({ resolveWithObject: true });
  const b = await sharp(png).removeAlpha().raw().toBuffer({ resolveWithObject: true });
  let d = 0;
  const A = a.data, B = b.data, n = Math.min(A.length, B.length);
  for (let i = 0; i < n; i++) d += Math.abs(A[i] - B[i]);
  return 1 - d / (a.info.width * a.info.height * 4 * 255);
}

const variants = JSON.parse(readFileSync(process.argv[2] ?? '/dev/null', 'utf8'));

for (const v of variants) {
  writeFileSync(ROOT + '/renderer/components/box-detail.mjs', readFileSync(ROOT + '/.iter/bd.mjs.bak', 'utf8'));
  let s = readFileSync(ROOT + '/renderer/components/box-detail.mjs', 'utf8');
  for (const [from, to] of v.repl) {
    if (from === '__NONE__') continue;
    if (!s.includes(from)) { console.log(v.name, 'PATCH MISS:', from.slice(0, 40)); process.exit(1); }
    s = s.replace(from, to);
  }
  writeFileSync(ROOT + '/renderer/components/box-detail.mjs', s);
  try { console.log(v.name, '->', await score()); } catch (e) { console.log(v.name, 'FAIL', String(e).slice(-120)); }
}
