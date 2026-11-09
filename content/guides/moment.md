# Moment: media-led Quiz contests

Moment is an MCQ variant (`source_type: "mc"`, `source_variant: "moment"`), not a hosted Game. It uses the existing Contest lifecycle, Space membership, enrollment, shared submissions, Puzzle grading, ranking and rewards. A normal Quiz can also carry optional image or video media; every Moment question must have one supported media reference.

## Prepare a source

Create an owned question sheet with `POST /v2/puzzle-sheets`. Each question uses the ordinary MCQ body with `number`, `puzzle_type: "mc"`, `question`, an `options` object, `solution`, positive `points`, and:

```json
{"media":{"type":"image","url":"https://your-media.example/frame.jpg"}}
```

Video uses the same shape:

```json
{"media":{"type":"video","url":"https://your-media.example/clip.mp4"}}
```

Set the sheet's root `format` to `moment`. Upload assets through your application's media service first; the API accepts HTTP(S) image/video references and rejects local file URIs and unsupported media types. One sheet may mix image and video questions. An empty sheet is a valid creator draft, but a playable Moment source must contain supported media on every MCQ question. Puzzle stores the references without downloading or transcoding the assets.

List sources with `GET /v2/puzzle-sheets?format=moment`. Cloning and owned-sheet sampling retain the format and media. The existing community categories, topic configuration and search-bank endpoints accept `format=moment`; community questions must be added to the bank and indexed before those topics appear. Moment has no runtime AI or automatic content generation.

## Host and play

Create an ordinary Contest with its Space, timing, rewards and MCQ question count:

```json
{"source_type":"mc","source_id":"<sheet-id>","source_variant":"moment","no_of_puzzles":5}
```

Read the Contest's existing source endpoint (`GET /v2/contest-puzzles/{contest-id}`). It retains media and omits correct answers. Existing MCQ source preparation remains asynchronous; do not interpret Contest creation as confirmation that Puzzle has finished preparing its source.

Render media according to `media.type`: show images uncropped and provide an appropriate player for video clips, alongside the question and selectable answer options. Loading failures should offer retry without losing selections. The current FansContest mobile UI implements image upload/display; video upload/playback is a separate client change.

Submit all selected option keys through **POST /v2/submissions**, with a Contest target and the usual MCQ `solutions` payload. Do not submit media again, create a Moment-specific endpoint, or grade on the client. Keep the same idempotency key for a retry and use the existing completion/receipt flow.

## Read results

Contest details, lists and feed Contest metadata expose the persisted `source_variant: "moment"` for labels through the normal Contest response. MCQ scoring remains unchanged. The MC module validates the variant at creation; there is no separate metadata lookup. Puzzle's graded sheet preserves image/video media so participants can review the question, their selected option and the correct answer after results become available.

Use `search_api` and `describe_operation` for the complete existing sheet, Contest, source and submission contracts. Tenant credentials remain in the BFF; acting platform identities and tenant ownership use the normal platform rules.
