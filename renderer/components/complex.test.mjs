import assert from 'node:assert/strict';
import { test } from 'node:test';

import renderBase from './base.mjs';
import renderCard from './card.mjs';
import renderEnemy from './enemy.mjs';
import renderOperator from './operator.mjs';

function imageRecorder() {
  const calls = [];
  return {
    calls,
    image: async (src, fallback) => {
      calls.push([src, fallback]);
      return `data:image/png;base64,${Buffer.from(String(src ?? fallback)).toString('base64')}`;
    },
  };
}

const card = {
  secretary: 'secretary.png', avatar: 'avatar.png', registeredOn: '2024-01-01', secretaryName: 'Amiya', secretaryEnName: 'Amiya',
  charCnt: 1, level: 1, name: 'Doctor', uid: '1', serverName: 'CN', resume: 'hello', equipCnt: 1, equipStage3Cnt: 1, equipOperatorCnt: 1,
  nationList: [{ name: 'rhodes', flag: 1 }], assistChars: [{ skinId: 'char_002_amiya#1', level: 1 }],
};
const base = { labor: { current: 1, total: 2 }, control: { level: 1, chars: [{ name: 'Amiya', avatar: 'char_002_amiya#1', AP: 50 }] }, dormitories: [], tradings: [], manufactures: [], powers: [], meeting: { level: 1, chars: [] }, hire: { level: 1, chars: [] }, training: { level: 1, chars: [] } };
const operator = { painting: 'painting.png', op: { name: 'Amiya', hp: '1', atk: '1', def: '1', res: '1', interval: '1', reDeploy: '1', cost: '1', block: '1', logo: 'RI', profession: 'CASTER', tags: 'tag', nameEn: 'Amiya' }, professionBranch: { name: 'Caster', desc: 'desc' }, talents: [{ evolve: 'E1', name: 'talent', desc: 'desc' }], buildingSkills: [{ icon: 'building.png', name: 'building', desc: 'desc' }], skills: [{ icon: 'skill.png', name: 'skill', desc: 'desc', spType: ['auto'], spInit: '0', spCost: '1', skillRange: '□' }] };
const enemy = { name: 'Slug', pic: 'enemy.png', desc: 'desc', enemyRace: 'race', enemyLevel: 'normal', attackType: 'melee', motion: 'ground', ability: 'none', levels: [{ desc: 'level', attackType: 'melee', motion: 'ground', hpRecovery: '0', hp: '1', atk: '1', def: '1', res: '0', ATKRadius: '1', weight: '1', moveSpeed: '1', interval: '1', damageRes: '0', elementRes: '0', ridicule: '0', point: '1', abnormal: 'none', skills: [{ name: 'hit', spInit: '0', spCost: '1', desc: 'desc' }], talent: 'talent' }] };

for (const [name, render, props] of [['card', renderCard, card], ['base', renderBase, base], ['operator', renderOperator, operator], ['enemy', renderEnemy, enemy]]) {
  test(`${name} resolves images and returns a flex root VNode`, async () => {
    const recorder = imageRecorder();
    const vnode = await render(props, { image: recorder.image });
    assert.equal(vnode.type, 'div');
    assert.equal(vnode.props.style.display, 'flex');
    assert.ok(vnode.props.children.length > 0);
    assert.ok(recorder.calls.length > 0);
    for (const [src] of recorder.calls) assert.ok(src);
  });
}
