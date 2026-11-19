# Integration recipes

These are implementation walkthroughs, not fixed request schemas. Obtain each operation's current parameters and response shape with `search_api` and `describe_operation` before writing the tenant integration.

## Community Word Chain or Word Reveal

Example request: “Let members create Word Chain contests and submit their complete board in our app, without a host.”

1. Read [community Guess](qaragon://guides/community-guess). Resolve the creator and participant PlatformIdentities through the BFF's existing authenticated user mapping.
2. Discover `/v2/venues`, `/v2/spaces`, `/v2/identities` and `/v2/spaces/{id}/memberships` as needed. Reuse existing records. Verify the participant has approved membership; apply any enrollment requirement from the Contest contract.
3. Discover `GET /v2/answer-boards` and `POST /v2/contests`. Select prepared content with `kind=word_chain` or `kind=word_reveal`. Create the scheduled Contest with the selected board ID, `source_type=guess`, `submission_policy=final_board`, `matching_rule=by_position` and `clue_strategy=initial_random`, plus the normal fields required by the live create schema. Use the returned `source_variant` for the UI label.
4. Read `GET /v2/contests/{id}/guesses` for the prompt, slot positions and saved clues. Keep given letters fixed while editing, and preserve unanswered positions as empty strings. Private creator configuration is separate from the participant projection.
5. Discover `POST /v2/submissions`. Target the Contest and send the complete `answers` list in its payload, retaining any grouping selection. Keep the action's idempotency key stable across retries. A 202 response opens the tenant's shared submission receipt/completed experience; final results wait for settlement.
6. Discover the participant result and scoreboard operations. Handle the documented pending response, then render finalized results. The app should show `submitted` while accepted work is still queued and keep hidden answers private until settlement.

Validate by submitting once, replaying the same key, attempting a second final entry, reading before settlement and reading after completion. Run live mutations only against the environment and accounts authorized for the task; deterministic BFF tests can cover these cases first.

## Contest result webhooks

Example request: “Notify our users when Qaragon has finished grading their Contest.”

1. Use `list_webhook_events` to find the public event, then `describe_webhook_event` for its envelope and data. `contest.completed` represents finalized grading; use public event names returned by the catalog.
2. Read [webhook delivery guidance](qaragon://guides/webhooks) and obtain current SDK setup with `get_sdk_setup`. Configure the subscription signing secret server-side, independently of the tenant API key.
3. Read the exact raw request bytes and verify `X-FC-Signature`, including the timestamp tolerance, with the SDK helper. Decode the verified envelope using its current schema and preserve batch/delivery metadata when queuing work. The [Go example](go-sdk-examples.md#verify-a-webhook) shows signature verification without discarding envelope fields.
4. Persist deduplication and processing intent before acknowledging success. A failed persistence/queue operation must remain retryable. Use the contract's event and batch identifiers; an entire multi-batch occurrence must not be collapsed into its first batch.
5. Resolve recipient PlatformIdentity IDs to tenant users, apply tenant notification preferences and send through the tenant's delivery provider. Keep these delivery concerns in the tenant BFF.

Check signature rejection, duplicate delivery and multiple batches of the same occurrence. Verify the tenant's public receiver and intended target environment separately from local unit tests.
