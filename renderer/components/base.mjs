import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';
const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const text = (value, style = {}) => h('span', { style }, String(value ?? ''));

async function operator(image, value) {
  return { ...value, src: await image(`https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/${encodeURIComponent(value.avatar ?? '')}.png`, fallback) };
}

export default async function render(props, { image }) {
  const rooms = [
    ['控制中枢', props.control],
    ...(props.dormitories ?? []).map((value) => ['宿舍', value]),
    ...(props.tradings ?? []).map((value) => ['贸易站', value]),
    ...(props.manufactures ?? []).map((value) => ['制造站', value]),
    ...(props.powers ?? []).map((value) => ['发电站', value]),
    ['会客室', props.meeting], ['办公室', props.hire], ['训练室', props.training],
  ];
  const prepared = await Promise.all(rooms.map(async ([title, room]) => ({ title, room, chars: await Promise.all((room?.chars ?? []).map((value) => operator(image, value)))})));
  const skill = props.training?.skill ? await image(`https://web.hycdn.cn/arknights/game/assets/char_skill/${encodeURIComponent(props.training.skill)}.png`, fallback) : null;
  const characterCard = (item) => {
    const ap = Math.max(0, Math.min(100, item.AP ?? 0));
    const color = ap <= 0 ? '#e42e20' : ap < 100 ? '#f0ab22' : '#3cd627';
    return box({ width: 170, height: 46, marginRight: 6, alignItems: 'center' },
      h('img', { src: item.src, width: 40, height: 40 }),
      box({ marginLeft: 4, flexDirection: 'column' },
        text(item.name, { fontSize: 13 }),
        box({ width: 150, height: 3, marginTop: 4, backgroundColor: '#444' },
          h('div', { style: { display: 'flex', width: `${ap}%`, height: 3, backgroundColor: color } }),
        ),
      ),
    );
  };
  const rightColumn = new Set(['制造站', '会客室', '训练室']);
  const roomCard = ({ title, room, chars }) => box({ width: title === '控制中枢' || title === '宿舍' ? 1105 : 550, height: 112, marginTop: 5, flexDirection: 'column', backgroundColor: '#21262f', border: '1px solid #21262f', borderRadius: 15, color: '#fff' },
    box({ position: 'relative', top: 7, left: rightColumn.has(title) ? 5 : 0, flexDirection: 'column' },
      box({ height: 49, marginLeft: 10, marginRight: 20, alignItems: 'center', justifyContent: 'space-between' },
        text(`${title} Lv.${room?.level ?? 0}`, { fontSize: 19, fontWeight: 600 }),
        title === '训练室' && skill
          ? box({ alignItems: 'center' }, text(`Lv.${room?.specializeLevel ?? ''}`), h('img', { src: skill, width: 30, height: 30, style: { marginLeft: 4 } }))
          : text(room?.strategy ?? room?.power ?? room?.comfort ?? room?.refreshCount ?? '', { color: '#8cd1ff' }),
      ),
      box({ marginLeft: 10, alignItems: 'center', flexWrap: 'wrap' }, chars.length ? chars.map(characterCard) : text('空无一人', { marginLeft: 10 })),
    ),
  );
  return box({ width: 1110, height: 612, flexDirection: 'column', overflow: 'hidden', backgroundColor: '#2b333d', fontFamily: 'NotoSansHans' },
    box({ height: 27, justifyContent: 'space-between', alignItems: 'center', color: '#fff' }, text('基建信息', { fontSize: 19, fontWeight: 600, marginLeft: 10 }), box({ alignItems: 'center', marginRight: 30 }, text(`${props.labor?.current ?? 0}/${props.labor?.total ?? 0}`), box({ width: 100, height: 3, marginLeft: 8, backgroundColor: '#555' }, h('div', { style: { display: 'flex', width: `${Math.min(100, 100 * (props.labor?.current ?? 0) / Math.max(1, props.labor?.total ?? 1))}%`, height: 3, backgroundColor: '#fff' } })))),
    box({ flexWrap: 'wrap', alignContent: 'flex-start' }, prepared.map(roomCard)),
  );
}
