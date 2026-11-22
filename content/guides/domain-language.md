# Qaragon domain language

Use these terms in requirements, public documentation, code and reviews across Qaragon's platform services and tenant integrations. This is the canonical domain glossary. It defines business meanings; schemas, lifecycle transitions and implementation support belong in the relevant contracts and rules guides. A term does not require a separate resource or database table.

Read the section for the capability being changed and the related concepts it uses. Add or refine terms here when the domain changes, and refresh the MCP bundle through `task mcp:guides` from the infra checkout. Detailed scoring and grouping rules are maintained in `docs/contest-rules.md` and published as `qaragon://guides/contest-rules`.

## Naming rules

- Reuse the canonical term whenever the business meaning is the same. Apply it to domain identifiers, API properties, errors, tests and documentation.
- Before introducing a domain term, search this glossary and the relevant existing code. Explain the distinct behavior or relationship it represents relative to the closest existing concept. If there is no distinction, use the existing term.
- Record a genuinely new concept here before implementing it. Resolve ambiguous meanings with the user instead of adding a convenient synonym.
- Tenant product labels may differ from Qaragon names; document their mapping at the integration boundary. Technical terms such as database connection pools retain their technical meaning.

## Platform, tenants and identities

| Term | Meaning |
| --- | --- |
| **Platform / Qaragon** | The shared contest infrastructure offering tenant-isolated identities, engagement capabilities, Contest primitives and composed experiences through its API. |
| **Tenant / Brand** | An organization integrating Qaragon into its own product. The Tenant is the ownership and data-isolation boundary; it controls its users, application permissions, notifications and reward fulfillment. |
| **Tenant application** | The product through which a tenant's users discover, create and participate in activities. Its user experience and authentication belong to the tenant. |
| **Tenant BFF** | The tenant-owned backend for its application. It authenticates users, authorizes their actions, maps their identities and calls Qaragon with tenant credentials. |
| **Platform operator** | The role managing platform-level administration, including tenant approval. It is distinct from an administrator inside a tenant's application. |
| **Tenant API key** | A tenant credential authorizing its server to call Qaragon. It authenticates the tenant rather than an individual fan. |
| **Fan / Member** | An end user of a tenant's product. Use PlatformIdentity when referring to that person's platform record, and Space membership when referring to their standing in a Space. |
| **PlatformIdentity / PI** | A tenant-scoped actor record used to identify a person in Qaragon activities. It contains the platform identity, profile and geographic information needed by the platform. |
| **Identity mapping** | The tenant's association between its own user identifier and a PlatformIdentity. The two identifiers need not be equal. |
| **Acting identity** | The PlatformIdentity on whose behalf an authorized tenant request is made. The tenant proves that its authenticated user may act for that identity. |
| **Identity metadata** | Opaque references or attributes stored on a PlatformIdentity for the tenant's own use. They do not establish authorization or create platform behavior. |
| **Venue** | A tenant-owned organizing location for Spaces. Venue organization sits inside tenant isolation. |

## Spaces and access

| Term | Meaning |
| --- | --- |
| **Space** | A tenant-owned gathering place under a Venue, with an owner, membership, access policy, geographic availability and optional expiry. |
| **Space owner** | The identity responsible for managing a Space and its membership. |
| **Space membership** | An identity's standing in a Space, including whether its membership is pending, approved or denied. Approved membership is required for Contest submissions from that Space. |
| **Public Space** | A Space allowing direct joining without an access code or owner approval, subject to its other admission constraints. |
| **Restricted Space** | A Space requiring a valid access code to join, subject to its other admission constraints. |
| **Private Space** | A Space whose membership requires owner approval and which is excluded from ordinary public discovery. |
| **Space code** | A human-facing lookup reference for a Space. Its identity and persisted associations use the Space ID. |
| **Access code** | The secret used to satisfy a Restricted Space's joining requirement. It is separate from the Space's lookup code. |
| **Space capacity** | The configured limit on membership admitted into a Space. It does not configure a Contest's winner count or group size. |
| **Space expiry** | The point after which new joining and new Contest creation stop. Existing Contests keep their own lifecycle and participation windows. |
| **Region** | A named geographic area, currently a country or continent, used to express availability or eligibility. |
| **Regional Space** | A Space whose availability is limited to its configured regions. |
| **Worldwide / Regionless Space** | A Space with no regional restriction. It may have its own access policy and membership requirements. |
| **Allowed regions** | An optional Contest entry restriction within a worldwide Space. Multiple selected regions are alternatives; other eligibility and access requirements still apply. |
| **Channel** | A former platform name and a possible tenant-facing label for an ongoing Space. It is not a separate Qaragon resource or Space kind. |
| **Room** | A possible tenant-facing label for a Space used for an activity. Qualify a mechanic's own room, such as a Code Swap room, when referring to that different concept. |

## Contests and play

| Term | Meaning |
| --- | --- |
| **Contest** | One configured activity with participation rules, content, a lifecycle and a result. It can have one participant or many. |
| **Contest type / primitive** | The reusable mechanism governing a Contest's valid submissions, grading and native result. Examples include MC, Prediction, Vote and Guess. |
| **Contest source** | The prepared content or configuration a Contest uses, such as a question sheet, prediction slip or answer board. It is distinct from the Contest instance that people enter. |
| **Source type** | The primitive governing the Contest. The public Contest's `source_type` identifies this mechanism, such as `mc`, `prediction` or `guess`. |
| **Source variant** | A named experience within a primitive, such as Moment within MC or Word Reveal within Guess. It does not introduce another Contest type or participation path. |
| **Managed Contest** | A Contest whose lifecycle and use are controlled by an enclosing experience. Its participant actions and results remain under that experience's contract. |
| **Contest Space** | The Space associated with the Contest itself. Authorized hosting can make the same Contest available through additional Spaces. |
| **Contest hosting** | An authorized association allowing another Space to display a Contest or admit participation in it, according to the hosting permissions. It does not create a copy of the Contest. |
| **Participation Space** | The Space through which a participant enters a Contest. It can differ from the Contest's own Space when hosting permits it. |
| **Cross-tenant hosting** | An authorized hosting arrangement in which the additional Space belongs to another tenant. Each tenant retains its own ownership boundary. |
| **Self-play** | Play in which a participant progresses under the Contest's rules without a human host directing their actions. Participants can play alone or independently against others. |
| **Personal Contest** | A Contest reserved for one PlatformIdentity. It can belong to an existing Space; it does not imply a personal Space. |
| **Shared Contest** | A Contest admitting multiple eligible participants. In self-play, each participant progresses independently within the same configured challenge. |
| **On-demand** | Created or made available in response to a request to play. It describes availability, not a Contest type. |
| **Challenge** | A product-facing name for an activity someone can play. It does not introduce another Qaragon aggregate between Contest and Participation. |
| **Hosted play** | Coordinated play in which a human host directs pacing or transitions for the participants. It can happen in person or remotely. |
| **Game definition** | A named, versioned composition of playable mechanics, including rules for progression and the overall outcome. |
| **Game catalog** | The set of first-party Game definitions offered for tenants to use through the Game API. |
| **Game** | One playthrough of a Game definition, with its roster, progress and result. An existing Contest does not need a Game wrapper merely to be playable. |
| **Game module** | Reusable playable rules with their own inputs, actions, state and outcome, usable by Game definitions. |
| **Stage** | One occurrence of a playable module within a Game definition. A Stage is part of that Game's orchestration. |
| **Tournament** | A competition connecting distinct Contests into rounds with advancement and an overall result. |
| **Tournament Round** | One step in a Tournament, associated with a Contest and its advancement rules. |
| **Game Round** | A repeated competitive phase within a Game's rules. Qualify the word “round” when the distinction from a Tournament Round matters. |

Self-play and personal are different distinctions. A shared daily word board is self-play; so is a fresh board prepared for one person. A scheduled Contest can be self-play, and on-demand creation does not make a Contest personal by itself.

Several questions, answer slots or guesses can form one Contest. A linked competition over separate self-play Contests belongs in a Tournament. Choosing “Play again” for a fresh independent Contest does not create a Tournament. This does not remove rounds or Stages from existing hosted Games.

## People and access

| Term | Meaning |
| --- | --- |
| **Creator** | The actor responsible for creating and configuring a Contest or Game. Creation does not imply a role in directing play. |
| **Host** | The person authorized to direct a hosted Game's runtime transitions. A creator may also be its host. |
| **Host control** | An authorized request to direct the Game's phase, reveal or progression under its rules. It is separate from a player's submitted gameplay action. |
| **Participant** | A PlatformIdentity admitted to take part in a Contest. Personal play still has an identified participant. |
| **Player** | A person on a Game's active playing roster. A Game can distinguish players from its host and audience. |
| **Roster** | The identified people committed to a Game, Stage or team, including an order when turn-taking depends on it. |
| **Audience** | People observing a Game who can contribute the audience interactions its rules permit. Audience participation is distinct from the players' actions. |
| **Audience poll** | A Game-defined question or vote offered to its audience. The Game determines how the audience result influences its experience. |
| **Enrollment** | Registration for a Contest whose mechanic requires an enrolled roster. It is distinct from joining a Space and submitting an answer. |
| **Participation eligibility** | The Contest's rules determining who may enter, such as participation history or allowed regions. |
| **Admission** | The decision whether an identity may enter through a given Space, accounting for access, membership, enrollment, eligibility and any experience qualification. |
| **Eligibility preset** | A named bundle of optional Contest eligibility rules, such as new members only or participants who have not won. |
| **Experience qualification** | A mandatory entry condition imposed by an enclosing experience, such as advancing from the preceding Tournament Round. It is separate from optional Contest eligibility. |
| **Enrollment exemption** | An experience-provided waiver of separate Contest enrollment because the experience supplies the roster. Other access and qualification requirements remain. |
| **Action validity** | Whether a particular submission is permitted by the mechanic, including its actor, input, phase, timing and attempt limits. Being eligible to enter does not make every action valid. |

## Contest lifecycle

| Term | Meaning |
| --- | --- |
| **Contest schedule** | The configured opening and closing times for participation. An enrollment window, when required, is separate. |
| **Enrollment window** | The period during which an identity may register for a Contest that requires enrollment. |
| **Created Contest** | A Contest that has been configured and exists before its participation phase starts. |
| **Awaiting approval** | A Contest state waiting for the approval required by its flow. It is not an open participation state. |
| **Started / Open Contest** | A Contest in its active participation phase. Individual submissions must still satisfy its window and other rules. |
| **Judging** | The phase in which entries are evaluated by the configured judging method. |
| **Closed Contest** | A Contest whose participation phase has ended while final processing may still be pending. |
| **Completed Contest** | A Contest whose results have been finalized. Completion is distinct from merely reaching the closing time. |
| **Paused Contest** | A Contest whose active flow has been temporarily suspended. Its permitted continuation depends on the type's lifecycle rules. |
| **Cancelled Contest** | A Contest terminated through cancellation rather than ordinary completion. |
| **Settlement** | Finalizing a Contest's native results, ranks and winners under its rules. It can follow the end of participation. |

These state names describe different facts. Each Contest type defines which states and transitions its flow uses.

## Actions and results

| Term | Meaning |
| --- | --- |
| **Submission** | One durably accepted participant action and its processing outcome. Different mechanics permit different actions and execution behavior. |
| **Submission number** | A Contest-local reference identifying a submission's order. It is separate from the participant's rank. |
| **Submission receipt** | The acknowledgement or processing outcome for one accepted submission. A receipt may precede the Contest's final result. |
| **Idempotent retry / Replay** | Repeating the same request identity to obtain the existing submission acknowledgement or outcome, without creating another gameplay attempt. |
| **Immediate execution** | Processing in which the submission's committed outcome is available in the submission response. |
| **Queued execution** | Processing in which acceptance is acknowledged before a worker applies the submission. Acceptance alone does not establish a final score. |
| **Deferred evaluation** | Processing in which accepted input waits for an external evaluation before its outcome is applied. It is distinct from a gameplay miss or timeout. |
| **Submitted answer** | One word or phrase supplied by a participant. A final-board Submission can contain several submitted answers to the same prompt. |
| **Submission policy** | The configured rule governing whether a participant submits repeated individual answers or one final answer list. It is separate from answer matching and clue presentation. |
| **Final-board submission** | One immutable Submission containing the participant's answer list for a board, graded at Contest closure. |
| **Guess attempt** | An accepted submission containing an answer guess. A miss is an attempt; replaying the same accepted request is not another attempt. |
| **Participation** | One participant's current record and accumulated result for one Contest. Repeated submissions update it. |
| **Participation result** | The Contest's assessment of a participant, including its applicable score, ranks, winning outcome and graded breakdown. |
| **Graded sheet** | A breakdown relating a participant's submitted inputs to their evaluated answers and awarded points. It is distinct from an unplayed content sheet's private answer key. |
| **Provisional result** | An assessment that may still change before final settlement, such as a live Prediction score. |
| **Pending results** | The condition in which final outcomes are not yet available, even if submission recording or the participation window has finished. |
| **Play progress** | The participant's current playable state, including any start time, deadline, discovered answers and completion reason. It describes progress within a Participation, not a new top-level session. |
| **Claim** | The successful recognition of an answer under the Contest type's exclusivity rule. |
| **Claim scope** | Whether an answer can be claimed once by each participant or once across the entire Contest. This is independent of whether a Contest is personal or shared. |
| **Play window** | The Contest's opening and closing times, shared by its participants. |
| **Play duration** | An optional time allowance measured from a particular participant's authoritative start. It is different from the Contest's play window. |
| **Score** | Points earned under the Contest's grading rules. |
| **Grading** | Evaluating submitted input under the Contest type's rules to produce its native assessment, such as correctness and earned points. |
| **Raw score** | The score earned before any configured score adjustment. |
| **Adjusted score** | The score used for comparison after applicable adjustments, such as submission-zone deductions. |
| **Submission time** | The authoritative time a participant's submission is accepted. It is distinct from time spent playing or time spent waiting for processing. |
| **Submission zone** | A band determined by a participant's submission time relative to submissions in the same Contest group. |
| **Zone penalty** | The score deduction configured for a submission zone. It affects the adjusted score. |
| **Zone Advantage** | The MC/Quiz/Moment creation setting controlling the strength of group-scoped submission-zone deductions. It encourages prompt submission; it does not configure ranking tie-breakers and is unavailable on other primitives. |
| **Ranking** | Ordering competitors using the applicable comparison and tie-breaking rules. It is distinct from grading input or selecting winners. |
| **Tie-breaking** | Resolving equal ranking values using an additional declared criterion. Whether ties remain shared is part of the applicable ranking/winner rule. |
| **Winning places** | The configured number of positions from which winners may be selected in a Contest group. Qualification and tie rules determine the actual winner count. |
| **Completion** | The end of play for a participant or the settlement of a Contest. Always specify which one when both exist. |
| **Winner** | A participant, option or side selected by the Contest's or enclosing experience's winner rule. Finishing a challenge does not automatically mean winning it. |
| **Reward** | The benefit or recognition offered to winners, including bragging rights without a material prize. Reward fulfillment is separate from winner selection. |

“Submission processed,” “participant finished” and “Contest completed” describe different facts. A processed guess can leave the participant playing, and one participant finishing a Shared Contest does not end everyone else's play.

For a final-board Contest, processed submission recording means the final answer list is stored; it does not mean the list has been graded. An editable local answer draft is distinct from the final accepted Submission.

## Groups and representation

| Term | Meaning |
| --- | --- |
| **Contest group** | Participants within one Contest who compete for the same winning places. With one group, all participants compete together. |
| **Grouping type** | The policy assigning participants to Contest groups: system assignment or open participant choice. Group count and winning places are separate settings. |
| **System grouping** | Qaragon assigns each ordinary participant to the least-loaded Contest group when their first Participation is recorded. The assignment stays fixed. Tournament roster preassignments take precedence. |
| **Open grouping** | Participants choose their Contest group when submitting. Group visibility follows Contest visibility; it is not a separate permission. |
| **Group preassignment** | Establishing a participant's Contest group before play, such as revealing later Tournament Round groups from the qualified roster. It does not itself grant qualification. |
| **Winners per group** | The winning-place configuration for each Contest group, interpreted with the applicable qualification and tie rules. |
| **Team** | A named side a participant can represent. An experience can also use teams as its actual competitive rosters; state which meaning applies. |
| **Team Set** | The named teams participants may represent in a Contest. Representation and competitive grouping are independent concepts. |
| **System Team Set** | A predefined reusable Team Set owned at tenant level and available across its Spaces. |
| **Creator Team Set** | A reusable Team Set defined by a creator for a Space. |
| **Team preference** | An identity's remembered choice of team in a Team Set, used when recording its representation for participation. |
| **Represented team** | The team recorded on an individual's Participation. It is distinct from that person's competitive Contest group. |
| **Team result** | A comparison of represented teams under a declared aggregation rule, such as how many individual winners represented each team. |
| **Code Swap room** | The cooperative group whose members exchange codes and complete check-ins together. It belongs to the Code Swap mechanic and is distinct from a competitive Contest group or a Space. |

## Tournaments and standings

| Term | Meaning |
| --- | --- |
| **Tournament advancement** | Qualifying for the next Tournament Round through the preceding round's Contest outcome. It is an experience rule above the primitive. |
| **Advance per group** | The Tournament Round's configured number of advancing places per competitive group, aligned with that round Contest's winning-place configuration. |
| **Tournament qualification** | The requirement to belong to the permitted set for a round. Later rounds currently derive this from winners of the preceding round Contest. |
| **Tournament template** | A suggested sequence of round group counts and advancing places used to configure a Tournament. It is distinct from the created Tournament and its round Contests. |
| **Tournament bracket** | The presentation of competitive groups and progression through Tournament Rounds. It does not replace the actual qualification rule. |
| **Contest scoreboard** | The ranked participant results for one Contest, optionally presented by competitive group. |
| **Contest global rank** | A participant's position across all ranked participants in that Contest under its ranking rule. “Global” here does not refer to all tenants. |
| **Group rank** | A participant's position within their competitive Contest group. |
| **Leaderboard** | Standings accumulated from participation across Contests for a defined scope and period. It is distinct from one Contest's scoreboard. |
| **Leaderboard period** | The interval summarized by standings, such as a week, month, year or lifetime. |
| **Leaderboard visibility** | The Space owner's control over who can read its standings. It does not alter participation outcomes. |
| **Buddy Board** | An invitation-based private comparison of selected people's results in an existing Contest, with its own winning-place and reward configuration. It does not create another Contest or change the underlying Contest result. |
| **Buddy Board invitation** | A request to join a Buddy Board. Accepting it establishes board membership, independently of Space membership and Contest admission. |

## Content

| Term | Meaning |
| --- | --- |
| **Content item** | Reusable prepared material such as an answer board or word chain, independent of any particular Contest or participant's progress. |
| **Question / Puzzle** | One content challenge with its prompt, permitted response, answer key and any points or media. Several questions can belong to one Contest. |
| **Question sheet / Puzzle sheet** | An ordered reusable collection of questions used as content for MC Contests. It is distinct from a participant's graded sheet. |
| **Sheet format** | The content presentation classification of a question sheet, currently Quiz or Moment. It maps to the Contest's appropriate source variant without changing the MC primitive. |
| **Answer key** | The accepted answers and point values used to evaluate submissions. Its visibility follows the mechanic's reveal rules. |
| **Media** | An image or video attached to a question or submitted as a Creative entry. Its role depends on the Contest type. |
| **Difficulty** | A classification of content challenge used in selection and presentation. Its meaning depends on the content family. |
| **Answer board** | A prompt and an ordered set of accepted answers, with any prepared context or answer values. It does not imply teams, steals or a host. |
| **Answer** | A word or phrase that a Guess Contest accepts. It is a guess target, rather than a selectable Vote option. |
| **Answer position** | An answer's place in the board. Its display position is distinct from its identity and from whether position affects grading. |
| **Matching rule** | The rule comparing submitted answers with accepted answers: any order matches the answer set, while by position compares corresponding slots. |
| **Clue strategy** | The rule that projects an unrevealed answer into its visible clue, such as hidden text, letter slots, an initial or selected letters. |
| **Word Reveal** | A community self-play Guess Contest using a prepared prompt, masked answer slots and one final answer-list submission. Its board kind is `word_reveal`; Fans Say names the hosted Game with teams, turns and stages. |
| **Word chain** | An ordered sequence in which adjacent words form meaningful connections, possibly within a theme. Word Chain is also the product name for the hosted and community challenges using this mechanic. The content relationship does not itself require participants to submit in that order. |
| **Content library** | A managed collection of reusable content and its review state. Candidates may await review; only approved items are offered for published selection. Reusing an item does not reuse a previous Contest's claims or scores. |
| **Global content** | Platform-owned catalog content available across tenants. This is distinct from tenant-owned content whose visibility may be public within its permitted scope. |
| **Content visibility** | Who may discover or use a content item, according to the catalog's permissions. It is distinct from a Space's access policy. |

Community self-play answer boards and hosted Games can reuse the same prepared content while keeping their own participation and progression rules. Each created activity retains its own committed content and results.

## Contest primitives and variants

| Term | Meaning |
| --- | --- |
| **MC / Quiz** | The primitive in which participants answer multiple-choice questions against a prepared answer key. Quiz is the ordinary question-led experience. |
| **Moment** | A media-led variant of MC in which each question requires an image or video; the text prompt may be omitted. It retains the MC participation mechanism. |
| **Prediction** | The primitive in which participant predictions are evaluated against the resolved outcomes of selected events and items. |
| **Guess** | The primitive evaluating submitted words or phrases against a configured answer board, with matching, claims, clues and submission policy determining its behavior. |
| **Vote** | The primitive collecting final selections from configured options and determining winning options under its winner strategy. A submitted vote is immutable. |
| **CheckIn / Code Swap** | The primitive in which enrolled participants are assigned cooperative rooms and submit exchanged codes to complete their check-ins. Code Swap is the tenant-facing game name. |
| **Creative** | The primitive collecting media entries and determining results through a configured judging method. |
| **Vote option** | A selectable target in a Vote Contest. An option need not represent an identity or a participant. |
| **Vote selection limit** | How many different options a voter may select in one final submission. It is separate from the number of winning options. |
| **Tally** | The accumulated counts for voting options. Counts and the configured winner strategy determine option outcomes. |
| **Top N / Least N** | Vote strategies selecting the highest-count or lowest-count options, respectively, using the primitive's tie rules. |
| **Creative entry** | A participant's image or video offered for evaluation in a Creative Contest. |
| **Judging method** | The mechanism evaluating Creative entries, currently juror scoring or community judgments. |
| **Juror / Jury** | An opted-in evaluator, or the assigned group of evaluators, responsible for scoring Creative entries. |
| **Juror assignment** | A request or commitment for a juror to judge a particular Contest, independently of being an entry participant. |
| **Community ballot** | A served pair of Creative entries offered to an eligible judge for a single comparison. It is distinct from a Vote Contest's participant submission. |
| **Rating model** | The rule converting judgments into an entry's comparative rating. It is separate from how judges are selected. |

## Prediction events and slips

| Term | Meaning |
| --- | --- |
| **Prediction event** | A scheduled occurrence with outcomes that participants can predict, such as a football match or a creator-defined event. |
| **Event family** | A classification identifying the domain of an event, such as football, used to interpret its items and outcomes. |
| **Event provider** | The source supplying event information and results. It can be an external catalog provider or a creator defining an event. |
| **Event ingestion** | Bringing event information from a provider into the reusable prediction catalog. It is separate from creating a Contest using those events. |
| **Event template** | A reusable definition of the prediction items available for a kind of event. |
| **Prediction item** | One question about an event's outcome, such as the match winner or number of goals. An event may expose several items. |
| **Prediction option** | A permitted selection for an item, with its applicable point value. |
| **Prediction slip** | A reusable selection of events and per-event items configured for a Prediction Contest. It is distinct from the participant's submitted selections. |
| **Curated slip** | A prepared prediction slip offered for reuse rather than assembled independently by each creator. |
| **Double Up** | A slip item setting that doubles that item's awarded point value. It is independent of submission-zone deductions. |
| **Event result** | The recorded outcome used to determine correct predictions for an event. Result approval and Contest settlement are separate steps. |
| **Event result approval** | Confirmation that a recorded outcome is ready to be used for grading, automatically under the provider's policy or through authorized review. |

## Content preparation and AI

| Term | Meaning |
| --- | --- |
| **Creator brief** | The creator's request describing the desired subject, audience and character of prepared content. It is input to preparation rather than a live gameplay action. |
| **Content profile** | Structured selection attributes such as topic tags, language, difficulty, minimum age and cultural context used to match library content to a brief. |
| **Content draft** | Prepared content awaiting the creator's approval and commitment to a Game. It can be revised before creation under the draft flow. |
| **Content approval** | An authorized decision that prepared content may be used for a stated purpose. A creator accepts a private Game draft; library publication requires the library's approval policy. Model review and successful generation do not themselves establish approval. |
| **Committed content** | The prepared questions, accepted answers and values fixed for a created activity under its rules. Reusing the source library later does not change that commitment. |
| **Content generation** | Producing candidate questions and answer boards during preparation, potentially with an AI model. |
| **Content classification** | Assigning a structured profile to a brief or content item so it can be matched with reusable material. |
| **Content review** | Assessing candidate content against its structural and game-specific quality rules before it is offered for approval. |
| **Content replacement / Rotation** | Replacing a draft question or board during preparation. It is separate from altering a created activity's committed content. |
| **AI provider** | The service supplying model inference. It provides model execution rather than ownership of Contest or Game rules. |
| **AI model** | The selected model used for generation, classification, review or evaluation. Different preparation tasks can use the same model. |
| **Live judging** | Evaluating a gameplay input against committed content during play, potentially using an external model. It is separate from generating or reviewing the content. |

Generated library content may be pending review, approved or rejected. Under
the initial publication policy, a human content administrator must approve an
item before it becomes available for published selection. A generation job can
finish successfully while its output remains pending review. Later automatic
approval can use the same publication decision; selection continues to require
approved content. The [content-review guide](qaragon://guides/content-review)
defines the administrative API and publication gate.

## Discovery, feeds and social relationships

| Term | Meaning |
| --- | --- |
| **Discovery** | Finding available Spaces, content or activities. Being discoverable does not establish that a person is admitted to participate. |
| **Contest Highlight / Featured Contest** | A tenant-selected, ordered marketing placement for a Contest. Highlighting it does not grant access or participation eligibility. |
| **Event Highlight** | An event selected for prominent display for a UTC day in the platform's shared catalog. It is prepared independently of a tenant's Contest Highlights. |
| **Feed item** | A published presentation of an activity, announcement, social post or insight for the feed's permitted audience. |
| **Feed audience** | The recipients or membership scope permitted to see a feed item. It is distinct from the source Contest's participation rules. |
| **Insight** | A derived observation about participation, results or community activity, potentially presented as a feed item. |
| **Analytics** | Aggregated measures describing platform or community activity within the permitted scope. |
| **Follow / Stan** | A one-way relationship in which one identity follows another. Following does not create Space membership. |
| **Block** | A restriction from one identity toward another, potentially scoped to a Space. Its effects are enforced by the affected access or interaction rules. |
| **Message** | An addressed platform message with its own sender, recipient, content and status. Tenant delivery channels remain a separate concern. |
| **Notification** | A tenant-facing or user-facing notice about an occurrence. The tenant owns delivery to its users through push, email or its own inbox. |

## Rewards and sponsorships

| Term | Meaning |
| --- | --- |
| **Reward fulfillment** | Delivering an offered reward to its entitled recipients. It is the tenant's responsibility and is separate from Qaragon determining winners. |
| **Bragging rights** | Recognition without a material reward. It still permits an activity to have results and winners. |
| **Sponsor** | The organization or person offering support for Contests under a sponsorship arrangement. |
| **Sponsorship** | An arrangement linking a sponsor to one Contest, with its amount, agreement and acceptance state. |
| **Sponsorship offer** | A proposed sponsorship awaiting the relevant acceptance. An offer alone does not establish an accepted sponsorship. |
| **Sponsorship term** | A reusable versioned condition that can be included in an agreement. |
| **Sponsorship agreement** | The amount, terms snapshot, branding conditions and acceptance governing a particular sponsorship. |
| **Terms snapshot** | The agreed versions and text of conditions preserved with an agreement independently of later edits to reusable terms. |
| **Sponsorship package** | A budget and constraints supporting several Contest sponsorships. It is distinct from each Contest's individual sponsorship. |
| **Package agreement** | The per-Contest funding, Contest count and conditions under which a sponsorship package is used. |
| **Amount** | A monetary value together with its currency. An amount alone does not indicate that money has been transferred. |

## Events and tenant integrations

| Term | Meaning |
| --- | --- |
| **Internal event** | An occurrence communicated between platform components. Its topic or payload is not automatically part of the tenant contract. |
| **Public webhook event** | A named, documented platform occurrence available to tenants through the public webhook catalog. Its contract is distinct from an internal broker topic. |
| **Event occurrence** | One business fact being reported, independently of the number of notifications or delivery retries it produces. |
| **Webhook subscription** | A tenant's registration of a receiving endpoint and selected public event types. It is unrelated to Space membership. |
| **Webhook delivery** | The signed notification of a public event to one subscription. Retries retain that delivery's identity. |
| **Webhook envelope** | The common delivery metadata surrounding an event's typed business data, including tenant and delivery identity. |
| **Webhook batch group** | All batches publishing one tenant event occurrence, identified together so receivers can recognize their common source. |
| **Webhook batch** | One bounded portion of a tenant event's audience and payload, with a stable batch identity within its batch group. Deliveries to different subscriptions retain their own delivery identities. |
| **Webhook signing secret** | The subscription credential used to verify the authenticity and integrity of webhook deliveries. It is separate from the tenant API key. |
| **SDK** | A generated or maintained client package implementing access to Qaragon's published API contract. It does not define additional Contest rules. |
| **Qaragon MCP** | The public documentation server exposing API contracts, vocabulary and developer guidance to coding assistants. It does not perform tenant API actions. |
| **Developer plugin** | The packaged MCP connection and integration skill guiding a coding assistant in a tenant's repository. It uses the same published contracts and glossary as other developers. |

## Terms to avoid

- **IRL versus online** for the mechanic distinction: hosted play can also happen online.
- **Solo** as a synonym for all self-play: shared self-play has multiple competitors.
- **Standalone** as a synonym for self-play: existing standalone Game definitions can still require a host and a team.
- **Host** for every creator: self-play has a creator but no runtime host requirement.
- **Session** or **run** as a new platform aggregate when referring to an existing Participation's progress.
- **Mode** without qualification: play style, participant admission, claim scope and submission execution are different concerns.
- **Round** for each typed guess or answer slot: an attempt and a round are different concepts.
- **Global** without its scope: a Contest global rank and global shared content describe different boundaries.
- **Subscription** for every relationship: distinguish Space membership, Contest enrollment, following and webhook subscriptions.
- **Team** as a synonym for a competitive group: establish whether it means representation or a Game's playing roster.
- **Event** without qualification when ambiguous: a Prediction event and a public webhook event are different concepts.

## Detailed guidance

Use [Contest rules](qaragon://guides/contest-rules) for scoring and grouping behavior, [tenant integration](qaragon://guides/tenant-bff) for the platform boundary and [webhooks](qaragon://guides/webhooks) for public delivery. [Community Guess](qaragon://guides/community-guess), [Moment](qaragon://guides/moment) and the individual hosted Game guides define their respective flows. Public operation schemas remain the authority for field shapes and permitted request values.

## Automated Contest Scheduler and shared content preparation

The **Automated Contest Scheduler** creates one ordinary Contest in a Space on selected days when that Space has no ordinary, non-cancelled Contest for that day. It selects one enabled type at random and uses existing content; missing content skips the day without backfill. Existing Contests take precedence, including manual ones.

**Scheduler settings** are the tenant's opt-in configuration: Space, creator, local creation/start clocks, time zone, weekdays and eligible primitive/variant pairs. A **scheduler run** records one attempted local date and its selected type/outcome, including a skipped or failed day. It is execution history, not another Contest or reward lifecycle. **Contest schedule** retains its lifecycle meaning: one Contest's participation opening and closing times. The public automation resource is `/v2/contest-schedulers`, distinct from those participation times.

The Daily date uses the opening for Quiz/word Contests and the submission deadline for Prediction, whose entry opens immediately. **FansContest Daily** is Branda's presentation of Contests in its configured General Spaces; tenants own their separate participants/results even when source content is shared.

**Library generation** is Qaragon's scheduled growth of its shared content library. Each enabled run requests a configured community-board batch independently of library size, previous selections or pending reviews. Valid, deduplicated candidates are appended pending publication approval. Generation, approval and the Automated Contest Scheduler are independent; Daily selects approved compatible boards randomly without usage accounting. See [Automated Contest Scheduler](qaragon://guides/recurring-contests) and [publication review](qaragon://guides/content-review).
