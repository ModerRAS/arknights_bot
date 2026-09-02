import assert from 'node:assert/strict';
import { test } from 'node:test';
import { h } from './lib/h.mjs';
import { contractPage, mockImage, renderPage, runContractSmoke } from './runner.mjs';

test('h returns a flattened Satori-compatible VNode', () => {
  const vnode = h('div', { id: 'root' }, ['one', [null, false, h('span', null, 'two')]], 'three');
  assert.deepEqual(vnode, {
    type: 'div',
    props: {
      id: 'root',
      children: ['one', { type: 'span', props: { children: ['two'] } }, 'three'],
    },
  });
});

test('page render receives async image context and reuses its string shape', async () => {
  const calls = [];
  const resolved = 'data:image/png;base64,fixture';
  const image = async (src, fallback) => {
    calls.push([src, fallback]);
    return resolved;
  };
  const vnode = await renderPage(
    contractPage,
    { title: 'test', avatar: 'a.png', avatarFallback: 'fallback.png' },
    image,
  );
  const avatar = vnode.props.children[1];
  assert.equal(vnode.type, 'div');
  assert.equal(avatar.type, 'img');
  assert.equal(avatar.props.src, resolved);
  assert.equal(vnode.props.style.backgroundImage, `url(${resolved})`);
  assert.deepEqual(calls, [['a.png', 'fallback.png']]);
});

test('mock image resolves a data URI and accepts fallback', async () => {
  const resolved = await mockImage(undefined, 'fallback.png');
  assert.match(resolved, /^data:image\/png;base64,/);
  assert.equal(resolved, `data:image/png;base64,${Buffer.from('fallback.png').toString('base64')}`);
});

test('runner contract smoke completes', async () => {
  const vnode = await runContractSmoke();
  const avatar = vnode.props.children[1];
  assert.equal(vnode.props.children[0].props.children[0], 'contract');
  assert.match(avatar.props.src, /^data:image\//);
  assert.equal(vnode.props.style.backgroundImage, `url(${avatar.props.src})`);
});
