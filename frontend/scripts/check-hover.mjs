// Fails when a raw palette class is used as a hover state in .vue files.
// Raw hover classes (bg-canvas, bg-surface, text-bee-deep, bg-graphite) blend
// into the surface in one of the themes. Use the semantic hover tokens instead:
// bg-hover-surface, bg-hover-surface-strong, text-hover-accent, bg-hover-ink,
// or the .hover-row / .hover-link / .hover-button classes.
import { readdirSync, readFileSync } from 'node:fs';
import { dirname, extname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const SOURCE_ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '../src');
const banned = [
  'hover:bg-canvas',
  'hover:bg-surface',
  'hover:text-bee-deep',
  'hover:bg-graphite',
];

const files = [];
for (const name of readdirSync(SOURCE_ROOT, { recursive: true })) {
  if (extname(String(name)) === '.vue') {
    files.push(join(SOURCE_ROOT, String(name)));
  }
}

const violations = [];

for (const file of files) {
  const lines = readFileSync(file, 'utf8').split('\n');

  lines.forEach((line, index) => {
    for (const token of banned) {
      if (line.includes(token)) {
        violations.push(`${relative(process.cwd(), file)}:${index + 1}: ${token}`);
      }
    }
  });
}

if (violations.length > 0) {
  console.error('Banned hover classes found — use hover tokens (bg-hover-surface, bg-hover-surface-strong, text-hover-accent, bg-hover-ink):');
  for (const violation of violations) {
    console.error(`  ${violation}`);
  }

  process.exitCode = 1;
}
