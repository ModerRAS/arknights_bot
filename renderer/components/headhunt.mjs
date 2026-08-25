import { h } from '../lib/h.mjs';
const fallback='assets/common/amiya.png';
// Frozen Headhunt.tmpl geometry, measured from the Playwright baseline:
// .bg inline-blocks pitch 98.6 = 95px card + one whitespace space (fit so that
// round(25+i*98.6) reproduces all 10 measured frame x-positions); box 95x370
// (270 content + 100 padding-top) at y=130; back_N.png cover-fills the padding box
// (110x230 -> 176.96x370, top-left anchored, clipped to 95); portrait 95x190 at
// content top y=230; rarity/profession are abspos at static position centered
// (x+10) with profession margin-top 190 -> y=420.
const PITCH = 98.6;

export default async function render(props, { image }) {
  const [bg, cards] = await Promise.all([
    image('assets/headhunt/bg.png', fallback),
    Promise.all((props ?? []).map(async (item) => ({
      item,
      back: await image(`assets/headhunt/back_${item.rarity}.png`, fallback),
      portrait: await image(item.thumbURL, fallback),
      profession: await image(`assets/headhunt/${item.profession}.png`, fallback),
      rarity: await image(`assets/headhunt/Rarity_${item.rarity}.png`, fallback),
    }))),
  ]);
  return h(
    'div',
    {
      style: {
        width: 1049, height: 576, paddingLeft: 25, position: 'relative', display: 'flex',
        backgroundImage: `url(${bg})`, backgroundSize: '1024px 576px',
        fontFamily: 'NotoSansHans',
      },
    },
    cards.map(({ back, portrait, profession, rarity }, i) => {
      const x = Math.round(25 + i * PITCH);
      return h(
        'div',
        { style: { position: 'absolute', left: x, top: 130, width: 95, height: 370, display: 'flex' } },
        h(
          'div',
          { style: { position: 'absolute', left: 0, top: 0, width: 95, height: 370, overflow: 'hidden', display: 'flex' } },
          h('img', { src: back, width: 176.96, height: 370, style: { position: 'absolute', left: 0, top: 0 } }),
        ),
        h('img', { src: portrait, width: 95, height: 190, style: { position: 'absolute', left: 0, top: 100 } }),
        h('img', { src: rarity, width: 75, height: 20, style: { position: 'absolute', left: 10, top: 100 } }),
        h('img', { src: profession, width: 75, height: 76, style: { position: 'absolute', left: 10, top: 290 } }),
      );
    }),
  );
}
