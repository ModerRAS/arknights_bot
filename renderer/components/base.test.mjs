import assert from 'node:assert/strict';
import { test } from 'node:test';

import renderBase from './base.mjs';

test('Base places character morale and progress in the room card', async () => {
  const vnode = await renderBase({
    control: { level: 1, chars: [{ name: 'A', avatar: 'a', AP: 0 }] },
  }, { image: async (src, fallback) => String(src ?? fallback) });
  const room = vnode.props.children[1].props.children[0];
  const card = room.props.children[0].props.children[1].props.children[0];
  const progress = card.props.children.at(-1);

  assert.equal(card.props.children[1].type, 'svg');
  assert.equal(progress.props.style.position, 'absolute');
  assert.equal(progress.props.style.left, 0);
  assert.equal(progress.props.style.bottom, 0);
  assert.equal(progress.props.children[0].props.style.width, '0%');
});

test('Base renders office refresh count with icon and defaults to 3 and preserves title geometry', async () => {
  const findOffice = (vnode) => {
    const roomsBox = vnode.props.children[1];
    const cards = roomsBox.props.children;
    return cards.find((card) => {
      const inner = card.props.children[0];
      const titleRow = inner.props.children[0];
      const left = titleRow.props.children[0];
      const val = Array.isArray(left.props.children) ? left.props.children.join('') : left.props.children;
      return String(val).includes('办公室');
    });
  };
  const check = async (hireProps, expectedText) => {
    const vnode = await renderBase({ hire: hireProps }, { image: async (src, fallback) => String(src ?? fallback) });
    const officeCard = findOffice(vnode);
    assert.ok(officeCard, 'office card not found');
    const titleRow = officeCard.props.children[0].props.children[0];
    // geometry
    assert.equal(titleRow.props.style.height, 49);
    assert.equal(titleRow.props.style.marginLeft, 10);
    assert.equal(titleRow.props.style.marginRight, 20);
    assert.equal(titleRow.props.style.alignItems, 'center');
    assert.equal(titleRow.props.style.justifyContent, 'space-between');
    const left = titleRow.props.children[0];
    const leftText = Array.isArray(left.props.children) ? left.props.children.join('') : left.props.children;
    assert.match(String(leftText), /办公室 Lv\.\d+/);
    assert.equal(left.props.style.fontSize, 19);
    assert.equal(left.props.style.fontWeight, 600);
    const right = titleRow.props.children[1];
    assert.equal(right.type, 'div');
    assert.equal(right.props.style.alignItems, 'center');
    const svg = right.props.children[0];
    assert.equal(svg.type, 'svg');
    assert.equal(svg.props.width, 20);
    assert.equal(svg.props.height, 20);
    assert.equal(svg.props.viewBox, '0 0 48 48');
    assert.equal(svg.props.fill, 'none');
    const paths = svg.props.children;
    assert.equal(paths.length, 3);
    assert.equal(paths[0].props.d, 'M4 34H44');
    assert.equal(paths[1].props.d, 'M42 39L21 5');
    assert.equal(paths[2].props.d, 'M6 39L27 5');
    for (const p of paths) {
      assert.equal(p.props.stroke, '#ffffff');
      assert.equal(p.props.strokeWidth, 4);
      assert.equal(p.props.strokeLinecap, 'round');
      assert.equal(p.props.strokeLinejoin, 'round');
    }
    const span = right.props.children[1];
    assert.equal(span.type, 'span');
    const txt = Array.isArray(span.props.children) ? span.props.children.join('') : span.props.children;
    assert.equal(String(txt), expectedText);
  };
  await check({ level: 5, refreshCount: 7 }, '刷新次数7');
  await check({ level: 2 }, '刷新次数3');
  await check({ level: 2, refreshCount: 0 }, '刷新次数0');
  // props.hire source
  const vnode = await renderBase({ hire: { level: 1, refreshCount: 2 } }, { image: async (src, fallback) => String(src ?? fallback) });
  const officeCard = findOffice(vnode);
  const right = officeCard.props.children[0].props.children[0].props.children[1];
  const txt = Array.isArray(right.props.children[1].props.children) ? right.props.children[1].props.children.join('') : right.props.children[1].props.children;
  assert.equal(String(txt), '刷新次数2');
});
