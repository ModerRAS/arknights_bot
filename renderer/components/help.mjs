import { h } from '../lib/h.mjs';

const fallback = 'assets/common/amiya.png';

// geometry measured off the frozen Playwright baseline (CSS px @1.5 scale):
// banner.png 660x239 @ y10; h1 32px bold ink @ (20,148..178); hint p @ (20,204);
// groups: label bar 660x40 (+10 above), bar text 16px @ left 25 line-top 8,
// cards 152x53.33 border-box grid (margins 15/10), 4 per row, row pitch 69.33
const BIND_SVG = 'data:image/svg+xml;base64,' + Buffer.from(
  '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="#fff" viewBox="0 0 16 16"><path d="M11 6a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/><path fill-rule="evenodd" d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm8-7a7 7 0 0 0-5.468 11.37C3.242 11.226 4.805 10 8 10s4.757 1.225 5.468 2.37A7 7 0 0 0 8 1z"/></svg>').toString('base64');

const card = (cmd) => h('div', { style: { width: 152, height: 53.33, border: '1px solid #fff', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, position: 'relative', display: 'flex', flexShrink: 0, marginLeft: 10, marginTop: 15.7, WebkitTextStrokeWidth: 0.5, WebkitTextStrokeColor: '#fff' } },
  h('div', { style: { position: 'absolute', left: 6, top: 3.5, fontSize: 15, display: 'flex', WebkitTextStrokeWidth: 0.5, WebkitTextStrokeColor: '#fff' } }, `${cmd.Cmd} ${cmd.Param}`.trim()),
  cmd.IsBind ? h('img', { src: BIND_SVG, width: 16, height: 16, style: { position: 'absolute', right: 10, top: 4 } }) : null,
  h('div', { style: { position: 'absolute', left: 6, top: 33, fontSize: 15, display: 'flex', WebkitTextStrokeWidth: 0.5, WebkitTextStrokeColor: '#fff' } }, cmd.Desc));

const group = (label, title, cmds) => h('div', { style: { display: 'flex', flexDirection: 'column', marginTop: 10 } },
  h('div', { style: { height: 40, display: 'flex', flexDirection: 'column', position: 'relative' } },
    h('img', { src: label, width: 660, height: 40 }),
    h('div', { style: { position: 'absolute', left: 25, top: 12.7, fontSize: 16, display: 'flex' } }, title)),
  h('div', { style: { display: 'flex', flexWrap: 'wrap', width: 660 } }, cmds.map(card)));

export default async function render(props, { image }) {
  const [banner, label, bg, amiya] = await Promise.all([
    image('assets/help/banner.png', fallback),
    image('assets/help/label.png', fallback),
    image('assets/help/bg.jpg', fallback),
    image('assets/help/amiya.png', fallback),
  ]);
  return h('div', { style: { width: 660, height: 1366, display: 'flex', flexDirection: 'column', fontFamily: 'NotoSansHans', fontSize: 16, color: '#fff', position: 'relative', backgroundImage: `url(${bg})`, backgroundSize: 'cover' } },
    h('img', { src: banner, width: 660, height: 239, style: { position: 'absolute', left: 0, top: 10 } }),
    h('div', { style: { position: 'absolute', left: 20, top: 147, fontSize: 32, fontWeight: 700, display: 'flex', WebkitTextStrokeWidth: 1.0, WebkitTextStrokeColor: '#fff' } }, '使用说明'),
    h('div', { style: { position: 'absolute', left: 20, top: 208, display: 'flex', alignItems: 'center', gap: 3 } },
      h('img', { src: BIND_SVG, width: 16, height: 16 }), '为需要绑定角色的指令'),
    h('div', { style: { position: 'absolute', left: 0, top: 249, width: 660, height: 1366 - 249, display: 'flex', flexDirection: 'column', backgroundImage: `linear-gradient(rgba(0,0,0,0.8), rgba(0,0,0,0.8)), url(${amiya})`, backgroundSize: '100% 100%, 660px auto', backgroundPosition: 'center top', backgroundRepeat: 'no-repeat' } },
      group(label, '私聊指令', props.PrivateCmds ?? []),
      group(label, '普通指令', props.PublicCmds ?? []),
      group(label, '管理员指令', props.AdminCmds ?? [])));
}
