---
name: astro-development-cycle
description: Use when implementing any feature, bug fix, or refactor. Runs an iterative develop-review-test cycle (minimum 3 iterations) with developer prompts at every decision point. Keeps code concise and under 15 cognitive complexity per SonarQube standards. Ends with a summary of changes and potential risks.
---

# Astro Development Cycle

Iterative development workflow that cycles through three phases — implement, self-review, and local verification — a minimum of 3 times. Every iteration tightens the code toward production-readiness. The developer is prompted for clarification at every non-trivial decision.

## Input — Mandatory Intake Prompt

Before doing ANY work, you MUST ask the developer these four questions and wait for answers. Do NOT proceed until all four are answered.

```
Before I begin, I need to understand the task clearly:

1. **Intention** — What do you want to accomplish? (e.g., add feature, fix bug, refactor, adapt to review feedback)

2. **Location** — Which file(s) or folder(s) should I work in? (e.g., `internal/domain/creditmanager/service/`, specific file path, or "you figure it out based on the domain")

3. **Expected Result** — What should the code do when this is done? (e.g., "new endpoint returns paginated credit history", "bug where balance goes negative is fixed")

4. **Don'ts** — Anything I should NOT do? (e.g., "don't touch the migration", "don't modify existing tests", "don't refactor unrelated code", "don't change the public API signature")
```

**Wait for the developer to answer all four.** If any answer is vague, ask a follow-up to clarify before proceeding. Partial answers like "just do it" or "you decide" are not acceptable for questions 1 and 3 — push back politely and explain why you need clarity.

Once all four are answered, summarize your understanding back to the developer in one short paragraph and ask: **"Is this correct? (yes / correct me)"**

Only after confirmation, proceed to Phase 0.

---

## Phase 0 — Load Context & Plan

### Step 0.1 — Load knowledge base

1. `CLAUDE.md` — already in context
2. `docs/invariants/` — glob `docs/invariants/*.md`, load relevant domain files
3. `docs/testspecs/` — glob `docs/testspecs/**/*.md`, skip if none

### Step 0.2 — Identify scope

Determine from the task:
- Target domain(s) — must be one of: `admanager` | `adindexer` | `creditmanager` | `budgetmanager` | `tracker` | `reporting` | `displaysearch` | `displaybrowse` | `approvalmanager`
- Affected layers (entity, service, repository, handler, subscriber, migration, etc.)
- Domain invariant rules that apply

### Step 0.3 — Present plan and ask for confirmation

Present:

```
Development Plan:
- Domain: <domain>
- Affected layers: <list>
- Applicable invariants: <list of RULE-XXX-NN>
- Estimated iterations: <3–5, based on complexity>

Proceed with this plan?
```

**Wait for developer confirmation.** If the developer corrects the scope, update the plan accordingly.

---

## Agent Roles

This skill uses **separate agents** with distinct responsibilities. The orchestrator (you) coordinates handoffs — never mix roles.

| Agent | Role | Can modify code? | Can read code? |
|-------|------|------------------|----------------|
| **Orchestrator** (you) | Coordinates phases, presents findings, handles developer interaction | No | Yes |
| **Worker Agent** | Implements code, applies fixes | Yes | Yes |
| **Review Agent 1** | Business Correctness review | No | Yes |
| **Review Agent 2** | Performance review | No | Yes |
| **Review Agent 3** | Maintainability review | No | Yes |

**Separation rule:** The Worker Agent MUST NOT self-review. The Review Agents MUST NOT modify code. The Orchestrator delegates — it does not implement or review directly.

---

## Iteration Loop (minimum 3 cycles)

Each iteration runs all three phases in sequence. After each iteration, report progress and ask whether to continue.

```
Iteration 1 of N
┌──────────────────────────────────────────────┐
│  Phase 1: Worker Agent — Implement / Fix      │
│  Phase 2: Review Agents — Parallel Self-Review │
│  Phase 3: Verification — lint / test / mock    │
└──────────────────────────────────────────────┘
```

---

### Phase 1 — Worker Agent (Implement / Fix)

Dispatch a **Worker Agent** via the Agent tool. This agent writes code — it does NOT review its own output.

#### Worker Agent Context

Pass the Worker Agent:
- The development plan from Phase 0 (domain, layers, invariants)
- The developer's intake answers (intention, location, expected result, don'ts)
- `CLAUDE.md` rules and patterns
- On subsequent iterations: the findings list from the previous Phase 2 and Phase 3

#### Worker Agent — First Iteration Instructions

```
You are the Worker Agent. Your job is to IMPLEMENT code, not review it.

Read existing code in every affected layer before writing anything.
Implement bottom-up following the layer order:
1. Migration (if needed)
2. Entity
3. DTO + Mapper
4. Repository interface + implementation
5. Domain errors
6. Service interface + implementation
7. gRPC handler (if applicable)
8. Subscriber handler (if applicable)
9. Tests for each layer

Code Complexity Rules (SonarQube <15):
- Flatten nested control flow with guard clauses (early returns)
- Extract complex conditional logic into named boolean variables
- Break functions exceeding ~40 lines into focused helpers (only if reused or significantly reduces complexity)
- Avoid deeply nested if/else/switch (max 3 levels)
- Replace complex switch/case chains with maps where applicable

Conciseness Rules:
- No dead code, no commented-out code
- No single-use trivial helpers (under ~10 lines called from exactly one place)
- No over-engineered abstractions for one-time operations
- Prefer 3 similar lines over a premature abstraction
- No unnecessary error wrapping — let errors bubble up per layer rules

DO NOT review your own output. Just implement and report what files you created/modified.

Return: list of files created or modified with a 1-line description of each.
```

#### Worker Agent — Subsequent Iteration Instructions

```
You are the Worker Agent. Your job is to FIX issues found by the Review Agents.

Findings to address:
<paste findings list from previous Phase 2>

Verification failures to address:
<paste failures from previous Phase 3>

For each finding:
1. Read the current file state at the reported location
2. Apply the fix
3. Report what you changed as a before/after summary

DO NOT skip CRITICAL or HIGH findings.
DO NOT review your own fixes — that's the Review Agents' job.

Return: list of fixes applied with before/after summary for each.
```

#### Orchestrator — After Worker Agent Returns

Present the Worker Agent's output to the developer:

```
Worker Agent completed. Files changed:
- <file>: <description>
- ...

Proceeding to review phase. The Worker Agent will NOT review its own code —
3 independent Review Agents will inspect the changes.
```

---

### Phase 2 — Review Agents (Parallel Self-Review)

Dispatch **3 Review Agents in a single message** (parallel) via the Agent tool. These agents inspect the Worker Agent's output — they do NOT modify code.

#### Shared Context for All Review Agents

Pass each Review Agent:
- The list of files created/modified by the Worker Agent
- `CLAUDE.md` rules
- `docs/invariants/<domain>.md` content
- `docs/testspecs/` content (if any)
- The developer's don'ts from intake (review agents should not flag things the developer explicitly excluded)

#### Review Agent 1: Business Correctness

```
You are Review Agent 1 — Business Correctness. You inspect code for rule violations. You do NOT modify code.

Read each file the Worker Agent created or modified. For each file, check:

- Domain invariants from docs/invariants/<domain>.md — cite specific RULE-XXX-NN
- CLAUDE.md invariants: no gRPC imports in service layer, clock.Now() not time.Now(),
  funcall.GetCallerName() called once per function, no logging in service layer,
  domain errors only (var ErrXxx = errors.New(...))
- Audit columns: deleted_by, updated_by, created_by must always be set on mutations
- Cross-domain interaction: must go through /integration gRPC, never direct service calls
- Subscriber handlers: must be thin, delegate all business logic to domain service
- Transaction safety: related mutations (ledger + balance, etc.) must be in same DB transaction
- Repository interfaces use entity types, never DTOs
- Match against testspecs if relevant testspec exists for the affected domain

For each finding, return:
- Severity: CRITICAL or HIGH
- File path and line number
- 1-sentence summary
- Rule or invariant being violated (cite by name)
- Suggested fix as a before/after code block (max ~15 lines)

Return findings as a JSON array:
[{"severity": "CRITICAL", "file": "path", "line": 42, "summary": "...", "rule": "RULE-XXX-NN", "fix": "..."}]

If no findings: return an empty array [].
```

#### Review Agent 2: Performance

```
You are Review Agent 2 — Performance. You inspect code for performance issues. You do NOT modify code.

Read each file the Worker Agent created or modified. For each file, check:

- N+1 DB calls: loops that call DB per iteration instead of batching
- JOIN without DISTINCT/GROUP BY producing duplicate rows
- Missing pagination.PageSizeNoLimit on Scan* calls (unbounded full-table scans)
- O(n²) linear scans where a map/index already exists or could be built once
- Redundant DB reads (fetching the same row multiple times within one request)
- Missing DB indexes implied by new WHERE/ORDER BY columns

For each finding, return:
- Severity: HIGH or MEDIUM
- File path and line number
- 1-sentence summary
- Category (N+1, O(n²), missing index, etc.)
- Suggested fix as a before/after code block (max ~15 lines)

Return findings as a JSON array:
[{"severity": "HIGH", "file": "path", "line": 67, "summary": "...", "category": "N+1", "fix": "..."}]

If no findings: return an empty array [].
```

#### Review Agent 3: Maintainability

```
You are Review Agent 3 — Maintainability. You inspect code for quality and complexity issues. You do NOT modify code.

Read each file the Worker Agent created or modified. For each file, check:

- Cognitive complexity per function (target <15, flag >=10 for awareness, flag >=15 as must-fix)
- Dead code: switch/case blocks with empty cases, unreachable branches
- Misleading names: functions named as queries that return mutation results, inverted boolean conventions
- Over-decomposition: single-use helper functions (~10-15 lines) called from exactly one place
- Magic indices: positional array access (Ads[2].Products[1]) in tests instead of named variables
- Struct names that describe implementation instead of intent
- Import hygiene: unused imports, wrong-layer imports (e.g., gRPC codes in domain service)
- Magic numbers without constants
- Duplicate code that should be extracted

For each finding, return:
- Severity: MEDIUM or LOW
- File path and line number
- 1-sentence summary
- Category (complexity, dead code, naming, etc.)
- Suggested fix as a before/after code block (max ~15 lines)

Return findings as a JSON array:
[{"severity": "MEDIUM", "file": "path", "line": 30, "summary": "...", "category": "complexity", "fix": "..."}]

If no findings: return an empty array [].
```

#### Orchestrator — Deduplication

After all 3 Review Agents return:

1. Collect all findings into a flat list
2. Merge findings that point to the same file+line from different agents (keep the more severe label, combine descriptions)
3. Discard exact duplicates (identical file+line+issue)
4. Sort: Critical → High → Medium → Low, then alphabetically by file within each severity

#### Orchestrator — Present Findings

Format the deduplicated findings as a table:

```
Iteration N — Self-Review Findings (3 independent review agents):

| # | Severity | Agent | File | Line | Finding | Rule/Category |
|---|----------|-------|------|------|---------|---------------|
| 1 | CRITICAL | Correctness | service/create_ad.go | 42 | Missing deleted_by on bulk delete | RULE-ADM-03 |
| 2 | HIGH | Performance | service/create_ad.go | 67 | N+1 DB call in loop | N+1 |
| 3 | MEDIUM | Maintainability | service/create_ad.go | 30 | Cognitive complexity ~18 | complexity |
| 4 | LOW | Maintainability | service/create_ad.go | 12 | Unused import | import hygiene |

CRITICAL/HIGH findings will be sent to the Worker Agent in the next iteration.
MEDIUM findings are recommended — should I include them? (yes / skip / defer)
LOW findings are optional — should I include them? (yes / skip)
```

**Wait for developer response on MEDIUM and LOW items.** CRITICAL and HIGH are always included in the next iteration's fix list.

---

### Phase 3 — Local Verification

The orchestrator runs local checks directly (no agent needed).

#### Step 3.1 — Regenerate mocks (if interfaces changed)

```bash
make mock
```

If `make mock` fails, report the error and **ask the developer** how to proceed before attempting a fix.

#### Step 3.2 — Run linter

```bash
make lint
```

Collect all lint errors. For each error:
- Classify as auto-fixable (import ordering, formatting) or requires manual intervention
- **Ask the developer:** "Lint found N errors. Auto-fix the M formatting issues and show the remaining K for your review?"

#### Step 3.3 — Run tests

```bash
make test
```

If tests fail:
- Identify which test files and test cases failed
- Read the failing test to understand what it asserts
- Read the implementation code the test covers
- **Present the failure to the developer:**

```
Test failure: TestGetCreditBalance_Success
File: service/get_credit_balance_test.go:42
Error: expected 10000, got 0

Likely cause: <analysis>

Proposed fix: <description>

Apply this fix? (yes / no / let me look at it first)
```

**Never silently fix a test failure.** Always present the failure and proposed fix to the developer first.

#### Step 3.4 — Report verification results

```
Iteration N — Verification Results:

| Check | Status | Details |
|-------|--------|---------|
| make mock | PASS / FAIL | <details if failed> |
| make lint | PASS / FAIL (N errors) | <summary of errors> |
| make test | PASS / FAIL (N failures) | <list of failing tests> |

All checks passed — ready for next iteration.
— OR —
N issues found — these will be sent to the Worker Agent in the next iteration.
```

---

### End-of-Iteration Checkpoint

After Phase 3, present the iteration summary:

```
Iteration N Complete:
- Worker Agent: <N files created/modified>
- Review Agents: <N findings> (X critical, Y high, Z medium, W low)
- Verification: mock <PASS/FAIL>, lint <PASS/FAIL>, test <PASS/FAIL>

Remaining issues for next iteration: <count>
Minimum iterations remaining: <max(0, 3 - current_iteration)>

Continue to iteration N+1? (yes / stop here / adjust scope)
```

**Wait for developer response.**

- **"yes"** — proceed to next iteration (dispatch Worker Agent with findings list)
- **"stop here"** — skip to Final Summary (Phase 4)
- **"adjust scope"** — developer provides new direction, update plan accordingly

#### When to stop iterating

The cycle must run **at least 3 iterations**. After iteration 3, recommend stopping when:
- All CRITICAL and HIGH findings are resolved
- `make lint` and `make test` both pass
- No new findings from the Review Agents

If after 5 iterations there are still CRITICAL/HIGH findings, **escalate to the developer:**
> "After 5 iterations, these issues remain unresolved: <list>. These may require architectural changes or clarification on requirements. How would you like to proceed?"

---

## Phase 4 — Final Summary

After the last iteration, produce a comprehensive summary.

### Step 4.1 — Change inventory

```
## Development Cycle Summary

### Task
<1-sentence description of what was implemented>

### Iterations Completed: N

### Files Changed
| File | Action | Description |
|------|--------|-------------|
| internal/domain/<domain>/entity/foo.go | Created | New entity for Foo |
| internal/domain/<domain>/service/create_foo.go | Created | Service method implementation |
| ... | | |

### Tests Added/Modified
| Test File | Cases | Coverage |
|-----------|-------|----------|
| service/create_foo_test.go | 4 (happy + 3 error paths) | Happy path, validation, repo error, edge case |
| ... | | |
```

### Step 4.2 — Resolved findings across iterations

```
### Findings Resolved
| Iteration | Severity | Finding | Resolution |
|-----------|----------|---------|------------|
| 1 | CRITICAL | Missing audit column | Added deleted_by to bulk delete |
| 2 | HIGH | N+1 in loop | Replaced with batch GetAdsByIDs |
| 2 | MEDIUM | Complexity 18 | Extracted validation into helper |
| ... | | | |
```

### Step 4.3 — Potential bugs and risks

This is the most important section. List everything the developer should be aware of:

```
### Potential Bugs & Risks

**Known Risks (address before merging):**
- [ ] <Description of risk> — `file.go:line` — <why this is risky>
- [ ] <Description of risk> — <context>

**Edge Cases Not Covered by Tests:**
- <Scenario that could fail in production but has no test>
- <Concurrent access pattern that wasn't tested>

**Domain Invariant Concerns:**
- <Any invariant that is partially satisfied or has a caveat>
- <Any financial safety rule that depends on correct caller behavior>

**Performance Considerations:**
- <Query patterns that may not scale>
- <Missing indexes that should be added if data grows>

**Dependencies & Assumptions:**
- <External service behavior assumed but not verified>
- <Cache TTL assumptions>
- <PubSub ordering assumptions>
```

### Step 4.4 — Recommended next steps

```
### Recommended Next Steps
1. Review the potential risks above before committing
2. Run `make test` one final time after any manual adjustments
3. Consider adding integration tests for: <list scenarios>
4. If creating a PR, run `/review-pr-ads` for external review
```

---

## Common Mistakes

| Mistake | Fix |
|---|---|
| Silently fixing code without asking the developer | Always show the change and ask for approval |
| Making confident architectural decisions | Ask the developer — "Should I use pattern A or B?" |
| Skipping iterations because "it looks good" | Minimum 3 iterations, no exceptions |
| Fixing test failures by weakening assertions | Fix the implementation, not the test — ask the developer if unsure |
| Auto-fixing lint errors without showing them | Show the errors first, ask to auto-fix the safe ones |
| Stopping after `make lint` passes but skipping `make test` | All three checks (mock, lint, test) run every iteration |
| Making a design decision when two valid approaches exist | Present both options with tradeoffs, let the developer choose |
| Assuming a domain invariant doesn't apply | Check `docs/invariants/<domain>.md` — if ambiguous, ask |
| Writing code that is correct but has complexity >=15 | Refactor before moving to Phase 2 — complexity is a Phase 1 concern |
| Reporting the same finding across multiple iterations without progress | Escalate after 2 iterations of the same finding — the approach may need to change |
| Worker Agent reviewing its own code | Worker implements only — dispatch Review Agents separately for inspection |
| Review Agents modifying code | Review Agents report only — fixes go to Worker Agent in the next iteration |
| Running all 3 Review Agents sequentially | Always dispatch all 3 in a single parallel message for speed |
| Skipping deduplication of review findings | Merge findings at same file+line from different agents before presenting |
