import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { Resvg } from '@resvg/resvg-js';
import satori from 'satori';

import renderBase from './base.mjs';
import renderCard from './card.mjs';
import renderEnemy from './enemy.mjs';
import renderOperator from './operator.mjs';
import { createAssetLoader } from '../lib/assets.mjs';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const font = await readFile(path.join(repoRoot, 'assets/font/NotoSansHans-Regular.ttf'));
const image = createAssetLoader({ repoRoot });

const fixtures = [
  ['card', renderCard, 1280, 720, { secretary: 'missing.png', avatar: 'missing.png', registeredOn: '2024-01-01', secretaryName: '阿米娅', secretaryEnName: 'Amiya', charCnt: 1, level: 1, name: 'Doctor', uid: '1', serverName: 'CN', resume: 'hello', equipCnt: 1, equipStage3Cnt: 1, equipOperatorCnt: 1, nationList: [], assistChars: [] }],
  ['base', renderBase, 1110, 612, { labor: { current: 1, total: 2 }, control: { level: 1, chars: [] }, dormitories: [], tradings: [], manufactures: [], powers: [], meeting: { level: 1, chars: [] }, hire: { level: 1, chars: [] }, training: { level: 1, chars: [] } }],
  ['operator', renderOperator, 1200, 800, { painting: 'missing.png', op: { name: '阿米娅', hp: '1', atk: '1', def: '1', res: '1', interval: '1', reDeploy: '1', cost: '1', block: '1', logo: '罗德岛', profession: 'CASTER', tags: '输出', nameEn: 'Amiya' }, professionBranch: { name: '术师', desc: '法术伤害' }, talents: [], buildingSkills: [], skills: [] }],
  ['enemy', renderEnemy, 656, 318, { name: '源石虫', pic: 'missing.png', desc: '描述', enemyRace: '感染生物', enemyLevel: '普通', attackType: '近战', motion: '地面', ability: '无', levels: [] }],
];

function assertFlexDivs(vnode) {
  if (vnode == null || typeof vnode !== 'object') return;
  const children = vnode.props?.children ?? [];
  if (vnode.type === 'div' && children.length > 1) {
    assert.equal(vnode.props.style?.display, 'flex', 'multi-child div must explicitly display:flex');
  }
  for (const child of children) assertFlexDivs(child);
}

for (const [name, render, width, height, props] of fixtures) {
  test(`${name} real Satori smoke rasterizes a PNG`, async () => {
    const vnode = await render(props, { image });
    assertFlexDivs(vnode);
    const svg = await satori(vnode, { width, height, fonts: [{ name: 'NotoSansHans', data: font, weight: 400, style: 'normal' }] });
    const png = new Resvg(svg, { fitTo: { mode: 'width', value: width } }).render().asPng();
    assert.ok(svg.startsWith('<svg'));
    assert.equal(png.subarray(1, 4).toString(), 'PNG');
    assert.ok(png.length > 1024);
  });
}
