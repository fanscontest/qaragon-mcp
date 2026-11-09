# Moment: image-led Quiz contests

Moment is an MCQ variant (`source_type: "mc"`, `source_variant: "moment"`), not a hosted Game. It uses the existing Contest lifecycle, Space membership, enrollment, shared submissions, Puzzle grading, ranking and rewards. A normal Quiz can also carry optional images; every Moment question must have one.

## Prepare a source

Create an owned question sheet with `POST /v2/puzzle-sheets`. Each question uses the ordinary MCQ body with `number`, `puzzle_type: "mc"`, `question`, an `options` object, `solution`, positive `points`, and:

```json
{"media":{"type":"image","url":"https://your-media.example/frame.jpg"}}
```

Set the sheet's root `format` to `moment`. Upload images through your application's media service first; local file URIs are not accepted. Images are the only supported medium. An empty sheet is a valid creator draft, but a playable Moment source must contain image-bearing questions.

List sources with `GET /v2/puzzle-sheets?format=moment`. Cloning and owned-sheet sampling retain the format and images. The existing community categories, topic configuration and search-bank endpoints accept `format=moment`; community questions must be added to the bank and indexed before those topics appear. Moment has no runtime AI or automatic content generation.

## Host and play

Create an ordinary Contest with its Space, timing, rewards and MCQ question count:

```json
{"source_type":"mc","source_id":"<sheet-id>","source_variant":"moment","no_of_puzzles":5}
```

Read the Contest's existing source endpoint (`GET /v2/contest-puzzles/{contest-id}`). It retains images and omits correct answers. Existing MCQ source preparation remains asynchronous; do not interpret Contest creation as confirmation that Puzzle has finished preparing its source.

Show each image uncropped, its question and selectable answer options. Images may have different aspect ratios; loading failures should offer retry. The mobile experience can be immersive without changing scoring or timing.

Submit all selected option keys through **POST /v2/submissions**, with a Contest target and the usual MCQ `solutions` payload. Do not submit media again, create a Moment-specific endpoint, or grade on the client. Keep the same idempotency key for a retry and use the existing completion/receipt flow.

## Read results

Contest details, lists and feed Contest metadata expose the persisted `source_variant: "moment"` for labels through the normal Contest response. MCQ scoring remains unchanged. The MC module validates the variant at creation; there is no separate metadata lookup. Puzzle's graded sheet preserves images so participants can review the question, their selected option and the correct answer after results become available.

Use `search_api` and `describe_operation` for the complete existing sheet, Contest, source and submission contracts. Tenant credentials remain in the BFF; acting platform identities and tenant ownership use the normal platform rules.
