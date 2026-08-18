import { h } from '../lib/h.mjs';

const fallback = '/assets/common/amiya.png';

const box = (style, ...children) => h('div', { style: { display: 'flex', ...style } }, ...children);
const label = (text, style = {}) => h('span', { style }, String(text ?? ''));

export default async function render(props, { image }) {
  const [background, secretary, avatar, level, nameCard, head, assistIcon, moduleBg, moduleIcon] = await Promise.all([
    image('/assets/card/bg.png'),
    image(props.secretary, fallback),
    image(props.avatar, fallback),
    image('/assets/card/level_bg.png'),
    image('/assets/card/name_card_short.png'),
    image('/assets/card/headicon_back.png'),
    image('/assets/card/assist_icon.png'),
    image('/assets/card/module_collection_bg.png'),
    image('/assets/card/module_collection_bg_icon.png'),
  ]);
  const assists = await Promise.all((props.assistChars ?? []).slice(0, 3).map(async (item) => ({
    ...item,
    avatar: await image(`https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/${encodeURIComponent(item.skinId ?? '')}.png`, fallback),
  })));
  const nations = await Promise.all((props.nationList ?? []).map(async (item) => ({
    ...item,
    icon: await image(`/assets/card/${item.name}.png`, fallback),
  })));

  const info = box({ position: 'absolute', top: 20, left: 20, flexDirection: 'column', color: '#fff', width: 330 },
    box({ height: 24, width: 132, backgroundColor: '#0098dc', color: '#111', alignItems: 'center' }, label('入职日'), label(props.registeredOn ?? props.regTime, { marginLeft: 4, backgroundColor: '#fff', paddingLeft: 3, paddingRight: 3 })),
    label(props.secretaryName, { fontSize: 24, marginTop: 54 }),
    label(props.secretaryEnName, { fontSize: 17 }),
    label('DATA PROVIDED BY PRTS', { fontSize: 12, marginTop: 46 }),
    box({ marginTop: 22, alignItems: 'center', backgroundColor: '#0098dc', width: 160, justifyContent: 'center' }, label('雇佣干员进度', { fontSize: 17, fontWeight: 600 })),
    label(props.charCnt, { fontSize: 55 }),
    box({ width: 320, flexWrap: 'wrap' }, nations.map((nation) => h('img', { src: nation.icon, width: 30, height: 30, style: { opacity: nation.flag === -1 ? 0.2 : 1 } }))),
  );

  const profile = box({ position: 'absolute', top: 70, right: 16, width: 638, height: 223, backgroundImage: `url(${nameCard})`, color: '#fff' },
    h('img', { src: head, width: 147, height: 149, style: { position: 'absolute', top: 30, left: 30 } }),
    h('img', { src: avatar, width: 130, height: 130, style: { position: 'absolute', top: 43, left: 48 } }),
    h('img', { src: level, width: 84, height: 84, style: { position: 'absolute', top: 5, left: 30 } }),
    box({ position: 'absolute', top: 13, left: 47, width: 50, flexDirection: 'column', alignItems: 'center' }, label(props.level, { fontSize: 20 }), label('LV', { fontSize: 14 })),
    label(props.name, { position: 'absolute', top: 50, left: 200, fontSize: 30 }),
    box({ position: 'absolute', top: 100, left: 200, gap: 6 }, label(`ID ${props.uid}`, { padding: '0 5px', backgroundColor: 'rgba(0,0,0,.2)', borderRadius: 10 }), label(props.serverName, { padding: '0 5px', backgroundColor: 'rgba(0,0,0,.2)', borderRadius: 10 })),
  );

  const assistsPanel = box({ position: 'absolute', top: 350, right: 16, width: 605, height: 200, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', alignItems: 'center', padding: '0 20px', color: '#a3a3a2' },
    box({ width: 110, flexDirection: 'column', alignItems: 'center' }, h('img', { src: assistIcon, width: 54, height: 64 }), label('助战干员', { fontSize: 17 }), label('SUPPORT UNIT', { fontSize: 13 })),
    ...assists.map((item) => box({ width: 150, height: 150, marginLeft: 8, position: 'relative', alignItems: 'center', justifyContent: 'center', backgroundColor: '#222' }, h('img', { src: item.avatar, width: 130, height: 130 }), label(`LV\n${item.level}`, { position: 'absolute', left: 15, top: 2, fontSize: 17, color: '#fff', textAlign: 'center' }))),
  );

  const modules = box({ position: 'absolute', right: 16, bottom: 18, width: 605, height: 145, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', color: '#a3a3a2', justifyContent: 'flex-end', alignItems: 'center', overflow: 'hidden' },
    h('img', { src: moduleBg, width: 500, height: 145, style: { position: 'absolute', right: 0, opacity: 0.7 } }),
    h('img', { src: moduleIcon, width: 140, height: 130, style: { position: 'absolute', left: 8, opacity: 0.25 } }),
    ...[[props.equipCnt, '总收集模组'], [props.equipStage3Cnt, 'STAGE3模组'], [props.equipOperatorCnt, '拥有模组干员']].map(([count, title]) => box({ width: 145, flexDirection: 'column', alignItems: 'center' }, label(count, { fontSize: 42, fontWeight: 600 }), label(title, { fontSize: 16 }))),
  );

  return box({ width: 1280, height: 720, position: 'relative', overflow: 'hidden', backgroundImage: `url(${background})`, fontFamily: 'NotoSansHans', backgroundColor: '#0098dc' },
    h('img', { src: secretary, height: 720, style: { position: 'absolute', left: 64, top: 0, opacity: 0.96 } }), info, profile,
    box({ position: 'absolute', top: 295, right: 16, width: 605, height: 70, borderRadius: 15, backgroundColor: 'rgba(0,0,0,.6)', color: '#a3a3a2', alignItems: 'center', padding: '0 20px' }, label(props.resume || '暂未设置签名', { fontSize: 19 })),
    assistsPanel, modules,
  );
}
