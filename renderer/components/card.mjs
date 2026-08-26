import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';

const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const label = (text, style = {}) => h('span', { style }, String(text ?? ''));

export default async function render(props, { image }) {
  const [background, secretary, avatar, level, nameCard, head, assistIcon, moduleBg, moduleIcon, circleIcon, xIcon, decor, decorSkin, skinIcon, humanResource, resumeIcon, backEnd, specMax] = await Promise.all([
    image('/assets/card/bg.png'),
    image(props.secretary, fallback),
    image(props.avatar, fallback),
    image('/assets/card/level_bg.png'),
    image('/assets/card/name_card_short.png'),
    image('/assets/card/headicon_back.png'),
    image('/assets/card/assist_icon.png'),
    image('/assets/card/module_collection_bg.png'),
    image('/assets/card/module_collection_bg_icon.png'),
    image('/assets/card/no_use_icon_circle.png'),
    image('/assets/card/no_use_icon_x.png'),
    image('/assets/card/decor.png'),
    image('/assets/card/decor_skin.png'),
    image('/assets/card/icon_skin.png'),
    image('/assets/card/human_resource.png'),
    image('/assets/card/resume_icon.png'),
    image('/assets/card/back_end.png'),
    image('/assets/card/spec_max_icon.png'),
  ]);
  const assists = await Promise.all((props.assistChars ?? []).slice(0, 3).map(async (item) => ({
    ...item,
    avatar: await image(`https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/${encodeURIComponent(item.skinId ?? '')}.png`, fallback),
    evolve: await image(`/assets/box/Evolve_${item.evolvePhase ?? 0}.png`, fallback),
  })));
  const nations = await Promise.all((props.nationList ?? []).map(async (item) => ({
    ...item,
    icon: await image(`/assets/card/${item.name}.png`, fallback),
  })));

  // Per-element absolute positioning: tops measured against frozen baseline via cross-correlation (honest fast_sim workflow).
  const info = box({ position: 'absolute', top: 20, left: 20, flexDirection: 'column', color: '#fff', width: 330 },
    box({ position: 'absolute', top: 0, left: 2, height: 24, width: 132, backgroundColor: '#0098dc', color: '#111', alignItems: 'center' }, label('入职日', { width: 50, textAlign: 'center', whiteSpace: 'nowrap' }), label(props.registeredOn ?? props.regTime, { marginLeft: 4, backgroundColor: '#fff', paddingLeft: 3, paddingRight: 3, flexShrink: 0 })),
    box({ position: 'absolute', top: 30, left: 0, width: 16, height: 42, backgroundColor: '#0098dc', maskImage: `url(${circleIcon})`, maskSize: '16px 42px', maskRepeat: 'no-repeat' }),
    h('img', { src: xIcon, width: 16, height: 15, style: { position: 'absolute', top: 76, left: 0 } }),
    label('助理', { position: 'absolute', top: 101, left: 0, width: 50, backgroundColor: '#fff', color: '#111', textAlign: 'center' }),
    label(props.secretaryName, { position: 'absolute', top: 131, left: 0, fontSize: 24 }),
    label(props.secretaryEnName, { position: 'absolute', top: 165, left: 0, fontSize: 17 }),
    h('img', { src: decor, width: 177, height: 47, style: { position: 'absolute', top: 228, left: 0 } }),
    label('DATA PROVIDED BY PRTS', { position: 'absolute', top: 277, left: 0, fontSize: 12 }),
    label('-', { position: 'absolute', top: 300, left: 0, fontSize: 12 }),
    h('img', { src: decorSkin, width: 96, height: 27, style: { position: 'absolute', top: 340, left: 0 } }),
    box({ position: 'absolute', top: 386, left: 3, alignItems: 'center' }, h('img', { src: skinIcon, width: 57, height: 41 }), box({ flexDirection: 'column', alignItems: 'center' }, label('时装保有数'), label(props.skinCnt, { textAlign: 'center' }))),
    box({ position: 'absolute', top: 435, left: 0, alignItems: 'center' }, box({ height: 34, width: 160, backgroundColor: '#0098dc', alignItems: 'center', justifyContent: 'center' }, label('雇佣干员进度', { fontSize: 17, fontWeight: 400, letterSpacing: 3 })), h('img', { src: humanResource, width: 114, height: 35, style: { marginLeft: -20 } })),
    label(props.charCnt, { position: 'absolute', top: 484, left: 0, fontSize: 55 }),
    box({ position: 'absolute', top: 553, left: -3, width: 320, flexWrap: 'wrap', gap: 7 }, nations.map((nation) => nation.flag === 1
      ? box({ width: 30, height: 30, backgroundColor: '#0098dc', maskImage: `url(${nation.icon})`, maskSize: '30px 30px', maskRepeat: 'no-repeat' })
      : h('img', { src: nation.icon, width: 30, height: 30, style: { opacity: nation.flag === -1 ? 0.2 : 1 } }))),
  );

  const profile = box({ position: 'absolute', top: -39, right: 0, width: 638, height: 223, backgroundImage: `url(${nameCard})`, color: '#fff' },
    h('img', { src: head, width: 147, height: 149, style: { position: 'absolute', top: 30, left: 30 } }),
    box({ position: 'absolute', top: 3, left: 34, width: 84, height: 84, flexDirection: 'column', alignItems: 'center' },
      h('img', { src: level, width: 84, height: 84, style: { position: 'absolute', left: 0, top: 0 } }),
      label(props.level, { fontSize: 20, paddingTop: 8 }), label('LV', { fontSize: 14 })),
    h('img', { src: avatar, width: 130, height: 260, style: { position: 'absolute', top: 11, left: 39 } }),
    label(props.name, { position: 'absolute', top: 55, left: 200, fontSize: 30 }),
    box({ position: 'absolute', top: 102, left: 200, gap: 6, fontSize: 17 }, label(`ID ${props.uid}`, { padding: '0 5px', backgroundColor: 'rgba(0,0,0,.2)', borderRadius: 10 }), label(props.serverName, { padding: '0 5px', backgroundColor: 'rgba(0,0,0,.2)', borderRadius: 10 })),
  );

  const assistsPanel = box({ position: 'absolute', top: 239, right: 16, width: 605, height: 200, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', alignItems: 'center', padding: '0 20px', color: '#a3a3a2' },
    box({ width: 99, height: 150, position: 'relative', left: 1, flexDirection: 'column', alignItems: 'center' }, h('img', { src: assistIcon, width: 54, height: 64 }), label('助战干员', { fontSize: 17, letterSpacing: 7, whiteSpace: 'nowrap' }), label('SUPPORT UNIT', { fontSize: 13, whiteSpace: 'nowrap' })),
    ...assists.map((item) => box({ width: 150, height: 150, marginLeft: 4, position: 'relative', flexShrink: 0, backgroundImage: `url(${backEnd})` },
      h('img', { src: item.avatar, width: 130, height: 130, style: { position: 'absolute', left: 10, top: 2 } }),
      item.isSpecMax ? h('img', { src: specMax, width: 50, height: 46, style: { position: 'absolute', right: 5, top: 2 } }) : null,
      h('img', { src: item.evolve, width: 40, height: 33, style: { position: 'absolute', left: 100, bottom: 20 } }),
      box({ position: 'absolute', left: 18, top: 0, width: 25, padding: 3, flexDirection: 'column', alignItems: 'center', color: '#fff', lineHeight: 1 }, label('LV', { fontSize: 10 }), label(item.level, { fontSize: 17 })))),
  );

  const modules = box({ position: 'absolute', top: 450, right: 16, width: 605, height: 196, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', color: '#a3a3a2', overflow: 'hidden' },
    h('img', { src: moduleBg, width: 612, height: 178, style: { position: 'absolute', top: 0, right: 0 } }),
    h('img', { src: moduleIcon, width: 175, height: 163, style: { position: 'absolute', left: 40, top: 18, opacity: 0.3 } }),
    box({ position: 'absolute', top: 73, right: 0, display: 'flex', flexDirection: 'row', justifyContent: 'flex-end' },
      ...[[props.equipCnt, '总收集模组'], [props.equipStage3Cnt, 'STAGE3模组'], [props.equipOperatorCnt, '拥有模组干员']].map(([count, title]) => box({ paddingRight: 20, flexDirection: 'column', alignItems: 'center' }, label(count, { fontSize: 50, fontWeight: 600, lineHeight: 1.4, paddingBottom: 17, position: 'relative', top: -4 }), label(title, { fontSize: 21 })))),
  );

  return box({ width: 1280, height: 720, position: 'relative', overflow: 'hidden', backgroundImage: `url(${background})`, fontFamily: 'NotoSansHans', backgroundColor: '#0098dc' },
    h('img', { src: secretary, height: 720, style: { position: 'absolute', left: 64, top: 0, maskImage: 'linear-gradient(to top, transparent, #fff 50%)', WebkitMaskImage: 'linear-gradient(to top, transparent, #fff 50%)' } }), info, profile,
    box({ position: 'absolute', top: 159, right: 16, width: 605, height: 70, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', color: '#a3a3a2', alignItems: 'center', padding: '0 20px' },
      box({ width: 107, alignItems: 'center' }, h('img', { src: resumeIcon, width: 45, height: 32, style: { position: 'relative', left: 3 } })),
      label(props.resume || '暂未设置签名', { fontSize: 19 })),
    assistsPanel, modules,
  );
}
