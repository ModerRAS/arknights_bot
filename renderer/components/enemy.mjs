import { h } from '../lib/h.mjs';

const WIDTH = 656;
const BORDER = '1px solid #595858';
const fallback = '/assets/common/amiya.png';
const baseTracks = [283, 124, 124, 125];
const levelTracks = [109, 109, 109, 109, 110, 110];

const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const text = (value, style = {}) => h('div', { style: { display: 'flex', width: '100%', justifyContent: 'center', alignItems: 'center', ...style } }, String(value ?? ''));
const cell = (width, style, ...children) => box({ width, boxSizing: 'border-box', border: BORDER, ...style }, ...children);
const row = (tracks, values, style = {}) => box({ width: WIDTH, ...style }, ...values.map((value, index) => cell(tracks[index], { alignItems: 'center', justifyContent: 'center' }, value)));
const fullCell = (style, ...children) => cell(WIDTH, style, ...children);
const metric = (tracks, pairs, height = 28) => pairs.length === tracks.length
  ? box({ width: WIDTH, flexDirection: 'column' }, row(tracks, pairs.map(([name]) => text(name, { fontSize: 16, whiteSpace: 'nowrap', textAlign: 'center' })), { height }), row(tracks, pairs.map(([, value]) => text(value, { textAlign: 'center' })), { height }))
  : row(tracks, pairs.flatMap(([name, value]) => [text(name, { fontSize: 16, whiteSpace: 'nowrap', textAlign: 'center' }), text(value, { textAlign: 'center' })]), { height });
const baseMetric = (pairs, values, height = 28) => row(baseTracks, (values ? pairs.map(([, value]) => text(value, { textAlign: 'center' })) : pairs.map(([name]) => text(name, { fontSize: 16, whiteSpace: 'nowrap', textAlign: 'center' }))), { height });
const levelSpanRow = (first, rest, style = {}) => box({ width: WIDTH, ...style }, cell(levelTracks[0], { alignItems: 'center', justifyContent: 'center' }, first), cell(levelTracks.slice(1).reduce((sum, value) => sum + value, 0), { alignItems: 'center', justifyContent: 'center' }, rest));

function levelRows(item, index) {
  const rows = [
    box({ width: WIDTH, flexDirection: 'column', color: '#fff' },
      fullCell({ height: 41, alignItems: 'center', justifyContent: 'center' }, text(`级别${index}`, { fontSize: 22, fontWeight: 600, justifyContent: 'center', textAlign: 'center' })),
      fullCell({ height: 34, alignItems: 'center', justifyContent: 'center' }, text('描述', { fontSize: 22, fontWeight: 600, justifyContent: 'center', textAlign: 'center' })),
      fullCell({ minHeight: 34, padding: 7 }, text(item.desc, { justifyContent: 'flex-start', alignItems: 'flex-start' })),
      metric(levelTracks, [['攻击方式', item.attackType], ['行动方式', item.motion], ['生命恢复速度', item.hpRecovery]]),
      metric(levelTracks, [['最大生命值', item.hp], ['攻击力', item.atk], ['防御力', item.def], ['法抗', item.res], ['攻击范围半径', item.ATKRadius], ['重量', item.weight]]),
      metric(levelTracks, [['移动速度', item.moveSpeed], ['攻击间隔 ', item.interval], ['损伤抗性', item.damageRes], ['元素抗性', item.elementRes], ['基础嘲讽等级', item.ridicule], ['进点损失', item.point]]),
      levelSpanRow(text('异常抗性', { fontSize: 12, whiteSpace: 'nowrap', textAlign: 'center' }), text(item.abnormal), { height: 28 }),
    ),
  ];
  const skills = item.skills ?? [];
  if (skills.length > 0) {
    rows.push(box({ width: WIDTH, flexDirection: 'column', color: '#fff' },
      fullCell({ height: 34, alignItems: 'center', justifyContent: 'center' }, text('技能', { fontSize: 20, fontWeight: 600, justifyContent: 'center', textAlign: 'center' })),
      box({ width: WIDTH, height: 28 }, cell(levelTracks[0], { alignItems: 'center', justifyContent: 'center' }, text('名称', { textAlign: 'center' })), cell(levelTracks[1], { alignItems: 'center', justifyContent: 'center' }, text('冷却时间', { textAlign: 'center' })), cell(levelTracks[2], { alignItems: 'center', justifyContent: 'center' }, text('技力消耗', { textAlign: 'center' })), cell(levelTracks.slice(3).reduce((sum, value) => sum + value, 0), { alignItems: 'center', justifyContent: 'center' }, text('效果', { textAlign: 'center' }))),
      ...skills.map((skill) => box({ width: WIDTH, minHeight: 34 }, cell(levelTracks[0], { alignItems: 'center', justifyContent: 'center' }, text(skill.name, { whiteSpace: 'nowrap', textAlign: 'center' })), cell(levelTracks[1], { alignItems: 'center', justifyContent: 'center' }, text(skill.spInit, { textAlign: 'center' })), cell(levelTracks[2], { alignItems: 'center', justifyContent: 'center' }, text(skill.spCost, { textAlign: 'center' })), cell(levelTracks.slice(3).reduce((sum, value) => sum + value, 0), { alignItems: 'center', justifyContent: 'center' }, text(skill.desc)))),
    ));
  }
  if (item.talent) {
    rows.push(box({ width: WIDTH, flexDirection: 'column', color: '#fff' },
      fullCell({ height: 34, alignItems: 'center', justifyContent: 'center' }, text('天赋&能力', { fontSize: 20, fontWeight: 600, justifyContent: 'center', textAlign: 'center' })),
      fullCell({ minHeight: 34, padding: 7 }, text(item.talent, { justifyContent: 'flex-start', alignItems: 'flex-start' })),
    ));
  }
  return rows;
}

export default async function render(props, { image }) {
  const picture = await image(props.pic, fallback);
  const levels = Array.isArray(props.level) ? props.level : (Array.isArray(props.levels) ? props.levels : []);
  const rootStyle = {
    width: WIDTH,
    flexDirection: 'column',
    backgroundColor: '#323332',
    fontFamily: 'NotoSansHans',
    color: '#fff',
    ...(props.autoHeight === true ? { minHeight: 318 } : { height: 318, overflow: 'hidden' }),
  };
  const base = box({ width: WIDTH, flexDirection: 'column' },
    fullCell({ height: 42, alignItems: 'center', justifyContent: 'center' }, text(props.name, { color: '#fff', fontSize: 25, fontWeight: 600, justifyContent: 'center', textAlign: 'center' })),
    box({ width: WIDTH, height: 162 },
      cell(baseTracks[0], { alignItems: 'center', justifyContent: 'center' }, h('img', { src: picture, width: 158, height: 158, style: { margin: '1px 6px' } })),
      cell(baseTracks.slice(1).reduce((sum, value) => sum + value, 0), { alignItems: 'center', justifyContent: 'center' }, text(props.desc, { justifyContent: 'center', textAlign: 'center' })),
    ),
    baseMetric([['种类', props.enemyRace], ['地位级别', props.enemyLevel], ['攻击方式', props.attackType], ['行动方式', props.motion]], false),
    baseMetric([['种类', props.enemyRace], ['地位级别', props.enemyLevel], ['攻击方式', props.attackType], ['行动方式', props.motion]], true),
    fullCell({ height: 34, alignItems: 'center', justifyContent: 'center' }, text('能力', { fontSize: 20, fontWeight: 600, justifyContent: 'center', textAlign: 'center' })),
  );
  return box(rootStyle,
    text(props.ability, { height: 24, alignItems: 'center', justifyContent: 'flex-start' }),
    base,
    ...levels.flatMap(levelRows),
  );
}
