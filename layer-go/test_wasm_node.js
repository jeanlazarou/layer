// Smoke test for the WebAssembly build, run with Node:
//   ./build-wasm.sh && node test_wasm_node.js
// Exercises the same JSON API the visual editor uses.
const fs = require('fs');
const path = require('path');
const assert = require('assert');

const PUBLIC = path.join(__dirname, '../visual-editor/frontend/public');
require(path.join(PUBLIC, 'wasm_exec.js'));

async function main() {
  const go = new Go();
  const { instance } = await WebAssembly.instantiate(
    fs.readFileSync(path.join(PUBLIC, 'layer.wasm')),
    go.importObject,
  );
  go.run(instance);

  const run = (src, vars) => JSON.parse(globalThis.executeLayer(src, vars ? JSON.stringify(vars) : undefined));
  const describe = (src) => JSON.parse(globalThis.describeLayer(src));
  const split = (src) => JSON.parse(globalThis.parseLayerFile(src));
  let failures = 0;
  const check = (name, fn) => {
    try {
      fn();
      console.log(`✓ ${name}`);
    } catch (e) {
      failures++;
      console.log(`✗ ${name}\n    ${e.message}`);
    }
  };

  const fact = `@layers [core, debug]
@layer core {
  fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)
}
@layer debug {
  fn factorial(n) =
    let _ = log("factorial({n})") in
      core.factorial(n)
}
factorial(3)`;

  check('runs a program and captures logs', () => {
    const r = run(fact);
    assert.strictEqual(r.success, true, JSON.stringify(r.errors));
    assert.strictEqual(r.result, '6');
    assert.strictEqual(r.kind, 'value');
    assert.deepStrictEqual(r.logs, ['[INFO] factorial(3)', '[INFO] factorial(2)', '[INFO] factorial(1)', '[INFO] factorial(0)']);
  });

  check('reports located runtime errors with a trace', () => {
    const r = run('@layers [core]\n@layer core {\n  fn f(x) = x + 1\n}\nf("a")');
    assert.strictEqual(r.success, false);
    assert.strictEqual(r.errors[0].kind, 'runtime');
    assert.strictEqual(r.errors[0].line, 3);
    assert.match(r.errors[0].message, /operator \+/);
    assert.strictEqual(r.errors[0].trace, 'core.f');
  });

  check('reports all static check errors at once', () => {
    const r = run('@layers [core]\n@layer core {\n  fn f(x) = g(x)\n  fn h(x) = k(x)\n}\n1');
    assert.strictEqual(r.success, false);
    assert.deepStrictEqual(r.errors.map(e => [e.kind, e.line]), [['check', 3], ['check', 4]]);
  });

  check('binds UI input values as variables (no string splicing)', () => {
    const src = '@layers [core]\n@layer core {\n  fn square(x) = x * x\n}\nsquare(num)';
    assert.strictEqual(run(src, { num: 7 }).result, '49');
    const quoted = run('@layers [core]\ngreeting', { greeting: 'say "hi" {there}' });
    assert.strictEqual(quoted.result, 'say "hi" {there}');
  });

  check('marks UI results', () => {
    const r = run('@layers [core, ui]\nview({title: "a < b", children: [text("x & y")]})');
    assert.strictEqual(r.kind, 'ui');
    assert.match(r.result, /title="a &lt; b"/);
    assert.match(r.result, /x &amp; y/);
  });

  check('deep recursion works despite the small browser call stack', () => {
    // Without goroutine stack hopping this crashed the Go runtime at ~2000.
    const r = run('@layers [core]\n@layer core {\n  fn count(n) = if n == 0 then 0 else 1 + count(n - 1)\n}\ncount(15000)');
    assert.strictEqual(r.success, true, JSON.stringify(r.errors));
    assert.strictEqual(r.result, '15000');
  });

  check('runaway recursion is an error, and the runtime survives it', () => {
    const r = run('@layers [core]\n@layer core {\n  fn loop(n) = loop(n + 1)\n}\nloop(0)');
    assert.strictEqual(r.success, false);
    assert.match(r.errors[0].message, /maximum call depth/);
    assert.strictEqual(run('1 + 1').result, '2');
  });

  check('describes the stack and resolution order', () => {
    const d = describe(fact);
    assert.strictEqual(d.success, true);
    assert.strictEqual(d.description.blendMode, 'override');
    assert.deepStrictEqual(d.description.layers.map(l => [l.name, l.pure]), [['core', true], ['debug', false]]);
    assert.deepStrictEqual(d.description.resolutions, [{ name: 'factorial', layers: ['debug', 'core'], mode: 'override' }]);
  });

  check('per-function blend modes: describe, split, and run', () => {
    const src = '@layers [core, data, cache] blend: {fetch: fallback}\n@layer core {\n  fn fetch(k) = data.Some(k)\n}\n@layer cache {\n  fn fetch(k) = data.None()\n}\ndata.unwrap(fetch("a"), "?")';
    assert.deepStrictEqual(describe(src).description.resolutions, [{ name: 'fetch', layers: ['cache', 'core'], mode: 'fallback' }]);
    assert.deepStrictEqual(split(src).blendModes, { fetch: 'fallback' });
    assert.strictEqual(run(src).result, 'a');
  });

  check('splits files, including impure and inactive layers', () => {
    const f = split('@layers [core, cache] blend: fallback\n@layer<impure> cache {\n  fn a() = 1\n}\n@layer hidden {\n  fn b() = 2\n}\na()');
    assert.strictEqual(f.success, true);
    assert.deepStrictEqual(f.enabledLayers, ['core', 'cache']);
    assert.deepStrictEqual(f.declaredLayers, ['cache', 'hidden']);
    assert.strictEqual(f.purity.cache, 'impure');
    assert.strictEqual(f.blendMode, 'fallback');
    assert.strictEqual(f.layers.hidden, 'fn b() = 2');
    assert.strictEqual(f.mainExpression, 'a()');
  });

  if (failures > 0) {
    console.log(`\n${failures} check(s) failed`);
    process.exit(1);
  }
  console.log('\nAll WASM checks passed');
  process.exit(0);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
