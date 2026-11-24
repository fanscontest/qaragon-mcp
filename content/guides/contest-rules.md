# Contest scoring, grouping and winner rules

The glossary is published as `qaragon://guides/domain-language` and maintained in `docs/domain-language.md`. This guide explains the rules and distinguishes them from current implementation support. API schemas describe field shapes; they do not replace these business meanings.

## Establish the mechanic before assessing configuration

For each setting, establish its intended effect, its scope, the Contest type's support and the behavior a test must demonstrate. An ignored field is evidence of a support gap; it does not establish that the setting is unsuitable for that mechanic. Conversely, a shared create field being accepted does not prove a type implements its effect.

Grading, score adjustment, ranking, tie-breaking, winner selection and reward fulfillment are separate decisions. Preserve their meanings when deciding which controls to expose. A hosted Game may define its own team outcome above a primitive's native result.

## Pre-enrollment admission

Pre-enrollment uses the shared admission rules for every Contest type. Before recording registration, check Space access and membership, creator blocks, mandatory experience qualification and the tenant-enabled configured eligibility rules. The enrollment window and the requirement for the Contest to use enrollment also apply. Registration does not require an existing enrollment; actual participation still does unless the experience supplies an enrollment exemption.

The `member_can_enroll` advisory uses the same access and eligibility decision as the enrollment endpoint. A rejected registration creates no enrollment and cannot enter a downstream enrolled roster, such as Code Swap room formation. Enrollment remains separate from accepted participation: registering does not bypass the existing first-submission admission check or establish durable entry.

## Submission zones

Zone Advantage is available only for MC/Quiz and its Moment variant. It encourages prompt submission where external tools could supply answers. Other primitives reject supplied `zone_advantage`; Prediction ranks raw earned percentages and does not apply zone deductions. Zone configuration is derived by the MC module.

Zones compare authoritative submission timestamps within the same Contest group. A participant's personal start time and queue-processing duration do not determine their zone. With one group, the comparison includes the whole Contest. With several groups, each group's turnout and submission times determine only that group's zone boundaries.

Adding, removing or changing submissions in group B must leave group A's zone assignment and deductions unchanged. This applies to both system and open grouping. Global leaderboard positions may still change as results from other groups enter the leaderboard.

Zone Advantage controls the strength of the deduction configuration. It does not configure ranking tie-breakers. Setting it to zero disables zone deductions while leaving the applicable ranking rule intact.

For MC and Moment, the configuration uses green/blue/orange/red bands of 30%/20%/20%/30%. Deductions are generated from Zone Advantage and the question count: `ceil(zone_index × zone_advantage × 100 / question_count)`, with indices 0–3. Non-green deductions are capped at 32%; if orange exceeds 25%, red uses the orange deduction. Small-group rounding and equal submission timestamps use the existing boundary algorithm. The group-scope correction preserves that algorithm.

The deduction is a proportion of the earned percentage score:

```text
zone deduction = raw percentage score × zone deduction percentage / 100
adjusted percentage score = raw percentage score − zone deduction
```

Raw earned points remain available in the result. Zone Advantage can allow someone with fewer correct answers to outrank a later submitter; it does not guarantee a fixed number of wrong answers will still win.

For example, these two entries belong to a larger group:

| Participant | Raw score | Zone deduction rate | Adjusted score |
| --- | --- | --- | --- |
| Alice | 80% | 0% | 80% |
| Bob | 90% | 20% | 72% |

Alice ranks above Bob on adjusted score. No tie-break is needed for these entries. Another group's earlier submissions must not move Alice or Bob into different zones.

## Ranking and winning places

The MC/Moment result flow is:

1. Grade answers to obtain raw points and a raw percentage score.
2. Determine each participant's zone within their Contest group and apply its deduction.
3. Rank adjusted percentage scores descending. Earlier accepted submission time resolves equal adjusted scores.
4. Assign global and group ranks. Select winners using group rank, `winners_per_group` and the minimum adjusted score of 25%.

Zone Advantage zero still leaves step 3's submission-time tie-break in effect. A queue delay does not replace the accepted timestamp with processing time.

Prediction uses the same percentage-score, submission-time and group-winning-place rules without step 2.

For these types, three groups with two winning places each permit up to six qualifying winners. Empty groups and participants below the minimum score can reduce that count. Other primitives have their own native winner rules; the shared winner-count field alone does not define how ties or qualification work.

## Grouping, teams and Code Swap rooms

`no_of_groups` sets the number of groups in a Contest. A count of one means one group. `system` delegates assignment to Qaragon; `open` lets participants choose their group. In ordinary system-grouped Contests, assignment chooses the least-loaded group on the first Participation recorded through shared submission processing. The assignment remains fixed through settlement; there is no closing-time regrouping. Tournament roster preassignments take precedence when present.

MC/Quiz/Moment, Prediction, Guess and Creative support either grouping policy. Groups, members and scores use the same Contest access and result-availability rules as one group; there is no separate group visibility permission. Empty groups remain visible for participant choice. A first open-group submission must select an existing group before acceptance; the shared Processor persists that association. Later attempts retain the recorded assignment. Code Swap and Vote use one Contest group.

A Team Set lets a participant represent a named team independently of their Contest group. Current Contest team results count individual winners by represented team and show each team's share of all winners. Individual winners without a represented team remain in that denominator. Team representation does not change individual grading, winning places or rewards.

Code Swap uses a different cooperative mechanic: members exchange codes within a room, and the room qualifies when its members have completed their check-ins. The first qualifying rooms win as whole rooms. `room_size` and the number of winning rooms determine the shared `winners_per_group` value. For example, two winning rooms of three people configure six winning participants when both rooms are full. Code Swap uses one competitive Contest group; its rooms are not generic Contest groups.

## Supported configuration

| Type or variant | Zone Advantage | Competitive grouping | Native winner rule |
| --- | --- | --- | --- |
| MC / Quiz / Moment | Optional; zero disables deductions. | One group, system or open. | Group winning places; minimum adjusted score 25%; equal scores ordered by accepted submission time. |
| Prediction | Unavailable. | One group, system or open. | Group winning places; minimum raw percentage 25%; equal scores ordered by accepted submission time. |
| Guess final-board, including Word Chain / Word Reveal | Unavailable. | One group, system or open. | Positive scores at the winning cutoff qualify, including ties; time does not break score ties. |
| Creative | Unavailable. | One group, system or open. | Judging scores ranked within each recorded group; ties share ranks. Existing judging-method qualification/quorum still applies. |
| CheckIn / Code Swap | Unavailable. | One Contest group; cooperative rooms remain separate. | First qualifying rooms win as whole rooms; `winners_per_group` must be a positive whole multiple of room size. |
| Vote | Unavailable. | One group. | Option winners selected by `top_n` or `least_n`; shared participant winner-count edits cannot change this immutable option strategy. |

Create and permitted updates delegate configuration validation to the owning module. Updates retain the existing lifecycle/timing restrictions. Creative computes judging scores once over the existing judgment set, then derives global and group ranks without changing the rating model or comparison set. Grouping does not alter Team Sets, rewards or primitive tie policies.

## Maintenance and verification

The canonical sources are Core's `docs/domain-language.md` and this guide. The infra task `mcp:guides` copies them unchanged into the MCP bundle; `mcp:contract` runs that task when preparing the developer catalog. MCP publishes both resources. Its tenant-integration prompt directs agents to the relevant glossary sections and includes the Contest rules. The developer plugin reads them from MCP rather than maintaining another set of rules.

When changing grading, zones, grouping or winner selection, update the relevant rule and add a behavior test. The Content regression tests cover group independence, single-group boundaries, disabled deductions and final ranking/winner selection. MCP tests verify that both documents can be read and are included in integration prompts.
