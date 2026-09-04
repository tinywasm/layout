---
PLAN: "refactor!: layout crudview tests use typed view.Backend doubles, then row-count overlay"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Do NOT run `gopush` or `codejob`.

# PLAN — execution queue for `layout`

> If you were told to "execute the plan described in docs/PLAN.md", execute
> **ALL the plans below, in order (top to bottom)**. Each plan is
> self-contained; finish one (its acceptance criteria green) before starting
> the next. Never mix changes from one plan into another.

| Order | Plan | Subject |
|-------|------|---------|
| 1 | [PLAN_VIEW_BACKEND.md](PLAN_VIEW_BACKEND.md) | refactor!: crudview tests use typed view.Backend doubles (view v0.3.0 follow-up) |
| 2 | [PLAN_ROWCOUNT_OVERLAY.md](PLAN_ROWCOUNT_OVERLAY.md) | refactor: crudview drops its row-count overlay (needs components listselect.Header; see its Dispatch order note) |

After completing all plans, run `gotest ./...` one final time: everything green.

> Gate: plan 2 carries its own dispatch-order note (it needs a `components`
> version containing `listselect.Header` in this repo's `go.mod`). If that
> gate does not hold when you reach it, STOP after plan 1 with everything
> green and report plan 2 as still pending — do not force it.
