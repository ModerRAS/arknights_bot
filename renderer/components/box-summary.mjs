import { h } from '../lib/h.mjs';

const fallback = 'assets/common/amiya.png';
const heads = ['全部干员', '六星干员', '五星干员', '四星干员'];
const metrics = [
  ['招募干员数量', 'allCharCnt', 'star6CharCnt', 'star5CharCnt', 'star4CharCnt'],
  ['精英阶段2干员', 'allEvolvePhase2Cnt', 'star6EvolvePhase2Cnt', 'star5EvolvePhase2Cnt', 'star4EvolvePhase2Cnt'],
  ['专精三技能数量', 'allSkill10Cnt', 'star6Skill10Cnt', 'star5Skill10Cnt', 'star4Skill10Cnt'],
  ['专精二技能数量', 'allSkill9Cnt', 'star6Skill9Cnt', 'star5Skill9Cnt', 'star4Skill9Cnt'],
  ['专精一技能数量', 'allSkill8Cnt', 'star6Skill8Cnt', 'star5Skill8Cnt', 'star4Skill8Cnt'],
  ['三级模组数量', 'allEquipStage3Cnt', 'star6EquipStage3Cnt', 'star5EquipStage3Cnt', 'star4EquipStage3Cnt'],
  ['二级模组数量', 'allEquipStage2Cnt', 'star6EquipStage2Cnt', 'star5EquipStage2Cnt', 'star4EquipStage2Cnt'],
  ['一级模组数量', 'allEquipStage1Cnt', 'star6EquipStage1Cnt', 'star5EquipStage1Cnt', 'star4EquipStage1Cnt'],
];

// legacy Chrome table model measured off frozen baseline:
// 27px metric rows on 32px pitch (border-spacing 20x5 + 16px line box),
// auto-layout columns are non-uniform: left/width per column (th centers confirm edges)
const cols = [[20, 209.5], [249.5, 195.5], [466, 195.5], [683, 196]];
const bold = { fontWeight: 700, WebkitTextStrokeWidth: 1.0, WebkitTextStrokeColor: '#fff' };
const row = { height: 27, marginBottom: 5, display: 'flex', flexDirection: 'column', position: 'relative' };
const cellAt = ([left, width], extra) => ({ position: 'absolute', left, width, top: 0, bottom: 0, display: 'flex', alignItems: 'center', fontSize: 16, ...extra });

export default async function render(props, { image }) {
  const [label, avatars] = await Promise.all([
    image('assets/help/label.png', fallback),
    Promise.all((props.missingChars ?? []).map((item) => image(item.skinId, fallback))),
  ]);
  const metricRow = ([title, ...keys]) => h('div', { style: row },
    keys.map((key, i) => h('div', { style: { ...cellAt(cols[i], { justifyContent: 'space-between', padding: '0 11px', borderBottom: '1px solid #fff' }) } },
      h('span', null, title), h('span', null, String(props[key] ?? '')))));
  return h('div', { style: { width: 900, height: 482, display: 'flex', flexDirection: 'column', backgroundColor: '#2e3031', color: '#fff', fontFamily: 'NotoSansHans', fontSize: 16, position: 'relative', WebkitTextStrokeWidth: 0.25, WebkitTextStrokeColor: '#fff' } },
    h('img', { src: label, width: 900, height: 60 }),
    props.name ? h('div', { style: { position: 'absolute', left: 25, top: 11.5, fontSize: 30, display: 'flex' } }, `Dr ${props.name}`) : null,
    h('div', { style: { position: 'absolute', top: 65, left: 0, width: 900, display: 'flex', flexDirection: 'column' } },
      h('div', { style: { height: 26, marginBottom: 5, display: 'flex', flexDirection: 'column', position: 'relative' } },
        heads.map((x, i) => h('div', { style: { ...cellAt(cols[i], { justifyContent: 'center', ...bold }) } }, x))),
      metrics.map(metricRow),
      h('div', { style: { height: 26, display: 'flex', justifyContent: 'center', alignItems: 'center', ...bold } }, '未招募干员'),
      avatarRows(avatars)));
}

// legacy td is text-align: justify: full lines stretch to both edges, last line keeps natural word spacing
const AVATAR = 40;
const NATURAL_GAP = 3.5;
const CONTENT_W = 858; // 900 - 21 left - 21 right
function avatarRows(avatars) {
  const perRow = Math.max(1, Math.floor((CONTENT_W + NATURAL_GAP) / (AVATAR + NATURAL_GAP)));
  const rows = [];
  for (let i = 0; i < avatars.length; i += perRow) rows.push(avatars.slice(i, i + perRow));
  if (!rows.length) return null;
  return h('div', { style: { display: 'flex', flexDirection: 'column', gap: 6, paddingLeft: 21, marginTop: 6 } },
    rows.map((chunk, ri) => ri < rows.length - 1
      ? h('div', { style: { display: 'flex', justifyContent: 'space-between', width: CONTENT_W } }, chunk.map((src) => h('img', { src, width: AVATAR, height: AVATAR })))
      : h('div', { style: { display: 'flex', gap: NATURAL_GAP } }, chunk.map((src) => h('img', { src, width: AVATAR, height: AVATAR })))));
}
