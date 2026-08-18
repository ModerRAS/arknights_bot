import assert from 'node:assert/strict';
import { test } from 'node:test';

import box from './box.mjs';
import boxDetail from './box-detail.mjs';
import boxSummary from './box-summary.mjs';
import depot from './depot.mjs';
import headhunt from './headhunt.mjs';
import help from './help.mjs';
import missing from './missing.mjs';
import recruit from './recruit.mjs';

const image = async () => 'data:image/png;base64,AA==';
const char = { skinId: 'char_002_amiya#1', name: '阿米娅', level: 90, evolvePhase: 2, potentialRank: 5, rarity: 5, profession: 'WARRIOR' };
const missingChar = { skinId: 'https://media.prts.wiki/a/a0/half.png', name: '阿米娅', rarity: 5, profession: 'WARRIOR' };
const operator = { name: '阿米娅', avatar: 'https://media.prts.wiki/3/36/avatar.png', thumbURL: 'https://media.prts.wiki/a/a0/half.png', rarity: 5, profession: 'WARRIOR' };

const pages = [
  ['help', help, { PrivateCmds: [{ Cmd: '/bind', Desc: '绑定角色', Param: '', IsBind: false }], PublicCmds: [{ Cmd: '/help', Desc: '使用说明', Param: '', IsBind: false }], AdminCmds: [{ Cmd: '/news', Desc: '动态推送', Param: '', IsBind: false }] }],
  ['box', box, { name: '测试博士', chars: Array.from({ length: 11 }, () => char) }],
  ['box-detail', boxDetail, [{ ...char, id: 'char_002_amiya#1', skills: [{ id: 'skcom_magic_rage[3]', level: 10 }], equips: [{ id: 'original', level: 1 }] }]],
  ['missing', missing, { name: '测试博士', chars: Array.from({ length: 11 }, () => missingChar) }],
  ['box-summary', boxSummary, { name: '测试博士', allCharCnt: '1/1', star6CharCnt: '1/1', star5CharCnt: '0/1', star4CharCnt: '0/1', missingChars: Array.from({ length: 24 }, () => missingChar) }],
  ['recruit', recruit, [{ tags: ['高级资深干员', '输出'], operators: Array.from({ length: 12 }, () => operator) }, { tags: ['术师干员'], operators: [operator, operator] }]],
  ['headhunt', headhunt, Array.from({ length: 10 }, () => operator)],
  ['depot', depot, Array.from({ length: 11 }, () => ({ name: '龙门币', count: '100000', icon: 'https://media.prts.wiki/thumb/lmd.png', sortId: 1 }))],
];

for (const [name, render, props] of pages) {
  test(`${name} renders an explicit flex VNode with injected assets`, async () => {
    const vnode = await render(props, { image });
    assert.equal(vnode.type, 'div');
    assert.equal(vnode.props.style.display, 'flex');
    assert.ok(vnode.props.children.length > 0);
  });
}
