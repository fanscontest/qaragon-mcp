# Find the Minority: tenant integration guide

Find the Minority is a Qaragon-owned game API for a host-led social deduction game. A tenant BFF owns its app experience, end-user authentication, identity creation, Space membership, and any streaming-platform integration. Qaragon owns game state, phases, ballots, role assignment, scoring, and reveal data.

The game works without an audience. Audience voting is optional and uses a separate Space. The API does not connect to TikTok, YouTube, Twitch, or read their chat or gifts.

Use the public OpenAPI contract for exact request and response schemas. This guide explains the workflow. Do not access the Game's internal Vote contests through Contest endpoints.

## Caller identity and access

Authenticate BFF requests using the tenant's Qaragon API key. Send `X-Acting-As: <platform-identity-id>` for the Platform Identity acting on that request. Qaragon verifies the identity belongs to the tenant and forwards it to the service as `X-Platform-Identity-ID`; tenants should use the public acting-as header, not the internal forwarded header. The BFF must authenticate its own user and authorize that user to act for the selected identity. Never send the tenant API key from a browser or mobile app.

The host creates the game using a Player Space that the host can manage. The host and selected players must be approved members of that Space. If audience voting is enabled, configure a different public Audience Space in the same tenant; every viewer who votes must first have a Platform Identity and approved membership there. Public Space access does not remove the membership requirement for participation.

The host selects 5–10 Player Space members for each round; the host cannot be selected as a player. A person in the roster who does not answer is an abstainer: they receive no role and cannot vote in the accusation phase. Only respondents receive roles and become accusation options.

## Create and play a game

All paths below are relative to `/v2`. Responses use the standard Qaragon `data` envelope.

1. **Create the game.** `POST /games` with `{"type":"find_the_minority","player_space_id":"..."}`. Add `audience_space_id` to enable an audience. Store the returned Game ID in the tenant BFF; there is no Game list endpoint.
2. **Start a round.** The host calls `POST /games/{id}/rounds` with a question, exactly two options with IDs `A` and `B`, and the round's `player_identity_ids`. The IDs must be distinct approved members of the Player Space. This command requires an `Idempotency-Key`. There are no automatic phase timers; the host advances the game when ready.
3. **Collect private answers.** Each selected player reads `GET /games/{id}` and submits exactly one answer using `POST /games/{id}/participate`, with `{"option_ids":["A"]}` or `{"option_ids":["B"]}`. A submitted ballot is final. The response means it was accepted into the shared submission queue; refresh the Game to observe the persisted answer. The host sees only the submitted count, not the answers.
4. **Lock answers.** The host sends `{"action":"lock_answers"}` to `POST /games/{id}/actions`. Qaragon closes answer acceptance and assigns roles from persisted responses. If the response is `409 BALLOTS_PROCESSING`, accepted submissions are still being processed. Retry the action with a new `Idempotency-Key` after they settle.
5. **Continue based on the outcome.** For a normal unequal split, respondents privately learn their Majority or Minority role and the round enters discussion. An equal split, unanimous answers, or no answers has no Minority: call `reveal_next` once to complete that special round. Special rounds award no points and skip audience and accusation voting.
6. **Optionally run the audience poll.** For a normal round with an Audience Space, the host can open voting with `open_audience_voting`. Each eligible viewer submits one final suspect choice through the same Game participation endpoint, using the respondent's Platform Identity ID as the option. Viewers cannot change a submitted vote. The host closes the poll with `close_audience_voting`; if it returns `409 BALLOTS_PROCESSING`, retry with a new key after accepted ballots settle. The host may show or hide the live tally to players and audience separately with `set_board_visibility`, and may reveal the frozen result with `reveal_audience_verdict`.
7. **Run player accusations.** The host calls `open_accusations`. If the audience poll was configured but never opened, this action closes it automatically. If it was opened, it must be closed first. Respondents then submit one private accusation each through `/participate`, selecting another respondent; self-votes are rejected. The host calls `lock_accusations`. If it returns `409 BALLOTS_PROCESSING`, retry with a new key after accepted ballots settle.
8. **Reveal and continue.** The host calls `reveal_next` three times. The stages reveal the Minority option and counts, the uniquely top-accused player's answer (or no accused player on a tie), then all respondents' answers. Points are applied at the final stage. Start another round only after the previous round is complete, or call `end` to end the Game.

Host commands use `POST /games/{id}/actions` with an `action` value: `lock_answers`, `open_audience_voting`, `close_audience_voting`, `set_board_visibility`, `reveal_audience_verdict`, `open_accusations`, `lock_accusations`, `reveal_next`, or `end`. `set_board_visibility` also accepts `board_to_players` and `board_to_audience` booleans.

## Views, privacy, and scoring

Use `GET /games/{id}` to refresh the view after a command or ballot. The Game API has no Game-specific real-time subscription; the tenant app should refresh the view or use its own notification mechanism.

- Only the host can issue host commands. The host sees ballot counts and the audience tally, but not hidden player answers or roles.
- A player sees their own answer and role. Other players' answers and roles remain hidden until the staged reveal.
- Audience members must be approved members of the Audience Space. They see the audience tally only when the host enables audience visibility or reveals the verdict.
- The audience poll is advisory. It does not affect the Minority assignment, player accusation tally, or score.
- Each respondent who correctly accuses a Minority player earns one point. Each Minority respondent who is not the unique top accused player earns one point. A tie for top accusation means no one is the group’s accused player. Split, unanimous, and no-answer rounds award no points. Players are never eliminated.

The `phase`, `audience_poll_state`, and `reveal_stage` in the Game view indicate which action is currently available. Invalid transitions return `409 GAME_PHASE_CONFLICT`. Other errors include `400 INVALID_GAME_REQUEST`, `403 FORBIDDEN`, and `404 GAME_NOT_FOUND`; ballot eligibility and duplicate-submission errors use the shared submission error codes.

## Retries and implementation boundary

`/rounds` and `/actions` require an `Idempotency-Key`. `/games` and `/participate` accept the key for safe retries; it is optional there. Use a unique key for each logical request and reuse it only to replay that same request. When a command returns `BALLOTS_PROCESSING`, retry as a new command with a new key. The participant submission itself remains one final ballot per identity per phase.

The Game resolves its active Vote internally and sends accepted ballots through Qaragon's shared participation and submission pipeline. It returns Game views and submission receipts, not child Contest IDs. Do not create, read, tally, close, or submit against the internal child Vote contests directly.

## MCP discovery

In the Qaragon MCP, use `search_api` with `Find the Minority` or `/games`, then `describe_operation` for the exact contract. Read `qaragon://guides/find-the-minority` for this workflow. The MCP is documentation-only; tenant BFFs make authenticated API calls themselves.
