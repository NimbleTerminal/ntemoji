import assert from 'node:assert/strict';
import { test } from 'node:test';
import { emojiFontCandidates, loadEmojiFace } from '../site/font-demo.js';

const faces = [
  { file: 'NotoColorEmoji.ttf', tech: 'color-CBDT', weight: 400, style: 'normal', bytes: 10673480 },
  { file: 'NotoColorEmoji-sbix.ttf', tech: 'color-sbix', weight: 400, style: 'normal', bytes: 10845476 },
];
function css(...technologies) {
  return { supports: query => query === 'font-tech(variations)' || technologies.some(tech => query === `font-tech(${tech})`) };
}
function browser(CSS, reject = []) {
  const attempts = [], registered = [];
  return { CSS, attempts, registered,
    FontFace: class {
      constructor(family, src, options) { Object.assign(this, { family, src, options }); }
      async load() {
        attempts.push(this);
        if (reject.some(file => this.src.includes(file))) throw Error('font load failed');
        return this;
      }
    },
    document: { fonts: { add: font => registered.push(font) } },
  };
}

test('Safari selects sbix without attempting the CBDT download', async () => {
  const env = browser(css('color-sbix'));
  assert.equal(env.attempts.length, 0);
  assert.equal((await loadEmojiFace(faces, env)).tech, 'color-sbix');
  assert.equal(env.attempts.length, 1);
  assert.match(env.attempts[0].src, /NotoColorEmoji-sbix\.ttf/);
  assert.equal(env.registered[0].options.unicodeRange, 'U+0021-10FFFF');
});
test('browsers supporting both formats prefer the unmodified CBDT font', async () => {
  const env = browser(css('color-CBDT', 'color-sbix'));
  assert.equal((await loadEmojiFace(faces, env)).tech, 'color-CBDT');
  assert.equal(env.attempts.length, 1);
});
test('an unsupported browser downloads neither font', async () => {
  const env = browser(css());
  assert.deepEqual(emojiFontCandidates(faces, env.CSS), []);
  assert.equal(await loadEmojiFace(faces, env), null);
  assert.equal(env.attempts.length, 0);
});
test('older browsers fall through from a rejected CBDT font to sbix', async () => {
  for (const CSS of [undefined, {}, { supports: () => false }]) {
    const env = browser(CSS, ['NotoColorEmoji.ttf']);
    assert.equal((await loadEmojiFace(faces, env)).tech, 'color-sbix');
    assert.equal(env.attempts.length, 2);
    assert.equal(env.registered.length, 1);
  }
});
test('a supported format failing to load does not block the next candidate', async () => {
  const env = browser(css('color-CBDT', 'color-sbix'), ['NotoColorEmoji.ttf']);
  assert.equal((await loadEmojiFace(faces, env)).tech, 'color-sbix');
  assert.equal(env.registered.length, 1);
});
test('all failed loads report failure and register no font', async () => {
  const env = browser(css('color-CBDT', 'color-sbix'), faces.map(face => face.file));
  await assert.rejects(loadEmojiFace(faces, env), AggregateError);
  assert.equal(env.registered.length, 0);
});
