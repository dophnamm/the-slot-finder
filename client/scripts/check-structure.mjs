// Checks the component folder convention under src/ (see .claude/skills/structuring-client-components).
import { readdirSync, statSync } from 'node:fs'
import { join, relative, sep } from 'node:path'

const SRC = new URL('../src/', import.meta.url).pathname
// Folders whose direct children are component or feature folders.
const ROOTS = ['app', 'pages', 'features', 'shared/components']
// app/ also holds bootstrap files that are not components.
const ROOT_FILES = { app: ['main.tsx', 'router.tsx'] }
const ALLOWED_FILES = /^(index\.ts|component\.tsx|types\.ts|mock-data\.ts|constants\.ts|utils\.ts|api\.ts|use-[a-z0-9-]+\.ts)$/
const KEBAB = /^[a-z0-9]+(-[a-z0-9]+)*$/
// owner (1) → components/child (2) → components/grandchild (3)
const MAX_COMPONENTS_DEPTH = 2

const problems = []

function checkFolder(dir, componentsDepth) {
  const name = dir.split(sep).at(-1)
  const entries = readdirSync(dir)
  const rel = relative(SRC, dir)

  if (!KEBAB.test(name)) problems.push(`${rel}: folder names are kebab-case`)
  if (componentsDepth > MAX_COMPONENTS_DEPTH) {
    problems.push(`${rel}: components nest at most 3 layers deep; flatten it or move it to shared/components`)
  }
  if (entries.includes('component.tsx') && !entries.includes('index.ts')) {
    problems.push(`${rel}: a component folder needs an index.ts that re-exports it`)
  }

  for (const entry of entries) {
    const path = join(dir, entry)
    if (statSync(path).isDirectory()) {
      if (entry === 'components') {
        for (const child of readdirSync(path)) {
          const childPath = join(path, child)
          if (statSync(childPath).isDirectory()) checkFolder(childPath, componentsDepth + 1)
          else problems.push(`${relative(SRC, childPath)}: put each child component in its own folder under components/`)
        }
      } else {
        problems.push(`${relative(SRC, path)}: only a components/ folder may sit inside a component or feature folder`)
      }
    } else if (!ALLOWED_FILES.test(entry)) {
      problems.push(`${relative(SRC, path)}: unexpected file name (allowed: index.ts, component.tsx, types.ts, mock-data.ts, constants.ts, utils.ts, api.ts, use-*.ts)`)
    }
  }
}

for (const root of ROOTS) {
  const rootPath = join(SRC, root)
  for (const entry of readdirSync(rootPath)) {
    const path = join(rootPath, entry)
    if (statSync(path).isDirectory()) checkFolder(path, 0)
    else if (!(ROOT_FILES[root] ?? []).includes(entry)) {
      problems.push(`${relative(SRC, path)}: put it in its own folder (${root}/<name>/component.tsx)`)
    }
  }
}

if (problems.length) {
  console.error(`Structure check failed:\n${problems.map((p) => `  - ${p}`).join('\n')}`)
  process.exit(1)
}
console.log('Structure check passed.')
