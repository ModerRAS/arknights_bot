import { h } from '../lib/h.mjs';

const fallback = 'assets/common/amiya.png';

// geometry measured off the frozen Playwright baseline (CSS px @1.5 scale)
// column centers (header th == icon centers): 51.7 / 129.3 / 184 / 291.3 / 425.7
const COLS = [51.7, 129.3, 184, 291.3, 425.7];
const HEADER_H = 35;
const ROW_H = 75;
const bold = { fontWeight: 700 };

// inline group: n icons of 50px with ~3.4px word space, centered on cx
function inlineCenters(cx, n, pitch) {
  const groupWidth = n * 50 + (n - 1) * (pitch - 50);
  const left = cx - groupWidth / 2;
  return Array.from({ length: n }, (_, j) => left + 25 + j * pitch);
}

export default async function render(props, { image }) {
  const rows = await Promise.all((props ?? []).map(async (item) => ({
    item,
    avatar: await image(`https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/${encodeURIComponent(item.id)}.png`, fallback),
    evolve: await image(`assets/box/Evolve_${item.evolvePhase}.png`, fallback),
    potential: await image(`assets/box/Potential_${item.potentialRank}.png`, fallback),
    skills: await Promise.all((item.skills ?? []).map(async (skill) => ({ ...skill, src: await image(`https://web.hycdn.cn/arknights/game/assets/char_skill/${encodeURIComponent(skill.id)}.png`, fallback) }))),
    equips: await Promise.all((item.equips ?? []).map(async (equip) => ({ ...equip, src: await image(`https://web.hycdn.cn/arknights/game/assets/uniequip/type/icon/${encodeURIComponent(equip.id)}.png`, fallback) }))),
  })));

  const at = (cx, top, children, extra = {}) => h('div', { style: { position: 'absolute', left: cx - 50, width: 100, top, display: 'flex', justifyContent: 'center', ...extra } }, children);
  const iconWithLv = (src, lv, iconTop, lvTop, hgt = 50) => h('div', { style: { position: 'absolute', left: 0, width: 100, top: 0, bottom: 0, display: 'flex', flexDirection: 'column', alignItems: 'center' } },
    h('img', { src, width: 50, height: hgt, style: { position: 'absolute', top: iconTop } }),
    h('div', { style: { position: 'absolute', top: lvTop, fontSize: 14, display: 'flex' } }, `LV${lv}`));

  const header = h('div', { style: { height: HEADER_H, display: 'flex', flexDirection: 'column', position: 'relative' } },
    COLS.map((cx, ci) => at(cx, 4.3, ['干员', '等级', '潜能', '技能', '模组'][ci], { height: 23.7, alignItems: 'center', fontSize: 16, ...bold })));

  const body = rows.map(({ item, avatar, evolve, potential, skills, equips }, ri) => {
    const skillCenters = inlineCenters(COLS[3], skills.length || 1, 53.4);
    const equipCenters = inlineCenters(COLS[4], equips.length || 1, 53);
    return h('div', { style: { height: ROW_H, display: 'flex', flexDirection: 'column', position: 'relative', ...(ri > 0 ? { borderTop: '1px solid #1f1f1f' } : {}) } },
      // operator: avatar 50x50 at (3.3, 7.7) + name 12px left 63.3 v-centered
      h('img', { src: avatar, width: 50, height: 50, style: { position: 'absolute', left: 3.3, top: 8.3 } }),
      h('div', { style: { position: 'absolute', left: 63.3, top: 0, bottom: 0, display: 'flex', alignItems: 'center', fontSize: 12 } }, item.name),
      // evolve: 50x41.3 (75x62 asset) top-aligned at +0.5; LV ink at +50
      h('div', { style: { position: 'absolute', left: COLS[1] - 50, width: 100, top: 0, bottom: 0, display: 'flex', flexDirection: 'column', alignItems: 'center' } },
        h('img', { src: evolve, width: 50, style: { position: 'absolute', top: -1 } }),
        h('div', { style: { position: 'absolute', top: 46, fontSize: 14, display: 'flex' } }, `LV${item.level}`)),
      at(COLS[2], 0, h('img', { src: potential, width: 50, height: 50, style: { position: 'absolute', top: 8 } })),
      ...skills.map((skill, i) => at(skillCenters[i], 0, iconWithLv(skill.src, skill.level, -1.5, 54))),
      ...equips.map((equip, i) => at(equipCenters[i], 0, iconWithLv(equip.src, equip.level, 4, 47.5))));
  });

  return h('div', { style: { width: 481, height: 186, display: 'flex', flexDirection: 'column', backgroundColor: '#2e3031', color: '#fff', fontFamily: 'NotoSansHans', fontSize: 14, position: 'relative', WebkitTextStrokeWidth: 0.35, WebkitTextStrokeColor: '#fff' } }, header, body);
}
