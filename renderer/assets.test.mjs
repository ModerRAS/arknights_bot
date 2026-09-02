import assert from 'node:assert/strict';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { AssetError, createAssetLoader } from './lib/assets.mjs';

const repoRoot = fileURLToPath(new URL('..', import.meta.url));
const fixture = 'assets/calendar/bg.png';
const tinyPng = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=', 'base64');

function loader(options = {}) {
  return createAssetLoader({ repoRoot, ...options });
}

test('loads repository-local images and fonts as data URIs', async () => {
  const load = loader();
  const image = await load(fixture);
  assert.match(image, /^data:image\/png;base64,/);
  const font = await load('assets/font/NotoSansHans-Regular.ttf');
  assert.match(font, /^data:font\/ttf;base64,/);
  assert.equal((await load('/assets/calendar/bg.png')), image);
});

test('rejects local traversal and unsupported local files', async () => {
  const load = loader();
  await assert.rejects(load('../package.json'), (cause) => cause.code === 'ASSET_LOCAL_ESCAPE');
  await assert.rejects(load('renderer/package.json'), (cause) => cause.code === 'ASSET_MIME_UNSUPPORTED');
});

test('accepts bounded data URIs and rejects oversized payloads', async () => {
  const load = loader({ limits: { maxBytes: 100 } });
  const uri = `data:image/png;base64,${tinyPng.toString('base64')}`;
  assert.equal(await load(uri), uri);
  await assert.rejects(load(`data:image/png;base64,${Buffer.concat([tinyPng, Buffer.alloc(100)]).toString('base64')}`), (cause) => cause.code === 'ASSET_TOO_LARGE');
});

test('requires HTTPS and validates remote MIME and redirect targets', async () => {
  const load = loader({
    fetch: async () => new Response('ok', { headers: { 'content-type': 'text/plain' } }),
  });
  await assert.rejects(load('http://example.test/image.png'), (cause) => cause.code === 'ASSET_REMOTE_PROTOCOL');
  await assert.rejects(load('https://example.test/image.png'), (cause) => cause.code === 'ASSET_MIME_UNSUPPORTED');

  const redirected = loader({
    fetch: async () => new Response(null, {
      status: 302,
      headers: { location: 'http://example.test/image.png' },
    }),
  });
  await assert.rejects(redirected('https://example.test/image.png'), (cause) => cause.code === 'ASSET_REMOTE_PROTOCOL');
});

test('bounds remote bytes and deduplicates concurrent fetches', async () => {
  let calls = 0;
  const load = loader({
    fetch: async (url) => {
      calls += 1;
      return new Response(tinyPng, {
        headers: { 'content-type': 'image/png' },
      });
    },
  });
  const [one, two] = await Promise.all([
    load('https://example.test/image.png'),
    load('https://example.test/image.png'),
  ]);
  assert.equal(one, two);
  assert.equal(calls, 1);
  const limited = loader({
    limits: { maxRemoteBytes: 3 },
    fetch: async () => new Response(new Uint8Array([1, 2, 3, 4]), {
      headers: { 'content-type': 'image/png', 'content-length': '4' },
    }),
  });
  await assert.rejects(limited('https://example.test/too-large.png'), (cause) => cause.code === 'ASSET_TOO_LARGE');
});

test('uses fallback and retains structured diagnostics', async () => {
  const diagnostics = [];
  const load = loader({ onDiagnostic: (entry) => diagnostics.push(entry) });
  const value = await load('assets/does-not-exist.png', fixture);
  assert.match(value, /^data:image\/png;base64,/);
  assert.deepEqual(diagnostics.map((entry) => entry.kind), ['asset_materialized', 'asset_fallback']);
  assert.equal(diagnostics[0].source, fixture);
  assert.match(diagnostics[0].materializedSha256, /^[a-f0-9]{64}$/);
  const fallbackDiagnostics = diagnostics.filter((entry) => entry.kind === 'asset_fallback');
  assert.equal(fallbackDiagnostics.length, 1);
  assert.equal(fallbackDiagnostics[0].source, 'assets/does-not-exist.png');
  assert.equal(fallbackDiagnostics[0].code, 'ASSET_NOT_FOUND');
  assert.equal(fallbackDiagnostics[0].usedFallback, true);
  assert.equal(fallbackDiagnostics[0].provenance, 'repository-local');
  assert.match(fallbackDiagnostics[0].materializedSha256, /^[a-f0-9]{64}$/);
  assert.equal(load.diagnostics().length, 1);
  assert.equal(load.stats().failures, 1);
  assert.equal(load.stats().fallbacks, 1);
});

test('reports fallback failure without leaking a VNode or raw response', async () => {
  const load = loader();
  await assert.rejects(load('assets/nope.png', 'assets/also-nope.png'), (cause) => {
    assert.ok(cause instanceof AssetError);
    assert.equal(cause.code, 'ASSET_FALLBACK_FAILED');
    assert.equal(cause.retryable, false);
    return true;
  });
});
