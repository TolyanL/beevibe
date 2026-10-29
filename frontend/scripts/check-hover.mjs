// Fails on two hover hazards in .vue files.
//
// 1. A raw palette class used as a hover state (bg-canvas, bg-surface,
//    text-bee-deep, bg-graphite) blends into the surface in one of the themes.
//    Use the semantic hover tokens instead: bg-hover-surface,
//    bg-hover-surface-strong, text-hover-accent, bg-hover-ink, or the
//    .hover-row / .hover-link / .hover-button classes.
// 2. The same hover property declared twice on one element — e.g. class has
//    hover:text-ink and the active branch adds hover:text-canvas. Which one
//    wins depends on Tailwind's CSS order, not on the order in the attribute,
//    so the label of an active pill can end up invisible. Put the hover state
//    in a single branch of the ternary.
import { readdirSync, readFileSync } from 'node:fs'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const SOURCE_ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '../src')
const banned = ['hover:bg-canvas', 'hover:bg-surface', 'hover:text-bee-deep', 'hover:bg-graphite']

const MULTI_SEGMENT = [
  'border-t',
  'border-r',
  'border-b',
  'border-l',
  'inset-x',
  'inset-y',
  'outline-offset',
  'ring-offset',
  'scale-x',
  'scale-y',
  'skew-x',
  'skew-y',
  'space-x',
  'space-y',
  'translate-x',
  'translate-y',
]

const TAG = /<([A-Za-z][\w.:-]*)((?:"[^"]*"|'[^']*'|[^>"'])*)>/g
const ATTR = /(^|\s)(:?)class\s*=\s*("([^"]*)"|'([^']*)')/g
const LITERAL = /'([^']*)'|"([^"]*)"|`([^`]*)`/g

const files = []
for (const name of readdirSync(SOURCE_ROOT, { recursive: true })) {
  if (extname(String(name)) === '.vue') {
    files.push(join(SOURCE_ROOT, String(name)))
  }
}

function propertyGroup(utility) {
  const prefixed = MULTI_SEGMENT.find(
    (prefix) => utility === prefix || utility.startsWith(`${prefix}-`),
  )
  return prefixed ?? utility.split('-')[0]
}

function classTokens(attributes) {
  const tokens = []
  for (const attribute of attributes.matchAll(ATTR)) {
    const value = attribute[4] ?? attribute[5] ?? ''
    const chunks =
      attribute[2] === ':'
        ? [...value.matchAll(LITERAL)].map((lit) => lit[1] ?? lit[2] ?? lit[3] ?? '')
        : [value]
    for (const chunk of chunks) {
      tokens.push(...chunk.split(/\s+/))
    }
  }
  return tokens.filter(Boolean)
}

const violations = []

for (const file of files) {
  const source = readFileSync(file, 'utf8')
  const path = relative(process.cwd(), file)
  const lineAt = (index) => source.slice(0, index).split('\n').length

  source.split('\n').forEach((line, index) => {
    for (const token of banned) {
      if (line.includes(token)) {
        violations.push(`${path}:${index + 1}: ${token}`)
      }
    }
  })

  for (const tag of source.matchAll(TAG)) {
    const hovered = new Map()
    for (const token of classTokens(tag[2] ?? '')) {
      if (!token.startsWith('hover:')) continue
      const utility = token.slice('hover:'.length)
      const group = propertyGroup(utility)
      if (!hovered.has(group)) hovered.set(group, new Set())
      hovered.get(group).add(utility)
    }
    for (const [group, values] of hovered) {
      if (values.size < 2) continue
      violations.push(
        `${path}:${lineAt(tag.index)}: <${tag[1]}> declares hover:${group}-* twice → ${[...values].join(', ')}`,
      )
    }
  }
}

if (violations.length > 0) {
  console.error(
    'Hover classes are not theme-safe — use hover tokens (bg-hover-surface, bg-hover-surface-strong, text-hover-accent, bg-hover-ink) and keep one hover state per property:',
  )
  for (const violation of violations) {
    console.error(`  ${violation}`)
  }

  process.exitCode = 1
}
