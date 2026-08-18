import { h } from '../lib/h.mjs';

const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });

const imageEntry = (entry, avatar, includeCount) => h('div', { style: style({ width: 230, height: 150, padding: '12px 4px', alignItems: 'center', gap: includeCount ? 3 : 12 }) },
  h('img', { src: avatar, width: 100, height: 100 }),
  h('div', { style: style({ height: 110, flexDirection: 'column', gap: includeCount ? 12 : 8, color: '#eee', ...(includeCount ? { transform: 'translateY(9px)' } : {}) }) },
    h('span', { style: { fontSize: includeCount ? 16 : 17 } }, entry.name, entry.isNew ? h('span', { style: { marginLeft: 6, color: 'red', fontSize: 10 } }, 'New') : null),
    h('span', { style: { fontSize: 15 } }, entry.date),
    h('span', { style: { fontSize: 15 } }, entry.pool),
    includeCount ? h('span', { style: { fontSize: 15 } }, `花费${entry.count}抽`) : null,
  ),
);

const listBox = (title, entries, resolved, includeCount) => h('div', { style: style({ width: 465, minHeight: 250, flexWrap: 'wrap', alignContent: 'flex-start', padding: 0, border: '1px solid #000', borderRadius: 20, backgroundColor: '#1f1e1e' }) },
  h('div', { style: style({ transform: 'translateY(15px)', width: '100%', height: 38, justifyContent: 'center', color: '#eee', fontSize: 20 }) }, title),
  ...entries.map((entry, index) => imageEntry(entry, resolved[index], includeCount)),
);

const v6Pie = (rarities) => {
  const total = rarities.reduce((sum, item) => sum + item.count, 0) || 1;
  let offset = 0;
  return rarities.map((item) => {
    const start = offset;
    offset += item.count / total * 360;
    const radius = 128;
    const a = (start - 90) * Math.PI / 180;
    const b = (offset - 90) * Math.PI / 180;
    return h('path', { d: `M 150 150 L ${150 + radius * Math.cos(a)} ${150 + radius * Math.sin(a)} A ${radius} ${radius} 0 ${offset - start > 180 ? 1 : 0} 1 ${150 + radius * Math.cos(b)} ${150 + radius * Math.sin(b)} Z`, fill: item.color });
  });
};

const v11PoolCard = (pools) => {
  const values = [...pools].reverse();
  const max = Math.max(1, ...values.map((pool) => pool.count));
  const barWidth = 195;
  return h('div', { style: style({ position: 'absolute', left: 660, top: 150, width: 300, height: 250, border: '1px solid #000', borderRadius: 20, backgroundColor: '#1f1e1e', overflow: 'hidden' }) },
    h('div', { style: style({ position: 'absolute', left: 0, top: 10, width: '100%', height: 24, justifyContent: 'center', color: '#eee', fontSize: 20 }) }, '卡池分布(最近10个)'),
    h('div', { style: style({ position: 'absolute', left: 0, top: 33, width: 300, height: 250, color: '#fff', fontSize: 14 }) },
      ...values.map((pool, index) => {
        const top = 42 + index * 100;
        const width = pool.count / max * barWidth;
        return h('div', { style: style({ position: 'absolute', left: 0, top, width: 300, height: 36, alignItems: 'center' }) },
          h('span', { style: { width: 75, textAlign: 'right', paddingRight: 5 } }, pool.poolName),
          h('div', { style: style({ width, height: 36, backgroundColor: '#5470c6' }) }),
          h('span', { style: { marginLeft: 8 } }, String(pool.count)),
        );
      }),
    ),
  );
};

export default async function render(props, { image }) {
  const [header, footer, ...avatars] = await Promise.all([
    image('/assets/gacha/header.png'),
    image('/assets/gacha/footer.png'),
    ...(props.chars || []).map((entry) => image(entry.avatar, '/assets/common/amiya.png')),
    ...(props.star6Info || []).map((entry) => image(entry.avatar, '/assets/common/amiya.png')),
  ]);
  const chars = avatars.slice(0, (props.chars || []).length);
  const star6 = avatars.slice(chars.length);
  const rarities = props.rarities || [];
  const pie = v6Pie(rarities);

  return h('div', { style: style({ width: 1000, height: 882, overflow: 'hidden', flexDirection: 'column', fontFamily: 'NotoSansHans', backgroundColor: '#0c0d0c' }) },
    h('div', { style: style({ position: 'relative', width: 1000, height: 400, flexShrink: 0, backgroundImage: `url(${header})` }) },
      h('span', { style: { position: 'absolute', left: 320, top: 50, color: '#eee', fontSize: 32, fontWeight: 700 } }, props.name),
      h('span', { style: { position: 'absolute', left: 250, top: 112, color: '#eee', fontSize: 28, fontWeight: 700 } }, `共${props.total}抽`, h('span', { style: { fontSize: 23 } }, `(${props.period})`)),
      h('div', { style: style({ position: 'absolute', left: 20, top: 150, width: 300, height: 250, padding: 12, flexDirection: 'column', border: '1px solid #000', borderRadius: 20, backgroundColor: '#1f1e1e' }) },
        h('span', { style: style({ width: '100%', justifyContent: 'center', alignItems: 'center', color: '#eee', fontSize: 20 }) }, '星级分布'),
        h('div', { style: style({ alignItems: 'center', gap: 8 }) },
          h('div', { style: style({ width: 100, flexDirection: 'column', gap: 12, transform: 'translateY(13px)', color: '#fff', fontSize: 14 }) }, ...rarities.map((item) => h('span', null, `${item.label}  ${item.percent.toFixed(2)}%`))),
          h('svg', { width: 180, height: 180, viewBox: '0 0 300 300' }, ...pie),
        ),
      ),
      h('div', { style: style({ position: 'absolute', left: 347, top: 150, width: 300, height: 250, padding: 22, flexDirection: 'column', gap: 20, border: '1px solid #000', borderRadius: 20, backgroundColor: '#1f1e1e', color: '#eee' }) },
        h('span', { style: style({ width: '100%', justifyContent: 'center', alignItems: 'center', fontSize: 20 }) }, '星级分布'),
        ...(props.averages || []).map((item) => h('div', { style: style({ justifyContent: 'space-between', fontSize: 18 }) }, h('span', null, `${item.count}个${item.label}`), h('span', null, `${item.avg}抽/个`))),
      ),
      v11PoolCard(props.pools || []),
    ),
    h('div', { style: style({ width: 1000, height: 213, padding: '40px 20px 0', gap: 26, backgroundColor: '#0c0d0c' }) },
      listBox('新获得干员(至多显示20个)', props.chars || [], chars, false),
      listBox('获得六星干员(至多显示20个)', props.star6Info || [], star6, true),
    ),
    h('div', { style: style({ position: 'relative', width: 1000, height: 269, flexShrink: 0, backgroundImage: `url(${footer})` }) },
      h('span', { style: { position: 'absolute', left: 530, top: 190, color: '#eee', fontSize: 32, fontWeight: 700 } }, props.today),
    ),
  );
}
