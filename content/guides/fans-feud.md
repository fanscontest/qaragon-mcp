# Fans Feud: Game API

Fans Feud composes two independently playable modules above the Guess primitive:

| Game `type` | Play | Prepared questions | Reward winners |
| --- | --- | --- | --- |
| `fans_feud` | Feud Showdown → Feud Rush | Four main boards, one tiebreak, five Rush questions | The qualified roster if Rush succeeds |
| `feud_showdown` | Competitive match only | Four main boards and one tiebreak | The qualifying team |
| `feud_rush` | Two timed runs only | Five Rush questions | The supplied roster if Rush succeeds |

Each playthrough has one Game ID, one final result and one configured Qaragon reward. Child Contests are managed implementation details and have reward `none`. Teams and individual guess scores do not create additional root rewards. Completion uses the existing `game.completed` event/webhook with winning Platform Identity IDs; fulfillment remains tenant-owned.

Hosts and players are tenant Platform Identities. Create their identities and approve their Space memberships through the existing platform flow. Competitive play uses two equal teams of 2–5 members. Rush takes one roster of 2–5, with two distinct finalists selected from it. The creator hosts and cannot be a player because they preview the answer key. Other approved Space members can read the safe view.

## Prepare and commit content

1. `POST /v2/games/drafts`, body `{"type":"fans_feud","brief":"A game for friends"}`. Briefs contain 1–300 Unicode characters. Feud MVP content is English and suitable for a broad audience.
2. The creator-only response contains `id`, `revision`, `type`, `source`, and the full `boards`, `tiebreak` and `rush` content, including answer IDs/text and points. Read it with `GET /v2/games/drafts/{id}`.
3. Replace one whole question with `POST /v2/games/drafts/{id}/rotate`, body `{"revision":1,"board_index":0}`. Set `all:true` to replace the entire draft. Indices follow main boards, tiebreak, then Rush questions. They are local to the selected definition: full 0–9, Showdown 0–4, Rush 0–4.
4. Commit the approved revision:

```json
{
  "type": "fans_feud",
  "player_space_id": "<space-id>",
  "draft_id": "<approved-draft-id>",
  "draft_revision": 1,
  "teams": [["<a1>", "<a2>"], ["<b1>", "<b2>"]],
  "reward": {"reward_type":"none","value":"0"}
}
```

Use `type:"feud_showdown"` for the main match alone. To create standalone Rush, use a `feud_rush` draft, omit `teams`, and supply `player_identity_ids` plus `finalist_identity_ids` (two distinct members of that roster). Content and rosters are committed at creation; the host cannot edit answers, points or AI rulings during play. Full play selects its finalists from the qualified roster when opening Rush.

Each ordinary question has six distinct canonical answers, positive strictly descending integer points totaling 100. The tiebreak has one answer worth 100. These are generated game weights, not observed survey results or measured popularity.

## Shared routes and safe views

- `POST /v2/games`, `GET /v2/games/{id}` and `GET /v2/games?type=...` handle creation, safe state and cursor discovery. Unfiltered discovery includes all accessible definitions once.
- `POST /v2/games/{id}/actions` handles the host's progression, scoring and reveals.
- **Every player buzz, play/pass choice, guess and Rush pass uses `POST /v2/submissions`.** There is no game-specific participant endpoint.
- `GET /v2/submissions/{id}` and the existing target attempt list expose committed processing status/results to the tenant BFF.

Game views include `type`, committed `definition`, active `stage`, `state`, `viewer_role`, content source, configured reward and eventual result. `showdown` contains the current round, team scores and completed rounds. `rush` contains the current finalist, prompt, clock and revealed run results. The active Stage has its occurrence ID, module ID and rules revision; `action_version` is a separate concurrency token.

Hidden future questions, locked answer keys, managed child Contest IDs, and AI continuations never appear in Game views or public receipts. The host's in-game view reveals answers at the same pace as other viewers. Draft preview remains creator-only.

## Feud Showdown

Four rounds use multipliers 1×, 1×, 2× and 3×. Highest total qualifies. Equal totals require the prepared single-answer buzzer tiebreak; a correct answer qualifies the buzzing team, while a miss/timeout qualifies the opponent. The tiebreak adds no match points.

Host actions require the current `showdown.round.id` as `board_id` and its `action_version`:

- `start`: open the current round's buzzer.
- `open_turn`: open the next actor's 15-second answer window, including after a failed inference is released.
- `expire`: resolve an elapsed answer window as a miss, provided no accepted judgment is pending.
- `next_round`: move from a completed round to the next prepared round, which the host then starts.
- `open_tiebreak`: prepare the tie round after tied match totals; the host then starts it.
- `open_rush`: full Fans Feud only, after qualification. Also supply `finalist_identity_ids` for two distinct qualified members.

Face-off representatives rotate through each roster by round. Only the displayed `faceoff_identity_ids` may buzz. The first accepted buzzer opens that actor's initial 15-second window. A top-ranked answer wins play/pass immediately; otherwise the opposing representative can beat it. If both miss, the next pair faces off. Both accepted face-off answers contribute to the pot.

The winner of the face-off submits the play/pass choice. Control rotates through that team's roster, starting after its face-off representative. A miss, already-revealed answer or timeout adds a strike. Three strikes give the opposing team's first roster member one steal answer. A successful steal adds its value to the pot; a failed steal awards the pot to the controlling team. The round winner receives the pot times its multiplier. Finding all answers awards the pot to the controlling team.

Player examples:

```json
{"target":{"type":"game","id":"<game-id>"},"payload":{"action":"buzz","board_id":"<round-id>","action_version":2}}
```

```json
{"target":{"type":"game","id":"<game-id>"},"payload":{"action":"choose","play":true,"board_id":"<round-id>","action_version":4}}
```

```json
{"target":{"type":"game","id":"<game-id>"},"payload":{"action":"guess","guess":"refrigerator","board_id":"<round-id>","action_version":6}}
```

Use a unique `Idempotency-Key` for each intended action and retain it when retrying that same action. Buzzes and choices return HTTP 200 with a committed receipt. Showdown guesses consistently return **202 Accepted**, including exact matches. The Game displays `judging` until the worker completes. Read the Submission or Game to obtain the committed result.

The server accepts a guess against its actor, board/version and server deadline before external evaluation. AI identifies an equivalent answer on the immutable board or returns no match; Go looks up the stored points, claims the answer, updates Participation and advances the module in one transaction. The original input is retained. A client-supplied answer ID cannot bypass matching.

A timely accepted guess does not become late because AI takes longer. Its pending turn cannot expire or advance. Provider errors are processing failures eligible for worker retry; they are not misses. If retries exhaust, the failed submission releases that same turn for a host-controlled reopening, without a strike or point change. The host cannot override the AI ruling.

Receipts use `result.type:"feud_showdown"` with the action and safe round snapshot. A processed guess also carries the primitive `guess` outcome and individual score. These are separate from the Game's team totals. Retrying the POST preserves the original 202 acknowledgement; a Submission GET reports its current processed/failed status.

## Feud Rush

The host selects two distinct finalists from the supplied/qualified roster. Each answers the same five questions: the first gets 20 seconds and the second 25. A combined 200 points wins the challenge. Both play before the outcome is decided, even if the first run reaches the target.

Host actions use `rush.stage_id` as `board_id` and the current `rush.action_version`:

1. `start_run` opens the active finalist's one continuous countdown.
2. The player submits answers or passes through `/v2/submissions`, using the current `rush.prompt_id` and `action_version`.
3. All answers being recorded, or the server deadline expiring, freezes that run. `expire` can finalize an elapsed run explicitly; the persisted deadline sweep also freezes it.
4. `score_run` makes one batch judgment outside the SQL transaction, then commits only if that frozen run/version still matches. A provider failure leaves it frozen for a retry.
5. `reveal` exposes one scored answer at a time. After five reveals, the next run becomes ready, or the challenge completes after the second run.

Example answer and pass payloads:

```json
{"target":{"type":"game","id":"<game-id>"},"payload":{"action":"guess","guess":"Dog","board_id":"<prompt-id>","action_version":2}}
```

```json
{"target":{"type":"game","id":"<game-id>"},"payload":{"action":"pass","board_id":"<prompt-id>","action_version":3}}
```

Captures return HTTP 200 immediately, with `result.type:"feud_rush"`, an outcome (`recorded`, `passed` or `duplicate`), and the updated safe `game` view. **Recorded does not mean correct.** No model call interrupts answer collection. Passed questions return after the other questions while time remains. Recorded answers cannot be replaced.

For the second finalist, text identical to the first finalist's answer after lowercasing and trimming outer whitespace returns `duplicate`. That prompt remains current and the clock continues; use the new action version for the replacement attempt. Equivalent wording discovered later by batch judging scores zero if it matches the first finalist's canonical answer.

The waiting second finalist's Game view hides the first finalist's active questions, guesses, revealed answers and scores until the second run finishes and is scored. Tenant BFFs must preserve this isolation: the platform authenticates the tenant on attempt reads, so do not forward another finalist's raw Submission history or creator draft previews to them. A display that needs the neutral broadcast view should use a separate approved spectator identity.

Rush success maps to every identity in its committed reward roster, including teammates who were not finalists. No synthetic Submissions or Participations are created for those teammates.

## Content library and AI configuration

Approved generated questions append to a shared Feud PostgreSQL library. Brief classification reuses existing topic tags when they fit and is cached per normalized brief. Question selection prefers matching topics and rotates by use count. Creator-specific generation contributes reusable questions too; the original private brief is not saved in library rows. A library hit needs no generation/review call.

Content administrators can explicitly generate fresh sets:

```http
POST /v2/admin/games/fans-feud/library/generate
Idempotency-Key: <unique-request-key>
Content-Type: application/json

{"type":"fans_feud","brief":"Everyday life","sets":2}
```

The response is 202 with a generation ID. Poll `GET /v2/admin/games/fans-feud/library/generations/{id}` for progress. Each approved set commits independently; a later failure keeps earlier questions. Queued generation has no automatic paid retries. There is no library-generation cron. Completed queue status expires after 24 hours; the approved library does not.

Use the existing `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` configuration. The same selected provider supports Feud generation, review, classification and live judging with separate typed prompts. Both credentials being set selects OpenAI, as with Fans Say. No new Feud-specific provider variables are needed.

Without credentials, development uses labeled fixed sample content and an exact-only SampleJudge. Samples cannot populate the shared approved library. Provider errors never substitute sample content or invented guesses. Live semantic quality and latency require credentialed evaluation before release; offline tests exercise the flow with sample/fake providers.

## Architecture and rollout

`gametype.FansFeud`, `gametype.FeudShowdown` and `gametype.FeudRush` are the public definitions. Deterministic mechanics live in `modules/feudshowdown` and `modules/feudrush`; the containing `fansfeud` package owns content, persistence, AI adaptation, Game operations and final rewards. Modules know nothing about their containing Game or provider transport.

New module-owned Game/draft/library tables are registered in startup migrations. Existing Fans Say Games do not need conversion for this addition. Deploy the API and its submission/deadline/library workers together, and regenerate consumer contracts/SDKs after reviewing the final API. Branda mobile/BFF screens, streaming integrations, disputes, autonomous hosting, configurable variants, model self-hosting and a general YAML composition runner remain outside this MVP.
