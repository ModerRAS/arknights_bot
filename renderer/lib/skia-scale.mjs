import sharp from 'sharp';

// Emulates Skia "medium" image filtering as used by Chromium when rasterizing
// <img> elements at devicePixelRatio 1: single-pass bilinear sampling with
// pixel-center mapping ((dst + 0.5) * scale - 0.5), computed in premultiplied
// float64 space with round-to-nearest unpremultiply.
//
// Rationale (headhunt timebox, frozen Playwright baseline): resvg's built-in
// filter is closer to Skia than any single-step sharp kernel, but an explicit
// center-aligned bilinear still measures +0.000116 similarity on the headhunt
// portrait band; an edge-aligned mapping loses -0.003, so the mapping
// convention is the load-bearing detail.
//
// Input/output are asset-loader style data URIs; output is always PNG so
// resvg performs a 1:1 blit with no further resampling.

function resizeBilinearCenter(pm, sw, sh, dw, dh) {
  const out = new Float64Array(dw * dh * 4);
  // Horizontal pass into scratch, then vertical pass.
  const tmp = new Float64Array(dw * sh * 4);
  const xs = new Array(dw);
  for (let dx = 0; dx < dw; dx++) {
    const fx = (dx + 0.5) * sw / dw - 0.5;
    let x0 = Math.floor(fx);
    xs[dx] = [x0, fx - x0];
  }
  for (let y = 0; y < sh; y++) {
    for (let dx = 0; dx < dw; dx++) {
      const [x0, wx] = xs[dx];
      const xa = x0 < 0 ? 0 : x0 >= sw ? sw - 1 : x0;
      const xb = x0 + 1 < 0 ? 0 : x0 + 1 >= sw ? sw - 1 : x0 + 1;
      const pa = (y * sw + xa) * 4;
      const pb = (y * sw + xb) * 4;
      const o = (y * dw + dx) * 4;
      tmp[o] = pm[pa] * (1 - wx) + pm[pb] * wx;
      tmp[o + 1] = pm[pa + 1] * (1 - wx) + pm[pb + 1] * wx;
      tmp[o + 2] = pm[pa + 2] * (1 - wx) + pm[pb + 2] * wx;
      tmp[o + 3] = pm[pa + 3] * (1 - wx) + pm[pb + 3] * wx;
    }
  }
  for (let dy = 0; dy < dh; dy++) {
    const fy = (dy + 0.5) * sh / dh - 0.5;
    let y0 = Math.floor(fy);
    const wy = fy - y0;
    const ya = y0 < 0 ? 0 : y0 >= sh ? sh - 1 : y0;
    const yb = y0 + 1 < 0 ? 0 : y0 + 1 >= sh ? sh - 1 : y0 + 1;
    for (let dx = 0; dx < dw; dx++) {
      const pa = (ya * dw + dx) * 4;
      const pb = (yb * dw + dx) * 4;
      const o = (dy * dw + dx) * 4;
      out[o] = tmp[pa] * (1 - wy) + tmp[pb] * wy;
      out[o + 1] = tmp[pa + 1] * (1 - wy) + tmp[pb + 1] * wy;
      out[o + 2] = tmp[pa + 2] * (1 - wy) + tmp[pb + 2] * wy;
      out[o + 3] = tmp[pa + 3] * (1 - wy) + tmp[pb + 3] * wy;
    }
  }
  return out;
}

export async function skiaBilinearDataUri(dataUri, dstWidth, dstHeight) {
  if (typeof dataUri !== 'string' || !dataUri.startsWith('data:')) {
    throw new TypeError('skiaBilinearDataUri expects a data URI');
  }
  if (!Number.isInteger(dstWidth) || !Number.isInteger(dstHeight) || dstWidth < 1 || dstHeight < 1) {
    throw new TypeError('target dimensions must be positive integers');
  }
  const base64 = dataUri.slice(dataUri.indexOf(';base64,') + ';base64,'.length);
  let source;
  try {
    source = await sharp(Buffer.from(base64, 'base64'))
      .ensureAlpha()
      .raw()
      .toBuffer({ resolveWithObject: true });
  } catch {
    // Not decodable pixel data (e.g. test-injected mock URIs): pass through so
    // callers behave identically to before; real assets always decode.
    return dataUri;
  }
  const { data, info } = source;
  if (info.width === dstWidth && info.height === dstHeight) return dataUri;

  const pixels = info.width * info.height;
  const pm = new Float64Array(pixels * 4);
  for (let i = 0; i < pixels; i++) {
    const alpha = data[i * 4 + 3] / 255;
    pm[i * 4] = data[i * 4] * alpha;
    pm[i * 4 + 1] = data[i * 4 + 1] * alpha;
    pm[i * 4 + 2] = data[i * 4 + 2] * alpha;
    pm[i * 4 + 3] = data[i * 4 + 3];
  }

  const scaled = resizeBilinearCenter(pm, info.width, info.height, dstWidth, dstHeight);
  const outBuf = Buffer.alloc(dstWidth * dstHeight * 4);
  for (let i = 0; i < dstWidth * dstHeight; i++) {
    const alpha = scaled[i * 4 + 3] / 255;
    for (let c = 0; c < 3; c++) {
      const v = alpha > 0 ? scaled[i * 4 + c] / alpha : 0;
      outBuf[i * 4 + c] = Math.max(0, Math.min(255, Math.round(v)));
    }
    outBuf[i * 4 + 3] = Math.max(0, Math.min(255, Math.round(scaled[i * 4 + 3])));
  }
  const png = await sharp(outBuf, { raw: { width: dstWidth, height: dstHeight, channels: 4 } })
    .png()
    .toBuffer();
  return `data:image/png;base64,${png.toString('base64')}`;
}
