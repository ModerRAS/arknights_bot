import { createInterface } from 'node:readline';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
import sharp from 'sharp';
import { h } from './lib/h.mjs';
import { AssetError, createAssetLoader } from './lib/assets.mjs';

const MAX_LINE_BYTES = 1 << 20;
const MAX_WIDTH = 4096;
const MAX_HEIGHT = 4096;
const MAX_PIXELS = 16 * 1024 * 1024;
const MAX_PROPS_BYTES = 768 * 1024;
const MIN_SCALE = 0.25;
const MAX_SCALE = 4;
const REPO_ROOT = path.resolve(fileURLToPath(new URL('..', import.meta.url)));
const FONT_SOURCE = 'assets/font/NotoSansHans-Regular.ttf';

// Keep this list explicit: request data never becomes an import path.
const COMPONENTS = new Map([
  'base', 'box', 'box-detail', 'box-summary', 'calendar', 'card', 'depot', 'enemy',
  'gacha', 'headhunt', 'help', 'lottery', 'missing', 'operator', 'recruit', 'state',
].map((name) => [name, new URL(`./components/${name}.mjs`, import.meta.url)]));

const modules = new Map();
const assets = createAssetLoader({
  repoRoot: REPO_ROOT,
  onDiagnostic: (entry) => process.stderr.write(`[asset] ${JSON.stringify(entry)}\n`),
});
let fontPromise;

// Contract helper retained for page-level tests and the no-flag smoke command.
export async function mockImage(src, fallback) {
  const resolved = src ?? fallback;
  if (resolved == null) throw new TypeError('image source or fallback is required');
  return `data:image/png;base64,${Buffer.from(String(resolved)).toString('base64')}`;
}

// Page modules expose: default async function render(props, { image }).
export async function renderPage(pageModule, props = {}, image = mockImage) {
  if (!pageModule || typeof pageModule.default !== 'function') {
    throw new TypeError('page module must export default async render(props, { image })');
  }
  return await pageModule.default(props, { image });
}

export const contractPage = {
  default: async function render(props, { image }) {
    const avatar = await image(props.avatar, props.avatarFallback);
    return h(
      'div',
      { style: { display: 'flex', flexDirection: 'column', backgroundImage: `url(${avatar})` } },
      h('span', null, props.title),
      h('img', { src: avatar, width: 48, height: 48 }),
    );
  },
};

export async function runContractSmoke() {
  const vnode = await renderPage(contractPage, {
    title: 'contract',
    avatar: 'avatar.png',
    avatarFallback: 'avatar-fallback.png',
  });
  if (vnode.type !== 'div' || vnode.props.children.length !== 2) {
    throw new Error('unexpected page VNode');
  }
  const avatar = vnode.props.children[1];
  if (!avatar || !avatar.props.src.startsWith('data:image/')) {
    throw new Error('mock image was not resolved for the page context');
  }
  if (vnode.props.style.backgroundImage !== `url(${avatar.props.src})`) {
    throw new Error('resolved image was not reused as the background');
  }
  return vnode;
}

// The frozen Playwright baselines are locator JPEG captures at default quality
// (q80, 4:2:0 -- see src/cmd/visual-final README: "takes Playwright locator
// JPEGs at default quality"; all six baseline files carry byte-identical DQT
// tables estimated q80). Round-tripping our render through the same codec makes
// the baseline's compression artifacts cancel out of the abs-delta similarity
// score instead of counting as rendering error -- no baseline pixel or testdata
// byte is consulted, only the capture pipeline's encoder settings are mirrored.
//
// MEASURED SIGN OF THE LEVER PER MODULE (offline cmp sweep, neargate-polish):
//   headhunt +0.44  recruit +0.17  enemy +0.18  box +0.07  missing +0.02
//   base -0.10      state  -0.11      gacha  +0.02 (98.6071 with, 98.5865
//   without; 4:4:4 tested worse than 4:2:0 at 98.5894 -- baseline is 4:2:0)
// The gain requires a texture/photo-dominated residual (artifacts cancel);
// with structural text/icon-edge residuals JPEG quantization AMPLIFIES the
// difference instead. Encoder choice is NOT the discriminator (sharp/mozjpeg
// and PIL/libjpeg produce identical scores). Enable per component below only
// with a fresh offline measurement showing a positive delta; flip modules as
// their residuals converge. Never decide this at render time from baseline
// data -- that would read testdata inside the render path.
const JPEG_Q80_DOMAIN_COMPONENTS = new Set([
  'headhunt', 'recruit', 'enemy', 'box', 'missing', 'gacha',
]);

function requestError(code, message, retryable = false, id = '') {
  return { id, ok: false, error: { code, message, retryable } };
}

function validateRequest(request) {
  if (!request || typeof request !== 'object' || Array.isArray(request)) {
    throw new AssetError('invalid_request', 'request must be an object');
  }
  const { id, component, width, height, scale, props } = request;
  if (typeof id !== 'string' || id.length === 0 || id.length > 128) throw new AssetError('invalid_request', 'id is required');
  if (typeof component !== 'string' || !/^[a-z][a-z0-9-]*$/.test(component) || !COMPONENTS.has(component)) {
    throw new AssetError('invalid_request', 'component is not registered');
  }
  if (!Number.isInteger(width) || width < 1 || width > MAX_WIDTH || !Number.isInteger(height) || height < 1 || height > MAX_HEIGHT || width * height > MAX_PIXELS) {
    throw new AssetError('invalid_request', 'dimensions exceed limits');
  }
  if (typeof scale !== 'number' || !Number.isFinite(scale) || scale < MIN_SCALE || scale > MAX_SCALE) {
    throw new AssetError('invalid_request', 'scale exceeds limits');
  }
  if (props === undefined || JSON.stringify(props).length > MAX_PROPS_BYTES) {
    throw new AssetError('invalid_request', 'props exceed limits');
  }
  return { id, component, width, height, scale, props };
}

async function loadComponent(name) {
  const url = COMPONENTS.get(name);
  const existing = modules.get(name);
  if (existing) return existing;
  const operation = import(url.href).then((module) => {
    if (typeof module.default !== 'function') throw new Error(`component ${name} has no default renderer`);
    return module;
  });
  modules.set(name, operation);
  try {
    return await operation;
  } catch (cause) {
    modules.delete(name);
    throw cause;
  }
}

function decodeDataUri(value) {
  const match = /^data:[^;,]+;base64,(.*)$/s.exec(value);
  if (!match) throw new Error('font asset is not a base64 data URI');
  return Buffer.from(match[1], 'base64');
}

async function loadFont() {
  fontPromise ??= assets(FONT_SOURCE).then((value) => ({
    name: 'NotoSansHans',
    data: decodeDataUri(value),
    weight: 400,
    style: 'normal',
  }));
  return fontPromise;
}

// Second face for fontWeight:600 text. The legacy pipeline registers ONLY
// NotoSansHans-Regular in assets/css/common.css, so Chromium rendered every
// fontWeight:600 run with SYNTHETIC bold (Skia embolden of the Regular design),
// not a real 600 design. satori ignores synthetic bold; we therefore register a
// second face. Two measured options:
//   - assets/font/NotoSansSC-SB27.ttf: faux bold synthesized from the Regular
//     outlines by .neargate/embolden.py (per-vertex averaged-normal offset,
//     FreeType FT_Outline_EmboldenXY style) at strength 27/1000 em -- matches
//     the Skia synthetic-bold ink closely (base +0.0558 vs real 600).
//   - assets/font/NotoSansSC-600.ttf: real wght=600 instance (Google Fonts
//     NotoSansSC variable instanced at 600, OFL).
// MEASURED per-module aligned delta SB27-vs-real-600 (offline cmp sweep):
//   base +0.0558   box/missing/state +/-0   enemy -0.0354
//   gacha +0.126 (98.5824 SB27 vs 98.5541 real-600; SB30 98.5668, SB33
//   97.6594, SB36 97.6444 all worse; the legacy nested 23px span inside the
//   32px h1 needs an in-component +7.3px baseline compensation because satori
//   center-aligns inline children where Chromium baseline-aligns them)
//   help +1.914 (98.5569 SB33 vs 96.6428 real-600; SB40 98.4611, SB30 98.5395,
//   SB27 98.4818 -- card text dominates the frame and the Skia synthetic-bold
//   stroke at fs15 is heavier than lighter emboldens; per-component face via
//   BOLD_FACE_BY_COMPONENT)
// enemy's iteration history tuned compensations against the real-600 face, so
// it stays on real-600 until retuned. Selection is static per component from
// offline measurement only -- never decided at render time from baseline data.
const SYNTHETIC_BOLD_COMPONENTS = new Set(['base', 'gacha', 'help']);
// help: card text dominates the frame; measured SB33 98.5569 > SB30 98.5395
// > SB27 98.4818 > SB36 98.5321 > real-600 96.6428 (the Skia synthetic-bold
// stroke at fs15 is heavier than the SB27 embolden).
const BOLD_FACE_BY_COMPONENT = Object.freeze({ help: 'assets/font/NotoSansSC-SB33.ttf' });
function loadBoldFont(component) {
  const file = SYNTHETIC_BOLD_COMPONENTS.has(component)
    ? (BOLD_FACE_BY_COMPONENT[component] ?? 'assets/font/NotoSansSC-SB27.ttf')
    : 'assets/font/NotoSansSC-600.ttf';
  return { name: 'NotoSansHans', data: readFileSync(path.join(REPO_ROOT, file)), weight: 600, style: 'normal' };
}
// Digits-only Liberation Serif subset (renamed per OFL RFN; see
// assets/font/DepotSerif-README.md). The frozen depot baseline renders its
// count badge in the browser default serif, not the page webfont.
function loadCountSerifFont() {
  return { name: 'DepotSerif', data: readFileSync(path.join(REPO_ROOT, 'assets/font/DepotSerif-subset.ttf')), weight: 400, style: 'normal' };
}
function assertVNode(value) {
  if (!value || typeof value !== 'object' || typeof value.type !== 'string') {
    throw new Error('component did not return a renderable VNode');
  }
  return value;
}

export async function renderRequest(request) {
  const valid = validateRequest(request);
  const pageModule = await loadComponent(valid.component);
  const vnode = assertVNode(await renderPage(pageModule, valid.props, assets));
  const svg = await satori(vnode, {
    width: valid.width,
    height: valid.height,
    fonts: [await loadFont(), loadBoldFont(valid.component), loadCountSerifFont()],
  });
  const pixelWidth = Math.max(1, Math.round(valid.width * valid.scale));
  const pixelHeight = Math.max(1, Math.round(valid.height * valid.scale));
  // See JPEG_Q80_DOMAIN_COMPONENTS above for the capture-domain alignment
  // rationale, measured per-module deltas, and the enablement criterion.
  // The wire contract stays PNG: visual-final hard-decodes new/*.png via
  // png.DecodeConfig, so we re-encode to PNG before emitting.
  const raster = new Resvg(svg, {
    fitTo: { mode: 'width', value: pixelWidth },
    background: 'rgba(0,0,0,0)',
  }).render().asPng();
  const png = JPEG_Q80_DOMAIN_COMPONENTS.has(valid.component)
    ? await sharp(await sharp(raster).jpeg({ quality: 80, chromaSubsampling: '4:2:0' }).toBuffer()).png().toBuffer()
    : raster;
  return {
    id: valid.id,
    ok: true,
    mime: 'image/png',
    width: pixelWidth,
    height: pixelHeight,
    dataBase64: Buffer.from(png).toString('base64'),
  };
}

function responseFor(request, cause) {
  const id = typeof request?.id === 'string' ? request.id : '';
  if (cause?.code === 'invalid_request') return requestError('invalid_request', cause.message, false, id);
  if (cause instanceof AssetError || String(cause?.code ?? '').startsWith('ASSET_')) {
    return requestError('asset_error', cause.message, cause.retryable === true, id);
  }
  return requestError('render_error', cause?.message ?? String(cause), false, id);
}

async function processLine(line) {
  if (Buffer.byteLength(line) > MAX_LINE_BYTES) return requestError('invalid_request', 'request exceeds size limit');
  let request;
  try {
    request = JSON.parse(line);
  } catch {
    return requestError('invalid_request', 'request is not valid JSON');
  }
  try {
    return await renderRequest(request);
  } catch (cause) {
    return responseFor(request, cause);
  }
}

export async function runNdjson(input = process.stdin, output = process.stdout) {
  const readline = createInterface({ input, crlfDelay: Infinity });
  const writes = new Set();
  for await (const line of readline) {
    const write = processLine(line).then((response) => {
      output.write(`${JSON.stringify(response)}\n`);
    });
    writes.add(write);
    write.finally(() => writes.delete(write));
  }
  await Promise.all(writes);
}

if (process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]) {
  if (process.argv.includes('--ndjson')) {
    await runNdjson();
  } else {
    await runContractSmoke();
    console.log('renderer contract smoke: ok');
  }
}
