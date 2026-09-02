import assert from 'node:assert/strict';
import { test } from 'node:test';

import renderGacha from './gacha.mjs';

// Valid 1x1 PNG data URI: the component prescales backgrounds through sharp.
const PNG_URI = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+ip1sAAAAASUVORK5CYII=';
const image = async () => PNG_URI;

const baseProps = {
  name: 'test',
  total: 15,
  period: '2025-01-01 00:00:00——2025-01-02 00:00:00',
  today: '2025年01月03日',
  rarities: [
    { label: '6星', count: 1, percent: 6.66, color: 'rgba(244,110,30,1)' },
    { label: '5星', count: 2, percent: 13.33, color: 'rgba(247,171,55,1)' },
    { label: '4星', count: 4, percent: 26.66, color: 'rgba(161,53,246,1)' },
    { label: '3星', count: 8, percent: 53.33, color: 'rgba(109,116,126,1)' },
  ],
  averages: [
    { label: '6星', count: 1, avg: 15 },
    { label: '5星', count: 2, avg: 7.5 },
    { label: '4星', count: 4, avg: 3.75 },
    { label: '3星', count: 8, avg: 1.87 },
  ],
  pools: [{ poolName: 'test', count: 5 }],
  chars: [
    { name: '阿米娅', avatar: 'a', date: '2025-01-01', pool: 'test', isNew: true },
    { name: '能天使', avatar: 'b', date: '2025-01-02', pool: 'test', isNew: false },
  ],
  star6Info: [
    { name: '阿米娅', avatar: 'a', date: '2025-01-01', pool: 'test', count: 5, isNew: true },
    { name: '能天使', avatar: 'b', date: '2025-01-02', pool: 'test', count: 5, isNew: false },
  ],
};

// Geometry pinned to the frozen Playwright baseline (pixel-measured at DPR 1.5).
test('Gacha averages card at measured left 346 with pinned rows', async () => {
  const vnode = await renderGacha(baseProps, { image });
  assert.equal(vnode.props.style.width, 1000);
  assert.equal(vnode.props.style.height, 882);
  // root: [header bg, name, total, card1, card2, card3, listBox1, listBox2, footer]
  const averagesCard = vnode.props.children[4];
  assert.equal(averagesCard.type, 'div');
  assert.equal(averagesCard.props.style.left, 346);
  assert.equal(averagesCard.props.style.top, 150);
  assert.equal(averagesCard.props.style.width, 302);
  // first averages row (after the title div) at the measured column
  const firstRow = averagesCard.props.children[1];
  assert.equal(firstRow.props.style.left, 34.4);
});

test('Gacha list entries: second entry offset 229.7, avatar pinned', async () => {
  const vnode = await renderGacha(baseProps, { image });
  const listBox1 = vnode.props.children[6];
  assert.equal(listBox1.props.style.left, 20);
  assert.equal(listBox1.props.style.top, 441);
  const entry1 = listBox1.props.children[1];
  const entry2 = listBox1.props.children[2];
  assert.equal(entry1.props.style.left, 0);
  assert.equal(entry2.props.style.left, 229.7);
  const avatar = entry1.props.children[0];
  assert.equal(avatar.props.style.left, 3.3);
  assert.equal(avatar.props.style.top, 8.7);
  assert.equal(avatar.props.width, 100);
});

test('Gacha legend: icon squares + text at measured positions', async () => {
  const vnode = await renderGacha({ ...baseProps, chars: [], star6Info: [] }, { image });
  const pieCard = vnode.props.children[3];
  // card children: [title, ...legend pairs (icon, text)..., svg]
  const icon = pieCard.props.children[1];
  const text = pieCard.props.children[2];
  assert.equal(icon.props.style.left, 8.3);
  assert.equal(icon.props.style.top, 93.1);
  assert.equal(icon.props.style.width, 8);
  assert.equal(icon.props.style.height, 8);
  assert.equal(text.props.style.left, 21);
  assert.match(text.props.children[0], /6星/);
});

test('Gacha rasterizes without error', async () => {
  const { readFile } = await import('node:fs/promises');
  const path = await import('node:path');
  const { fileURLToPath } = await import('node:url');
  const satori = (await import('satori')).default;
  const { Resvg } = await import('@resvg/resvg-js');
  const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
  const font = await readFile(path.join(repoRoot, 'assets/font/NotoSansHans-Regular.ttf'));
  const dataUri = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+ip1sAAAAASUVORK5CYII=';
  const imageDataUri = async () => dataUri;
  const vnode = await renderGacha({
    name: 'test',
    total: 1,
    period: '2025-01-01 00:00:00——2025-01-02 00:00:00',
    today: '2025年01月03日',
    rarities: [{ label: '6星', count: 1, percent: 100, color: 'rgba(244,110,30,1)' }],
    averages: [{ label: '6星', count: 1, avg: 1 }],
    pools: [],
    chars: [],
    star6Info: [],
  }, { image: imageDataUri });
  const svg = await satori(vnode, { width: 1000, height: 882, fonts: [{ name: 'NotoSansHans', data: font, weight: 400, style: 'normal' }] });
  assert.match(svg, /<svg width="1000" height="882"/);
  const png = new Resvg(svg, { fitTo: { mode: 'width', value: 1000 } }).render().asPng();
  assert.ok(png.length > 1024);
});
