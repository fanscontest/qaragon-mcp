# Automated Contest Scheduler

The Automated Contest Scheduler creates one ordinary Contest in a Space on
selected days when that Space does not already have one for that day. It uses
existing content and skips the day when suitable content is unavailable.

Each tenant opts in through **scheduler settings**: Space, creator, time zone,
daily creation clock, Contest timing and eligible types. Qaragon stores those
settings and runs the jobs. FansContest Daily is Branda's brand and Home
presentation. A Contest's own schedule remains its participation opening and
closing times.

## Daily behavior

At the configured local morning time:

1. Check the target Space for an ordinary, non-cancelled Contest for today,
   regardless of its creator or when it was created. If one exists, skip.
2. Otherwise randomly choose **one** enabled Contest type.
3. Ask that content owner for usable material. Create one ordinary Contest if
   possible; otherwise skip the day. Do not try another type, generate content,
   wait for an approval or backfill later that day.

Quiz/word editions use their opening date. Prediction opens immediately and
closes at the earliest selected event kickoff; the shared Prediction module
rejects an already-passed submission deadline during normal create/update validation; its submission-deadline date
identifies its edition. A Prediction opened yesterday for today's fixture must
therefore count today, including in the skip check and Home list.

The run records the selected type and `preparing`, `created`, `skipped` or
`failed` outcome, with a unique tenant/Space/local-date constraint. A committed
attempt prevents another execution that day, including after missing content
becomes available. Creation checks for a manually added Contest again after
source preparation. Disabled configurations and past openings do no work.
There is no rolling 24-hour creation window or five-minute creation sweep.

Asynq's existing periodic-task manager reconciles enabled configuration once a
minute and registers actual daily cron entries in each configured time zone.
That reconciliation only updates cron configuration; it does not attempt
Contest creation. Each daily task has automatic retries disabled. Database
locking and dated runs protect against scheduler replicas and queue replay.
An interrupted/failed attempt is acceptable for this MVP; it does not start a
recovery workflow or keep retrying until it creates something.

## Content ownership

| Selection | Source |
| --- | --- |
| Quiz (`mc`) | Puzzle's existing topic/category and per-difficulty availability APIs, then its adhoc community-sheet preparation. Sample up to the configured question count. |
| Prediction (`prediction`) | Puzzle prepares a tenant-owned slip from the first usable saved event Highlight before the tenant's day ends. It returns only the slip ID, event count and submission deadline. |
| Word Chain (`guess` / `word_chain`) | Published shared chain boards. |
| Word Reveal (`guess` / `word_reveal`) | Published multi-answer, general/global-culture boards suitable for all ages in the configured language. |

Code Swap, Moment and Creative are excluded from this automation. Hosted Games
are separate experiences. Prediction needs a usable published event Highlight and
normal prediction questions. Missing Highlights or questions skip the day.
Quiz uses the existing tenant question bank, not AI generation. Word Chain
currently requires English. `language` selects the word-board library;
Quiz availability is determined by Puzzle's existing topic bank.

`platform/content/puzzlecatalog` calls Puzzle's existing APIs over the private
service network with the verified tenant and configured creator identity.
Uman does not query OpenSearch, Puzzle SQL or API-Football. Puzzle owns topic
sampling, sheet/slip preparation and source validation. `PUZZLE_SERVICE_URL`
defaults to `http://puzzle:4000`; override only for a different private topology.
Prediction preparation is an internal service operation, unavailable through the
public gateway or tenant API contract. Puzzle owns Highlight order, event
materialization, usable-question selection and ordinary slip creation. Uman
receives only the prepared slip's ID, event count and submission deadline. No
suitable Highlight skips the day. Preparation performs no ingestion or model calls.

## Configure scheduler settings

Use the tenant key and the actual creator's `X-Acting-As`. The caller must have
ordinary Contest creation authority in the selected Space. One configuration
is allowed per tenant/Space. Its creator can list/update it; other tenants are
not opted in automatically. Qaragon platform-admin rights are not needed to
configure a tenant's own automation.

```http
POST /v2/contest-schedules
Content-Type: application/json

{
  "space_id": "TENANT_GENERAL_SPACE_ID",
  "title": "FansContest Daily",
  "time_zone": "Africa/Lagos",
  "creation_time": "04:00",
  "start_time": "19:00",
  "weekdays": [0, 1, 2, 3, 4, 5, 6],
  "duration_minutes": 10,
  "question_count": 8,
  "contest_types": [
    {"source_type": "mc"},
    {"source_type": "prediction"},
    {"source_type": "guess", "source_variant": "word_chain"},
    {"source_type": "guess", "source_variant": "word_reveal"}
  ],
  "language": "en",
  "winners_per_group": 1,
  "enabled": false
}
```

When the tenant is ready to enable the Automated Contest Scheduler, send
`PUT /v2/contest-schedules/{id}` with the returned settings ID and the full
settings above, changing `enabled` to `true`. This enables future daily runs at
the configured creation time.

`GET /v2/contest-schedules` reads the caller's scheduler settings.
`PUT /v2/contest-schedules/{id}` replaces settings, including `enabled: false`
to stop future tasks. It does not cancel existing Contests or an already-claimed
run. Weekdays use Sunday=0 through Saturday=6; empty means every day. The normal
opening must follow the creation clock. Prediction uses an immediate opening
and its source deadline instead of the normal start/duration.

All automatic editions use bragging rights, one competitive group and the
configured number of winners. Word boards use Guess final-board submission,
positional matching and fixed initial/random clues. Normal `/v2/submissions`,
Participation, grading, results, access rules and existing notification
preferences apply. No Daily opt-out or new notification campaign is added.

## Architecture and transaction

`experiences/recurringcontests` owns tenant configuration, dated attempts and
content selection. `platform/contests/creation.Creator` prepares/persists ordinary
Contests for both HTTP and automation. Application wiring binds repositories,
content clients, queue infrastructure and existing lifecycle/feed callbacks.
Lower platform packages remain independent of this experience and concrete Games.

A short transaction claims the dated attempt. Remote source preparation runs
outside SQL transactions. The final transaction rechecks the Space and commits
the Contest, primitive content, accepted self-hosting, created outbox event,
word-board usage and successful run together. Lifecycle/feed hooks follow
commit, using existing delivery behavior. A prepared Puzzle sheet/slip can
remain if later Contest creation fails; no distributed rollback is introduced.

## Independent content preparation and review

Global content replenishment is independent of the daily job and tenant settings.
Qaragon platform admins manage:

- `GET /v2/admin/answer-boards/replenishment`
- `PUT /v2/admin/answer-boards/replenishment/{word_chain|word_reveal}`

A policy specifies `brief`, `language`, `difficulty`, `minimum_approved`,
`pending_limit`, `sets_per_run` and `enabled`. Policies are disabled unless
explicitly configured. The 02:00 UTC sweep counts unused approved selectable
boards and queues a bounded batch only below that reserve and the pending
backlog threshold. The threshold is checked before the batch; one batch can
exceed it, and later batches pause. Stable kind/date task IDs retained for
48 hours and disabled paid retries prevent repeated automatic batches.

The existing owning workers generate three chains per Link Rush set, or six
seven-answer boards plus a hosted tiebreak per Seven Clues set. Deduplication
can reduce new rows. The single-answer tiebreak is excluded from community
selection/reserve counts. Generated candidates stay pending human approval;
a generation finishing does not create a Contest or publish a question.
See [publication review](qaragon://guides/content-review).

## Home delivery and FansContest bootstrap

`GET /v2/contest-schedules/contests?region=NG` returns ordinary Contests for
today from the tenant's configured country, containing-continent and Worldwide
Spaces. It includes manual and automatic editions. Omit `region` for Worldwide
only. Cancelled, deleted and managed Contests are omitted. Disabling automation
does not hide existing editions. This marketing read grants no participation
permission. Branda relays member/public `/contests/daily` to this operation.

Home shows personal message → Featured → FansContest Daily → feed. Daily cards
reuse the Activity left-accent presentation, open Contest Details and show
normal timing/status. Daily IDs are removed from Featured on that page. Empty
or failed Daily reads leave the feed usable.

Branda's startup bootstrap explicitly opts **FansContest Nigeria only** into the
settings above with `enabled: true`, using the configured General Nigeria Space
and its actual owner as creator. It waits for that environment's General Space
mapping import and Qaragon API readiness without blocking the mobile backend.
It creates once using a stable idempotency key; any existing configuration for
that Space is retained, including changed title/time or a later disable choice.
There is no hardcoded email, tenant ID or staging Space ID in application code.

Deploy Puzzle with the highlighted-slip preparation operation before Uman
uses it, then deploy the matching Qaragon schema/routes, gateway and Branda. Once the General
Space map exists, no manual configuration API call is needed for FansContest's
initial Daily setup. Other tenants/Spaces remain opt-in. Provisioning, reviewer
assignment, approval of stock and existing queue health checks are in the
production checklist. No paid model call or live deployment was performed while
implementing this feature. Renaming Uman and paid-feature billing remain deferred.
