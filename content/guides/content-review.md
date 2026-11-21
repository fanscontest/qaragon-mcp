# Shared answer-board publication review

Word Chain and Word Reveal content libraries are owned by Qaragon and shared
across tenants. Preparing a board and approving its publication are separate
operations. AI validation is useful review input; it is not editorial approval.

## Prepare candidates

Use the existing platform-admin generation operations:

- `POST /v2/admin/games/word-chain/library/generate`
- `POST /v2/admin/games/fans-say/library/generate`

The latter supplies the clue-board library used by Word Reveal as well as
hosted Fans Say. Generation requests retain their existing brief, set shape,
quantity, idempotency and progress contracts. They append content; they do not
reset the library. Each validated set is saved with `review_status:
"pending_review"`. The generation job can complete while human review is
outstanding. Failed later sets do not delete earlier candidates.

## Review and publish

Authenticate with the tenant key and the actual reviewer's `X-Acting-As`
Platform Identity. Qaragon must have explicitly granted that identity its
platform-admin role. Tenant/app-admin status alone grants no access. Do not
replace a human caller with a privileged system identity in the tenant BFF.

```http
GET /v2/admin/answer-boards?kind=word_chain&status=pending_review&limit=20
GET /v2/admin/answer-boards/word_chain:BOARD_ID
PUT /v2/admin/answer-boards/word_chain:BOARD_ID/review
Content-Type: application/json

{"status":"approved"}
```

Use `kind=word_reveal` for clue boards. List status defaults to `pending_review`;
`approved` and `rejected` are also supported. Pass the last returned namespaced
ID as `after` to continue; the limit is 1–100. Administrative list/detail reads
include the prompt, ordered complete answers, topic/profile metadata, source,
creation time and current review decision. These responses are not cacheable.
They must never be forwarded to participant views.

The review write accepts only `approved` or `rejected`. It records
`reviewed_by` from the authenticated Platform Identity and `reviewed_at` from
the server clock; clients do not supply either audit value. Repeating the
same decision keeps its audit. A new decision replaces the current publication
state. Re-importing the same content does not reset that decision.

Only approved Word Chain/Word Reveal items appear in the ordinary
`GET /v2/answer-boards` catalog, resolve by `board_id` during Contest creation,
or get selected from a shared library for a hosted draft. Pending/rejected IDs
cannot bypass that rule. Existing Contests and Games retain their committed
answers, clues, results and progression after source-library review changes.

A creator can still prepare and accept a private hosted Game draft through
the existing flow. Its generated content enters the shared library as pending;
accepting that private draft does not approve shared publication.

## Initial review client and rollout

FansContest temporarily exposes Question review under Approvals. Branda
relays all operations as its signed-in caller. The entry appears only after a
Qaragon-authorized review read succeeds, rather than checking the app's admin
flag. Qaragon authorizes every read and write independently of UI visibility.

Deploy the Uman schema and verified-identity admin-header correction, then
Branda's relays, before testing the app. Normal migration adds review columns
to `word_chain_library` and `fans_say_library`; existing rows default to pending.
Review them before enabling automatic Contest selection. Newly inserted curated
starter definitions explicitly declare approval. Seeding does not overwrite or
approve an existing record, and no migration bulk-approves historical AI content.

No new provider key or generated SDK release is needed for these opaque BFF
relays. Publish the updated OpenAPI/MCP contract before external integrations
depend on the administrative operations. Optional [bounded replenishment and recurring Contests](qaragon://guides/recurring-contests)
use these same workers and approval decisions. Global replenishment remains
disabled until explicitly configured. Tenant daily jobs opt in separately;
FansContest bootstraps Nigeria without enabling paid generation. Prediction, Quiz, Moment and Creative ingestion
are outside this approval change.
