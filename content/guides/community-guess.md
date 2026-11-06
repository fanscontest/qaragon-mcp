# Community word-board contests

Word Links, Fans Say and Fans Feud community variants are ordinary `guess` Contests. They use one prepared answer board and one final answer-list Submission per participant. These are separate from the hosted Games with similarly named content.

## Select prepared content

`GET /v2/answer-boards?kind=fans_say&limit=20` returns board IDs, prompts and answer counts. Supported kinds are `word_links`, `fans_say` and `fans_feud`. Pass the final returned ID as `after` for another page. The catalog never returns accepted answer text. Creation and gameplay perform no AI calls; generation/review is a separate content-preparation operation.

Libraries are platform-owned and shared across tenants. Existing hosted draft preparation and community selection use the same content tables. The release adds curated starter boards without replacing existing library records. Empty selections remain empty; there is no runtime generation fallback.

## Create an ordinary Contest

Use the shared `POST /v2/contests` route with normal Space, title, timing, reward and grouping properties, plus:

```json
{
  "source_type": "guess",
  "board_id": "fans_say:community-kitchen-v1",
  "submission_policy": "final_board",
  "matching_rule": "by_position",
  "clue_strategy": "initial"
}
```

The server resolves and freezes the selected board's prompt, ordered answers and point values. Reusing a board creates independent Contests and results. Use an idempotency key for retryable creation. A missing/unknown board returns a create error; no model is called.

Recommended tenant presets:

| Variant | Clues | Matching |
| --- | --- | --- |
| Word Links | `length` | `by_position` |
| Fans Say | `initial` | `by_position` |
| Fans Feud | `hidden` | `any_order` |

Display positions are explicit zero-based positions. `by_position` grades the corresponding slots. `any_order` matches the answer set and scores each accepted answer at most once. Chain relationships belong to the prepared content; they introduce no live unlocking or host transitions. Clue strategy is independent of matching.

`initial` shows the first character of each word and masks the remaining characters. `length` masks characters while preserving word separators. `hidden` reveals neither letters nor length. Masks are computed server-side and remain stable. The creator-only configuration read remains trusted server-side data; never forward its hidden answer key to a participant app.

## Read and submit

`GET /v2/contests/{id}/guesses` returns the prompt, selected rules, ordered `slots`, `submitted` and `closed`. Each slot has a position and a masked `clue`. Answers become available after board settlement. `submitted` includes a durable accepted Submission even before its queue worker records Participation, so reopening the screen does not offer another entry.

Each participant edits their draft in the app and sends one final list:

```http
POST /v2/submissions
Idempotency-Key: one-final-board-entry

{
  "target": {"type": "contest", "id": "<contest-id>"},
  "payload": {"answers": ["Fridge", "Oven", "Sink", "Kettle", "Plate", "Fork", "Cup"]}
}
```

Send one entry per slot; keep unanswered slots as empty strings. Open-grouping Contests can carry `contest_group_name` in the payload after group selection. Matching ignores case and surrounding whitespace, with no synonyms or semantic judging.

The endpoint returns **202 Accepted**. Its shared worker records the final list in the existing Participation and marks submission processing complete. A processed recording receipt can contain `type: "guess_board", status: "awaiting_results"`; it is not a final grade. Retrying the same key replays the original acknowledgement. Another final entry is rejected, including while the first is pending.

Approved Space membership, normal admission and the shared opening/closing window still apply. There are no per-person timers, per-word requests or server-side draft records. For a guest-capable app, browsing can use the tenant's ordinary public viewer; submission still requires its authenticated participant identity.

## Closure and results

Scheduled closure, or the creator's existing Guess close operation, waits for already-accepted recording work. Guess then grades the final lists, stores a typed breakdown in `Participation.result.graded_sheet`, ranks positive scores and includes cutoff ties under the configured winner count. Accepted-on-time work is not made late by a delayed worker.

Result/scoreboard reads return `409 RESULTS_PENDING` until final settlement. Normal Contest completion and reward/webhook behavior apply. A completed graded sheet has `type: "guess_board"`, total `score`, `possible_points`, and per-input `answers` containing `position`, `submitted`, `correct`, awarded `points` and any corresponding `expected` answer. The settled board read supplies the complete accepted answer set.

Read the participant's normal Contest result or grouped scoreboard after completion. Each person has one Participation, and repeat settlement cannot accumulate points. A recorded zero before grading is not a final loss.

## Existing hosted Guess

`submission_policy: "repeated"` remains the default for existing Guess integrations. It accepts individual `guess` input, returns immediate 200 outcomes, and uses configured claims/attempt caps. The final-board policy reuses its answer model and native settlement; it does not change hosted turns, teams or Game orchestration.

The MCP guide is available as `qaragon://guides/community-guess`. Search the public API for `/answer-boards`, `/contests/{id}/guesses` and `/submissions` for exact schemas. Tenant BFFs keep their API keys server-side and enforce their own user permissions.
