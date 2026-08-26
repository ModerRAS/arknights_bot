// Operator card — faithful rebuild of template/Operator.tmpl (frozen Playwright baseline).
// Geometry measured from the frozen baseline at DPR 1.5 (CSS px = image px / 1.5)
// and from quirks-mode Chrome layout of the template itself:
// - attr table: 3 rows x 26px on 32px pitch? no: rows 26px at y=20/46/72, label 102 + value 72
// - potential table at y=118: header 26 + 2 data rows 26, table 139 wide
// - pb block at (10,482.39) 300x155.09, names block at (10,652.48) 300x107.52 (bottom 5% = 40)
// - talent y=20 rows 20px, col edges rel 600: [164.3, 186.7, 249]
// - bsk y=60 header 20 + row 38, col edges rel: [53.6, 119.4, 102, 325]
// - skill y=118 72 tall (20+32+20), col edges rel: [116.6, 452, 31.4]
// Chrome UA quirks-mode metrics: line-height normal = 1.5em (NotoSansHans hhea),
// h1 32/48, h3 18.72/28.08, h5 13.28/19.92, td padding 1px.
import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';
const box = (style, ...children) => h('div', { style: { position: 'absolute', display: 'flex', ...style } }, ...children);
const text = (value, style = {}) => h('span', { style }, String(value ?? ''));

const GRAY_ROW = 'rgba(176,177,177,.7)';
const DARK_ROW = 'rgba(63,63,62,.99)';
const HEADER_BG = 'rgba(0,0,0,.8)';
const HEADER_FG = 'rgba(255,255,255,.8)';
const PANEL_BG = 'rgba(239,238,239,.4)';
const VALUE_BG = 'rgba(239,238,239,.8)';
const VALUE_FG = 'rgba(0,0,0,.8)';

// attr table: rows y=20/46/72, label cells 102 wide at x=0/174/348, value cells 72 at x=102/276/450
const ATTR_ROWS = [
  [['最大生命值', 'hp'], ['攻击力', 'atk'], ['防御力', 'def']],
  [['法抗', 'res'], ['攻击间隔', 'interval'], ['再部署时间', 'reDeploy']],
  [['阻挡数', 'block'], ['部署费用', 'cost'], ['所属', 'logo']],
];
const cell = (x, y, w, bg, fg, value) => box({ left: x, top: y, width: w, height: 26, backgroundColor: bg, paddingLeft: 1, paddingTop: 1.33 },
  text(value, { color: fg, fontSize: 16, lineHeight: '24px', whiteSpace: 'nowrap' }));

export default async function render(props, { image }) {
  const [background, painting, professionIcon, rarityImg, building, skill, potentialIcons] = await Promise.all([
    image('/assets/operator/bg.png'), image(props.painting, fallback),
    image(`/assets/box/${(props.op?.profession ?? 'CASTER')}.png`, fallback),
    image(`/assets/box/Rarity_${(props.op?.rarity ?? 0)}.png`, fallback),
    image(props.buildingSkills?.[0]?.icon, fallback), image(props.skills?.[0]?.icon, fallback),
    Promise.all((props.potentials ?? []).map((p) => image(`/assets/box/Potential_${p.rank}.png`, fallback))),
  ]);
  const data = props.op ?? {};
  const branch = props.professionBranch ?? {};

  const attrCells = [];
  ATTR_ROWS.forEach((row, ri) => row.forEach(([label, key], ci) => {
    attrCells.push(cell(ci * 174, 20 + ri * 26, 102, HEADER_BG, HEADER_FG, label));
    attrCells.push(cell(ci * 174 + 102, 20 + ri * 26, 72, VALUE_BG, VALUE_FG, data[key]));
  }));

  // potential table: header 26 at y=118, data rows 26 at y=144/170, table 139 wide
  const potentialRows = (props.potentials ?? []).map((p, i) => box({ left: 0, top: 146 + i * 26, width: 139, height: 26, backgroundColor: PANEL_BG },
    h('img', { src: potentialIcons[i], width: 20, height: 18.67, style: { position: 'absolute', left: 3, top: 3.67 } }),
    text(p.desc, { position: 'absolute', left: 27, top: 3.67, fontSize: 16, lineHeight: '24px', whiteSpace: 'nowrap' })));

  // pb block: measured (10,482.39) 300x155.09
  const pb = box({ left: 10, top: 482.39, width: 300, height: 155.09, backgroundColor: PANEL_BG },
    h('img', { src: professionIcon, width: 75, height: 75, style: { position: 'absolute', left: 3, top: 3 } }),
    h('img', { src: rarityImg, width: 75, height: 21, style: { position: 'absolute', left: 82, top: 4 } }),
    text(branch.name, { position: 'absolute', left: 82, top: 30.11, fontSize: 18.72, lineHeight: '28.08px', fontWeight: 600, whiteSpace: 'nowrap' }),
    text(data.tags, { position: 'absolute', left: 3, top: 82, fontSize: 16, lineHeight: '24px', whiteSpace: 'nowrap' }),
    text(branch.desc, { position: 'absolute', left: 3, top: 110, fontSize: 13.28, lineHeight: '19.92px', fontWeight: 600, whiteSpace: 'nowrap' }));

  // names block: measured (10,652.48) 300x107.52
  const names = box({ left: 10, top: 652.48, width: 300, height: 107.52, backgroundColor: PANEL_BG },
    text(data.name, { position: 'absolute', left: 3, top: 3, fontSize: 32, lineHeight: '48px', fontWeight: 600 }),
    text(props.attackRange, { position: 'absolute', left: 101.3, top: 3, fontSize: 32, lineHeight: '48px' }),
    box({ left: 3, top: 81.12, height: 18.72, backgroundColor: '#000', paddingLeft: 1 },
      text(data.code, { color: '#fff', fontSize: 18.72, lineHeight: '18.72px', fontWeight: 600, whiteSpace: 'nowrap' })),
    text(data.nameEn, { position: 'absolute', left: 102.3, top: 76.44, fontSize: 18.72, lineHeight: '28.08px', fontWeight: 600, whiteSpace: 'nowrap' }));

  // talent table: header 20 at y=20, data row 20 at y=40; col edges 764.3 / 951
  const talentHeader = (title) => box({ left: 600, top: 20, width: 600, height: 20, backgroundColor: HEADER_BG, paddingLeft: 1 },
    text(title, { color: HEADER_FG, fontSize: 12, lineHeight: '18px' }));
  const talentRow = (props.talents ?? []).map((t) => box({ left: 600, top: 40, width: 600, height: 20, backgroundColor: GRAY_ROW },
    text(t.evolve, { position: 'absolute', left: 1, top: 1, fontSize: 12, lineHeight: '18px', color: '#000', whiteSpace: 'nowrap' }),
    text(t.name, { position: 'absolute', left: 165.3, top: 1, fontSize: 12, lineHeight: '18px', color: '#000', whiteSpace: 'nowrap' }),
    text(t.desc, { position: 'absolute', left: 352, top: 1, fontSize: 12, lineHeight: '18px', color: '#000' })));

  // bsk table: header 20 at y=60, row 38 at y=80; col edges 653.6 / 773 / 875
  const bskRows = (props.buildingSkills ?? []).map((item, i) => item.desc
    ? box({ left: 600, top: 80 + i * 38, width: 600, height: 38, backgroundColor: GRAY_ROW },
      text(item.evolve, { position: 'absolute', left: 1, top: 1, fontSize: 12, lineHeight: '18px', color: '#000', whiteSpace: 'nowrap' }),
      i === 0 ? h('img', { src: building, width: 36, height: 36, style: { position: 'absolute', left: 93.3, top: 1 } }) : null,
      text(item.name, { position: 'absolute', left: 174, top: 1, fontSize: 12, lineHeight: '18px', color: '#000', whiteSpace: 'nowrap' }),
      text(item.desc, { position: 'absolute', left: 276, top: 1, fontSize: 12, lineHeight: '18px', color: '#000' }))
    : box({ left: 600, top: 80 + i * 38, width: 600, height: 38, backgroundColor: GRAY_ROW },
      text(item.evolve, { position: 'absolute', left: 1, top: 1, fontSize: 12, lineHeight: '18px', color: '#000' })));

  // skill table: 600x72 at (600,118), dark bg; col1 [0,116.6] col2 [116.6,568.6] col3 [568.6,600] (rel)
  const skillRows = (props.skills ?? []).map((item, i) => {
    const top = 118 + i * 72;
    return box({ left: 600, top, width: 600, height: 72, backgroundColor: DARK_ROW },
      i === 0 ? h('img', { src: skill, width: 50, height: 50, style: { position: 'absolute', left: 33.3, top: 1 } }) : null,
      box({ left: 0, top: 53, width: 116.6, height: 18, justifyContent: 'center', alignItems: 'center' },
        text(item.name, { fontSize: 12, lineHeight: '18px', color: '#fff', whiteSpace: 'nowrap' })),
      box({ left: 116.6, top: 1, width: 452, height: 18, justifyContent: 'center', alignItems: 'center' },
        ...(item.spType ?? []).map((sp) => text(sp, { fontSize: 12, lineHeight: '18px', color: '#fff', whiteSpace: 'nowrap', marginRight: 5 })),
        h('div', { style: { display: 'flex', height: 12, backgroundColor: '#808080', borderRadius: 3, paddingLeft: 5, paddingRight: 5, marginRight: 8.3 } },
          text(`技力${item.spInit}/${item.spCost}`, { fontSize: 12, lineHeight: '12px', color: '#fff', whiteSpace: 'nowrap' })),
        item.duration ? h('div', { style: { display: 'flex', height: 12, backgroundColor: '#808080', borderRadius: 3, paddingLeft: 5, paddingRight: 5 } },
          text(`持续时间${item.duration}s`, { fontSize: 12, lineHeight: '12px', color: '#fff', whiteSpace: 'nowrap' })) : null),
      text(item.desc, { position: 'absolute', left: 128.93, top: 37, fontSize: 12, lineHeight: '18px', color: '#fff' }),
      box({ left: 568.6, top: 1, width: 31.4, height: 70, justifyContent: 'center', alignItems: 'center' },
        text(item.skillRange, { fontSize: 12, lineHeight: '18px', color: '#fff', whiteSpace: 'nowrap' })));
  });

  return box({ left: 0, top: 0, width: 1200, height: 800, backgroundImage: `url(${background})`, backgroundSize: 'cover', fontFamily: 'NotoSansHans', fontSize: 16, color: '#000', overflow: 'hidden' },
    h('img', { src: painting, width: 650, height: 650, style: { position: 'absolute', left: 60, top: 150 } }),
    ...attrCells,
    box({ left: 0, top: 120, width: 139, height: 26, backgroundColor: HEADER_BG, paddingLeft: 1, paddingTop: 1.33 },
      text('潜能提升', { color: HEADER_FG, fontSize: 16, lineHeight: '24px' })),
    ...potentialRows,
    pb, names,
    talentHeader('天赋'),
    ...talentRow,
    box({ left: 600, top: 60, width: 600, height: 20, backgroundColor: HEADER_BG, paddingLeft: 1 },
      text('基建技能', { color: HEADER_FG, fontSize: 12, lineHeight: '18px' })),
    ...bskRows,
    ...skillRows,
  );
}
