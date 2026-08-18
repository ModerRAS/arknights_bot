import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';
const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const text = (value, style = {}) => h('span', { style }, String(value ?? ''));

export default async function render(props, { image }) {
  const picture = await image(props.pic, fallback);
  const valueRow = (pairs) => {
    const cellWidth = pairs.length === 4 ? 164 : 109;
    return box({ width: 656, minHeight: 34, borderBottom: '1px solid #595858', color: '#fff' }, ...pairs.map(([name, value]) => box({ width: cellWidth, flexDirection: 'column', alignItems: 'center', justifyContent: 'center', borderRight: '1px solid #595858', padding: 3 }, text(name, { fontSize: 12, whiteSpace: 'nowrap' }), text(value, { marginTop: 3 }))));
  };
  const level = (item, index) => box({ width: 656, flexDirection: 'column', color: '#fff', borderBottom: '1px solid #595858' },
    text(`级别${index}`, { padding: 7, fontSize: 22, fontWeight: 600, textAlign: 'center', borderBottom: '1px solid #595858' }),
    text(item.desc, { minHeight: 30, padding: 7, borderBottom: '1px solid #595858' }),
    valueRow([['攻击方式', item.attackType], ['行动方式', item.motion], ['生命恢复速度', item.hpRecovery], ['最大生命值', item.hp], ['攻击力', item.atk], ['防御力', item.def]]),
    valueRow([['法抗', item.res], ['攻击范围半径', item.ATKRadius], ['重量', item.weight], ['移动速度', item.moveSpeed], ['攻击间隔', item.interval], ['损伤抗性', item.damageRes]]),
    valueRow([['元素抗性', item.elementRes], ['基础嘲讽等级', item.ridicule], ['进点损失', item.point], ['异常抗性', item.abnormal]]),
    ...(item.skills ?? []).map((skill) => box({ padding: 7, flexDirection: 'column', borderTop: '1px solid #595858' }, text(`${skill.name}  ${skill.spInit}/${skill.spCost}`, { fontWeight: 600 }), text(skill.desc, { marginTop: 3 }))),
    item.talent ? box({ padding: 7, flexDirection: 'column', borderTop: '1px solid #595858' }, text('天赋&能力', { fontWeight: 600 }), text(item.talent, { marginTop: 3 })) : null,
  );
  return box({ width: 656, height: 318, overflow: 'hidden', flexDirection: 'column', backgroundColor: '#323332', fontFamily: 'NotoSansHans', paddingTop: 24 },
    text(props.name, { height: 40, color: '#fff', fontSize: 25, fontWeight: 600, textAlign: 'center', paddingTop: 7 }),
    box({ minHeight: 90, color: '#fff', borderTop: '1px solid #595858', borderBottom: '1px solid #595858' }, box({ width: 283, justifyContent: 'center' }, h('img', { src: picture, width: 158, height: 158, style: { margin: '1px 6px' } })), text(props.desc, { padding: 10, flex: 1 })),
    valueRow([['种类', props.enemyRace], ['地位级别', props.enemyLevel], ['攻击方式', props.attackType], ['行动方式', props.motion]]),
    box({ padding: 6, color: '#fff', borderBottom: '1px solid #595858' }, text('能力：', { fontWeight: 600 }), text(props.ability)),
    ...(props.level ?? props.levels ?? []).map(level),
  );
}
