# Fans Say: Game API

Fans Say composes Seven Clues and Ten in Sixty. Each is also offered as a standalone Game using the same playable modules and Guess primitive:

| `type` | Play | Draft content | Winning reward roster |
| --- | --- | --- | --- |
| `fans_say` | Competitive main match → cooperative finale | Six main boards, one tiebreak, four finale prompts | The qualifying team, if it clears the finale |
| `seven_clues` — Seven Clues | Competitive match only, including its tiebreak | Six main boards and one tiebreak | The resolved winning team |
| `ten_in_sixty` — Ten in Sixty | Cooperative challenge only | Four prompts requiring 1, 2, 3 and 4 answers | The supplied roster, if it clears the challenge |

Each playthrough has one Game ID, one final result and one configured reward. The definition is committed at creation. Reward funding and fulfillment remain tenant-owned. Generation defaults to explicitly labeled sample content; hosted generation/review can be enabled by environment configuration.

Go integrations use `gametype.FindTheMinority`, `gametype.FansSay`, `gametype.SevenClues` and `gametype.TenInSixty` from `experiences/games/gametype`. The two reusable mechanics live in `modules/sevenclues` and `modules/teninsixty`. The API accepts only the published type IDs above; discarded working names have no aliases.

Hosts and players are tenant Platform Identities. The host creates the Game in an authorized Space and remains outside the roster. Competitive play requires two equal teams of 2–5 approved members. Finale-only play requires one ordered roster of 2–5 approved members. Other approved members may watch the safe view but cannot submit player actions. Every player action uses `POST /v2/submissions`.

## Approve and create

Draft briefs are trimmed and must contain 1–300 Unicode characters. Language is a valid language/locale code, such as `en` or `en-NG`; it defaults to `en` and is normalized before generation. Invalid preferences return `400 INVALID_GAME_REQUEST` before AI calls.

1. `POST /v2/games/drafts` with `{"type":"fans_say","brief":"A game for friends"}` returns a creator-only draft with `id`, `revision`, `source`, `boards` (six ordered prompts with seven answers each), `tiebreak` (one prompt with one accepted answer), and `finale` (four prompts requiring 1, 2, 3 and 4 answers). Board indices 0–1 belong to round 1, 2–3 to round 2, and 4–5 to round 3. Each team receives a separate board in each round.
2. Read the complete preview with `GET /v2/games/drafts/{id}`. Replace one prompt and all its answers using `POST /v2/games/drafts/{id}/rotate`, body `{"revision":1,"board_index":3}`. Indices 0–5 select main boards, index 6 the tiebreak, and indices 7–10 the four finale prompts. Rotation replaces only the selected prompt and its complete answer set. Other boards stay unchanged. Rotation checks the reviewed revision again after generation, which runs outside database locks.
3. Commit the approved revision with `POST /v2/games`:

   ```json
   {
     "type": "fans_say",
     "player_space_id": "<space-id>",
     "draft_id": "<draft-id>",
     "draft_revision": 2,
     "reward": {"reward_type":"money","value":"100","currency":"NGN"},
     "teams": [["<a1>", "<a2>"], ["<b1>", "<b2>"]]
   }
   ```

Creation atomically locks all six boards, the tiebreak and finale content, and the roster, consumes the draft and creates the first board's managed Guess Contest. Changed or consumed revisions return 409. Future boards are stored privately, and each receives its own managed Guess Contest when the host advances to it. No content or roster editing is available after creation. `content_source` identifies sample content in Game reads.

## Standalone creation and views

To host Seven Clues, use `type:"seven_clues"` in both the draft request and Game creation. Supply `teams` exactly as in the full-match example. The Game completes automatically when main scores produce a winner, or when its tiebreak resolves. It creates no finale Contests and does not accept `open_finale`.

To host Ten in Sixty, first request a `ten_in_sixty` draft, then create:

```json
{
  "type":"ten_in_sixty",
  "player_space_id":"<space-id>",
  "draft_id":"<finale-draft-id>",
  "draft_revision":1,
  "player_identity_ids":["<player-1>","<player-2>"]
}
```

Do not supply `teams` for finale-only play or `player_identity_ids` for competitive play. Creation rejects a draft belonging to another definition or revision. Finale-only creation prepares its four managed Guess Contests and returns the challenge in `ready` state. The host calls `start_finale_prompt` using the returned `finale.current_prompt_id` and `finale.action_version`; no main match or `open_finale` action is required.

Draft rotation indices enumerate only included questions: full match uses 0–10, main-only uses 0–6, finale-only uses 0–3. `all:true` regenerates only that definition's content. Library selection, generation schemas, review and correction use the same content plan.

Game reads include `type`, `definition:{"id":"seven_clues","version":1}` and the active `stage` with its module ID/rules version. Full and main-only views expose competitive board/team fields. Finale-only views expose `player_identity_ids` and `finale`, with no fabricated `board`, opposing team or finalist index. `finale.team` is present only for a full match's qualified team. `action_version` is the current turn's concurrency token, separate from the committed definition/module revisions.

Use `GET /v2/games?type=seven_clues` or `type=ten_in_sixty` to filter discovery. An unfiltered list includes all accessible offerings once. Owner/player/audience roles and tenant isolation follow the same rules as full Fans Say.

## Shared question library and AI workflow

All accepted live-generated questions and their complete answer boards enter Qaragon's shared PostgreSQL library, including those generated for a creator's brief. Library rows contain question content, suitability metadata, provenance and usage, with no tenant/creator identity or original brief. Game content remains an immutable snapshot; changing library inventory cannot change an existing Game. Development sample content is never imported.

For a new brief, AI interprets topics, the youngest audience age and cultural context. Language and difficulty remain application-controlled preferences (`en`/`medium` by default). The classifier receives existing category names and is asked to reuse them when appropriate, or propose a new concise category if none fits. Names are normalized for casing/spacing and deduplicated; semantic synonyms are model decisions, not guaranteed by normalization. The interpretation is cached under a digest of normalized brief/language/difficulty. A repeat brief avoids that classification call; different briefs with similar tags can reuse the same questions. Only category names are supplied from inventory, not other creators' briefs or questions.

Generated questions include a `profile` with `topics`, `language`, `difficulty`, `minimum_age` and `culture`. Code filters library candidates by language, difficulty, age compatibility and general/matching cultural context, then ranks by topic overlap and prior use. Selection keeps the highest-ranked version of each distinct prompt before applying the slot limit. Individual questions are reviewed and tagged when generated. Matching is an approximation of a free-text brief; the creator still reviews the assembled draft before approval. No Elasticsearch or embedding dependency is required.

A full library hit needs no generation/review calls. A previously unseen brief normally adds one classification call. With insufficient inventory, one call generates a complete draft of eleven boards and one call reviews it. Reusable selections are retained and missing slots are filled from the new draft; every accepted new board is stored, including surplus content. Code validates the shape, answer counts, distinctness, metadata and question uniqueness before editorial review. Rejected boards receive targeted corrections, followed by whole-draft review, with two correction rounds maximum. A malformed overall shape can require regenerating the full draft. Provider failures, refusal or exhausted review prevent saving a partial draft.

Single-question rotation selects a suitable library board first; a miss uses one generation and one review call. Whole-draft regeneration and rotation retain rejected-question history so future replacements do not return previously rejected questions. Failure retains the existing draft and revision. Usage counts change atomically with saving the draft/replacement; reuse rotates inventory but does not consume it. There is no guarantee of a new question forever when inventory is finite. No membership restrictions are implemented.

`Generator` declares classification, complete-draft generation and single-board repair/rotation; `Reviewer` reviews a board or complete draft. Runtime wiring uses the same selected client/model for both. Starting-letter clues are derived from canonical answers by game code. AI review checks relevance, clarity, culture/age suitability, difficulty, metadata, near duplicates and unsupported survey claims; it is an opinion, not a correctness guarantee.

The normal request deadlines remain five minutes for draft operations. Models use a 4,096-token output cap for individual boards/reviews and an 8,192-token cap for complete-draft generation. Samples are fixed English development data and cannot infer categories or honor preferences semantically. Live content/model quality awaits credentialed evaluation.

## Content-team library generation

Library generation is a platform-admin operation, separate from creator drafts and participant submissions. It uses the existing administrator authorization and Asynq worker; nothing generates library content on a cron schedule.

```http
POST /v2/admin/games/fans-say/library/generate
Content-Type: application/json
Idempotency-Key: <unique-request-key>

{"brief":"Everyday life for Nigerian adults","language":"en-NG","difficulty":"medium","sets":2}
```

`sets` is a positive integer. Briefs share the creator-draft validation (1–300 characters), language parsing and difficulty values; omitted language/difficulty default to `en`/`medium`. The request returns `202 Accepted` with an ID, `status: queued`, `requested_sets` and zero initial progress. Poll:

```http
GET /v2/admin/games/fans-say/library/generations/{id}
```

The response reports `queued`, `running`, `completed` or `failed`, along with `generated_sets` and `generated_questions`. The optional `type` selects the set shape: `fans_say` (default) has eleven questions, `seven_clues` has seven, and `ten_in_sixty` has four. A two-set full-match request generates twenty-two questions regardless of existing inventory size. Each set goes through the same classification, generation, validation and AI review as creator drafts, excluding suitable existing library prompts. Development samples cannot enter the library.

Each approved set commits independently, then progress is recorded in the queue. If a later set fails, earlier sets remain reusable and the last reported progress remains available. Failed status includes the safe `CONTENT_GENERATION_FAILED` code; provider details stay out of the response. Infrastructure failures between saving a set and reporting progress can leave the displayed counts behind the saved content. Queued work uses the existing worker execution timeout, with a five-minute model-work budget per set. No automatic paid job retries run; submit another request deliberately if more content is needed.

An identical POST with the same `Idempotency-Key` replays the acknowledgement through the existing middleware; a changed request with that key is rejected. Completed queue status/results are retained for 24 hours, while library questions remain in PostgreSQL. This is operational status, not a permanent generation audit ledger. There is no new database table or participant submission path.

`FANS_SAY_LIBRARY_SCHEDULE` and `FANS_SAY_LIBRARY_PROFILES` are removed; only the existing AI credentials are needed. These are ordinary provider calls, not discounted provider Batch APIs. Platform maintainers can use `uman/docs/games/fans-say-ai-evaluation.md` in the platform repository for manual model comparison; that operator document is not a bundled MCP resource. No live model calls are needed to run the tests.

## Host controls and progression

Use `POST /v2/games/{id}/actions`. Every Fans Say host action requires the current `board_id`; a missing or stale ID returns 409. Use the existing command middleware's `Idempotency-Key` when retrying requests.

Start the first board with:

```json
{"action":"start","board_id":"<current-board-id>","team":0}
```

Team indices are 0 and 1. The host selects the first team once; that team starts rounds 1 and 2. The other team receives the second board of each round. The leader after round 2 starts round 3; if scores tie, the server selects and persists a random starting team. Subsequent `start` actions use that schedule; the `team` field no longer selects the team.

The first roster member starts their team's round-1 board, the second starts round 2, and the third starts round 3, wrapping to the first for two-player teams. Membership is rechecked before each board starts. The shared 30-second clock starts when the host starts the board.

After a board completes, advance with:

```json
{"action":"next_board","board_id":"<completed-board-id>"}
```

This atomically archives the completed board, prepares the next approved board and creates its managed Guess Contest. The new board is `ready`, with a new board ID. The host starts it separately, using its new ID. You cannot advance during timed play, expiration awaiting a steal, or stealing; nor can you advance past board 6. Delayed/concurrent advances cannot skip a board. If next-Contest creation fails, the board index, history and child records remain unchanged.

For full `fans_say`, the host plays all six boards. At the end, the Game becomes `awaiting_finale` with `finalist_team:0|1`, or `awaiting_tiebreak` with no finalist. These states do not award rewards or accept another main-board start. The host may open a tiebreak only in `awaiting_tiebreak`. A resolved main-game winner can enter the finale.

## Shared guess submissions

All player guesses use the shared endpoint:

```http
POST /v2/submissions
Idempotency-Key: unique-player-action

{
  "target": {"type":"game", "id":"<game-id>"},
  "payload": {
    "board_id":"<current-board-id>",
    "action_version":1,
    "guess":"Fridge"
  }
}
```

The Game and each board have distinct IDs. Clients use the current board ID/version rather than deriving them. Clients cannot select the managed Contest or override the participation Space.

HTTP 200 returns a committed receipt with `result.type:"seven_clues_guess"`. `result.guess` is the primitive outcome, answer points and cumulative individual score for this board's Contest. `result.board` is the safe board snapshot after that action. Its score/team totals are board-local; the Game read's top-level `team_scores` are match-wide.

| Round | Correct answer | Timed clear bonus | Seven-answer timed clear |
|---|---:|---:|---:|
| 1 | 100 | 1,000 | 1,700 |
| 2 | 200 | 2,000 | 3,400 |
| 3 | 300 | 3,000 | 5,100 |

- Exact matches ignore case and outer whitespace. New matches reveal and claim an answer once across that board's Contest. The same player keeps the turn.
- A miss or already-claimed answer awards zero. During timed play it rotates to the next teammate without changing the deadline; during stealing it ends the board.
- Guess owns matching, claims and individual answer points. Fans Say owns the round-dependent configuration, turns, team totals and timed-clear bonus. Bonuses are never individual Guess awards.
- Every accepted guess advances the action version. Blank, stale, out-of-turn, late or unapproved-member input is rejected without recording an attempt.
- An accepted request replay returns its original receipt, even after the host advances to another board or the main match ends. A new request against an old board is rejected.

## Stealing and reveals

After the 30-second main clock expires with answers left, reads show `board.state:"expired"`. The host may persist expiration with `{"action":"expire","board_id":"<current-board-id>"}`. Expiration does not settle the Contest or permit advancement: its remaining claims are needed for stealing.

The host opens the opposing team's steal:

```json
{"action":"open_steal","board_id":"<current-board-id>","submitter_identity_id":"<opposing-player-id>"}
```

The nominee must belong to the opposing locked roster and still have approved Space access. The host cannot open stealing early, change the nominee after opening, or reopen a completed board. The response has `board.state:"stealing"`, the nominated `active_identity_id` and a new version. `deadline` is omitted: stealing is untimed. Teammates may discuss, but only the nominee submits guesses.

Correct guesses continue until all remaining answers are found; the first miss or already-claimed answer ends stealing. The original team retains its points. Steals award the current round's answer value with no clear bonus.

Completion reveals all seven answers. An answer's `found` flag is true only when a player claimed it; display-only final reveals do not create claims or award points. Before completion, unfound answers remain letter clues.

## Reads and discovery

`GET /v2/games/{id}` returns:

- `state`: current match state; completed intermediate boards use `awaiting_next_board`.
- `round`: 1–3, and `board_index`: 0–5.
- `team_scores`: accumulated match totals, including the current board without counting completed history twice.
- `board`: current safe board, with its own score, team totals, ID, state, action version, active player and optional deadline.
- `completed_boards`: ordered safe snapshots of completed boards. Internal child Contest references and future content are excluded.
- `finalist_team`: present only after all six boards finish with a clear leading team.
- `last_guess`: the previous committed guess for the current board, cleared when the host prepares the next board.

Host, players and approved spectators receive the same safe board. Full future answers are available only through the host's private draft. Display a countdown only during timed play.

`GET /v2/games` lists hosted/joined Games; `?type=fans_say` or `?type=find_the_minority` filters the cursor collection. Summaries report match state rather than treating an intermediate completed board as a completed Game. Find the Minority's API remains unchanged.

`GET /v2/submissions/{id}` exposes the accepted Game target and typed result without managed IDs or answer keys. Generic Contest endpoints reject managed children, including previous boards' Contests.

## Sudden-death tiebreak

When all six main boards finish tied, the host opens the approved tiebreak using the final main board ID:

```json
{"action":"open_tiebreak","board_id":"<last-main-board-id>"}
```

The Game becomes `state:"tiebreak"`. Its `tiebreak` object contains its own `id`, prompt, `state:"buzzing"` and `action_version:1`. It does not expose the answer or a running clock. Only the first roster member of each team may buzz. The host remains outside both teams.

Buzzer presses use the shared endpoint and require an idempotency key:

```http
POST /v2/submissions
Idempotency-Key: unique-buzzer-action

{
  "target":{"type":"game","id":"<game-id>"},
  "payload":{"board_id":"<tiebreak-id>","action":"buzz","action_version":1}
}
```

The first eligible buzzer accepted under the Game lock wins. Competing, stale or ineligible buzzes return 409 without recording an attempt. The buzzer receipt is HTTP 200 with `result.type:"seven_clues_tiebreak"`, `action:"buzz"` and the safe `tiebreak` snapshot. It uses the shared durable Submission and Participation writer, but has no Guess result, claim or point award. It is a control action, not a fabricated word guess.

The tiebreak becomes `answering`, with the winning captain's `active_identity_id`, a new action version and `deadline` equal to buzzer acceptance time plus 15 seconds. The winner submits one typed guess through `/v2/submissions`, using that tiebreak ID, version, `action:"guess"` and `guess:"..."`. Guess supplies exact case/outer-whitespace-insensitive matching against the locked single answer. A correct answer sends their team to the finale; a miss sends the opponent. No second answer or rebuzz is allowed.

The answer receipt has the same `seven_clues_tiebreak` result type, `action:"guess"`, the typed primitive `guess` result and the resolved `tiebreak`. The main-match score remains tied: tiebreak matching uses one internal primitive point to distinguish a correct claim, without adding main-game points or rewards.

If no answer is accepted before the deadline, it counts as a miss. Qaragon automatically finalizes this through a recurring deadline worker, without another host or player request. The database retains the deadline; a restart does not lose it. Guesses are rejected at the deadline even if queue execution is delayed. During that delay, reads show `tiebreak.state:"expired"`; worker settlement then changes the Game to `awaiting_finale` with the opposing `finalist_team`. Duplicate sweeps are harmless, and an answer already accepted on time is allowed to finish atomically before expiry checks its committed state.

Once resolved, the single answer is revealed and `found` records whether the captain actually matched it. `last_tiebreak_action` retains the previous committed buzzer/guess result. Timeouts do not create guess submissions. Original receipts replay unchanged after any resolution. Matching, claim, Participation result, receipt, Game outcome and native child settlement roll back together on failure; timeout settlement is also transactional and retried from persisted due state.

The scheduler registers `fans-say:deadlines` every second using the existing Asynq infrastructure. It examines up to 100 due games per invocation, scopes each mutation to its owning tenant and continues other games if one settlement fails. Deadline indexes, partitioning and larger-scale scheduling/fanout are deferred; the hard submission deadline does not depend on poll latency.

## Bonus finale

After a clear main-game winner or resolved tiebreak, the host opens the finale using the final main board ID:

```json
{"action":"open_finale","board_id":"<last-main-board-id>"}
```

Opening atomically prepares four managed Guess Contests, one for each locked finale prompt. If any child cannot be created, none of them is committed and the Game stays `awaiting_finale`. The response has `state:"finale"` and a `finale` object with `team`, `state:"ready"`, a new `action_version`, `current_prompt_id`, `prompt_index:0`, `remaining_ms:60000` and no running deadline. Its `prompts` list is initially empty: no finale question or clue is exposed until the host starts it.

Start the current prompt with:

```json
{"action":"start_finale_prompt","board_id":"<current-prompt-id>","action_version":1}
```

Only the host may start prompts. The server rechecks finalist membership and the current ID/version. The first roster member starts the finale. Its state becomes `playing`, the active identity is exposed, and `deadline` is the start time plus the remaining active-play budget. Only this prompt is added to the public `prompts` list. The host cannot reset the budget or select another player.

Finalists guess through the unchanged `/v2/submissions` envelope, using the current prompt ID, finale action version, `action:"guess"` and their typed `guess`. Correct guesses keep the turn; a miss or already-claimed answer passes it to the next teammate. Every accepted action advances the shared finale version. The clock keeps running within a prompt; misses and correct guesses do not restart it.

A prompt is complete when all its required answers are found. Its managed child settles normally with reward `none`; the finale pauses, stores the exact remaining budget and passes the turn to the next teammate. The next prompt remains hidden until the host starts it. All four prompts share the original 60 seconds, irrespective of how long the host waits between them.

### One skip

The active finalist may submit:

```json
{
  "target":{"type":"game","id":"<game-id>"},
  "payload":{"board_id":"<current-prompt-id>","action":"skip","action_version":4}
}
```

An idempotency key is required. Only one skip is available across the entire finale. It moves the current prompt to the back of the unfinished queue, preserves its found answers and native claims, passes the turn and pauses the clock until the host starts the next prompt. The returned `skip_available` flag is false once used or when no other unfinished prompt exists. A skip cannot include a guess. It records a neutral shared Submission/Participation update with no claim or extra points.

After the intervening prompts finish, the skipped prompt returns with the same ID, preserved claims and remaining budget. Its action version is newer, so old guesses or host starts cannot operate on it. Replaying an already accepted skip still returns its original receipt.

### Results, privacy and automatic expiry

Finale receipts use `result.type:"ten_in_sixty"`, `action:"guess"|"skip"`, the originating `prompt_id`, a safe `finale` snapshot, and a primitive `guess` result only for guesses. `prompt_id` identifies where the action happened even if its result advances to another prompt. The Game read exposes the current `finale` and `last_finale_action`; the current view is authoritative when the last receipt predates a host transition or timeout.

The finale view also identifies the playable Stage: `"stage":{"id":"finale","module":{"id":"ten_in_sixty","version":1}}`. `action_version` remains the concurrency token for the current turn; the module version identifies the rules. The full-match `team` field retains the qualifying team's original index. Its ordered roster is committed at handoff. Internal Contest IDs, unrevealed answers and private snapshots remain excluded from the view.

Finding all 10 answers in time produces `finale.state:"won"`, Game `state:"complete"` and `winner_team` equal to the finalist team. Finale Guess claims use one primitive point per answer for native grading; these do not change main-game totals. The Game records the configured reward and every winning roster identity at completion; actual fulfillment belongs to the tenant.

If time expires during an open prompt, late guesses and skips are rejected immediately. Reads temporarily project `finale.state:"expired"` while the durable deadline worker finalizes `lost`, closes all unfinished children and completes the Game with no `winner_team`. No fake guess is written for a timeout. The same recurring worker handles both tiebreak and finale deadlines; paused finales are not due. A guess accepted on time is allowed to commit atomically even if processing finishes after the clock.

On success or timeout, answers are revealed only for prompts the host already opened. Unplayed prompts and their answers stay hidden. `found` distinguishes actual claims from end-of-game reveals, which award no points. Completed and partially answered skipped prompts remain visible as played history; internal Contest IDs and unopened content are never part of the public Game or receipt.

All guesses, skips, claims, Participation results, terminal native settlement and Game updates remain within the shared submission transaction. Concurrent guess/skip requests for one version accept at most one action. Failed settlement rolls the whole action back. Automatic expiry is likewise transactional, including closure of multiple unfinished children, and failed games remain due for retry. Larger-scale deadline indexing, fanout and scheduling remain deferred.

## Reward configuration and completed results

`POST /v2/games` accepts the existing Contest `reward` model: `reward_type`, `value`, optional `currency` and `redemption_condition`. Supported types are `none`, `money`, `data`, `airtime`, `surprise` and `other`. Omitting the reward uses `{"reward_type":"none","value":"0"}`. The configured reward is fixed at creation and appears in the Game read before play; it cannot be changed through host actions.

For full Fans Say or finale-only play, a finale win earns the full configured reward for **each member of its locked cooperative roster**. For main-only play, resolving the main winner earns that reward for every member of the winning team. It is not a team pot to divide. For example, a reward value of 100 with five winning members identifies five eligible winners at that value. All roster members qualify, including someone who never submitted a guess. Native child Contests retain reward `none`; the Game does not fabricate guesses or Participation rows to make non-guessing teammates look like primitive winners.

After either success or timeout, `GET /v2/games/{id}` includes an immutable `result`:

```json
{
  "outcome":"won",
  "winner_team":0,
  "winner_platform_identity_ids":["<a1>","<a2>"],
  "reward":{"reward_type":"money","value":"100","currency":"NGN","redemption_condition":""},
  "completed_at":"2026-10-29T12:00:00Z"
}
```

A failed finale has `outcome:"lost"`, an empty winner list and no `winner_team`. The reward remains the configured metadata; an empty list means nobody earned it. Type `none` still permits a game win and winner list, with no fulfillment owed. Before finale completion, `result` is absent, including after main-game victory and tiebreak resolution.

The result and one `uman.game.completed` outbox record commit atomically with the terminal shared submission or automatic timeout. Failure to persist the event rolls back the terminal claim, Participation/result, native child completion and Game outcome. Retries and accepted submission replays do not create another result or completion occurrence.

Tenants subscribe to public `game.completed` through the existing webhook subscription API. Its body includes `game_id`, `game_type`, `outcome`, optional `winner_team`, `winner_platform_identity_ids`, the configured `reward` and `completed_at`. The envelope supplies `tenant_id`; the body carries no payment details or managed child identifiers. The stable occurrence is `game.completed:<game-id>`. Delivery signatures, occurrence deduplication and HTTP retries use the existing webhook infrastructure; see [the webhook guide](qaragon://guides/webhooks). Use the MCP tools `list_webhook_events` and `describe_webhook_event` for the published event catalog, exact schemas and examples.

Qaragon supplies authoritative eligibility, not a transfer receipt. The tenant must validate/reserve funding when offering the reward, map its own Platform Identity IDs to fulfillment details, apply redemption conditions and perform transfers. Deduplicate fulfillment by Game ID and winner identity. No BFF/mobile payment behavior is changed by this backend slice, and existing SDKs should be regenerated after the Game contract stabilizes.

## Transactions and cutover

Immediate guesses follow the existing request identity → Game → Contest lock order. Shared admission samples acceptance time after locking. Guess grading, Participation, Submission receipt, board/history updates and native child settlement commit together. Failures roll them back together. Host advancement holds the same Game lock while creating the next Contest, configuration and hosting, so it cannot race an unfinished terminal guess or create duplicate next boards.

There is no new submission endpoint or Participation writer. Each completed child settles normally with reward `none`; main-game victory alone does not qualify anyone for the configured match reward. Tenant fulfillment integration, streaming and scaling remain following work.

The earlier single-board prototype required a development schema cutover: old rows did not contain approved content for the full match. That historical reset instruction does not apply to the playable-module extraction. Existing full matches are preserved.

For this release, stop the old Uman API and workers, back up `fans_say_games` and `fans_say_drafts`, run `uman/docs/rollouts/2026-11-fans-say-stage-cutover.sql` from the platform repository in one transaction, then start the new release. It pins existing full Games and drafts to `fans_say` revision 1, moves the main board/children/history/tiebreak into the main module's snapshot, and commits the finale roster/content/reference. The former `match` JSON column becomes `state`, reflecting that the same Game model also hosts a standalone cooperative challenge. The obsolete top-level board/Contest/receipt columns are dropped after their data is moved.

The script preserves clocks, prompt/Contest IDs, claims, Submissions and replay receipts, history, reward results and outbox occurrences. It also supplies stage metadata for the saved last finale action. Existing accepted work resumes against the same child IDs and action versions. The deadline sweep reads the new main-module path; paused finales retain their budget. The cutover is idempotent and does not reset a database or emit completion events. Unsupported definition/module revisions fail visibly instead of silently upgrading running Games.

## Assembled tenant journey

Run `go test ./journeys -run TestFansSay -count=1` to exercise the backend integration from a tenant's perspective. The journey registers identities, creates a public Space and joins both teams, previews and rotates sample content, creates a rewarded Game, plays all six boards with steals, resolves a tied match, and completes the finale after skipping and returning to a partially answered prompt. Every player action uses `/v2/submissions`. It also checks terminal receipt replay, completed-game discovery for host and players, and exclusion of an unrelated identity.

`TestFansSayStandaloneTenantJourneys` also covers main-only and finale-only creation, local draft rotation, Space membership, complete play, rewards, discovery and terminal receipt replay. The full-match journey feeds the actual committed completion outbox payload through the production webhook bridge, dispatcher, persistence and signed HTTP delivery. A local receiver rejects its first delivery, then accepts the retry; assertions cover signature verification, unchanged retry payload, event deduplication and tenant isolation. This confirms reward eligibility and notification, not payment execution.

The HTTP server starts at the trusted gateway header boundary; the separate API-key verifier and live Kafka/Asynq transport are not exercised. The test advances a main-board deadline directly to avoid waiting through discussion windows; all business transitions go through the API. Content generation remains explicitly sample data. Mobile/BFF flows and a real AI provider are separate follow-up work.

## Hosted AI configuration

After deploying the hosted-provider code and its infrastructure environment forwarding, set `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` in the desired Infisical environment and redeploy Uman. These are platform-wide credentials; no per-game provider or model variables are needed. An OpenAI key selects `gpt-6-luna`; otherwise an Anthropic key selects `claude-haiku-5-5`. OpenAI takes precedence when both keys are present. Generation and review use the same selected client and model; model selection is defined in code. No keys belong in requests or repository files.

When neither API key is set, development samples remain available when library inventory is unavailable; suitable cached library content is still reusable. Once credentials select a live provider, upstream errors fail with `503 CONTENT_GENERATION_UNAVAILABLE`; they do not switch provider or fall back to sample content. Review failure, refusal, invalid/truncated model responses and upstream errors likewise prevent draft persistence. Each call is capped at 30 seconds and 4,096 output tokens (8,192 for whole-draft generation) with no automatic HTTP retries; editorial corrections retain the existing two-attempt budget. Generation and rotation HTTP routes have a shared five-minute total context/write budget; other API routes keep their normal server timeout. Requests cancelled by callers abort provider work. Proxies/BFFs must also allow a sufficiently long response wait for these two routes.

Uman logs usage metadata only: provider/model, operation, prompt version, token counts, elapsed time and failure status. Briefs, boards, upstream error bodies and API keys are excluded. OpenAI calls use `store:false`; provider account-level retention policies still apply. Source metadata identifies the generation provider/model. Reviewer model and token usage are recorded per call.

The deployment must recreate Uman after changing Infisical values; editing secrets does not change a running container. Remove both API keys and redeploy to return future drafts to development samples. To use Anthropic when both keys are configured, remove the OpenAI key and redeploy. Existing Game content remains locked. Membership-based routing, generation quotas and throughput optimization are deferred. Platform maintainers can find the credential-ready comparison command in `uman/docs/games/fans-say-ai-evaluation.md` in the platform repository; no live models have been evaluated yet.

Provider implementations live in `platform/ai`, behind `Provider.Complete(ctx, Request)`. OpenAI and Anthropic own their request formats and response parsing in separate files, sharing HTTP transport and normalized token usage. The request carries instructions, JSON input and a result schema; it has no dependency on Games or Fans Say. `experiences/games/fanssay/ai/client.go` receives that interface through its constructor and owns Fans Say prompts, typed result validation and operation-level usage reporting. Its existing `Generator` and `Reviewer` contracts remain unchanged. Future AI features can use the same provider interface with their own prompts and schemas.

The application constructs the configured generator once at startup and injects it into Fans Say’s `DraftService`, which owns drafting, classification and question-library operations. The gameplay `Service`, shared contest factory and submission transaction wiring have no content-generation dependencies. Startup handles initialization errors directly; it does not install an always-failing generator. Draft handlers and the manually requested library worker reuse that instance; transaction-bound gameplay services do not receive it. Runtime provider failures still return `503 CONTENT_GENERATION_UNAVAILABLE`, while requests for other platform features continue normally. Provider construction makes no remote API calls.
