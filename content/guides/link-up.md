# Link Up integration guide

Qaragon offers three definitions through the shared Game API:

| `type` | Stages | Reward |
| --- | --- | --- |
| `link_up` | Link Clash → Link Rush | Qualified roster, if Rush succeeds |
| `link_clash` | Four-round competitive match | Winning team |
| `link_rush` | Three-chain cooperative challenge | Supplied roster, if successful |

Link Clash and Link Rush are independently playable modules. The full definition hands the winning team to Rush in its original roster order. Each Game has one completion/result/reward lifecycle. Managed Guess Contests implement the child chains and remain hidden behind the Game API.

## Identity, Space and preparation

The tenant BFF authenticates its users and calls Qaragon server-to-server with its tenant key. Use `X-Acting-As` for the host or current player. Every player needs a Platform Identity and approved membership of the player Space. The host needs authority to create games there and cannot play: they can preview the full answer key.

1. `POST /v2/games/drafts` with `{"type":"link_up","brief":"Everyday expressions for friends"}`. The creator-only response includes the draft ID, revision and complete chains.
2. Optionally `POST /v2/games/drafts/{id}/rotate` with `{"revision":1,"board_index":0}` or `{"revision":1,"all":true}`. Rotation returns a new revision. It is only available before creation.
3. `POST /v2/games` commits that draft, roster, Space and configured reward:

```json
{
  "type": "link_up",
  "player_space_id": "SPACE_ID",
  "draft_id": "DRAFT_ID",
  "draft_revision": 1,
  "teams": [["A", "B"], ["C", "D"]],
  "reward": {"reward_type": "none", "value": "0"}
}
```

For standalone `link_rush`, use `player_identity_ids` with 2–5 identities instead of `teams`. Competitive definitions require two equal teams of 2–5. A submitted draft revision is immutable after creation.

Preparation uses the additive shared content library before generating/reviewing compatible new content. Clash prepares eight chains (four main, three Speed Chains, one reserved tiebreak); Rush prepares three; full Link Up prepares eleven. AI prepares content before play. There is no live inference during guessing. Without a configured AI key, development uses deterministic sample content through the same draft flow. Never expose draft content to players or spectators.

## Shared participant endpoint

Every participant action uses **`POST /v2/submissions`** with a fresh `Idempotency-Key`. Every accepted Link action is processed immediately and returns HTTP 200 with a committed receipt. Its `result.type` is `link_up`; `result.game.type` identifies the selected definition. Host pacing actions use `POST /v2/games/{id}/actions`. For safe host-action retries, send a fresh `Idempotency-Key` for each logical action and reuse it only for the identical request. The host-action header is optional in the current API; Link participant submissions require it.

A player first selects the top or bottom available word:

```json
{
  "target": {"type": "game", "id": "GAME_ID"},
  "payload": {
    "action": "select",
    "chain_id": "CURRENT_CHAIN_ID",
    "action_version": 4,
    "direction": "top"
  }
}
```

The receipt's `result.game` supplies the newly revealed leading letter, committed selected position and new action version. Then submit:

```json
{
  "target": {"type": "game", "id": "GAME_ID"},
  "payload": {
    "action": "guess",
    "chain_id": "CURRENT_CHAIN_ID",
    "action_version": 5,
    "guess": "clip"
  }
}
```

Both actions share the existing Submission/Participation transaction. The Game validates actor, phase, chain, version and authoritative accepted time; Guess claims only a matching selected slot. Matching ignores capitalization and outer whitespace. Synonyms, plurals and typo correction are unsupported. A correct word elsewhere in the chain is a miss. Selection is final for that turn and exposes no further clue on retry. The final letter is never given automatically.

Reuse a request's original key only when retrying that identical request. Replay returns its original projection even after later transitions. Use the latest committed version for each new action. Invalid/stale/late actions fail without advancing the game. Do not send public answer IDs, participant identity overrides or Space overrides in Game submissions.

## Link Clash host and player flow

Read `game.clash.phase`, `chain_id`, `action_version`, `active_identity_id`, `team_scores` and any `deadline`.

| Phase | Next operation |
| --- | --- |
| `ready` | Host `start` |
| `awaiting_turn` | Host `open_turn` |
| `turn` | Active player `select`, then `guess` |
| `speed_ready` | Host `open_speed` |
| `speed` | Finishing player submits `speed` |
| `round_complete` | Host `next_round` |
| `tie_choice` | Designated player submits `choose` |
| `tie_ready` | Host `open_speed` |
| `tie` | Assigned team's designated player submits `speed` |
| Qualified in full Link Up | Host `open_rush` |

Host action example:

```json
{"action":"open_turn","chain_id":"CURRENT_CHAIN_ID","action_version":7}
```

Main chains have 6/6/6/7 words. The first three rounds award 100/200/300 per solved word. A correct guess retains team control; a miss switches teams. Each team's roster rotates after its attempts. The first starting team is randomly committed at creation; the trailing team starts later rounds, with alternating starts on a tie.

Round four's selection also includes `wager`: 100–500 in increments of 100, or up to 1,000 for the final missing word, bounded by the bank. The wager commits before the letter is shown. Correct adds it; miss/timeout deducts it. Timeout before selection deducts 100 and passes control. Reaching zero loses the match, including entering round four with a zero bank.

After each of the first three rounds, the finishing player submits the two middle Speed Chain words **together**:

```json
{
  "target":{"type":"game","id":"GAME_ID"},
  "payload":{"action":"speed","chain_id":"SPEED_CHAIN_ID","action_version":18,"answers":["cream","cheese"]}
}
```

The pair must be correct in order. It has one attempt and awards the round's 100/200/300 bonus, with no partial points or steal. Endpoints and first letters remain hidden until the host opens its 15-second clock. Main turns also have 15 seconds, including selection and guessing. Expiry is automatic; the host can also request `expire` after a deadline.

A tied final bank uses the reserved Speed Chain. The team that started round four chooses `{"action":"choose",...,"play":true}` to attempt it or `play:false` to assign it to the opponent. Its roster's next player submits. Success qualifies that team; failure qualifies the opponent. Banks do not change. This assignment is separate from ordinary guesses, which have no Pass action.

## Link Rush flow

Read `game.rush`: `state`, ordered `player_identity_ids`, `active_identity_id`, `chain_id`, `action_version`, `remaining_milliseconds`, `deadline` and the current chain.

- Host `start_chain` reveals the current endpoints and starts the remaining clock.
- The active player selects a direction, reads the clue, then guesses using the shared endpoint.
- Rotate after every guess, correct or incorrect. Selection does not rotate. Misses retain clues without a strike limit or extra time deduction.
- Completing a chain pauses the clock. The next player's rotation is preserved, while the next chain stays hidden until the host opens it.
- The 4/5/6-word chains share 60 seconds: nine missing words overall.
- Completing all three succeeds. Expiry automatically fails the challenge and reveals only the current chain, never unreached content.

## Reads, completion and operational setup

Use `GET /v2/games/{id}` for safe snapshots and `GET /v2/games?type=link_up` (or a standalone type) for accessible discovery. Receipts contain the same safe Game projection, so a BFF need not reread after every player action. Fanout/polling follows the tenant's existing approach; this feature adds no separate realtime transport.

`GET /v2/submissions/{id}` and attempt lists preserve validated normalized action inputs, including direction, wager and ordered pairs. These reads are tenant-authorized; the BFF must enforce its own user access and must not broadcast raw attempt history. Private content and inference continuations are never normalized participant input.

Tenants subscribe to the public `game.completed` webhook through the existing subscription API. A terminal Game publishes one completion occurrence through the outbox/webhook path. Its result provides winner identity IDs and the configured Qaragon reward. Full Link Up qualification alone issues no reward. Child Contest scores are primitive records; they do not replace the Game's team banks or final winners.

Startup migrates `link_up_games`, `link_up_drafts`, `link_up_library` and `link_up_brief_profiles`. The existing workers register a one-second deadline sweep, recovering persisted clocks after restarts. Deploy API and workers together. Release the matching MCP guide only after Uman with Link Up support (PR #326) is deployed in the target environment; merging the backend PR alone does not make its Game types available. Link uses the platform's existing AI provider/key/model environment; no new model service is required. Credentialed content-quality evaluation remains a release check.

Authorized platform content operators can enqueue additive prefill through `POST /v2/admin/games/link-up/library/generate` with `type`, `brief` and positive `sets`, and poll `/v2/admin/games/link-up/library/generations/{id}`. This is a manual operation, has no scheduled paid generation and never resets the library. Completed sets remain if a later set fails; automatic paid retries are disabled. The current queue's batch timeout still limits very large manual requests.
