// Generates the blend-mode animations (animated SVG, CSS keyframes only) used
// by the README, LANGUAGE_SPEC and the editor's Help dialog:
//
//   node docs/images/blend/generate.mjs
//
// Each animation runs one call through a three-layer stack: a token travels
// through the layers that run, the running layer lights up, each layer's
// outcome appears beside it, and a caption narrates. With reduced motion (or
// no animation support) the SVG shows its final frame, which summarises the
// outcome. CI checks that the committed SVGs match this script.
import { writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const OUT_DIR = dirname(fileURLToPath(import.meta.url))

const D = 10 // seconds per loop
const FADE = 0.25
const END = 9.4 // everything has faded out by D

const W = 700
const H = 336
const X0 = 230 // stack left
const X1 = 510 // stack right
const BOX_H = 46
const boxY = i => 104 + i * 60
const cy = i => boxY(i) + 23
const LEFT = i => [238, cy(i)] // where the token enters a layer
const RIGHT = i => [502, cy(i)] // where it leaves
const CALL = [194, 127]
const RESULT = [528, 127]
const RX = 528 // result column
const RW = 160

const esc = s =>
  String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
const pct = t => `${Math.max(0, Math.min(100, (t / D) * 100)).toFixed(2)}%`

const TAG_STYLES = {
  ok: { bg: '#dafbe1', fg: '#116329' },
  miss: { bg: '#fff1e5', fg: '#953800' },
  off: { bg: '#eaeef2', fg: '#57606a' },
  value: { bg: '#ddf4ff', fg: '#0969da' },
}

function render(spec) {
  const keyframes = []
  let n = 0
  const anim = name => `animation:${name} ${D}s linear infinite`

  // Fade in at t0, out at t1.
  const visible = (t0, t1) => {
    const name = `${spec.id}${n++}`
    const frames = ['0%{opacity:0}']
    if (t0 > 0) frames.push(`${pct(t0)}{opacity:0}`)
    frames.push(`${pct(t0 + FADE)}{opacity:1}`, `${pct(t1 - FADE)}{opacity:1}`, `${pct(t1)}{opacity:0}`, '100%{opacity:0}')
    keyframes.push(`@keyframes ${name}{${frames.join('')}}`)
    return anim(name)
  }

  // Move through [t, x, y] waypoints, visible only while moving.
  const route = points => {
    const name = `${spec.id}${n++}`
    const at = (x, y) => `transform:translate(${x}px,${y}px)`
    const [t0, x0, y0] = points[0]
    const [tn, xn, yn] = points[points.length - 1]
    const frames = [`0%{opacity:0;${at(x0, y0)}}`, `${pct(t0 - 0.05)}{opacity:0;${at(x0, y0)}}`]
    for (const [t, x, y] of points) frames.push(`${pct(t)}{opacity:1;${at(x, y)}}`)
    frames.push(`${pct(tn + 0.25)}{opacity:0;${at(xn, yn)}}`, `100%{opacity:0;${at(xn, yn)}}`)
    keyframes.push(`@keyframes ${name}{${frames.join('')}}`)
    return anim(name)
  }

  const parts = []

  // Header
  parts.push(
    `<text x="24" y="38" class="title">${esc(spec.title)}</text>`,
    `<text x="24" y="60" class="subtitle">${esc(spec.subtitle)}</text>`,
    `<text x="24" y="86" class="mono directive">${esc(spec.directive)}</text>`,
  )

  // Call and result areas
  parts.push(
    `<text x="24" y="108" class="label">CALL</text>`,
    `<rect x="24" y="114" width="170" height="26" rx="13" fill="#f6f8fa" stroke="#d0d7de"/>`,
    `<text x="109" y="131" text-anchor="middle" class="mono pill">${esc(spec.call)}</text>`,
    `<text x="${RX}" y="108" class="label">RESULT</text>`,
  )

  // Layers (top of the stack first)
  spec.layers.forEach((layer, i) => {
    const y = boxY(i)
    parts.push(
      `<rect x="${X0}" y="${y}" width="${X1 - X0}" height="${BOX_H}" rx="8" fill="#fff" stroke="#d0d7de"/>`,
    )
    for (const [t0, t1] of layer.active ?? []) {
      parts.push(
        `<rect x="${X0}" y="${y}" width="${X1 - X0}" height="${BOX_H}" rx="8" fill="${layer.color}" fill-opacity="0.12" stroke="${layer.color}" stroke-width="2" opacity="0" style="${visible(t0, t1)}" class="a"/>`,
      )
    }
    parts.push(
      `<rect x="${X0}" y="${y}" width="6" height="${BOX_H}" rx="3" fill="${layer.color}"/>`,
      `<text x="${X0 + 16}" y="${y + 18}" class="name">${esc(layer.name)}</text>`,
      `<text x="${X0 + 16}" y="${y + 36}" class="mono def">${esc(layer.def)}</text>`,
    )
    if (layer.tag) {
      const { text, kind, at } = layer.tag
      const { bg, fg } = TAG_STYLES[kind]
      const w = Math.round(text.length * 6.4 + 16)
      const x = X1 - 10 - w
      parts.push(
        `<g opacity="1" style="${visible(at, END)}" class="a">` +
          `<rect x="${x}" y="${y + 6}" width="${w}" height="17" rx="8.5" fill="${bg}"/>` +
          `<text x="${x + w / 2}" y="${y + 18.5}" text-anchor="middle" class="tag" fill="${fg}">${esc(text)}</text></g>`,
      )
    }
  })

  // Result: one value, or a list built up item by item
  if (spec.result.items) {
    const items = spec.result.items
    const h = 16 * items.length + 14
    parts.push(
      `<g opacity="1" style="${visible(items[0].at, END)}" class="a"><rect x="${RX}" y="114" width="${RW}" height="${h}" rx="8" fill="#ddf4ff" stroke="#0969da"/></g>`,
    )
    items.forEach((item, k) => {
      const prefix = k === 0 ? '[' : ' '
      const suffix = k === items.length - 1 ? ']' : ','
      parts.push(
        `<text x="${RX + 10}" y="${131 + k * 16}" class="mono result" opacity="1" style="${visible(item.at, END)}">${esc(prefix + item.text + suffix)}</text>`,
      )
    })
  } else {
    parts.push(
      `<g opacity="1" style="${visible(spec.result.at, END)}" class="a">` +
        `<rect x="${RX}" y="114" width="${RW}" height="26" rx="13" fill="#ddf4ff" stroke="#0969da"/>` +
        `<text x="${RX + RW / 2}" y="131" text-anchor="middle" class="mono result">${esc(spec.result.text)}</text></g>`,
    )
  }

  if (spec.note) {
    parts.push(`<text x="${X0}" y="290" class="note">${esc(spec.note)}</text>`)
  }

  // The travelling token
  parts.push(
    `<g opacity="0" style="${route(spec.route)}" class="a"><circle r="11" fill="#0969da" fill-opacity="0.18"/><circle r="5.5" fill="#0969da"/></g>`,
  )

  // Captions: the last one stays visible in the final (static) frame
  spec.captions.forEach(([t0, t1, text], k) => {
    const last = k === spec.captions.length - 1
    parts.push(
      `<text x="${W / 2}" y="316" text-anchor="middle" class="caption" opacity="${last ? 1 : 0}" style="${visible(t0, t1)}">${esc(text)}</text>`,
    )
  })

  const css = [
    'text{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Helvetica,Arial,sans-serif;fill:#24292f}',
    '.mono{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}',
    '.title{font-size:20px;font-weight:700}',
    '.subtitle{font-size:13px;fill:#57606a}',
    '.directive{font-size:12px;fill:#8250df}',
    '.label{font-size:10px;font-weight:600;letter-spacing:.08em;fill:#8c959f}',
    '.pill{font-size:12px}',
    '.name{font-size:13px;font-weight:600}',
    '.def{font-size:11.5px;fill:#57606a}',
    '.tag{font-size:11px;font-weight:600}',
    '.result{font-size:12px;fill:#0550ae}',
    '.caption{font-size:13.5px}',
    '.note{font-size:11px;fill:#8c959f}',
    ...keyframes,
    '@media (prefers-reduced-motion:reduce){*{animation:none!important}}',
  ].join('\n')

  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-labelledby="t d">
<title id="t">Layer blend mode: ${esc(spec.title)}</title>
<desc id="d">${esc(spec.description)}</desc>
<style>
${css}
</style>
<rect x="0.5" y="0.5" width="${W - 1}" height="${H - 1}" rx="12" fill="#ffffff" stroke="#d0d7de"/>
${parts.join('\n')}
</svg>
`
}

// Waypoints between layers go around the boxes: down the left or right side,
// or through the gap between two boxes.
const gap = i => boxY(i) + BOX_H + 7 // middle of the gap below box i

const specs = [
  {
    id: 'override',
    title: 'override',
    subtitle: 'The top definition answers; the ones below are shadowed.',
    directive: '@layers [core, formal, friendly]          # blend: override is the default',
    call: 'greet("Ada")',
    description:
      'A call to greet goes to the top layer, friendly, which answers "Hi Ada!". The definitions in formal and core are shadowed, but a layer can still call them explicitly, e.g. core.greet(n).',
    layers: [
      { name: 'friendly', color: '#e91e63', def: 'fn greet(n) = "Hi {n}!"', active: [[1.5, 3.2]], tag: { text: 'answers', kind: 'ok', at: 3.0 } },
      { name: 'formal', color: '#3f51b5', def: 'fn greet(n) = "Good day, {n}."', tag: { text: 'shadowed', kind: 'off', at: 3.7 } },
      { name: 'core', color: '#4caf50', def: 'fn greet(n) = "Hello {n}"', tag: { text: 'shadowed', kind: 'off', at: 4.0 } },
    ],
    route: [
      [0.5, ...CALL],
      [1.5, ...LEFT(0)],
      [3.0, ...RIGHT(0)],
      [3.6, ...RESULT],
    ],
    result: { text: '"Hi Ada!"', at: 3.6 },
    captions: [
      [0.2, 1.7, 'A call to greet enters at the top of the stack.'],
      [1.7, 3.7, 'friendly defines greet, so friendly answers.'],
      [3.7, END, 'formal and core are shadowed — a layer can still call core.greet(n) explicitly.'],
    ],
  },
  {
    id: 'chain',
    title: 'chain',
    subtitle: 'The bottom layer runs first; each layer above transforms the previous result.',
    directive: '@layers [core, discount, tax] blend: chain',
    call: 'total([10, 20])',
    description:
      'core.total runs first with the original arguments and returns 30; discount receives 30 and returns 25; tax receives 25 and returns 31.25, which is the result.',
    layers: [
      { name: 'tax', color: '#ff9800', def: 'fn total(x) = x * 1.25', active: [[5.3, 6.4]], tag: { text: '→ 31.25', kind: 'value', at: 6.1 } },
      { name: 'discount', color: '#009688', def: 'fn total(x) = x - 5', active: [[3.6, 4.7]], tag: { text: '→ 25', kind: 'value', at: 4.5 } },
      { name: 'core', color: '#4caf50', def: 'fn total(items) = sum(items)', active: [[1.8, 3.0]], tag: { text: '→ 30', kind: 'value', at: 2.8 } },
    ],
    route: [
      [0.5, ...CALL],
      [0.8, 214, 127],
      [1.6, 214, cy(2)],
      [1.8, ...LEFT(2)],
      [2.8, ...RIGHT(2)],
      [3.05, 518, cy(2)],
      [3.45, 518, cy(1)],
      [3.6, ...RIGHT(1)],
      [4.5, ...LEFT(1)],
      [4.7, 214, cy(1)],
      [5.1, 214, cy(0)],
      [5.3, ...LEFT(0)],
      [6.1, ...RIGHT(0)],
      [6.6, ...RESULT],
    ],
    result: { text: '31.25', at: 6.6 },
    captions: [
      [0.2, 3.0, 'chain starts at the bottom: core.total gets the original arguments → 30.'],
      [3.0, 6.4, 'Each layer above receives the previous result as its first argument.'],
      [6.4, END, "The top layer's result is the answer: 30 → 25 → 31.25."],
    ],
  },
  {
    id: 'collect',
    title: 'collect',
    subtitle: 'Every layer that defines the function runs; the answer is the list of results.',
    directive: '@layers [core, audit, console] blend: collect',
    call: 'notify("login")',
    description:
      'console, audit and core all run notify("login"), top of the stack first, and the result is the list ["console: login", "audit: login", "core: login"].',
    layers: [
      { name: 'console', color: '#3f51b5', def: 'fn notify(e) = "console: {e}"', active: [[1.3, 2.4]], tag: { text: 'ran', kind: 'ok', at: 2.2 } },
      { name: 'audit', color: '#795548', def: 'fn notify(e) = "audit: {e}"', active: [[3.0, 4.0]], tag: { text: 'ran', kind: 'ok', at: 3.8 } },
      { name: 'core', color: '#4caf50', def: 'fn notify(e) = "core: {e}"', active: [[4.6, 5.6]], tag: { text: 'ran', kind: 'ok', at: 5.4 } },
    ],
    route: [
      [0.5, ...CALL],
      [1.3, ...LEFT(0)],
      [2.2, ...RIGHT(0)],
      [2.45, 518, cy(0)],
      [2.85, 518, cy(1)],
      [3.0, ...RIGHT(1)],
      [3.8, ...LEFT(1)],
      [4.0, 214, cy(1)],
      [4.4, 214, cy(2)],
      [4.6, ...LEFT(2)],
      [5.4, ...RIGHT(2)],
    ],
    result: {
      items: [
        { text: '"console: login"', at: 2.3 },
        { text: '"audit: login"', at: 3.9 },
        { text: '"core: login"', at: 5.5 },
      ],
    },
    captions: [
      [0.2, 2.4, 'collect calls every layer that defines notify, top of the stack first…'],
      [2.4, 5.7, '…and keeps each result.'],
      [5.7, END, 'The answer is the list of all the results.'],
    ],
  },
  {
    id: 'fallback',
    title: 'fallback',
    subtitle: 'Try the layers from the top; the first result that is not None wins.',
    directive: '@layers [core, data, disk, memory] blend: fallback',
    call: 'fetch("user:7")',
    description:
      'memory is tried first and returns None (a miss), so disk is tried next and returns Some(42), which is the result. core is never called.',
    layers: [
      { name: 'memory', color: '#00acc1', def: 'fn fetch(k) = memo_get(k)', active: [[1.3, 2.4]], tag: { text: 'None', kind: 'miss', at: 2.2 } },
      { name: 'disk', color: '#ff9800', def: 'fn fetch(k) = disk_get(k)', active: [[3.3, 4.3]], tag: { text: 'Some(42)', kind: 'ok', at: 4.1 } },
      { name: 'core', color: '#4caf50', def: 'fn fetch(k) = compute(k)', tag: { text: 'not called', kind: 'off', at: 4.9 } },
    ],
    note: 'data is in the stack for Some/None but does not define fetch, so it takes no part.',
    route: [
      [0.5, ...CALL],
      [1.3, ...LEFT(0)],
      [2.2, ...RIGHT(0)],
      [2.5, 502, gap(0)],
      [3.0, 238, gap(0)],
      [3.3, ...LEFT(1)],
      [4.1, ...RIGHT(1)],
      [4.6, ...RESULT],
    ],
    result: { text: 'Some(42)', at: 4.6 },
    captions: [
      [0.2, 2.5, 'fallback tries the top layer first: memory misses and returns None…'],
      [2.5, 4.6, '…so the next layer down is tried: disk has the value.'],
      [4.6, END, 'The first result that is not None wins — core is never called.'],
    ],
  },
]

for (const spec of specs) {
  writeFileSync(join(OUT_DIR, `${spec.id}.svg`), render(spec))
  console.log(`wrote ${spec.id}.svg`)
}
