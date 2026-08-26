import { h } from '../lib/h.mjs';

const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });

export default async function render(props) {
  const colors = { empty: ['#222', '#333', '#666', 'rgba(255,255,255,.03)'], selected: ['#1e3a3f', '#00e5ff', '#00e5ff', 'rgba(0,229,255,.1)'], winner: ['#4a1c1c', '#ff3d00', '#ff3d00', 'rgba(255,61,0,.15)'] };
  return h('div', { style: style({ width: 982, height: 1111, padding: 40, flexDirection: 'column', fontFamily: 'NotoSansHans', color: '#fff', backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: 16, boxShadow: '0 10px 50px rgba(0,0,0,.8)' }) },
    h('div', { style: style({ height: 76, justifyContent: 'center', alignItems: 'flex-start' }) }, h('span', { style: { paddingBottom: 10, borderBottom: '2px solid #00e5ff', fontSize: 28, fontWeight: 700, letterSpacing: 4 } }, '选号详情')),
    h('div', { style: style({ width: 900, height: 900, flexWrap: 'wrap', gap: 8, marginBottom: 30 }) }, (props.cells || []).map((cell) => {
      const [background, border, accent, watermark] = colors[cell.state] || colors.empty;
      return h('div', { style: style({ position: 'relative', width: 82.8, height: 82.8, padding: 6, flexDirection: 'column', justifyContent: 'space-between', overflow: 'hidden', backgroundColor: background, border: `1px solid ${border}`, borderRadius: 6 }) },
        h('span', { style: { position: 'absolute', left: 0, top: 12, width: '100%', textAlign: 'center', fontSize: 32, fontWeight: 800, color: watermark } }, String(cell.number)),
        h('span', { style: { zIndex: 1, fontSize: 16, fontWeight: 700, color: accent } }, String(cell.number)),
        cell.state !== 'empty' ? h('div', { style: style({ zIndex: 1, flexDirection: 'column', gap: 4 }) }, h('span', { style: { fontSize: 12, fontWeight: 700, color: cell.state === 'winner' ? '#ff3d00' : '#fff' } }, cell.userName), h('span', { style: { fontSize: 10, color: 'rgba(255,255,255,.7)' } }, cell.userNumber)) : null,
      );
    })),
    h('div', { style: style({ width: '100%', justifyContent: 'center', alignItems: 'center', gap: 30, fontSize: 14 }) },
      ...[['empty', '未选择'], ['selected', '已占位'], ['winner', '中奖']].map(([kind, label]) => h('div', { style: style({ alignItems: 'center', gap: 8 }) }, h('span', { style: { display: 'flex', width: 16, height: 16, border: `1px solid ${colors[kind][1]}`, borderRadius: 4, backgroundColor: colors[kind][0] } }), h('span', null, label))),
    ),
  );
}
