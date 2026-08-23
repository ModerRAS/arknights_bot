import assert from 'node:assert/strict';
import { test } from 'node:test';

import renderGacha from './gacha.mjs';

const image = async (src, fallback) => String(src ?? fallback);

test('Gacha averages box left 340', async () => {
  const vnode = await renderGacha({
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
    chars: [{ name: '阿米娅', avatar: 'a', date: '2025-01-01', pool: 'test', isNew: true }],
    star6Info: [{ name: '能天使', avatar: 'b', date: '2025-01-02', pool: 'test', count: 5, isNew: false }],
  }, { image });
  assert.equal(vnode.props.style.width, 1000);
  assert.equal(vnode.props.style.height, 882);
  const header = vnode.props.children[0];
  const averagesBox = header.props.children[3];
  assert.equal(averagesBox.props.style.left, 340);
  assert.equal(averagesBox.props.style.top, 150);
});

test('Gacha imageEntry padding 8px 4px', async () => {
  const vnode = await renderGacha({
    name: 'test', total: 1, period: '2025-01-01 00:00:00——2025-01-02 00:00:00', today: '2025年01月03日',
    rarities: [{ label: '6星', count: 1, percent: 100, color: 'rgba(244,110,30,1)' }],
    averages: [{ label: '6星', count: 1, avg: 1 }], pools: [],
    chars: [{ name: '阿米娅', avatar: 'a', date: '2025-01-01', pool: 'test', isNew: false }],
    star6Info: [],
  }, { image });
  const article = vnode.props.children[1];
  const listBox = article.props.children[0];
  const entry = listBox.props.children[1];
  assert.equal(entry.props.style.padding, '8px 4px');
  assert.equal(entry.props.style.width, 230);
  assert.equal(entry.props.style.height, 150);
});

test('Gacha pie legend marginTop 13', async () => {
  const vnode = await renderGacha({
    name: 'test', total: 1, period: '2025-01-01 00:00:00——2025-01-02 00:00:00', today: '2025年01月03日',
    rarities: [{ label: '6星', count: 1, percent: 100, color: 'rgba(244,110,30,1)' }],
    averages: [{ label: '6星', count: 1, avg: 1 }], pools: [],
    chars: [], star6Info: [],
  }, { image });
  const header = vnode.props.children[0];
  const pieBox = header.props.children[2];
  const legendContainer = pieBox.props.children[1];
  const legend = legendContainer.props.children[0];
  assert.equal(legend.props.style.marginTop, 13);
  assert.equal(legend.props.style.transform, undefined);
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
