import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';
const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const text = (value, style = {}) => h('span', { style }, String(value ?? ''));

export default async function render(props, { image }) {
  const [background, painting, building, skill] = await Promise.all([
    image('/assets/operator/bg.png'), image(props.painting, fallback),
    image(props.buildingSkills?.[0]?.icon, fallback), image(props.skills?.[0]?.icon, fallback),
  ]);
  const data = props.op ?? {};
  const stat = (name, value) => box({ width: 170, height: 30 }, text(name, { width: 100, padding: '5px 8px', color: '#fff', backgroundColor: 'rgba(0,0,0,.8)' }), text(value, { width: 70, padding: '5px 8px', color: '#111', backgroundColor: 'rgba(239,238,239,.8)' }));
  const panel = (title, children, style = {}) => box({ width: 600, flexDirection: 'column', backgroundColor: 'rgba(63,63,62,.93)', color: '#fff', ...style }, box({ height: 20, padding: '5px 10px', alignItems: 'center', backgroundColor: 'rgba(0,0,0,.8)' }, text(title, { fontWeight: 600 })), ...children);
  return box({ width: 1200, height: 800, position: 'relative', overflow: 'hidden', backgroundImage: `url(${background})`, backgroundSize: 'cover', fontFamily: 'NotoSansHans' },
    h('img', { src: painting, height: 650, style: { position: 'absolute', left: 60, bottom: 0 } }),
    box({ position: 'absolute', left: 0, top: 20, width: 510, flexWrap: 'wrap' }, stat('最大生命值', data.hp), stat('攻击力', data.atk), stat('防御力', data.def), stat('法抗', data.res), stat('攻击间隔', data.interval), stat('再部署时间', data.reDeploy), stat('阻挡数', data.block), stat('部署费用', data.cost), stat('所属', data.logo)),
    box({ position: 'absolute', left: 10, bottom: 40, width: 330, flexDirection: 'column', color: '#111', backgroundColor: 'rgba(239,238,239,.55)' }, text(data.profession, { fontSize: 18, padding: 8 }), text(data.tags, { padding: 8 }), text(props.professionBranch?.name, { padding: 8 }), text(props.professionBranch?.desc, { padding: 8 }), text(data.name, { fontSize: 30, fontWeight: 600, padding: 8 }), text(data.nameEn, { padding: '0 8px 8px' })),
    box({ position: 'absolute', top: 20, right: 0, width: 600, flexDirection: 'column' },
      panel('天赋', (props.talents ?? []).map((item) => box({ padding: '2px 6px', backgroundColor: 'rgba(176,177,177,.7)' }, text(item.evolve, { width: 100 }), text(item.name, { width: 160 }), text(item.desc, { flex: 1 })))),
      panel('基建技能', (props.buildingSkills ?? []).map((item, index) => box({ padding: '2px 6px', alignItems: 'center', backgroundColor: 'rgba(176,177,177,.7)' }, index === 0 ? h('img', { src: building, width: 36, height: 36 }) : null, text(item.name, { marginLeft: 8, width: 130 }), text(item.desc, { flex: 1 })))),
      panel('技能', (props.skills ?? []).map((item, index) => box({ minHeight: 70, padding: 8, alignItems: 'center' }, index === 0 ? h('img', { src: skill, width: 50, height: 50 }) : null, box({ marginLeft: 10, flexDirection: 'column', flex: 1 }, text(item.name, { fontWeight: 600 }), text(`${(item.spType ?? []).join(' ')} 技力${item.spInit}/${item.spCost}${item.duration ? ` 持续${item.duration}s` : ''}`, { fontSize: 13 }), text(item.desc, { fontSize: 13 })), text(item.skillRange, { width: 90, textAlign: 'center' })))),
    ),
  );
}
