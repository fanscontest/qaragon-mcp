# Contest scoring, grouping and winner rules

The glossary is published as `qaragon://guides/domain-language` and maintained in `docs/domain-language.md`. This guide explains the rules and distinguishes them from current implementation support. API schemas describe field shapes; they do not replace these business meanings.

## Establish the mechanic before assessing configuration

For each setting, establish its intended effect, its scope, the Contest type's support and the behavior a test must demonstrate. An ignored field is evidence of a support gap; it does not establish that the setting is unsuitable for that mechanic. Conversely, a shared create field being accepted does not prove a type implements its effect.

Grading, score adjustment, ranking, tie-breaking, winner selection and reward fulfillment are separate decisions. Preserve their meanings when deciding which controls to expose. A hosted Game may define its own team outcome above a primitive's native result.

## Submission zones

Zones compare authoritative submission timestamps within the same Contest group. A participant's personal start time and queue-processing duration do not determine their zone. With one group, the comparison includes the whole Contest. With several groups, each group's turnout and submission times determine only that group's zone boundaries.

Adding, removing or changing submissions in group B must leave group A's zone assignment and deductions unchanged. This applies to both system and open grouping. Global leaderboard positions may still change as results from other groups enter the leaderboard.

Zone Advantage controls the strength of the deduction configuration. It does not configure ranking tie-breakers. Setting it to zero disables zone deductions while leaving the applicable ranking rule intact.

For MC, Moment and Prediction, the current configuration uses green/blue/orange/red bands of 30%/20%/20%/30%. Deductions are generated from Zone Advantage and the question count: `ceil(zone_index × zone_advantage × 100 / question_count)`, with indices 0–3. Non-green deductions are capped at 32%; if orange exceeds 25%, red uses the orange deduction. Small-group rounding and equal submission timestamps use the existing boundary algorithm. The group-scope correction preserves that algorithm.

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

The MC/Moment/Prediction result flow is:

1. Grade answers to obtain raw points and a raw percentage score.
2. Determine each participant's zone within their Contest group and apply its deduction.
3. Rank adjusted percentage scores descending. Earlier accepted submission time resolves equal adjusted scores.
4. Assign global and group ranks. Select winners using group rank, `winners_per_group` and the minimum adjusted score of 25%.

Zone Advantage zero still leaves step 3's submission-time tie-break in effect. A queue delay does not replace the accepted timestamp with processing time.

For these types, three groups with two winning places each permit up to six qualifying winners. Empty groups and participants below the minimum score can reduce that count. Other primitives have their own native winner rules; the shared winner-count field alone does not define how ties or qualification work.

## Grouping, teams and Code Swap rooms

`no_of_groups` divides a Contest into competitive pools. A count of one means one pool. `system` delegates assignment to Qaragon; `open` lets participants choose their pool. In ordinary system-grouped Contests, assignment currently chooses the least-loaded pool at submission. Tournament roster preassignments take precedence when present.

A Team Set lets a participant represent a named team independently of their Contest group. Current Contest team results count individual winners by represented team and show each team's share of all winners. Individual winners without a represented team remain in that denominator. Team representation does not change individual grading, winning places or rewards.

Code Swap uses a different cooperative mechanic: members exchange codes within a room, and the room qualifies when its members have completed their check-ins. The first qualifying rooms win as whole rooms. `room_size` and the number of winning rooms determine the shared `winners_per_group` value. For example, two winning rooms of three people configure six winning participants when both rooms are full. Code Swap uses one competitive Contest pool; its rooms are not generic winner pools.

## Current implementation support

| Type or variant | Current native result behavior | Configuration limits or gaps |
| --- | --- | --- |
| MC / Moment | Adjusted percentage score, then submission time; qualifying winning places per group. | System or open grouping. Submission zones are group-scoped. |
| Prediction | Same score-adjustment/ranking flow as MC. | Open grouping or one pool; system grouping is rejected. |
| Guess final-board, including Word Chain / Word Reveal | Raw score ranking within each group; positive scores at the winning cutoff qualify, including ties. | Zone configuration is currently accepted but unused. Submission time is recorded but does not break score ties. These are implementation facts to assess against the intended product behavior. |
| CheckIn / Code Swap | First qualifying cooperative rooms win as whole rooms. | One competitive pool; whole-room winner count. Zone configuration is unused. |
| Creative | Judging-method scores, currently ranked globally; tied scores share ranks. | Generic grouping and zone settings are not applied by settlement. |
| Vote | Winning options selected by the configured `top_n` or `least_n` strategy. | Its winners are options, rather than voting participants. |

This table records implementation support. It does not authorize silently removing configured behavior or changing a primitive's tie policy. Confirm intended behavior before changing those mechanics, and update this guide with the implementing change.

## Maintenance and verification

The canonical sources are Uman's `docs/domain-language.md` and this guide. The infra task `mcp:guides` copies them unchanged into the MCP bundle; `mcp:contract` runs that task when preparing the developer catalog. MCP publishes both resources. Its tenant-integration prompt directs agents to the relevant glossary sections and includes the Contest rules. The developer plugin reads them from MCP rather than maintaining another set of rules.

When changing grading, zones, grouping or winner selection, update the relevant rule and add a behavior test. The Puzzle regression tests cover group independence, single-pool boundaries, disabled deductions and final ranking/winner selection. MCP tests verify that both documents can be read and are included in integration prompts.
