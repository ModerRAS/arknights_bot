import { h } from '../lib/h.mjs';
import { skiaBilinearDataUri } from '../lib/skia-scale.mjs';
const fallback = 'assets/common/amiya.png';
// Frozen Depot.tmpl geometry, measured from the Playwright baseline (DPR 1.5):
// #main 850px, .item inline-flex width:80 flowing with one whitespace between
// items -> column pitch 83.59 (linear fit r<0.3px), first cell origin x=1.0;
// rows pitch 80 (icon 75 + inline descent), 10 items per row at 850px;
// .count abspos lands as a dark pill at cell-relative right edge 6.3 /
// top 52 with 12px white text (pill ~39.7x17.3).
const PITCH = 83.59;
const X0 = 0.33;
const ROW_H = 81.33;
const PER_ROW = 10;

export default async function render(props, { image }) {
  const items = await Promise.all((props ?? []).map(async (item, i) => ({
    item, i,
    icon: await skiaBilinearDataUri(await image(item.icon, fallback), 150, 150),
  })));
  return h('div', { style: { width: 850, height: 156, position: 'relative', display: 'flex', backgroundColor: '#2e3031', fontFamily: 'NotoSansHans' } },
    items.map(({ item, icon, i }) => {
      const x = X0 + (i % PER_ROW) * PITCH + 2.5;
      const y = Math.floor(i / PER_ROW) * ROW_H;
      return h('div', { style: { position: 'absolute', left: x, top: y, width: 75, height: 75, display: 'flex' } },
        h('img', { src: icon, width: 75, height: 75 }),
        h('div', { style: { position: 'absolute', right: 5.6, top: 52, display: 'flex', color: '#fff', backgroundColor: 'rgba(0,0,0,0.5)', fontFamily: 'DepotSerif', fontSize: 12, lineHeight: '17.3px' } }, item.count),
      );
    }));
}
