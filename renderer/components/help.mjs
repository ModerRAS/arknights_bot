import { h } from '../lib/h.mjs';
import sharp from 'sharp';

// Geometry frozen from the legacy Playwright capture (Help.tmpl + pixel
// measurement of the frozen baseline at dev time; the render path reads no
// testdata). Page 660x1366 CSS at frozen DPR 1.5.
//
// Measured layout (CSS px): banner img at y10 h239; h1 使用说明 fs32 bold ink
// (20,148); banner p icon 16px at (20,208) + fs16 text; 3 group label strips
// (label.png 660x40, titles fs16 at left 25, vertically centered); command
// cards 152x54 border-box (1px white border, radius 10), 4 per row at pitch
// 162x69.33, text fs16 fw600 two lines (ink +8.6/+34.8 from card top),
// person-circle icon for IsBind cards at right.

const dataUriToBuffer = (uri) => Buffer.from(String(uri).replace(/^data:[^;,]+;base64,/, ''), 'base64');
const toDataUri = (buf) => `data:image/png;base64,${buf.toString('base64')}`;
const prescaleCache = new Map();

// Skia-medium-quality upscale is bilinear (triangle); sharp has no triangle
// kernel, so the 1.5x upscale is done manually on raw RGBA. Pre-upscaling to
// the exact device resolution makes the embedded bitmap 1:1 at the frozen DPR.
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

// NotoSansHans hhea metrics: 880/-120/lineGap 500 -> Chromium normal LH 1.5em.
const lh = (fs) => ({ lineHeight: `${(1.5 * fs).toFixed(2)}px` });

// Bootstrap bi-person-circle, drawn at 16 CSS (banner) / 25 CSS (cards, the
// frozen baseline renders the card icons visually larger).
const personIcon = (size) => h('svg', { style: { display: 'flex' }, width: size, height: size, viewBox: '0 0 16 16', fill: '#fff' },
  h('path', { d: 'M11 6a3 3 0 1 1-6 0 3 3 0 0 1 6 0z' }),
  h('path', { d: 'M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm8-7a7 7 0 0 0-5.468 11.37C3.242 11.226 4.805 10 8 10s4.757 1.225 5.468 2.37A7 7 0 0 0 8 1z' }));

const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });

const cmdCard = (cmd) => h('div', { style: style({ position: 'relative', width: 152, height: 54, marginLeft: 10, marginTop: 15, border: '1px solid #fff', borderRadius: 10, color: '#fff', flexShrink: 0 }) },
  h('span', { style: { position: 'absolute', left: 5, top: 2.6, ...lh(15), fontSize: 15, fontWeight: 600, whiteSpace: 'nowrap' } }, `${cmd.Cmd}${cmd.Param ? ` ${cmd.Param}` : ''}`),
  cmd.IsBind ? h('div', { style: style({ position: 'absolute', right: 10, top: 5.7 }) }, personIcon(16)) : null,
  h('span', { style: { position: 'absolute', left: 5, top: 28.8, ...lh(15), fontSize: 15, fontWeight: 600, whiteSpace: 'nowrap' } }, cmd.Desc),
);

const group = (title, commands, labelHi) => [
  h('div', { style: { position: 'relative', width: 660, height: 40, marginTop: 10, display: 'flex', backgroundImage: `url(${labelHi})`, backgroundSize: '100% 100%' } },
    h('span', { style: { position: 'absolute', left: 25, top: 8.4, ...lh(16), fontSize: 16, color: '#fff' } }, title)),
  h('div', { style: style({ flexWrap: 'wrap', width: 660 }) },
    ...commands.map((cmd) => cmdCard(cmd))),
];

export default async function render(props, { image }) {
  const [banner, label, bg, amiya] = await Promise.all([
    image('assets/help/banner.png'),
    image('assets/help/label.png'),
    image('assets/help/bg.jpg'),
    image('assets/help/amiya.png'),
  ]);
  const [bannerHi, labelHi, amiyaHi] = await Promise.all([
    prescale(banner, 990, 359),
    prescale(label, 990, 60),
    prescale(amiya, 990, 1018),
  ]);

  return h('div', { style: style({ position: 'relative', width: 660, height: 1366, overflow: 'hidden', fontFamily: 'NotoSansHans', backgroundColor: '#050505', backgroundImage: `url(${bg})`, backgroundSize: 'cover' }) },
    h('div', { style: { position: 'absolute', left: 0, top: 249, width: 660, height: 679, display: 'flex', backgroundImage: `url(${amiyaHi})`, backgroundSize: '100% 100%' } }),
    h('div', { style: { position: 'absolute', left: 0, top: 249, width: 660, height: 1366 - 249, display: 'flex', backgroundImage: 'linear-gradient(rgba(0,0,0,0.8), rgba(0,0,0,0.8))' } }),
    h('div', { style: { position: 'absolute', left: 0, top: 10, width: 660, height: 239, display: 'flex', backgroundImage: `url(${bannerHi})`, backgroundSize: '100% 100%' } },
      h('span', { style: { position: 'absolute', left: 20, top: 129.4, ...lh(32), fontSize: 32, fontWeight: 700, color: '#fff' } }, '使用说明'),
      h('div', { style: style({ position: 'absolute', left: 20, top: 198.3 }) }, personIcon(16)),
      h('span', { style: { position: 'absolute', left: 37.3, top: 194, ...lh(16), fontSize: 16, color: '#fff' } }, '为需要绑定角色的指令')),
    h('div', { style: style({ position: 'absolute', left: 0, top: 249, width: 660, flexDirection: 'column' }) },
      group('私聊指令', props.PrivateCmds ?? [], labelHi),
      group('普通指令', props.PublicCmds ?? [], labelHi),
      group('管理员指令', props.AdminCmds ?? [], labelHi)),
  );
}
