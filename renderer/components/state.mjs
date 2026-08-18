import { h } from '../lib/h.mjs';

const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });

const clock = (extra = {}) => h('svg', { width: 16, height: 16, viewBox: '0 0 16 16', fill: 'currentColor', style: extra },
  h('path', { d: 'M8 3.5a.5.5 0 0 0-1 0V9a.5.5 0 0 0 .252.434l3.5 2a.5.5 0 0 0 .496-.868L8 8.71V3.5z' }),
  h('path', { d: 'M8 16A8 8 0 1 0 8 0a8 8 0 0 0 0 16zm7-8A7 7 0 1 1 1 8a7 7 0 0 1 14 0z' }),
);

const flowMeter = (label, value, top) => h('div', { style: style({ position: 'absolute', left: 460, top, width: 410, flexDirection: 'column', color: '#fff' }) },
  h('div', { style: style({ alignItems: 'center', height: 25, fontSize: 25 }) },
    h('span', null, label),
    clock({ marginLeft: 30, marginRight: 10 }),
    h('span', null, value.label),
    h('span', { style: { marginLeft: 130 } }, `${value.current}/${value.max}`),
  ),
  h('div', { style: style({ marginTop: 10, width: 410, height: 11, backgroundColor: '#999' }) },
    h('div', { style: style({ width: `${value.max ? value.current * 100 / value.max : 0}%`, height: '100%', backgroundColor: '#fff' }) }),
  ),
);

const fixedMeter = (label, value, top) => h('div', { style: style({ position: 'absolute', left: 460, top, width: 410, flexDirection: 'column', color: '#fff' }) },
  h('div', { style: style({ height: 25, justifyContent: 'space-between', alignItems: 'center', fontSize: 25 }) },
    h('span', null, label),
    h('span', null, `${value.label}     ${value.current}/${value.max}`),
  ),
  h('div', { style: style({ marginTop: 10, width: 410, height: 11, backgroundColor: '#999' }) },
    h('div', { style: style({ width: `${value.max ? value.current * 100 / value.max : 0}%`, height: '100%', backgroundColor: '#fff' }) }),
  ),
);

const row = (icon, title, value, iconLeft, iconTop, titleLeft, titleTop, valueLeft, valueTop) => [
  h('img', { src: icon, width: 42, height: 42, style: { position: 'absolute', left: iconLeft, top: iconTop } }),
  h('span', { style: { position: 'absolute', left: titleLeft, top: titleTop, fontSize: 25, color: '#fff' } }, title),
  h('span', { style: { position: 'absolute', left: valueLeft, top: valueTop, fontSize: 25, color: '#fff' } }, value),
];

export default async function render(props, { image }) {
  const [background, apIcon, campaignIcon, recruitIcon, tiredIcon, tradingIcon, manufactureIcon, avatar, training] = await Promise.all([
    image('/assets/state/bg.png'),
    image('/assets/state/ap.png'),
    image('/assets/state/campaign.png'),
    image('/assets/state/recruit.png'),
    image('/assets/state/tired_chars.png'),
    image('/assets/state/tradings.png'),
    image('/assets/state/manufactures.png'),
    image(props.avatar, '/assets/common/amiya.png'),
    props.training ? image(props.training.avatar, '/assets/common/amiya.png') : Promise.resolve(null),
  ]);

  return h('div', { style: style({ position: 'relative', width: 1092, height: 510, overflow: 'hidden', fontFamily: 'NotoSansHans', backgroundColor: '#2e3031', backgroundImage: `url(${background})` }) },
    h('img', { src: avatar, width: 54, height: 54, style: { position: 'absolute', left: 34, top: 34 } }),
    h('span', { style: { position: 'absolute', left: 98, top: 45, color: '#fff', fontSize: 30 } }, `Dr ${props.playerName}`),
    h('img', { src: apIcon, width: 75, height: 73, style: { position: 'absolute', left: 35, top: 146 } }),
    h('div', { style: style({ position: 'absolute', left: 146, top: 145, flexDirection: 'column', color: '#fff' }) },
      h('span', { style: { fontSize: 30 } }, `${props.ap.current}/${props.ap.max}`),
      h('span', { style: { marginTop: 10, fontSize: 21 } }, props.ap.label),
    ),
    fixedMeter('数据增补条', props.towerLower, 122),
    flowMeter('数据增补仪', props.towerHigher, 214),
    h('span', { style: { position: 'absolute', left: 945, top: 53, fontSize: 25, color: props.checkedIn ? '#5d9a00' : '#cd2828' } }, props.checkedIn ? '已签到' : '未签到'),
    h('img', { src: campaignIcon, width: 105, height: 108, style: { position: 'absolute', left: 925, top: 126 } }),
    h('span', { style: { position: 'absolute', left: 950, top: 249, color: '#fff', fontSize: 20 } }, `${props.reward.current}/${props.reward.max}`),
    h('span', { style: { position: 'absolute', left: 930, top: 218, color: '#fff', fontSize: 16, backgroundColor: 'rgba(0,0,0,.5)' } }, props.reward.label),
    ...row(recruitIcon, '公开招募', `${props.recruitment.current}/${props.recruitment.max}`, 49, 332, 120, 342, 330, 342),
    ...row(tiredIcon, '干员疲劳', String(props.tiredChars), 49, 430, 115, 438, 325, 438),
    ...row(tradingIcon, '订单进度', `${props.trading.current}/${props.trading.max}`, 460, 315, 520, 318, 780, 318),
    ...row(manufactureIcon, '制造进度', `${props.manufacture.current}/${props.manufacture.max}`, 460, 414, 520, 417, 780, 417),
    props.training ? h('div', { style: style({ position: 'absolute', left: 922, top: 307, width: 134, flexDirection: 'column', alignItems: 'center', color: '#fff' }) },
      h('img', { src: training, width: 130, height: 130 }),
      h('div', { style: style({ position: 'absolute', top: 105, width: 133, height: 25, backgroundColor: '#000', opacity: 0.5 }) }),
      h('div', { style: style({ position: 'absolute', top: 107, alignItems: 'center', fontSize: 16 }) }, clock({ marginRight: 4 }), h('span', null, props.training.label)),
      h('span', { style: { marginTop: 12, fontSize: 30 } }, '训练室'),
    ) : null,
  );
}
