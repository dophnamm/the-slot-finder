export const meta = {
  name: 'review-levels',
  description: 'Review the current change in 5 levels: L0 machine checks gate L1-L4 AI reviewers',
  whenToUse: 'Before opening a pull request. Args: the base branch (default main).',
  phases: [
    { title: 'L0 Automated Code Validator', detail: 'make check-server / check-client / check-frozen, no AI', model: 'haiku' },
    { title: 'L1-L4 AI review', detail: 'functional, maintainability, software design, architecture compliance' },
  ],
}

const base = (typeof args === 'string' ? args : args?.base) || 'main'
const L0_PHASE = 'L0 Automated Code Validator'
const AI_PHASE = 'L1-L4 AI review'
const L0_OPTS = { phase: L0_PHASE, model: 'haiku', effort: 'low' }
// Same paths filter as .github/workflows/check.yml: these files trigger both sides.
const SHARED_FILES = ['Makefile', '.github/workflows/check.yml']
const SEVERITY_ORDER = ['blocker', 'major', 'minor']

const CHANGE_SCHEMA = {
  type: 'object',
  properties: {
    merge_base: { type: 'string' },
    tracked: { type: 'array', items: { type: 'string' } },
    untracked: { type: 'array', items: { type: 'string' } },
  },
  required: ['merge_base', 'tracked', 'untracked'],
}

const RUN_SCHEMA = {
  type: 'object',
  properties: {
    results: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          command: { type: 'string' },
          exit_code: { type: 'integer' },
          output_tail: { type: 'string' },
        },
        required: ['command', 'exit_code', 'output_tail'],
      },
    },
  },
  required: ['results'],
}

const FINDINGS_SCHEMA = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          file: { type: 'string' },
          line: { type: 'integer' },
          severity: { type: 'string', enum: SEVERITY_ORDER },
          summary: { type: 'string' },
          evidence: { type: 'string' },
          fix: { type: 'string' },
        },
        required: ['file', 'line', 'severity', 'summary', 'evidence', 'fix'],
      },
    },
  },
  required: ['findings'],
}

const LEVELS = [
  {
    id: 'L1',
    name: 'AI Functional Reviewer',
    purpose: 'The code does what was asked: the requirements, the business rules and the edge cases.',
    scope: 'Behaviour, scenarios, and whether tests cover those scenarios.',
    guide: `Requirements come from the feature spec in docs/features/ (find it by the branch name from \`git branch --show-current\` or by the changed area; the list is docs/features/README.md) and its plan in docs/plans/. Business rules and metric formulas are in README.md section 7.
For each acceptance criterion the change touches: is it met, and does a test cover it? Look for missed edge cases: empty input, zero capacity, ISO week boundaries and the organization time zone, each role (planner, manager, specialist), stale versions (409), merges into the open auto re-plan draft.
When user-visible behaviour changed, check that the feature spec, server/README.md or client/README.md and docs/user-manual.md still describe it.`,
  },
  {
    id: 'L2',
    name: 'AI Maintainability Reviewer',
    purpose: 'New code is easy to read, change and test later.',
    scope: 'Inside one function or file: complexity, duplication, naming, testability.',
    guide: `Judge only new or changed code. Load the repo skills that match the changed files and apply them: writing-go-server-code for the server; structuring-client-components, styling-client-components and translating-client-text for the client.
Flag: long or deeply nested functions that hide the logic, copied blocks that should reuse an existing helper (name it), misleading names, literals the "never hardcode" rule says must be named constants, logic that cannot be unit tested without heavy setup.`,
  },
  {
    id: 'L3',
    name: 'AI Software Design Reviewer',
    purpose: 'Module split, responsibilities and dependency direction make sense.',
    scope: 'Inside the application: module, component, package, state.',
    guide: `Load the repo skills structuring-server-modules (server) and organizing-client-shared-code (client) when they match the changed files.
Server: handlers -> services -> repositories -> models, dtos only at the handler edge, one file per layer per module, internal/engine imports the stdlib only, providers wired through FX.
Client: pages compose features, features never import each other, shared code in shared/, utils/ and resources/, server state in TanStack Query, Zustand for UI and session state only, HTTP only through shared/api/http.ts.
Flag: logic in the wrong layer, a module doing another module's job, wrong-direction or cyclic imports, the same state held in two places, an abstraction with a single caller that does not need it.`,
  },
  {
    id: 'L4',
    name: 'AI Architecture Compliance',
    purpose: 'The code follows the approved architecture. Check compliance only; never redesign.',
    scope: 'The whole system: service boundaries, data ownership, integration, technology, security.',
    guide: `The approved architecture is README.md sections 2-4 (context, containers, server modules), 9 (cross-cutting concerns), 10 (design rules), 14 (the ADR table) and 15 (AI assistant), plus server/README.md sections 8 (settings) and 9 (AI assistant).
Check, among others: metrics are computed on the server as pure functions in internal/metrics; the planning engine stays pure; a human accepts every staffing draft; Keycloak JWT auth with RequireRole and AuthzService checks, CSRF on unsafe requests; a manager sees only the teams of the programs they manage; intake extraction masks customer names, emails and phone numbers before the brief reaches the LLM, and the draft summary sends none of them (README.md section 15, server/README.md section 9); no new service, datastore, queue or technology that no ADR covers.
Each finding names the ADR or README section it breaks. Disagreeing with an approved decision is not a finding.`,
  },
]

const LEVEL_MAP = [
  'L0 machine checks (format, lint, types, build, tests, frozen files) already passed.',
  ...LEVELS.map((level) => `${level.id} ${level.name}: ${level.scope}`),
].join(' ')

const reviewPrompt = (level, change) => `You are review level ${level.id}, ${level.name}, in a five-level code review of this repo.
Read-only: never edit files, never format, never run a git command that changes state.

The change: everything that differs from merge base ${change.merge_base} (base branch ${base}). Run \`git diff ${change.merge_base}\` for the tracked files (commits plus uncommitted work).
Changed tracked files: ${change.tracked.join(', ') || 'none'}.
New untracked files, part of the change, read them whole: ${change.untracked.join(', ') || 'none'}.

Purpose of your level: ${level.purpose}
Your scope: ${level.scope}
${level.guide}

Rules:
- Review the change and the code it directly touches. Read surrounding code to confirm each finding.
- Stay inside your level; anything else belongs to another level, so skip it. ${LEVEL_MAP} Never report lint, format or type issues.
- Report only problems you confirmed in the code, each with file, line, evidence (the quoted code or doc line) and a one-line fix. No praise, no nitpicks, no "might" findings.
- severity: blocker = wrong result, data loss, security hole or a broken approved decision; major = fix before merge; minor = worth fixing, does not block.
- An empty findings list is a valid answer.`

log(`Reviewing the change against ${base}.`)
phase(L0_PHASE)
const change = await agent(
  `Run these read-only git commands and report their output exactly. Do nothing else.
0. \`cd "$(git rev-parse --show-toplevel)"\` so every path is relative to the repo root.
1. \`git merge-base ${base} HEAD\` -> merge_base (the full sha).
2. \`git diff --name-only <merge_base>\` -> tracked (one path per line).
3. \`git ls-files --others --exclude-standard\` -> untracked (one path per line).`,
  { ...L0_OPTS, label: 'L0 change', schema: CHANGE_SCHEMA },
)
if (!change) return { base, error: 'L0 change agent failed' }

const files = [...change.tracked, ...change.untracked]
if (!files.length) {
  log(`No change against ${base}; nothing to review.`)
  return { base, merge_base: change.merge_base, files, l0: [], levels: [] }
}

const sharedChanged = files.some((f) => SHARED_FILES.includes(f))
const needsChecks = (side) => sharedChanged || files.some((f) => f.startsWith(`${side}/`))
const FROZEN_COMMAND = `make check-frozen BASE=${base}`
const commands = [
  ...(needsChecks('server') ? ['make check-server', FROZEN_COMMAND] : []),
  ...(needsChecks('client') ? ['make check-client'] : []),
]

let l0 = []
if (commands.length) {
  const run = await agent(
    `Run each command below from the repo root (\`cd "$(git rev-parse --show-toplevel)" && <command>\`) as its own Bash call with timeout 600000 (they may run in parallel). Do not fix anything, do not judge the output.
For each command report: the command exactly as given, its real exit code, and the last 30 lines of its output.
${commands.map((c) => `- \`${c}\``).join('\n')}`,
    { ...L0_OPTS, label: 'L0 checks', schema: RUN_SCHEMA },
  )
  if (!run) return { base, error: 'L0 check agent failed', commands }
  // The gate trusts exit codes only, and every expected command must have run.
  l0 = commands.map(
    (c) => run.results.find((r) => r.command === c) || { command: c, exit_code: -1, output_tail: 'not run' },
  )
  const failed = l0.filter((r) => r.exit_code !== 0)
  if (failed.length) {
    log(`L0 failed (${failed.map((r) => r.command).join(', ')}); L1-L4 skipped. Fix L0 first.`)
    return { base, merge_base: change.merge_base, files, l0, levels: [], skipped: ['L1', 'L2', 'L3', 'L4'] }
  }
  // server/scripts/check-frozen.sh prints this marker when a rule could not run.
  const frozen = l0.find((r) => r.command === FROZEN_COMMAND)
  if (frozen?.output_tail.includes('SKIP ')) log('L0 skipped a frozen-file rule; see the check-frozen output.')
} else {
  log('No server or client file changed; L0 has nothing to run.')
}

phase(AI_PHASE)
const reviews = await parallel(
  LEVELS.map(
    (level) => () => agent(reviewPrompt(level, change), { phase: AI_PHASE, label: `${level.id} ${level.name}`, schema: FINDINGS_SCHEMA }),
  ),
)

const levels = LEVELS.map((level, i) => {
  const review = reviews[i]
  if (!review) log(`${level.id} did not finish; its findings are missing.`)
  const findings = (review?.findings || []).sort(
    (a, b) => SEVERITY_ORDER.indexOf(a.severity) - SEVERITY_ORDER.indexOf(b.severity),
  )
  return { id: level.id, name: level.name, finished: Boolean(review), findings }
})

return { base, merge_base: change.merge_base, files, l0, levels }
