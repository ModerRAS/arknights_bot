import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';
const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const text = (value, style = {}) => h('span', { style }, String(value ?? ''));
const laborIcon = () => h('svg', { width: 20, height: 20, viewBox: '0 0 48 48', fill: 'none' },
  h('path', { d: 'M12 12L19 19M36 36L29 29', stroke: '#852cd3', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M36 12L29 19M12 36L19 29', stroke: '#852cd3', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('rect', { x: 19, y: 19, width: 10, height: 10, fill: '#852cd3', stroke: '#852cd3', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M36 19C37.3845 19 38.7378 18.5895 39.889 17.8203C41.0401 17.0511 41.9373 15.9579 42.4672 14.6788C42.997 13.3997 43.1356 11.9922 42.8655 10.6344C42.5954 9.2765 41.9287 8.02922 40.9497 7.05026C39.9708 6.07129 38.7235 5.4046 37.3656 5.13451C36.0078 4.86441 34.6003 5.00303 33.3212 5.53285C32.0421 6.06266 30.9489 6.95987 30.1797 8.11101C29.4105 9.26215 29 10.6155 29 12', stroke: '#852cd3', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M36 29C37.3845 29 38.7378 29.4105 39.889 30.1797C41.0401 30.9489 41.9373 32.0421 42.4672 33.3212C42.997 34.6003 43.1356 36.0078 42.8655 37.3656C42.5954 38.7235 41.9287 39.9708 40.9497 40.9497C39.9708 41.9289 38.7235 42.5954 37.3656 42.8655C36.0078 43.1356 34.6003 42.997 33.3212 42.4672C32.0421 41.9373 30.9489 41.0401 30.1797 39.889C29.4105 38.7378 29 37.3845 29 36', stroke: '#852cd3', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M12 29C10.6155 29 9.26216 29.4105 8.11101 30.1797C6.95987 30.9489 6.06266 32.0421 5.53285 33.3212C5.00303 34.6003 4.86441 36.0078 5.13451 37.3656C5.4046 38.7235 6.07129 39.9708 7.05026 40.9497C8.02922 41.9289 9.2765 42.5954 10.6344 42.8655C11.9922 43.1356 13.3997 42.997 14.6788 42.4672C15.9579 41.9373 17.0511 41.0401 17.8203 39.889C18.5895 38.7378 19 37.3845 19 36', stroke: '#852cd3', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
);

const moodIcon = (ap, color) => h('svg', { width: 18, height: 18, viewBox: '0 0 18 18', fill: 'none', style: { marginLeft: 2 } },
  h('circle', { cx: 9, cy: 9, r: 7, stroke: color, strokeWidth: 2 }),
  h('circle', { cx: 6.5, cy: 7, r: 1, fill: color }),
  h('circle', { cx: 11.5, cy: 7, r: 1, fill: color }),
  h('path', { d: ap <= 0 ? 'M6 13Q9 10 12 13' : ap < 100 ? 'M6 12H12' : 'M6 11Q9 14 12 11', stroke: color, strokeWidth: 2, strokeLinecap: 'round' }),
);
const comfortIcon = () => h('svg', { width: 20, height: 20, viewBox: '0 0 48 48', fill: 'none' },
  h('path', { d: 'M31 43C31 43 18 44 11 36C4 28 4 4 4 4C4 4 28 3 36 9C44 15 42 32 42 32', stroke: '#66c02f', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M44 44C44 44 32.8207 35.5515 26 28C19.1793 20.4485 16 13 16 13', stroke: '#66c02f', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M26 28L27 15M26 28L16 27', stroke: '#66c02f', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
);
const powerIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 48 48', fill: 'none' },
  h('path', { d: 'M13 14H6C4.89543 14 4 14.8954 4 16V32C4 33.1046 4.89543 34 6 34H13M31 34H38C39.1046 34 40 33.1046 40 32V16C40 14.8954 39.1046 14 38 14H31M22.002 14L17 24.0012H27.004L22 34', stroke: '#adfe2e', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M42 20H44C45.1046 20 46 20.8954 46 22V26C46 27.1046 45.1046 28 44 28H42V20Z', fill: '#adfe2e' }),
);
// ponytail: minimal office icon, 3 paths only, no extra abstraction
const officeIcon = () => h('svg', { width: 20, height: 20, viewBox: '0 0 48 48', fill: 'none' },
  h('path', { d: 'M4 34H44', stroke: '#ffffff', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M42 39L21 5', stroke: '#ffffff', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
  h('path', { d: 'M6 39L27 5', stroke: '#ffffff', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
);

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
    return box({ width: 170, height: 46, marginTop: 5, marginRight: 6, alignItems: 'center', position: 'relative' },
      h('img', { src: item.src, width: 40, height: 40 }),
      moodIcon(ap, color),
      text(item.name, { fontSize: 13, marginLeft: 2 }),
      box({ position: 'absolute', left: 0, bottom: 0, width: 150, height: 3, backgroundColor: '#444' },
        h('div', { style: { display: 'flex', width: `${ap}%`, height: 3, backgroundColor: color } }),
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
          : title === '贸易站'
            ? box({ alignItems: 'center' }, h('svg', { width: 20, height: 20, viewBox: '0 0 48 48', fill: 'none' },
                h('path', { d: 'M33.0499 7H38C39.1046 7 40 7.89543 40 9V42C40 43.1046 39.1046 44 38 44H10C8.89543 44 8 43.1046 8 42L8 9C8 7.89543 8.89543 7 10 7H16H17V10H31V7H33.0499Z', fill: 'none', stroke: '#8cd1ff', strokeWidth: 4, strokeLinejoin: 'round' }),
                h('rect', { x: 17, y: 4, width: 14, height: 6, stroke: '#8cd1ff', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                h('path', { d: 'M26.9996 19L19 27.0012H29.004L21.0003 35.0018', stroke: '#8cd1ff', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
              ), text(`${room?.strategy ?? ''}\u00a0${room?.current ?? 0}/${room?.total ?? 0}`, { color: '#8cd1ff' }))
            : title === '制造站'
              ? box({ alignItems: 'center' }, h('svg', { width: 20, height: 20, viewBox: '0 0 48 48', fill: 'none' },
                  h('path', { d: 'M44 14L24 4L4 14V34L24 44L44 34V14Z', stroke: '#d79d13', strokeWidth: 4, strokeLinejoin: 'round' }),
                  h('path', { d: 'M4 14L24 24', stroke: '#d79d13', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                  h('path', { d: 'M24 44V24', stroke: '#d79d13', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                  h('path', { d: 'M44 14L24 24', stroke: '#d79d13', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                  h('path', { d: 'M34 9L14 19', stroke: '#d79d13', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                ), text(`${room?.item ?? ''}\u00a0${room?.current ?? 0}/${room?.total ?? 0}`, { color: '#d79d13' }))
              : title === '会客室'
                ? box({ alignItems: 'center' }, room?.sharing ? box({ position: 'absolute', marginLeft: -200, color: '#eb9712', alignItems: 'center' }, h('svg', { width: 20, height: 20, viewBox: '0 0 48 48', fill: 'none' },
                    h('path', { d: 'M13.5 39.3706C16.3908 41.6439 20.0371 42.9999 24 42.9999C27.9629 42.9999 31.6092 41.6439 34.5 39.3706', stroke: '#ee810a', strokeWidth: 4 }),
                    h('path', { d: 'M19 9.74707C12.0513 11.8822 7 18.3511 7 25.9999C7 27.9247 7.31989 29.7748 7.9094 31.4999', stroke: '#ee810a', strokeWidth: 4 }),
                    h('path', { d: 'M29 9.74707C35.9487 11.8822 41 18.3511 41 25.9999C41 27.9247 40.6801 29.7748 40.0906 31.4999', stroke: '#ee810a', strokeWidth: 4 }),
                    h('path', { d: 'M43 36C43 37.3416 42.4716 38.5597 41.6117 39.4577C40.7015 40.4082 39.4199 41 38 41C35.2386 41 33 38.7614 33 36C33 33.9899 34.1861 32.2569 35.8967 31.4626C36.536 31.1657 37.2487 31 38 31C40.7614 31 43 33.2386 43 36Z', fill: 'none', stroke: '#ee810a', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                    h('path', { d: 'M15 36C15 37.3416 14.4716 38.5597 13.6117 39.4577C12.7015 40.4082 11.4199 41 10 41C7.23858 41 5 38.7614 5 36C5 33.9899 6.18614 32.2569 7.89667 31.4626C8.53604 31.1657 9.24867 31 10 31C12.7614 31 15 33.2386 15 36Z', fill: 'none', stroke: '#ee810a', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                    h('path', { d: 'M29 9C29 10.3416 28.4716 11.5597 27.6117 12.4577C26.7015 13.4082 25.4199 14 24 14C21.2386 14 19 11.7614 19 9C19 6.98991 20.1861 5.25686 21.8967 4.4626C22.536 4.16572 23.2487 4 24 4C26.7614 4 29 6.23858 29 9Z', fill: 'none', stroke: '#ee810a', strokeWidth: 4, strokeLinecap: 'round', strokeLinejoin: 'round' }),
                  ), text('线索交流开启中')) : null, text('线索'), ...(room?.board ?? []).map((value) => box({ width: 20, justifyContent: 'center', border: '1px solid #fff', borderRadius: 5, marginLeft: 4 }, text(value))))
                : title === '宿舍'
                  ? box({ alignItems: 'center' }, comfortIcon(), text(`舒适度${room?.comfort ?? 0}`, { color: '#66c02f' }))
                  : title === '发电站'
                    ? box({ alignItems: 'center' }, powerIcon(), text(room?.power ?? 0, { color: '#adfe2e' }))
                    : title === '办公室'
                      ? box({ alignItems: 'center' }, officeIcon(), text(`刷新次数${props.hire?.refreshCount ?? room?.refreshCount ?? 3}`, { marginLeft: 4 }))
                      : text(room?.strategy ?? room?.power ?? room?.comfort ?? room?.refreshCount ?? '', { color: '#8cd1ff' }),
      ),
      box({ marginLeft: 10, alignItems: 'center', flexWrap: 'wrap' }, chars.length ? chars.map(characterCard) : text('空无一人', { marginLeft: 10 })),
    ),
  );
  return box({ width: 1110, height: 612, flexDirection: 'column', overflow: 'hidden', backgroundColor: '#2b333d', fontFamily: 'NotoSansHans' },
    box({ height: 27, justifyContent: 'space-between', alignItems: 'center', color: '#fff' }, text('基建信息', { fontSize: 19, fontWeight: 600, marginLeft: 10 }), box({ flexDirection: 'column', alignItems: 'flex-start', marginRight: 30 }, box({ alignItems: 'center' }, laborIcon(), text(`${props.labor?.current ?? 0}/${props.labor?.total ?? 0}`)), box({ width: 100, height: 3, marginTop: 2, backgroundColor: '#555' }, h('div', { style: { display: 'flex', width: `${Math.min(100, 100 * (props.labor?.current ?? 0) / Math.max(1, props.labor?.total ?? 1))}%`, height: 3, backgroundColor: '#fff' } })))),
    box({ flexWrap: 'wrap', alignContent: 'flex-start' }, prepared.map(roomCard)),
  );
}
