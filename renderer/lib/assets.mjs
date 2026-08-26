import { createHash } from 'node:crypto';
import { readFile, realpath } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

export const DEFAULT_ASSET_LIMITS = Object.freeze({
  maxBytes: 16 * 1024 * 1024,
  maxRemoteBytes: 4 * 1024 * 1024,
  maxCacheEntries: 128,
  cacheTtlMs: 5 * 60 * 1000,
  maxRedirects: 2,
  timeoutMs: 5_000,
  maxSourceLength: 8_192,
  maxDecodedBytes: 16 * 1024 * 1024,
  maxDecodedWidth: 4096,
  maxDecodedHeight: 4096,
  maxDecodedPixels: 16 * 1024 * 1024,
});

const MIME_BY_EXTENSION = Object.freeze({
  '.avif': 'image/avif',
  '.gif': 'image/gif',
  '.jpeg': 'image/jpeg',
  '.jpg': 'image/jpeg',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.webp': 'image/webp',
  '.otf': 'font/otf',
  '.ttf': 'font/ttf',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
});

const ALLOWED_MIME = /^(?:image\/(?:avif|gif|jpeg|png|svg\+xml|webp)|font\/(?:otf|ttf|woff2?))$/;

// Fixture inputs are bounded aliases into the frozen cache, never network URLs.
const FROZEN_FIXTURE_CACHE_ALIASES = Object.freeze({
  'https://fixture-cache.invalid/card/secretary-painting.png': 'card-secretary-1024.png',
  'https://fixture-cache.invalid/card/player-avatar.png': 'card-player-portrait-180x360.png',
  'https://fixture-cache.invalid/operator/amiya-painting.png': 'operator-painting-1024.png',
  'https://fixture-cache.invalid/operator/building-skill-icon.png': 'operator-building-36.png',
  'https://fixture-cache.invalid/operator/skill-icon.png': 'operator-skill-128.png',
  'https://fixture-cache.invalid/enemy/originium-slug.png': 'enemy-originium-slug-158.png',
  'https://fixture-cache.invalid/recruit-amiya.png': 'amiya-avatar.webp',
  'https://fixture-cache.invalid/depot-lmd.png': 'depot-lmd.png',
});

export class AssetError extends Error {
  constructor(code, message, { retryable = false, manifestFatal = false, cause } = {}) {
    super(message, { cause });
    this.name = 'AssetError';
    this.code = code;
    this.retryable = retryable;
    this.manifestFatal = manifestFatal;
  }
}

function error(code, message, options) {
  return new AssetError(code, message, options);
}

function asLimits(options) {
  return {
    ...DEFAULT_ASSET_LIMITS,
    ...(options.limits ?? {}),
  };
}

function assertLimits(limits) {
  for (const key of ['maxBytes', 'maxRemoteBytes', 'maxCacheEntries', 'cacheTtlMs', 'maxRedirects', 'timeoutMs', 'maxSourceLength', 'maxDecodedBytes', 'maxDecodedWidth', 'maxDecodedHeight', 'maxDecodedPixels']) {
    if (!Number.isSafeInteger(limits[key]) || limits[key] < 0) {
      throw new TypeError(`asset limit ${key} must be a non-negative integer`);
    }
  }
}

function mimeForPath(filePath) {
  return MIME_BY_EXTENSION[path.extname(filePath).toLowerCase()] ?? null;
}

function assertMime(mime) {
  const normalized = String(mime ?? '').split(';', 1)[0].trim().toLowerCase();
  if (!ALLOWED_MIME.test(normalized)) {
    throw error('ASSET_MIME_UNSUPPORTED', `unsupported asset MIME type: ${normalized || 'missing'}`);
  }
  return normalized;
}

function dataUri(bytes, mime) {
  return `data:${mime};base64,${Buffer.from(bytes).toString('base64')}`;
}

function detectMime(bytes) {
  const value = Buffer.from(bytes);
  if (value.length >= 8 && value.subarray(0, 8).equals(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]))) return 'image/png';
  if (value.length >= 3 && value.subarray(0, 3).equals(Buffer.from([255, 216, 255]))) return 'image/jpeg';
  if (value.length >= 12 && value.subarray(0, 4).toString('ascii') === 'RIFF' && value.subarray(8, 12).toString('ascii') === 'WEBP') return 'image/webp';
  if (value.length >= 6 && (value.subarray(0, 6).toString('ascii') === 'GIF87a' || value.subarray(0, 6).toString('ascii') === 'GIF89a')) return 'image/gif';
  const text = value.subarray(0, 512).toString('utf8').replace(/^\\uFEFF/, '').trimStart();
  if (/^<svg(?:\\s|>)/i.test(text) || /^<\\?xml[\\s\\S]*<svg(?:\\s|>)/i.test(text)) return 'image/svg+xml';
  return null;
}

function validateImageBytes(bytes, declaredMime) {
  const mime = assertMime(declaredMime);
  if (!mime.startsWith('image/')) return mime;
  const detected = detectMime(bytes);
  if (!detected || detected !== mime) {
    throw error('ASSET_MAGIC_MISMATCH', `asset MIME ${mime} does not match its bytes`);
  }
  return detected;
}

let sharpPromise;
async function normalizeWebp(bytes, limits) {
  sharpPromise ??= import('sharp').then((module) => module.default ?? module);
  let sharp;
  try {
    sharp = await sharpPromise;
  } catch (cause) {
    throw error('ASSET_WEBP_DECODER', 'WebP decoder is unavailable', { retryable: false, cause });
  }
  let image;
  try {
    image = sharp(bytes, { failOn: 'error', limitInputPixels: limits.maxDecodedPixels });
    const metadata = await image.metadata();
    const width = metadata.width ?? 0;
    const height = metadata.height ?? 0;
    if (!width || !height || width > limits.maxDecodedWidth || height > limits.maxDecodedHeight || width * height > limits.maxDecodedPixels) {
      throw error('ASSET_DECODE_LIMIT', 'decoded WebP dimensions exceed limits');
    }
    const result = await image.png({ compressionLevel: 9, adaptiveFiltering: false, palette: false }).toBuffer({ resolveWithObject: true });
    if (result.data.byteLength > limits.maxDecodedBytes) throw error('ASSET_DECODE_LIMIT', 'decoded WebP exceeds byte limit');
    return { bytes: result.data, mime: 'image/png', width, height };
  } catch (cause) {
    if (cause instanceof AssetError) throw cause;
    throw error('ASSET_WEBP_DECODE', 'unable to decode WebP asset', { cause });
  }
}

async function materialize(bytes, mime, limits) {
  const actualMime = validateImageBytes(bytes, mime);
  let output = bytes;
  let outputMime = actualMime;
  if (actualMime === 'image/webp') {
    const normalized = await normalizeWebp(bytes, limits);
    output = normalized.bytes;
    outputMime = normalized.mime;
  }
  return {
    value: dataUri(output, outputMime),
    sha256: createHash('sha256').update(output).digest('hex'),
  };
}

function parseDataUri(source, limits) {
  const match = /^data:([^;,]+)(;base64)?,(.*)$/s.exec(source);
  if (!match) throw error('ASSET_INVALID_SOURCE', 'invalid data URI');
  const mime = assertMime(match[1]);
  let bytes;
  try {
    bytes = match[2]
      ? Buffer.from(match[3].replace(/\\s/g, ''), 'base64')
      : Buffer.from(decodeURIComponent(match[3]));
  } catch (cause) {
    throw error('ASSET_INVALID_SOURCE', 'invalid data URI payload', { cause });
  }
  if (bytes.byteLength > limits.maxBytes) throw error('ASSET_TOO_LARGE', `asset exceeds ${limits.maxBytes} bytes`);
  return { bytes, mime };
}

function isInside(root, candidate) {
  const relative = path.relative(root, candidate);
  return relative === '' || (relative !== '..' && !relative.startsWith(`..${path.sep}`) && !path.isAbsolute(relative));
}

async function resolveLocalPath(source, root) {
  let candidate;
  if (source.startsWith('file:')) {
    try {
      candidate = fileURLToPath(new URL(source));
    } catch (cause) {
      throw error('ASSET_INVALID_SOURCE', 'invalid local file URL', { cause });
    }
  } else {
    const relative = source.replace(/^\/+/, '');
    candidate = path.resolve(root, relative);
  }
  if (!isInside(root, candidate)) {
    throw error('ASSET_LOCAL_ESCAPE', 'local asset path escapes repository root');
  }
  let resolved;
  try {
    resolved = await realpath(candidate);
  } catch (cause) {
    throw error('ASSET_NOT_FOUND', `local asset not found: ${source}`, { cause });
  }
  if (!isInside(root, resolved)) {
    throw error('ASSET_LOCAL_ESCAPE', 'local asset symlink escapes repository root');
  }
  return resolved;
}

async function readLocal(source, root, limits) {
  const filePath = await resolveLocalPath(source, root);
  const mime = assertMime(mimeForPath(filePath));
  let bytes;
  try {
    bytes = await readFile(filePath);
  } catch (cause) {
    throw error('ASSET_READ_FAILED', `unable to read local asset: ${source}`, { cause });
  }
  if (bytes.byteLength > limits.maxBytes) throw error('ASSET_TOO_LARGE', `asset exceeds ${limits.maxBytes} bytes`);
  if (mime.startsWith('image/')) validateImageBytes(bytes, mime);
  return { bytes, mime };
}

async function readResponseBytes(response, limits) {
  const contentLength = Number(response.headers?.get?.('content-length'));
  if (Number.isFinite(contentLength) && contentLength > limits.maxRemoteBytes) {
    throw error('ASSET_TOO_LARGE', `remote asset exceeds ${limits.maxRemoteBytes} bytes`);
  }
  if (!response.body?.getReader) {
    const bytes = Buffer.from(await response.arrayBuffer());
    if (bytes.byteLength > limits.maxRemoteBytes) throw error('ASSET_TOO_LARGE', `asset exceeds ${limits.maxRemoteBytes} bytes`);
    return bytes;
  }
  const reader = response.body.getReader();
  const chunks = [];
  let total = 0;
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      total += value.byteLength;
      if (total > limits.maxRemoteBytes) {
        await reader.cancel();
        throw error('ASSET_TOO_LARGE', `remote asset exceeds ${limits.maxRemoteBytes} bytes`);
      }
      chunks.push(Buffer.from(value));
    }
  } finally {
    reader.releaseLock();
  }
  return Buffer.concat(chunks, total);
}

async function readRemote(source, limits, fetchImpl) {
  let current;
  try {
    current = new URL(source);
  } catch (cause) {
    throw error('ASSET_INVALID_SOURCE', `invalid remote URL: ${source}`, { cause });
  }
  if (current.protocol !== 'https:') throw error('ASSET_REMOTE_PROTOCOL', 'remote assets require HTTPS');

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), limits.timeoutMs);
  timer.unref?.();
  try {
    for (let redirect = 0; ; redirect += 1) {
      let response;
      try {
        response = await fetchImpl(current, { redirect: 'manual', signal: controller.signal });
      } catch (cause) {
        if (cause?.name === 'AbortError' || controller.signal.aborted) {
          throw error('ASSET_TIMEOUT', `remote asset timed out after ${limits.timeoutMs}ms`, { retryable: true, cause });
        }
        throw error('ASSET_FETCH_FAILED', `unable to fetch remote asset: ${source}`, { retryable: true, cause });
      }
      if (response.status >= 300 && response.status < 400) {
        if (redirect >= limits.maxRedirects) throw error('ASSET_REDIRECT_LIMIT', 'remote asset redirect limit exceeded');
        const location = response.headers?.get?.('location');
        if (!location) throw error('ASSET_REDIRECT_INVALID', 'remote asset redirect has no location');
        current = new URL(location, current);
        if (current.protocol !== 'https:') throw error('ASSET_REMOTE_PROTOCOL', 'remote asset redirects require HTTPS');
        continue;
      }
      if (!response.ok) {
        throw error('ASSET_HTTP', `remote asset returned HTTP ${response.status}`, {
          retryable: response.status >= 500,
        });
      }
      const mime = assertMime(response.headers?.get?.('content-type'));
      const bytes = await readResponseBytes(response, limits);
      if (mime.startsWith('image/')) validateImageBytes(bytes, mime);
      return { bytes, mime };
    }
  } finally {
    clearTimeout(timer);
  }
}

async function loadManifest(manifestSource, root) {
  const candidate = path.resolve(root, manifestSource);
  if (!isInside(root, candidate)) throw error('ASSET_MANIFEST_PATH', 'asset manifest escapes repository root', { manifestFatal: true });
  let manifestPath;
  try {
    manifestPath = await realpath(candidate);
  } catch (cause) {
    throw error('ASSET_MANIFEST_MISSING', 'asset manifest does not exist', { manifestFatal: true, cause });
  }
  if (!isInside(root, manifestPath)) throw error('ASSET_MANIFEST_PATH', 'asset manifest escapes repository root', { manifestFatal: true });
  let manifest;
  try {
    manifest = JSON.parse(await readFile(manifestPath, 'utf8'));
  } catch (cause) {
    throw error('ASSET_MANIFEST_INVALID', 'asset manifest is not valid JSON', { manifestFatal: true, cause });
  }
  if (manifest?.status !== 'frozen' || !Array.isArray(manifest.resources) || manifest.resources.length !== 26) {
    throw error('ASSET_MANIFEST_INVALID', 'asset manifest must be frozen and contain 26 resources', { manifestFatal: true });
  }
  const aliases = new Map();
  const cacheEntries = new Map();
  const cacheRoot = path.dirname(manifestPath);
  for (const entry of manifest.resources) {
    if (!entry || typeof entry.sourceURL !== 'string' || typeof entry.requestAlias !== 'string' || typeof entry.cachePath !== 'string' || typeof entry.sha256 !== 'string' || !Number.isSafeInteger(entry.bytes) || entry.bytes < 0) {
      throw error('ASSET_MANIFEST_INVALID', 'asset manifest entry is malformed', { manifestFatal: true });
    }
    let mime;
    try {
      mime = assertMime(entry.mime);
    } catch (cause) {
      throw error('ASSET_MANIFEST_MIME', `manifest MIME is unsupported: ${entry.mime}`, { manifestFatal: true, cause });
    }
    const cacheCandidate = path.resolve(cacheRoot, entry.cachePath);
    if (!isInside(cacheRoot, cacheCandidate) || !isInside(root, cacheCandidate)) {
      throw error('ASSET_MANIFEST_PATH', `manifest cache path escapes root: ${entry.cachePath}`, { manifestFatal: true });
    }
    let cachePath;
    try {
      cachePath = await realpath(cacheCandidate);
    } catch (cause) {
      throw error('ASSET_MANIFEST_CACHE_MISSING', `manifest cache is missing: ${entry.cachePath}`, { manifestFatal: true, cause });
    }
    if (!isInside(cacheRoot, cachePath) || !isInside(root, cachePath)) {
      throw error('ASSET_MANIFEST_PATH', `manifest cache path escapes root: ${entry.cachePath}`, { manifestFatal: true });
    }
    let bytes;
    try {
      bytes = await readFile(cachePath);
    } catch (cause) {
      throw error('ASSET_MANIFEST_CACHE_READ', `unable to read manifest cache: ${entry.cachePath}`, { manifestFatal: true, cause });
    }
    const digest = createHash('sha256').update(bytes).digest('hex');
    if (bytes.byteLength !== entry.bytes || digest !== entry.sha256.toLowerCase()) {
      throw error('ASSET_MANIFEST_HASH', `manifest cache integrity mismatch: ${entry.cachePath}`, { manifestFatal: true });
    }
    try {
      validateImageBytes(bytes, mime);
    } catch (cause) {
      throw error('ASSET_MANIFEST_MAGIC', `manifest cache MIME/magic mismatch: ${entry.cachePath}`, { manifestFatal: true, cause });
    }
    const resolved = {
      bytes,
      mime,
      cachePath,
      sha256: entry.sha256.toLowerCase(),
      byteLength: entry.bytes,
      canonicalSource: entry.requestAlias,
      provenance: 'frozen-manifest',
    };
    cacheEntries.set(path.basename(cachePath), resolved);
    for (const alias of [entry.sourceURL, entry.requestAlias]) {
      const prior = aliases.get(alias);
      if (prior && (prior.sha256 !== entry.sha256.toLowerCase() || prior.byteLength !== entry.bytes || prior.mime !== mime)) {
        throw error('ASSET_MANIFEST_INVALID', `manifest alias maps to conflicting content: ${alias}`, { manifestFatal: true });
      }
      if (!prior) aliases.set(alias, resolved);
    }
  }
  for (const [alias, cacheName] of Object.entries(FROZEN_FIXTURE_CACHE_ALIASES)) {
    const resolved = cacheEntries.get(cacheName);
    if (!resolved) throw error('ASSET_MANIFEST_INVALID', `fixture alias target is absent: ${cacheName}`, { manifestFatal: true });
    aliases.set(alias, { ...resolved, provenance: 'frozen-manifest-fixture-alias' });
  }
  return aliases;
}


function recordDiagnostic(diagnostics, onDiagnostic, entry) {
  diagnostics.push(entry);
  while (diagnostics.length > 64) diagnostics.shift();
  onDiagnostic?.(entry);
}

export function createAssetLoader(options = {}) {
  const limits = asLimits(options);
  assertLimits(limits);
  const root = path.resolve(options.repoRoot ?? process.cwd());
  const fetchImpl = options.fetch ?? globalThis.fetch;
  if (typeof fetchImpl !== 'function') throw new TypeError('fetch implementation is required for remote assets');
  const manifestSource = options.manifestPath ?? process.env.SATORI_ASSET_MANIFEST ?? null;
  const manifestPromise = manifestSource == null ? Promise.resolve(null) : loadManifest(manifestSource, root);
  const cache = new Map();
  const pending = new Map();
  const diagnostics = [];
  const materializations = [];
  let failures = 0;
  let fallbacks = 0;

  const prune = () => {
    if (limits.maxCacheEntries === 0) return;
    const now = Date.now();
    for (const [key, item] of cache) {
      if (item.expiresAt <= now) cache.delete(key);
    }
    while (cache.size >= limits.maxCacheEntries && cache.size > 0) cache.delete(cache.keys().next().value);
  };

  const load = async (src, fallback) => {
    const primary = src == null || src === '' ? null : String(src);
    const backup = fallback == null || fallback === '' ? null : String(fallback);
    if (!primary && !backup) throw error('ASSET_SOURCE_MISSING', 'image source or fallback is required');
    const resolve = async (source) => {
      if (source.length > limits.maxSourceLength) throw error('ASSET_INVALID_SOURCE', 'asset source is too long');
      const cached = cache.get(source);
      if (cached && cached.expiresAt > Date.now()) {
        cache.delete(source);
        cache.set(source, cached);
        return cached.value;
      }
      const existing = pending.get(source);
      if (existing) return existing;
      const operation = (async () => {
        const manifest = await manifestPromise;
        let result;
        const mapped = manifest?.get(source);
        if (mapped) {
          const materialized = await materialize(mapped.bytes, mapped.mime, limits);
          result = { ...materialized, provenance: mapped.provenance, canonicalSource: mapped.canonicalSource, cachePath: mapped.cachePath, manifestSource };
        } else if (manifest && /^[a-z][a-z\\d+.-]*:\/\//i.test(source)) {
          throw error('ASSET_MANIFEST_MISSING', 'remote asset is absent from the frozen manifest', { manifestFatal: true });
        } else if (source.startsWith('data:')) {
          const parsed = parseDataUri(source, limits);
          result = { ...(await materialize(parsed.bytes, parsed.mime, limits)), provenance: 'data-uri', canonicalSource: source, manifestSource };
        } else if (source.startsWith('https://')) {
          const remote = await readRemote(source, limits, fetchImpl);
          result = { ...(await materialize(remote.bytes, remote.mime, limits)), provenance: 'remote-network', canonicalSource: source, manifestSource };
        } else if (/^[a-z][a-z\\d+.-]*:/i.test(source) && !source.startsWith('file:')) {
          throw error('ASSET_REMOTE_PROTOCOL', 'only HTTPS remote assets are allowed');
        } else {
          const local = await readLocal(source, root, limits);
          result = { ...(await materialize(local.bytes, local.mime, limits)), provenance: 'repository-local', canonicalSource: source, manifestSource };
        }
        const entry = {
          kind: 'asset_materialized',
          source,
          canonicalSource: result.canonicalSource,
          materializedSha256: result.sha256,
          provenance: result.provenance,
          manifestSource: result.manifestSource,
          cachePath: result.cachePath,
        };
        materializations.push(entry);
        options.onDiagnostic?.(entry);
        if (limits.maxCacheEntries > 0) {
          prune();
          cache.set(source, { value: result, expiresAt: Date.now() + limits.cacheTtlMs });
        }
        return result;
      })();
      pending.set(source, operation);
      try {
        return await operation;
      } finally {
        pending.delete(source);
      }
    };

    try {
      const result = await resolve(primary ?? backup);
      return result.value;
    } catch (cause) {
      failures += 1;
      if (!backup || backup === primary || cause.manifestFatal) throw cause;
      try {
        const result = await resolve(backup);
        fallbacks += 1;
        recordDiagnostic(diagnostics, options.onDiagnostic, {
          kind: 'asset_fallback',
          source: primary,
          fallback: backup,
          canonicalSource: result.canonicalSource,
          materializedSha256: result.sha256,
          provenance: result.provenance,
          manifestSource: result.manifestSource,
          cachePath: result.cachePath,
          code: cause.code ?? 'ASSET_FAILED',
          message: cause.message,
          retryable: cause.retryable === true,
          usedFallback: true,
        });
        return result.value;
      } catch (fallbackCause) {
        throw error('ASSET_FALLBACK_FAILED', `asset and fallback failed: ${primary}`, {
          retryable: cause.retryable === true || fallbackCause.retryable === true,
          cause: fallbackCause,
        });
      }
    }
  };

  load.clearCache = () => cache.clear();
  load.ready = () => manifestPromise;
  load.diagnostics = () => diagnostics.slice();
  load.materializations = () => materializations.slice();
  load.stats = () => ({ cacheEntries: cache.size, pending: pending.size, diagnostics: diagnostics.length, materialized: materializations.length, failures, fallbacks, manifestSource });
  return load;
}

export async function loadAsset(source, fallback, options = {}) {
  return createAssetLoader(options)(source, fallback);
}
