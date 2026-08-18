import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { test } from 'node:test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('..', import.meta.url));
const runner = path.join(root, 'renderer', 'runner.mjs');

function pngDimensions(dataBase64) {
  const png = Buffer.from(dataBase64, 'base64');
  assert.deepEqual([...png.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
  return { width: png.readUInt32BE(16), height: png.readUInt32BE(20), bytes: png.byteLength };
}

const calendarProps = {
  weeks: [],
  weekdays: [],
  date: '2026-08-18',
  weekday: 'Tue',
  resource: '',
  chip: '',
};

test('NDJSON renders Satori SVG to exact PNG envelope and scale', () => {
  const requests = [
    { id: 'cold', component: 'calendar', width: 1920, height: 1080, scale: 1, props: calendarProps },
    { id: 'hot', component: 'calendar', width: 1920, height: 1080, scale: 1.5, props: calendarProps },
    { id: 'bad-scale', component: 'calendar', width: 100, height: 100, scale: 9, props: {} },
    { id: 'unknown', component: 'not-registered', width: 100, height: 100, scale: 1, props: {} },
  ];
  const result = spawnSync(process.execPath, [runner, '--ndjson'], {
    cwd: root,
    input: `${requests.map((request) => JSON.stringify(request)).join('\n')}\n`,
    encoding: 'utf8',
    maxBuffer: 8 * 1024 * 1024,
  });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stderr.trim(), '');
  const responses = result.stdout.trim().split('\n').map((line) => JSON.parse(line));
  assert.equal(responses.length, requests.length);
  const byId = new Map(responses.map((response) => [response.id, response]));
  assert.equal(byId.size, requests.length);
  const cold = byId.get('cold');
  const hot = byId.get('hot');
  const badScale = byId.get('bad-scale');
  const unknown = byId.get('unknown');
  assert.ok(cold && hot && badScale && unknown);
  assert.deepEqual(Object.keys(cold).sort(), ['dataBase64', 'height', 'id', 'mime', 'ok', 'width'].sort());
  assert.equal(cold.ok, true);
  assert.equal(cold.mime, 'image/png');
  const coldPng = pngDimensions(cold.dataBase64);
  assert.equal(cold.width, 1920);
  assert.equal(cold.height, 1080);
  assert.equal(coldPng.width, 1920);
  assert.equal(coldPng.height, 1080);
  assert.ok(coldPng.bytes > 0);
  const hotPng = pngDimensions(hot.dataBase64);
  assert.equal(hot.width, 2880);
  assert.equal(hot.height, 1620);
  assert.equal(hotPng.width, 2880);
  assert.equal(hotPng.height, 1620);
  assert.deepEqual(badScale, {
    id: 'bad-scale',
    ok: false,
    error: { code: 'invalid_request', message: 'scale exceeds limits', retryable: false },
  });
  assert.deepEqual(unknown, {
    id: 'unknown',
    ok: false,
    error: { code: 'invalid_request', message: 'component is not registered', retryable: false },
  });
});

test('NDJSON bad JSON and oversized requests stay on stdout envelope', () => {
  const oversized = `{"id":"oversized","component":"calendar","width":1,"height":1,"scale":1,"props":{"x":"${'x'.repeat(1 << 20)}"}}`;
  const result = spawnSync(process.execPath, [runner, '--ndjson'], {
    cwd: root,
    input: `{bad json\n${oversized}\n`,
    encoding: 'utf8',
    maxBuffer: 2 * 1024 * 1024,
  });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stderr.trim(), '');
  const responses = result.stdout.trim().split('\n').map((line) => JSON.parse(line));
  assert.deepEqual(responses[0], {
    id: '',
    ok: false,
    error: { code: 'invalid_request', message: 'request is not valid JSON', retryable: false },
  });
  assert.equal(responses[1].ok, false);
  assert.equal(responses[1].error.code, 'invalid_request');
});

test('no flag remains a finite contract smoke command', () => {
  const result = spawnSync(process.execPath, [runner], { cwd: root, encoding: 'utf8' });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout.trim(), 'renderer contract smoke: ok');
  assert.equal(result.stderr.trim(), '');
});
