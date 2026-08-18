import { h } from '../lib/h.mjs';

const fallback = 'assets/common/amiya.png';

function iconColumn(items) {
  return h('div', { style: { display: 'flex' } }, items);
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
  const header = h('div', { style: { height: 35, display: 'flex', alignItems: 'center', justifyContent: 'space-around', fontWeight: 700 } }, '干员', '等级', '潜能', '技能', '模组');
  const body = rows.map(({ item, avatar, evolve, potential, skills, equips }) => {
    const operator = h('div', { style: { width: 110, display: 'flex', alignItems: 'center' } }, h('img', { src: avatar, width: 50, height: 50 }), h('span', { style: { marginLeft: 5 } }, item.name));
    const level = h('div', { style: { width: 62, display: 'flex', flexDirection: 'column', alignItems: 'center' } }, h('img', { src: evolve, width: 50 }), `LV${item.level}`);
    const potentialIcon = h('div', { style: { width: 58, display: 'flex', justifyContent: 'center' } }, h('img', { src: potential, width: 50 }));
    const skillIcons = iconColumn(skills.map((skill) => h('div', { style: { display: 'flex', flexDirection: 'column', alignItems: 'center' } }, h('img', { src: skill.src, width: 50, height: 50 }), `LV${skill.level}`)));
    const equipIcons = iconColumn(equips.map((equip) => h('div', { style: { display: 'flex', flexDirection: 'column', alignItems: 'center' } }, h('img', { src: equip.src, width: 50 }), `LV${equip.level}`)));
    return h('div', { style: { height: 75, borderTop: '1px solid #1f1f1f', display: 'flex', alignItems: 'center', justifyContent: 'space-around' } }, operator, level, potentialIcon, skillIcons, equipIcons);
  });
  return h('div', { style: { width: 481, height: 186, display: 'flex', flexDirection: 'column', backgroundColor: '#2e3031', color: '#fff', fontFamily: 'NotoSansHans' } }, header, body);
}
