# Community word-board contests

Word Chain, Word Reveal and Fans Feud community variants are ordinary `guess` Contests. They use one prepared answer board and one final answer-list Submission per participant. Word Reveal is the self-play preset; Fans Say (`type=fans_say`) remains the hosted Game. The two reuse prepared question content without sharing Contest or Game state. Hosted Word Chain uses the Game API; its community counterpart uses an ordinary Guess Contest. Select the play style explicitly when requesting integration guidance.

## Select prepared content

`GET /v2/answer-boards?kind=word_reveal&limit=20` returns board IDs, prompts and answer counts. Supported kinds are `word_chain`, `word_reveal` and `fans_feud`. Pass the final returned ID as `after` for another page. The catalog never returns accepted answer text. Creation and gameplay perform no AI calls; generation/review is a separate content-preparation operation.

Libraries are platform-owned and shared across tenants. Existing hosted draft preparation and community selection use the same content tables. The release adds curated starter boards without replacing existing library records. Empty selections remain empty; there is no runtime generation fallback.

Contest details and lists include the persisted `source_variant` (`word_chain`, `word_reveal` or `fans_feud`), alongside `source_type: "guess"`. Feed contest blocks retain `contest_type: "guess"` and use the same `source_variant` field. Use that value for the user-facing experience name. Guess determines it from the selected board during creation; clients may supply the matching value, but a conflict is rejected. Reads do not query Guess configuration or infer a variant from a board reference. A custom Guess contest without a prepared board has no named variant and should use a generic label.

## Create an ordinary Contest

Use the shared `POST /v2/contests` route with normal Space, title, timing, reward and grouping properties, plus:

```json
{
  "source_type": "guess",
  "source_variant": "word_reveal",
  "board_id": "word_reveal:community-kitchen-v1",
  "submission_policy": "final_board",
  "matching_rule": "by_position",
  "clue_strategy": "initial_random"
}
```

The server resolves and freezes the selected board's prompt, ordered answers and point values. Reusing a board creates independent Contests and results. Use an idempotency key for retryable creation. A missing/unknown board returns a create error; no model is called.

Recommended tenant presets:

| Variant | Clues | Matching |
| --- | --- | --- |
| Word Chain | `initial_random` | `by_position` |
| Word Reveal | `initial_random` | `by_position` |
| Fans Feud | `hidden` | `any_order` |

Display positions are explicit zero-based positions. `by_position` grades the corresponding slots. `any_order` matches the answer set and scores each accepted answer at most once. Chain relationships belong to the prepared content; they introduce no live unlocking or host transitions. Clue strategy is independent of matching.

`initial_random` shows the first character of each word plus one randomly selected remaining position when that word has more than three characters. Words of three or fewer characters show only their initial. Spaces, hyphens and apostrophes keep their separator behavior. The extra position can be anywhere after the initial, including the last position. For example, a Contest may show `FRIDGE` as `F___G_` and `EGG` as `E__`.

Qaragon generates each random mask once during Contest creation and saves its `MaskedText` alongside that Contest's private answer. All participants and repeated reads use the saved mask; a refresh, a different server replica or a later masking-code change does not select another letter. The shared question library keeps complete reusable boards and does not store Contest-specific masks. A new Contest using the same board gets its own selection. Clients receive only the existing slot `clue` and must keep all given letters fixed while editing; they submit complete answer text, including those letters.

The existing `initial` strategy shows only the first character of each word, and `length` masks characters while preserving word separators. `hidden` reveals neither letters nor length. Those strategies retain their existing deterministic projection. Existing Contests keep their saved strategy and are not upgraded or backfilled. The creator-only configuration read remains trusted server-side data; never forward its hidden answer key to a participant app.

## Read and submit

`GET /v2/contests/{id}/guesses` returns the prompt, selected rules, ordered `slots`, `submitted` and `closed`. Use the ordinary Contest response's `source_variant` for the experience label; this rules endpoint has no separate type-identification field. Each slot has a position and a masked `clue`. Answers become available after board settlement. `submitted` includes a durable accepted Submission even before its queue worker records Participation, so reopening the screen does not offer another entry.

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
