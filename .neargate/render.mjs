// Local iteration helper (not committed): render selected modules from frozen specs
// via the resident runner protocol, writing PNGs for visual-regression comparison.
import { spawn } from 'node:child_process';
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import path from 'node:path';

const ids = process.argv[2] ? process.argv[2].split(',') : ['base', 'box', 'missing', 'state', 'enemy'];
const root = path.resolve(new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]):/, '$1:'));

// Load specs
const specs = new Map();
for (const name of ['static-specs.ndjson', 'dynamic-render-specs.ndjson', 'complex_render_specs.ndjson']) {
  const lines = readFileSync(path.join(root, 'specs', name), 'utf8').split('\n').filter(Boolean);
  for (const line of lines) {
    const s = JSON.parse(line);
    if (s.id) specs.set(s.id, s);
  }
}

const child = spawn(process.execPath, [path.join(root, 'renderer', 'runner.mjs'), '--ndjson'], {
  cwd: root,
  env: { ...process.env, SATORI_ASSET_MANIFEST: path.join(root, 'src', 'utils', 'media', 'testdata', 'visual', 'baseline', 'resource-manifest.json') },
  stdio: ['pipe', 'pipe', 'inherit'],
});
mkdirSync(path.join(root, '.neargate', 'new'), { recursive: true });

const pending = new Map();
let buf = '';
child.stdout.on('data', (chunk) => {
  buf += chunk;
  let idx;
  while ((idx = buf.indexOf('\n')) >= 0) {
    const line = buf.slice(0, idx);
    buf = buf.slice(idx + 1);
    if (!line.trim() || !line.startsWith('{')) continue;
    const msg = JSON.parse(line);
    const resolve = pending.get(msg.id);
    if (resolve) { pending.delete(msg.id); resolve(msg); }
  }
});

function request(spec) {
  return new Promise((resolve, reject) => {
    pending.set(spec.id, resolve);
    child.stdin.write(JSON.stringify({ id: spec.id, component: spec.component, width: spec.width, height: spec.height, scale: spec.scale, props: spec.props }) + '\n');
    setTimeout(() => reject(new Error(`timeout rendering ${spec.id}`)), 60000);
  });
}

try {
  for (const id of ids) {
    const spec = specs.get(id);
    if (!spec) throw new Error(`spec not found: ${id}`);
    const res = await request(spec);
    if (!res.ok) throw new Error(`${id}: ${res.error?.code} ${res.error?.message}`);
    writeFileSync(path.join(root, '.neargate', 'new', `${id}.png`), Buffer.from(res.dataBase64, 'base64'));
    console.log(`${id}: ${res.width}x${res.height} ok`);
  }
  console.log('done');
} finally {
  child.kill();
}
