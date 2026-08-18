import { createInterface } from 'node:readline';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
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
  onDiagnostic: (entry) => process.stderr.write(`[asset_fallback] ${JSON.stringify(entry)}\n`),
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
    fonts: [await loadFont()],
  });
  const pixelWidth = Math.max(1, Math.round(valid.width * valid.scale));
  const pixelHeight = Math.max(1, Math.round(valid.height * valid.scale));
  const png = new Resvg(svg, {
    fitTo: { mode: 'width', value: pixelWidth },
    background: 'rgba(0,0,0,0)',
  }).render().asPng();
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
