import { h } from '../lib/h.mjs';

// Capture-domain image alignment: the frozen baseline was captured by Chromium
// at DPR 1.5, which upscales the 1000px-wide header/footer PNGs with Skia
// bilinear filtering. The resident renderer's SVG rasterizer uses a different
// (sharper) kernel, so pre-upscaling to the exact device resolution with
// bilinear here makes the embedded bitmap 1:1 at the frozen DPR. The DPR is
// frozen per module in the baseline manifest (gacha: 1.5); no testdata is read.
import sharp from 'sharp';
const dataUriToBuffer = (uri) => Buffer.from(String(uri).replace(/^data:[^;,]+;base64,/, ''), 'base64');
const toDataUri = (buf) => `data:image/png;base64,${buf.toString('base64')}`;
const prescaleCache = new Map();
// Skia-medium-quality upscale is bilinear (triangle); sharp has no triangle
// kernel, so the 1.5x upscale is done manually on raw RGBA.
const bilinearResize = async (uri, width, height) => {
  const raw = await sharp(dataUriToBuffer(uri), { failOn: 'error' })
    .ensureAlpha()
    .raw()
    .toBuffer({ resolveWithObject: true });
  const { width: sw, height: sh, channels } = raw.info;
  const out = Buffer.alloc(width * height * channels);
  const sx = sw / width;
  const sy = sh / height;
  for (let y = 0; y < height; y++) {
    const fy = Math.min((y + 0.5) * sy - 0.5, sh - 1);
    const y0 = Math.max(0, Math.floor(fy));
    const y1 = Math.min(sh - 1, y0 + 1);
    const wy = fy - y0;
    for (let x = 0; x < width; x++) {
      const fx = Math.min((x + 0.5) * sx - 0.5, sw - 1);
      const x0 = Math.max(0, Math.floor(fx));
      const x1 = Math.min(sw - 1, x0 + 1);
      const wx = fx - x0;
      const di = (y * width + x) * channels;
      for (let c = 0; c < channels; c++) {
        const a = raw.data[(y0 * sw + x0) * channels + c];
        const b = raw.data[(y0 * sw + x1) * channels + c];
        const cc = raw.data[(y1 * sw + x0) * channels + c];
        const d = raw.data[(y1 * sw + x1) * channels + c];
        out[di + c] = (a * (1 - wx) + b * wx) * (1 - wy) + (cc * (1 - wx) + d * wx) * wy;
      }
    }
  }
  return toDataUri(await sharp(out, { raw: { width, height, channels } }).png().toBuffer());
};
const prescale = async (uri, width, height) => {
  const key = `${uri.slice(-64)}@${width}x${height}`;
  if (prescaleCache.has(key)) return prescaleCache.get(key);
  const out = await bilinearResize(uri, width, height);
  prescaleCache.set(key, out);
  return out;
};

// Geometry frozen from the legacy Playwright capture (Gacha.tmpl + pixel
// measurement of the frozen baseline at dev time; no testdata is read at
// render time). All coordinates are CSS px in page space, cards are
// absolutely positioned; text lines are placed by measured ink tops via
// textY (line box centering model, verified by render->measure iteration).
const LH = 1.448; // Chromium normal line-height for NotoSansHans (hhea metrics)
const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });
const snap = (v) => Math.round(v * 1.5) / 1.5; // Chromium snaps baselines to the device-pixel grid
const textY = (fs, inkTop, digitOnly = false) =>
  snap(inkTop - (LH * fs - (digitOnly ? 0.73 : 0.875) * fs) / 2);
const lh = (fs) => ({ lineHeight: `${(LH * fs).toFixed(2)}px` });

const card = (left, title, ...kids) => h('div', { style: style({ position: 'absolute', left, top: 150, width: 302, height: 252, border: '1px solid #000', borderRadius: 20, backgroundColor: '#1f1e1e' }) },
  h('div', { style: style({ position: 'absolute', left: 2, top: -1.3, width: '100%', height: 38, transform: 'translateY(6px)', justifyContent: 'center', alignItems: 'center', color: '#eee', fontSize: 20 }) }, title),
  ...kids,
);

// ECharts pie replica: center (191.5, 140.5) card-local, r 78, sectors from
// 12 o'clock clockwise in data order (startAngle 90).
const pieSectors = (rarities) => {
  const total = rarities.reduce((sum, item) => sum + item.count, 0) || 1;
  let offset = 0;
  const r = 78;
  return rarities.map((item) => {
    const start = offset;
    offset += (item.count / total) * 360;
    const a = ((start - 90) * Math.PI) / 180;
    const b = ((offset - 90) * Math.PI) / 180;
    return h('path', {
      d: `M 0 0 L ${(r * Math.cos(a)).toFixed(2)} ${(r * Math.sin(a)).toFixed(2)} A ${r} ${r} 0 ${offset - start > 180 ? 1 : 0} 1 ${(r * Math.cos(b)).toFixed(2)} ${(r * Math.sin(b)).toFixed(2)} Z`,
      fill: item.color,
    });
  });
};

const legend = (rarities) => rarities.map((item, i) => [
  h('div', { style: { position: 'absolute', left: 8.3, top: 93.1 + i * 32, width: 8, height: 8, display: 'flex', backgroundColor: item.color } }),
  h('span', { style: { position: 'absolute', left: 21, top: textY(14, 89.7 + i * 32), ...lh(14), color: '#fff', fontSize: 14 } }, `${item.label}\u00A0\u00A0${item.percent.toFixed(2)}%`),
  ]);

// ECharts horizontal bar replica. Plot (card-local): x 75.7..270.4,
// y 46..240.7; band 97.35, bar 65.3; category axis on the left with
// boundary ticks; white value labels right of each bar.
const barCard = (pools) => {
  const values = [...pools].reverse();
  const max = Math.max(1, ...values.map((pool) => pool.count));
  const plotX = 75.7;
  const plotW = 194.7;
  const plotY = 46;
  const band = 97.35;
  const barH = 65.3;
  return [
    ...values.map((pool, i) => {
      const center = plotY + band * (i + 0.5);
      const width = (pool.count / max) * plotW;
      return h('div', { style: style({ position: 'absolute', left: plotX, top: center - barH / 2, width, height: barH, display: 'flex', backgroundColor: '#5470c6' }) });
    }),
    h('div', { style: { position: 'absolute', left: 75.5, top: plotY, width: 1.5, height: band * values.length, display: 'flex', backgroundColor: '#ccc' } }),
    ...[0, 1, 2].slice(0, values.length + 1).map((i) =>
      h('div', { style: { position: 'absolute', left: 70.5, top: plotY + band * i - 0.75, width: 5, height: 1.5, display: 'flex', backgroundColor: '#ccc' } })),
    ...values.map((pool, i) => {
      const center = plotY + band * (i + 0.5);
      return h('span', { style: { position: 'absolute', left: 2.7, width: 61.7, textAlign: 'right', top: textY(12, center - 7.25), ...lh(12), color: '#fff', fontSize: 12 } }, pool.poolName);
    }),
    ...values.map((pool, i) => {
      const center = plotY + band * (i + 0.5);
      const left = plotX + (pool.count / max) * plotW + 4.6;
      return h('span', { style: { position: 'absolute', left, top: textY(12, center - 7.1, true), ...lh(12), color: '#fff', fontSize: 12 } }, String(pool.count));
    }),
  ];
};

// List entry: avatar 100x100 + absolute text lines (measured ink tops differ
// between the 3-line and 4-line boxes because the legacy table distributes
// the rowspan image height across rows).
const entry = (e, avatar, kind, i) => {
  const lines = kind === 'chars'
    ? { name: 17, date: 51.7, pool: 83.7 }
    : { name: 13.7, date: 43, pool: 68.3, count: 95.7 };
  return h('div', { style: style({ position: 'absolute', left: i === 0 ? 0 : 229.7, top: 53, width: 230, height: 197 }) },
    h('img', { src: avatar, width: 100, height: 100, style: { position: 'absolute', left: 3.3, top: 8.7 } }),
    h('span', { style: { position: 'absolute', left: 107, top: textY(16, lines.name), ...lh(16), color: '#eee', fontSize: 16 } },
      e.name,
      e.isNew ? h('span', { style: { marginLeft: 5, color: 'red', fontSize: 10 } }, 'New') : null),
    h('span', { style: { position: 'absolute', left: 107, top: textY(15, lines.date, true), ...lh(15), color: '#eee', fontSize: 15 } }, e.date),
    h('span', { style: { position: 'absolute', left: 107, top: textY(15, lines.pool), ...lh(15), color: '#eee', fontSize: 15 } }, e.pool),
    kind === 'star6' ? h('span', { style: { position: 'absolute', left: 107, top: textY(15, lines.count), ...lh(15), color: '#eee', fontSize: 15 } }, `花费${e.count}抽`) : null,
  );
};

const listBox = (title, entries, resolved, kind) => h('div', { style: style({ position: 'absolute', left: kind === 'chars' ? 20 : 511, top: 441, width: 465, height: 250, border: '1px solid #000', borderRadius: 20, backgroundColor: '#1f1e1e' }) },
  h('div', { style: style({ position: 'absolute', left: 2, top: -1.3, width: '100%', height: 38, transform: 'translateY(6px)', justifyContent: 'center', alignItems: 'center', color: '#eee', fontSize: 20 }) }, title),
  ...entries.map((e, i) => entry(e, resolved[i], kind, i)),
);

export default async function render(props, { image }) {
  const [header, footer, ...avatars] = await Promise.all([
    image('/assets/gacha/header.png'),
    image('/assets/gacha/footer.png'),
    ...(props.chars || []).map((e) => image(e.avatar, '/assets/common/amiya.png')),
    ...(props.star6Info || []).map((e) => image(e.avatar, '/assets/common/amiya.png')),
  ]);
  const [headerHi, footerHi] = await Promise.all([
    prescale(header, 1500, 600),
    prescale(footer, 1500, 404),
  ]);
  const chars = avatars.slice(0, (props.chars || []).length);
  const star6 = avatars.slice(chars.length);
  const rarities = props.rarities || [];

  return h('div', { style: style({ position: 'relative', width: 1000, height: 882, overflow: 'hidden', fontFamily: 'NotoSansHans', backgroundColor: '#0c0d0c' }) },
    h('div', { style: { position: 'absolute', left: 0, top: 0, width: 1000, height: 400, display: 'flex', backgroundSize: '100% 100%', backgroundImage: `url(${headerHi})` } }),
    h('span', { style: { position: 'absolute', left: 320, top: 51.3, ...lh(32), color: '#eee', fontSize: 32, fontWeight: 700 } }, props.name),
    h('span', { style: { position: 'absolute', left: 250, top: 97.3, ...lh(32), color: '#eee', fontSize: 32, fontWeight: 700 } },
      `共${props.total}抽`,
      h('span', { style: { fontSize: 23, position: 'relative', top: '7.3px' } }, `(${props.period})`)),
    // card 1: star distribution pie + legend
    card(20, '星级分布', ...legend(rarities),
      h('svg', { style: { position: 'absolute', left: 108.8, top: 60.5 }, width: 158, height: 158, viewBox: '-79 -79 158 158' }, ...pieSectors(rarities))),
    // card 2: averages table
    card(346, '星级分布',
      ...(props.averages || []).map((item, i) =>
        h('div', { style: style({ position: 'absolute', left: 34.4, top: textY(16, [65.3, 110, 155.3, 200][i]), ...lh(16), color: '#eee', fontSize: 16 }) },
          h('span', { style: { width: 131.3 } }, `${item.count}个${item.label}`),
          h('span', null, `${item.avg}抽/个`)))),
    // card 3: pool bars
    card(671, '卡池分布(最近10个)', ...barCard(props.pools || [])),
    listBox('新获得干员(至多显示20个)', props.chars || [], chars, 'chars'),
    listBox('获得六星干员(至多显示20个)', props.star6Info || [], star6, 'star6'),
    h('div', { style: { position: 'absolute', left: 0, top: 613, width: 1000, height: 269, display: 'flex', backgroundSize: '100% 100%', backgroundImage: `url(${footerHi})` } },
      h('span', { style: { position: 'absolute', left: 530, top: 190.7, ...lh(32), color: '#eee', fontSize: 32, fontWeight: 700 } }, props.today)),
  );
}
