import { h } from '../lib/h.mjs';
import sharp from 'sharp';

// Geometry frozen from the legacy Playwright capture (BoxSummary.tmpl + pixel
// measurement of the frozen baseline at dev time; the render path reads no
// testdata). Page 900x482 CSS at frozen DPR 1.5.
//
// Measured layout (CSS px): label.png strip 0..60 (prescaled 1350x90 blit);
// Dr title fs30 ink (28,12.7); th row fs16 bold centered at column centers
// 120/340/560/780, ink top 70; 8 metric rows: white border-bottom at
// y=122+32i spanning each 200px column, fs16 title/value at cell+11.3 and
// right edge cell+190, ink top border+12; 未招募干员 th centered at 450;
// missing avatars 40px, 19/row at (20.7,384) pitch 45.84x49.3.

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

const lh = (fs) => ({ lineHeight: `${(1.5 * fs).toFixed(2)}px` });
const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });

const METRICS = [
  ['招募干员数量', 'allCharCnt', 'star6CharCnt', 'star5CharCnt', 'star4CharCnt'],
  ['精英阶段2干员', 'allEvolvePhase2Cnt', 'star6EvolvePhase2Cnt', 'star5EvolvePhase2Cnt', 'star4EvolvePhase2Cnt'],
  ['专精三技能数量', 'allSkill10Cnt', 'star6Skill10Cnt', 'star5Skill10Cnt', 'star4Skill10Cnt'],
  ['专精二技能数量', 'allSkill9Cnt', 'star6Skill9Cnt', 'star5Skill9Cnt', 'star4Skill9Cnt'],
  ['专精一技能数量', 'allSkill8Cnt', 'star6Skill8Cnt', 'star5Skill8Cnt', 'star4Skill8Cnt'],
  ['三级模组数量', 'allEquipStage3Cnt', 'star6EquipStage3Cnt', 'star5EquipStage3Cnt', 'star4EquipStage3Cnt'],
  ['二级模组数量', 'allEquipStage2Cnt', 'star6EquipStage2Cnt', 'star5EquipStage2Cnt', 'star4EquipStage2Cnt'],
  ['一级模组数量', 'allEquipStage1Cnt', 'star6EquipStage1Cnt', 'star5EquipStage1Cnt', 'star4EquipStage1Cnt'],
];
const HEADS = ['全部干员', '六星干员', '五星干员', '四星干员'];
const CELL_LEFTS = [20, 240, 460, 680];

export default async function render(props, { image }) {
  const [labelRaw, avatars] = await Promise.all([
    image('assets/help/label.png'),
    Promise.all((props.missingChars ?? []).map((item) => image(item.skinId))),
  ]);
  const labelHi = await prescale(labelRaw, 1350, 90);

  return h('div', { style: style({ position: 'relative', width: 900, height: 482, overflow: 'hidden', backgroundColor: '#2e3031', color: '#fff', fontFamily: 'NotoSansHans' }) },
    h('div', { style: { position: 'absolute', left: 0, top: 0, width: 900, height: 60, display: 'flex', backgroundImage: `url(${labelHi})`, backgroundSize: '100% 100%' } }),
    props.name ? h('span', { style: { position: 'absolute', left: 27.5, top: 3.3, ...lh(30), fontSize: 30, color: '#fff' } }, `Dr ${props.name}`) : null,
    ...HEADS.map((head, i) => h('span', { style: { position: 'absolute', left: CELL_LEFTS[i], width: 200, textAlign: 'center', top: 65, ...lh(16), fontSize: 16, fontWeight: 700, color: '#fff' } }, head)),
    ...METRICS.map(([title, ...keys], row) =>
      keys.map((key, col) => {
        const left = CELL_LEFTS[col];
        return h('div', { style: style({ position: 'absolute', left, top: 122 + 32 * row, width: 200, height: 23, borderBottom: '1px solid #fff', justifyContent: 'space-between', alignItems: 'flex-end', padding: '0 10px', paddingBottom: 4 }) },
          h('span', { style: { ...lh(16), fontSize: 16 } }, title),
          h('span', { style: { ...lh(16), fontSize: 16 } }, String(props[key] ?? '')));
      })).flat(),
    h('span', { style: { position: 'absolute', left: 350, width: 200, textAlign: 'center', top: 352, ...lh(16), fontSize: 16, fontWeight: 700, color: '#fff' } }, '未招募干员'),
    (props.missingChars ?? []).length ? h('div', { style: style({ position: 'absolute', left: 20.7, top: 384, width: 871, flexWrap: 'wrap', columnGap: 5.84, rowGap: 9.3, alignItems: 'flex-start' }) },
      avatars.map((src) => h('img', { src, width: 40, height: 40 }))) : null,
  );
}
