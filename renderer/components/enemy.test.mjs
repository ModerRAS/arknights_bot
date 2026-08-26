import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { Resvg } from '@resvg/resvg-js';
import satori from 'satori';

import renderEnemy from './enemy.mjs';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const font = await readFile(path.join(repoRoot, 'assets/font/NotoSansHans-Regular.ttf'));

const image = async (src, fallback) => String(src ?? fallback);
const props = {
  name: '源石虫',
  pic: 'enemy.png',
  desc: '拥有较高防御力的感染生物。',
  enemyRace: '感染生物',
  enemyLevel: '普通',
  attackType: '近战',
  motion: '地面',
  ability: '免疫沉默',
};
const richLevel = {
  desc: '阶段描述',
  attackType: '近战',
  motion: '地面',
  hpRecovery: '0',
  hp: '100',
  atk: '20',
  def: '10',
  res: '0',
  ATKRadius: '1',
  weight: '1',
  moveSpeed: '1',
  interval: '2',
  damageRes: '0',
  elementRes: '0',
  ridicule: '0',
  point: '1',
  abnormal: '无',
  skills: [{ name: '冲撞', spInit: '0', spCost: '1', desc: '造成伤害' }],
  talent: '免疫沉默',
};

function widths(row) {
  return row.props.children.map((child) => child.props.style.width);
}

test('Enemy reconstructs the browser-repaired 8-track base table', async () => {
  const vnode = await renderEnemy({ ...props, level: [] }, { image });
  assert.equal(vnode.props.children[0].props.children[0], props.ability);
  const base = vnode.props.children[1];
  assert.equal(base.props.children.length, 5);
  assert.deepEqual(widths(base.props.children[2]), [283, 124, 124, 125]);
  assert.deepEqual(widths(base.props.children[3]), [283, 124, 124, 125]);
  assert.equal(base.props.children[0].props.children[0].props.children[0], props.name);
  assert.equal(base.props.children[4].props.children[0].props.children[0], '能力');
});

test('Enemy uses canonical level and explicit 6-track level rows', async () => {
  const vnode = await renderEnemy({ ...props, level: [richLevel] }, { image });
  assert.equal(vnode.props.children.length, 5);
  const level = vnode.props.children[2];
  assert.equal(level.props.children.length, 7);
  assert.deepEqual(widths(level.props.children[3]), [109, 109, 109, 109, 110, 110]);
  assert.deepEqual(widths(level.props.children[4].props.children[0]), [109, 109, 109, 109, 110, 110]);
  assert.deepEqual(widths(level.props.children[4].props.children[1]), [109, 109, 109, 109, 110, 110]);
  assert.deepEqual(widths(level.props.children[5].props.children[0]), [109, 109, 109, 109, 110, 110]);
  assert.deepEqual(widths(level.props.children[5].props.children[1]), [109, 109, 109, 109, 110, 110]);
  assert.deepEqual(widths(level.props.children[6]), [109, 547]);
});

test('Enemy auto-height is a one-pass natural-root contract', async () => {
  const fixed = await renderEnemy({ ...props, level: [] }, { image });
  const auto = await renderEnemy({ ...props, level: [], autoHeight: true }, { image });
  assert.equal(fixed.props.style.height, 318);
  assert.equal(fixed.props.style.overflow, 'hidden');
  assert.equal(auto.props.style.minHeight, 318);
  assert.equal(auto.props.style.height, undefined);
  assert.equal(auto.props.style.overflow, undefined);
});

test('Enemy rich auto-height VNode rasterizes without clipping or Resvg errors', async () => {
  const picture = `data:image/png;base64,${(await readFile(path.join(repoRoot, 'src/utils/media/testdata/visual/baseline/cache/enemy-originium-slug-158.png'))).toString('base64')}`;
  const vnode = await renderEnemy({ ...props, pic: picture, level: [richLevel], autoHeight: true }, { image: async (src) => src });
  const svg = await satori(vnode, {
    width: 656,
    height: 1024,
    fonts: [{ name: 'NotoSansHans', data: font, weight: 400, style: 'normal' }],
  });
  const png = new Resvg(svg, { fitTo: { mode: 'width', value: 984 } }).render().asPng();
  assert.match(svg, /<svg width="656" height="1024"/);
  assert.ok(png.length > 1024);
});
