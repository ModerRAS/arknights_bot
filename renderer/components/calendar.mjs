import { h } from '../lib/h.mjs';

const style = (extra = {}) => ({ display: 'flex', boxSizing: 'border-box', ...extra });

export default async function render(props, { image }) {
  const background = await image('/assets/calendar/bg.png');
  const rows = props.weeks?.length || 5;
  const cellHeight = (1040 - 40) / rows;
  return h('div', { style: style({ width: 1920, height: 1080, fontFamily: 'NotoSansHans', backgroundColor: '#fff', color: '#000' }) },
    h('div', { style: style({ width: 192, height: '100%', padding: '15px 20px', flexDirection: 'column', color: '#fff', backgroundImage: `url(${background})`, backgroundColor: '#141516', backgroundSize: 'contain', backgroundRepeat: 'no-repeat' }) },
      h('div', { style: style({ width: '100%', justifyContent: 'center', flexDirection: 'column', alignItems: 'center', fontSize: 23 }) }, props.date, h('span', { style: { marginTop: 8 } }, props.weekday)),
      h('div', { style: style({ marginTop: 350, flexDirection: 'column', gap: 20, fontSize: 18 }) },
        h('div', { style: style({ flexDirection: 'column', gap: 12 }) }, h('b', null, '资源关卡开放'), h('span', null, props.resource)),
        h('div', { style: style({ flexDirection: 'column', gap: 12 }) }, h('b', null, '芯片关卡开放'), h('span', null, props.chip)),
      ),
    ),
    h('div', { style: style({ width: 1728, height: '100%', padding: '0 15px', flexDirection: 'column' }) },
      h('div', { style: style({ height: 40, width: '100%' }) }, (props.weekdays || []).map((day, index) => h('div', { style: style({ width: '14.285714%', justifyContent: 'center', alignItems: 'center', color: index >= 5 ? '#e02d2d' : '#2c9bb3', fontSize: 23 }) }, day))),
      h('div', { style: style({ width: '100%', height: 1000, flexDirection: 'column' }) }, (props.weeks || []).map((week) => h('div', { style: style({ width: '100%', height: cellHeight, borderTop: '1px solid #c8cacc' }) }, week.map((day) => h('div', { style: style({ width: '14.285714%', height: '100%', padding: '18px 10px', flexDirection: 'column', alignItems: 'center', justifyContent: 'space-between', backgroundColor: day.today ? 'cornflowerblue' : '#fff', color: day.today ? '#fff' : (!day.inCurrentMonth ? '#bfbfbf' : (day.weekend ? '#e02d2d' : '#000')) }) },        h('span', { style: { fontSize: 23 } }, String(day.day).padStart(2, '0')),
        h('div', { style: style({ width: '100%', flexDirection: 'column', color: day.today ? '#fff' : '#616161', fontSize: 14, overflow: 'hidden' }) }, ...(day.events || []).map((event) => h('span', null, event))),
      ))))),
    ),
  );
}
