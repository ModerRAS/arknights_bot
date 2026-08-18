import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { test } from 'node:test';
import { createAssetLoader } from './lib/assets.mjs';

const repoRoot = path.resolve(fileURLToPath(new URL('..', import.meta.url)));
const manifestPath = path.join(repoRoot, 'src/utils/media/testdata/visual/baseline/resource-manifest.json');
const manifest = JSON.parse(await readFile(manifestPath, 'utf8'));
const pngBytes = await readFile(path.join(repoRoot, 'assets/calendar/bg.png'));
const pngUri = `data:image/png;base64,${pngBytes.toString('base64')}`;

function decode(uri) {
  assert.match(uri, /^data:image\/(?:png|jpeg);base64,/);
  return Buffer.from(uri.slice(uri.indexOf(',') + 1), 'base64');
}

test('frozen manifest has 26 exact cache entries and aliases hit without fetch', async () => {
  let fetches = 0;
  const load = createAssetLoader({
    repoRoot,
    manifestPath,
    fetch: async () => {
      fetches += 1;
      throw new Error('network must not be used for a manifest alias');
    },
  });
  assert.equal(manifest.resources.length, 26);
  assert.equal(new Set(manifest.resources.map((entry) => entry.cachePath)).size, 26);
  for (const entry of manifest.resources) {
    const value = await load(entry.requestAlias);
    assert.match(value, /^data:image\/(?:png|jpeg);base64,/);
    assert.ok(decode(value).byteLength > 0);
  }
  assert.equal(fetches, 0);
  assert.equal(load.stats().cacheEntries, new Set(manifest.resources.map((entry) => entry.requestAlias)).size);
});

test('frozen WebP aliases normalize to lossless PNG with expected dimensions and cache', async () => {
  const sharp = (await import('sharp')).default;
  let fetches = 0;
  const load = createAssetLoader({
    repoRoot,
    manifestPath,
    fetch: async () => {
      fetches += 1;
      throw new Error('network must not be used for a frozen WebP');
    },
  });
  const webps = manifest.resources.filter((entry) => entry.mime === 'image/webp');
  assert.ok(webps.length >= 4);
  const dimensions = new Set();
  for (const entry of webps) {
    const value = await load(entry.requestAlias);
    const bytes = decode(value);
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    dimensions.add(`${metadata.width}x${metadata.height}`);
    assert.ok(bytes.byteLength > 0);
    assert.equal(await load(entry.requestAlias), value);
  }
  assert.ok(dimensions.has('180x180'));
  assert.ok(dimensions.has('180x360'));
  assert.equal(fetches, 0);
});

test('manifest PNG and local JPEG paths pass through without re-encoding', async () => {
  const load = createAssetLoader({ repoRoot, manifestPath });
  const pngEntry = manifest.resources.find((entry) => entry.mime === 'image/png');
  const rawManifestPng = await readFile(path.join(path.dirname(manifestPath), pngEntry.cachePath));
  assert.equal(await load(pngEntry.requestAlias), `data:image/png;base64,${rawManifestPng.toString('base64')}`);
  const jpegPath = 'src/utils/media/testdata/visual/baseline/card-legacy.jpg';
  const rawJpeg = await readFile(path.join(repoRoot, jpegPath));
  assert.equal(await createAssetLoader({ repoRoot })(jpegPath), `data:image/jpeg;base64,${rawJpeg.toString('base64')}`);
});

async function makeManifestFixture() {
  const dir = await mkdtemp(path.join(repoRoot, 'renderer', '.manifest-test-'));
  const cache = path.join(dir, 'cache');
  await mkdir(cache);
  const entries = [];
  for (let index = 0; index < 26; index += 1) {
    const name = `fixture-${index}.png`;
    const bytes = pngBytes;
    await writeFile(path.join(cache, name), bytes);
    entries.push({
      sourceURL: `https://fixture.test/source/${index}.png`,
      requestAlias: `https://fixture.test/alias/${index}.png`,
      cachePath: `cache/${name}`,
      mime: 'image/png',
      bytes: bytes.byteLength,
      sha256: createHash('sha256').update(bytes).digest('hex'),
    });
  }
  const file = path.join(dir, 'resource-manifest.json');
  await writeFile(file, JSON.stringify({ status: 'frozen', resources: entries }));
  return { dir, file, entries };
}

test('manifest hash/path/missing aliases fail closed without network or fallback', async () => {
  const fixture = await makeManifestFixture();
  try {
    const hashManifest = JSON.parse(await readFile(fixture.file, 'utf8'));
    hashManifest.resources[0].sha256 = '0'.repeat(64);
    const hashPath = path.join(fixture.dir, 'hash.json');
    await writeFile(hashPath, JSON.stringify(hashManifest));
    let fetches = 0;
    const hashLoader = createAssetLoader({ repoRoot, manifestPath: hashPath, fetch: async () => { fetches += 1; } });
    await assert.rejects(hashLoader(fixture.entries[0].requestAlias), (cause) => cause.code === 'ASSET_MANIFEST_HASH' && cause.manifestFatal);
    assert.equal(fetches, 0);

    const pathManifest = JSON.parse(await readFile(fixture.file, 'utf8'));
    pathManifest.resources[0].cachePath = '../package.json';
    const pathFile = path.join(fixture.dir, 'path.json');
    await writeFile(pathFile, JSON.stringify(pathManifest));
    const pathLoader = createAssetLoader({ repoRoot, manifestPath: pathFile, fetch: async () => { fetches += 1; } });
    await assert.rejects(pathLoader(fixture.entries[0].requestAlias), (cause) => cause.code === 'ASSET_MANIFEST_PATH' && cause.manifestFatal);
    assert.equal(fetches, 0);

    const mimeManifest = JSON.parse(await readFile(fixture.file, 'utf8'));
    mimeManifest.resources[0].mime = 'text/plain';
    const mimeFile = path.join(fixture.dir, 'mime.json');
    await writeFile(mimeFile, JSON.stringify(mimeManifest));
    const mimeLoader = createAssetLoader({ repoRoot, manifestPath: mimeFile, fetch: async () => { fetches += 1; } });
    await assert.rejects(mimeLoader(fixture.entries[0].requestAlias, pngUri), (cause) => cause.code === 'ASSET_MANIFEST_MIME' && cause.manifestFatal);
    assert.equal(fetches, 0);

    const missingCacheManifest = JSON.parse(await readFile(fixture.file, 'utf8'));
    missingCacheManifest.resources[0].cachePath = 'cache/missing.png';
    const missingCacheFile = path.join(fixture.dir, 'missing-cache.json');
    await writeFile(missingCacheFile, JSON.stringify(missingCacheManifest));
    const missingCacheLoader = createAssetLoader({ repoRoot, manifestPath: missingCacheFile, fetch: async () => { fetches += 1; } });
    await assert.rejects(missingCacheLoader(fixture.entries[0].requestAlias), (cause) => cause.code === 'ASSET_MANIFEST_CACHE_MISSING' && cause.manifestFatal);
    assert.equal(fetches, 0);

    const missingLoader = createAssetLoader({ repoRoot, manifestPath: fixture.file, fetch: async () => { fetches += 1; } });
    await assert.rejects(missingLoader('https://fixture.test/not-listed.png', pngUri), (cause) => cause.code === 'ASSET_MANIFEST_MISSING' && cause.manifestFatal);
    assert.equal(fetches, 0);
  } finally {
    await rm(fixture.dir, { recursive: true, force: true });
  }
});

test('zero-entry cache remains bounded and does not retain public values', async () => {
  let fetches = 0;
  const load = createAssetLoader({
    repoRoot,
    limits: { maxCacheEntries: 0 },
    fetch: async () => {
      fetches += 1;
      return new Response(pngBytes, { headers: { 'content-type': 'image/png' } });
    },
  });
  const url = 'https://fixture.test/no-cache.png';
  assert.equal(await load(url), pngUri);
  assert.equal(await load(url), pngUri);
  assert.equal(fetches, 2);
  assert.equal(load.stats().cacheEntries, 0);
});

test('without a manifest, HTTPS behavior remains fetch-limited and cached', async () => {
  let fetches = 0;
  const load = createAssetLoader({
    repoRoot,
    fetch: async () => {
      fetches += 1;
      return new Response(pngBytes, { headers: { 'content-type': 'image/png' } });
    },
  });
  const url = 'https://fixture.test/no-manifest.png';
  const value = await load(url);
  assert.equal(value, pngUri);
  assert.equal(await load(url), value);
  assert.equal(fetches, 1);
});
