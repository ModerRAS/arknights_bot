import { h } from '../lib/h.mjs';

const fallback = 'assets/common/amiya.png';

// Geometry measured off the frozen Playwright baseline (CSS px @1.5 scale).
// The legacy page is a <table> with default 16px font everywhere and
// img{width:50px;height:auto}. Element tops measured from baseline ink
// (device px / 1.5): row1 top 47px, row pitch 117px (78css), avatar 64px,
// name ink 88px, evolve/skill box 47px, potential 66px, equip 65px,
// evolve LV ink 117px, skill LV ink 128px, equip LV ink 120px.
const COLS = [51.7, 129.3, 184, 291.3, 425.7];
const HEADER_H = 35;
const ROW_PITCH = 78;
const FONT = 16;

const T = {
  avatar: 42.7, // 64px
  name: 56, // name ink top 88px (16px font, ~4px glyph inset)
  evolve: 31.3, // 47px
  evolveLv: 75.3, // LV ink top 117px
  potential: 44, // 66px
  skill: 31.3, // 47px
  skillLv: 82.7, // LV ink top 128px
  equip: 43.3, // 65px
  equipLv: 77.3, // LV ink top 120px
};

const bold = { fontWeight: 700 };

// inline group: n icons of 50px with ~3.4px word space, centered on cx
function inlineCenters(cx, n, pitch) {
  const groupWidth = n * 50 + (n - 1) * (pitch - 50);
  const left = cx - groupWidth / 2;
  return Array.from({ length: n }, (_, j) => left + 25 + j * pitch);
}

const at = (cx, top, children, extra = {}) => h('div', { style: { position: 'absolute', left: cx - 50, width: 100, top, display: 'flex', justifyContent: 'center', ...extra } }, children);

export default async function render(props, { image }) {
  const rows = await Promise.all((props ?? []).map(async (item) => ({
    item,
    avatar: await image(`https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/${encodeURIComponent(item.id)}.png`, fallback),
    evolve: await image(`assets/box/Evolve_${item.evolvePhase}.png`, fallback),
    potential: await image(`assets/box/Potential_${item.potentialRank}.png`, fallback),
    skills: await Promise.all((item.skills ?? []).map(async (skill) => ({ ...skill, src: await image(`https://web.hycdn.cn/arknights/game/assets/char_skill/${encodeURIComponent(skill.id)}.png`, fallback) }))),
    equips: await Promise.all((item.equips ?? []).map(async (equip) => ({ ...equip, src: await image(`https://web.hycdn.cn/arknights/game/assets/uniequip/type/icon/${encodeURIComponent(equip.id)}.png`, fallback) }))),
  })));

  const iconWithLv = (src, lv, iconTop, lvTop, hgt) => h('div', { style: { position: 'absolute', left: 0, width: 100, top: 0, bottom: 0, display: 'flex', flexDirection: 'column', alignItems: 'center' } },
    h('img', { src, width: 50, ...(hgt ? { height: hgt } : {}), style: { position: 'absolute', top: iconTop } }),
    h('div', { style: { position: 'absolute', top: lvTop, fontSize: FONT, display: 'flex' } }, `LV${lv}`));

  const header = h('div', { style: { height: HEADER_H, display: 'flex', flexDirection: 'column', position: 'relative' } },
    COLS.map((cx, ci) => at(cx, 4, ['干员', '等级', '潜能', '技能', '模组'][ci], { height: 23.7, alignItems: 'center', fontSize: FONT, ...bold })));

  // one row of elements at canvas-absolute tops (ri=0 row1, ri=1 row2)
  const body = (rows ?? []).flatMap((r, ri) => {
    const y = ri * ROW_PITCH;
    const skillCenters = inlineCenters(COLS[3], r.skills.length || 1, 53.4);
    const equipCenters = inlineCenters(COLS[4], r.equips.length || 1, 53);
    return [
      h('img', { src: r.avatar, width: 50, height: 50, style: { position: 'absolute', left: 3.3, top: T.avatar + y } }),
      h('div', { style: { position: 'absolute', left: 60, top: T.name + y, fontSize: FONT, display: 'flex' } }, r.item.name),
      at(COLS[1], T.evolve + y, iconWithLv(r.evolve, r.item.level, 0, T.evolveLv - T.evolve)),
      at(COLS[2], T.potential + y, h('img', { src: r.potential, width: 50, height: 50, style: { position: 'absolute', top: 0 } })),
      ...r.skills.map((skill, i) => at(skillCenters[i], T.skill + y, iconWithLv(skill.src, skill.level, 0, T.skillLv - T.skill, 50))),
      ...r.equips.map((equip, i) => at(equipCenters[i], T.equip + y, iconWithLv(equip.src, equip.level, 0, T.equipLv - T.equip))),
    ];
  });

  return h('div', { style: { width: 481, height: 186, display: 'flex', flexDirection: 'column', backgroundColor: '#2e3031', color: '#fff', fontFamily: 'NotoSansHans', fontSize: FONT, position: 'relative', WebkitTextStrokeWidth: 0.35, WebkitTextStrokeColor: '#fff' } },
    header, body);
}