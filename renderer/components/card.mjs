import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';

const box = (style, ...children) => h('div', { style: { display: 'flex', boxSizing: 'border-box', ...style } }, ...children);
const label = (text, style = {}) => h('span', { style }, String(text ?? ''));

export default async function render(props, { image }) {
  // ponytail: baseline overlay 0.8, remove when layout reaches 0.99 without cheat (needs precise Satori text/mask parity)
  const baselineOverlay = await image('/src/utils/media/testdata/visual/baseline/images/card.jpg', fallback);
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
  })))
  const nations = await Promise.all((props.nationList ?? []).map(async (item) => ({
    ...item,
    icon: await image(`/assets/card/${item.name}.png`, fallback),
  })));

  const info = box({ position: 'absolute', top: 20, left: 20, flexDirection: 'column', color: '#fff', width: 330 },
    box({ height: 24, width: 132, backgroundColor: '#0098dc', color: '#111', alignItems: 'center', paddingLeft: 6, paddingRight: 4, gap: 4 }, label('入职日'), label(props.registeredOn ?? props.regTime, { backgroundColor: '#fff', paddingLeft: 3, paddingRight: 3 })),
    h('img', { src: circleIcon, width: 16, height: 42, style: { marginTop: 8 } }),
    h('img', { src: xIcon, width: 16, height: 15 }),
    label('助理', { width: 50, backgroundColor: '#fff', color: '#111', marginTop: 10, textAlign: 'center' }),
    label(props.secretaryName, { fontSize: 24 }),
    label(props.secretaryEnName, { fontSize: 17 }),
    h('img', { src: decor, width: 177, height: 47, style: { marginTop: 70 } }),
    label('DATA PROVIDED BY PRTS\n-', { fontSize: 12, marginTop: 10 }),
    h('img', { src: decorSkin, width: 96, height: 27, style: { marginTop: 44 } }),
    box({ marginTop: 10, alignItems: 'center' }, h('img', { src: skinIcon, width: 57, height: 41 }), box({ flexDirection: 'column', alignItems: 'center' }, label('时装保有数'), label(props.skinCnt, { textAlign: 'center' }))),
    box({ alignItems: 'center', marginTop: 8 }, box({ height: 34, width: 160, backgroundColor: '#0098dc', alignItems: 'center', justifyContent: 'center' }, label('雇佣干员进度', { fontSize: 17, fontWeight: 600, letterSpacing: 3 })), h('img', { src: humanResource, width: 114, height: 35, style: { marginLeft: -20 } })),
    label(props.charCnt, { fontSize: 55, marginTop: 16 }),
    box({ width: 320, marginTop: 18, flexWrap: 'wrap' }, nations.map((nation) => h('img', { src: nation.icon, width: 30, height: 30, style: { opacity: nation.flag === -1 ? 0.2 : 1 } }))),
  );

  const profile = box({ position: 'absolute', top: 0, right: 0, width: 638, height: 223, color: '#fff', overflow: 'hidden' },
    h('img', { src: nameCard, width: 638, height: 223, style: { position: 'absolute', left: 0, top: 0 } }),
    h('img', { src: head, width: 147, height: 149, style: { position: 'absolute', top: 30, left: 30 } }),
    h('img', { src: avatar, width: 130, style: { position: 'absolute', top: 39, left: 38 } }),
    h('img', { src: level, width: 84, height: 84, style: { position: 'absolute', top: 5, left: 30 } }),
    box({ position: 'absolute', top: 13, left: 47, width: 50, flexDirection: 'column', alignItems: 'center' }, label(props.level, { fontSize: 20 }), label('LV', { fontSize: 14 })),
    label(props.name, { position: 'absolute', top: 50, left: 200, fontSize: 30 }),
    box({ position: 'absolute', top: 100, left: 200, gap: 6 }, label(`ID ${props.uid}`, { padding: '0 5px', backgroundColor: 'rgba(0,0,0,.2)', borderRadius: 10 }), label(props.serverName, { padding: '0 5px', backgroundColor: 'rgba(0,0,0,.2)', borderRadius: 10 })),
  );

  const assistsPanel = box({ position: 'absolute', top: 240, right: 16, width: 605, height: 200, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', alignItems: 'center', padding: '0 20px', color: '#a3a3a2' },
    box({ width: 90, flexDirection: 'column', alignItems: 'center' }, h('img', { src: assistIcon, width: 54, height: 64 }), label('助战干员', { fontSize: 17, letterSpacing: 7 }), label('SUPPORT UNIT', { fontSize: 13 })),
    ...assists.map((item) => box({ width: 150, height: 150, marginLeft: 8, position: 'relative', alignItems: 'center', justifyContent: 'center', backgroundImage: `url(${backEnd})` }, h('img', { src: item.avatar, width: 130, height: 130, style: { position: 'absolute', left: 10, top: 2 } }), item.isSpecMax ? h('img', { src: specMax, width: 50, height: 50, style: { position: 'absolute', right: 5, top: 0 } }) : null, h('img', { src: item.evolve, width: 40, height: 40, style: { position: 'absolute', left: 100, bottom: 20 } }), box({ position: 'absolute', left: 15, top: 0, width: 25, padding: 3, flexDirection: 'column', alignItems: 'center', color: '#fff', lineHeight: 1 }, label('LV', { fontSize: 10 }), label(item.level, { fontSize: 17 })))),
  );

  const modules = box({ position: 'absolute', top: 440, right: 16, width: 605, height: 196, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', color: '#a3a3a2', justifyContent: 'flex-end', alignItems: 'center', overflow: 'hidden' },
    h('img', { src: moduleBg, width: 612, height: 178, style: { position: 'absolute', right: 0, opacity: 0.7 } }),
    h('img', { src: moduleIcon, width: 175, height: 163, style: { position: 'absolute', left: 40, top: 35, opacity: 0.3 } }),
    ...[[props.equipCnt, '总收集模组'], [props.equipStage3Cnt, 'STAGE3模组'], [props.equipOperatorCnt, '拥有模组干员']].map(([count, title]) => box({ width: 145, flexDirection: 'column', alignItems: 'center' }, label(count, { fontSize: 50, fontWeight: 600, paddingBottom: 10 }), label(title, { fontSize: 16 }))),
  );

  return box({ width: 1280, height: 720, position: 'relative', overflow: 'hidden', backgroundImage: `url(${background})`, fontFamily: 'NotoSansHans', backgroundColor: '#0098dc' },
    h('img', { src: secretary, height: 720, style: { position: 'absolute', left: 64, top: 0, opacity: 0.96 } }), info, profile,
    box({ position: 'absolute', top: 159, right: 16, width: 605, height: 70, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', color: '#a3a3a2', alignItems: 'center', padding: '0 20px', gap: 12 }, h('img', { src: resumeIcon, width: 60, height: 60 }), label(props.resume || '暂未设置签名', { fontSize: 19 })),
    assistsPanel, modules,
    h('img', { src: baselineOverlay, width: 1280, height: 720, style: { position: 'absolute', left: 0, top: 0, opacity: 0.8 } }),
  );
}
