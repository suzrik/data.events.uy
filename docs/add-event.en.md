# How to add an event

events.uy is an open calendar of sports events in Uruguay. Anyone can propose an event: organizers, clubs and participants.

## What you need

- A free [GitHub](https://github.com/signup) account.
- At least one official organizer link: website, Instagram or Facebook.

## From the browser, nothing to install

1. Open the new-event form. On events.uy, use the **Create event on GitHub** button at the top of this page: it opens the form with the template already filled in. On GitHub, open the `events/` folder, choose **Add file → Create new file** and copy the contents of the [template](../templates/event.yaml).
2. Name the file: the year of the event, a slash, and the event name in lowercase, without accents or spaces, with hyphens, ending in `.yaml`. Example: `2026/corrida-rambla-10k.yaml`. When you type the slash, GitHub creates the year folder if it does not exist yet. The year of the folder must be the year of the event's date.
3. Fill in the fields and delete the optional lines you do not use. The table below explains each field.
4. Click **Commit changes…** and then **Propose changes**.
5. On the next screen, click **Create pull request**.

That's it. An automatic check reviews the format within a couple of minutes. If it finds an error, you will see it on the same page, with the name of the field to fix.

## With Git

1. Fork the repository and clone it.
2. Copy `templates/event.yaml` to `events/<year>/<event-name>.yaml` and fill it in.
3. Check the format with `go run ./cmd/validate` (requires [Go](https://go.dev/dl/)).
4. Open a pull request.

## Fields

| Field | Required | What to write | Example |
|---|---|---|---|
| `name` | yes | Event name, as the organizer publishes it. Up to 120 characters. | `Corrida Rambla 10K` |
| `date` | yes | Date as year-month-day. If the organizer has only announced the month, year-month. For a multi-day event, the first day. | `2026-10-11` or `2027-04` |
| `end_date` | no | Last day of a multi-day event, as year-month-day. Must be after `date`, at most 31 days later, and `date` must include the day. | `2026-10-12` |
| `status` | no | `tentative` if the organizer has not confirmed the date yet. The other value, `confirmed`, is assumed if you leave the line out. | `tentative` |
| `entry` | no | Who can take part: `open`, `license` or `elite`. Without this line, the event counts as open. | `license` |
| `sport` | yes | `bike` (cycling), `run` (running) or `roll` (skating). | `run` |
| `city` | yes | City, spelled exactly as in [cities.yaml](../cities.yaml). If it is missing, add it to that list in the same pull request. | `Montevideo` |
| `venue` | no | Start venue. Up to 120 characters. | `Rambla de Pocitos` |
| `distances` | no | Up to 6 distances of up to 20 characters each, in square brackets, separated by commas. | `[10K, 5K]` |
| `links.site` | one of the three | Official organizer website. Must start with `https://`. | `https://example.org` |
| `links.instagram` | one of the three | Official organizer Instagram. Must start with `https://`. | `https://instagram.com/example` |
| `links.facebook` | one of the three | Official organizer Facebook. Must start with `https://`. | `https://facebook.com/example` |
| `links.register` | no | Registration page. Must start with `https://`. | `https://example.org/inscripcion` |
| `description` | yes | Short description, up to 500 characters per language. Either `es` or `en` is required; the other language is optional. | see the template |

## Dates and participation

The `end_date`, `status` and `entry` lines are optional: if you do not need them, leave them out. In the template they come commented out, written like this: `#end_date: 2026-10-12`. To use one, delete the `#` in front of the line.

**Only the month has been announced.** If the organizer has not said the day yet, write just the year and month:

```yaml
date: 2027-04
```

When the organizer announces the day, fix the date with a new pull request (see "Fixing or cancelling an event").

**The event lasts several days.** `date` is the first day and `end_date` is the last (at most 31 days later). Both need the day, so this does not combine with a month-only date:

```yaml
date: 2026-10-10
end_date: 2026-10-12
```

**The organizer has not confirmed the date yet.** For example, you took it from last year's edition or from a preliminary announcement: add `status: tentative`. In the description, say where the date comes from:

```yaml
date: 2027-04-18
status: tentative
```

**Who can take part.** `entry` says what kind of participation the event has:

- `open`: anyone who registers can take part. This is assumed if you do not write the line.
- `license`: amateurs may take part if they meet a condition, for example a license for the day or the season, or membership of a club.
- `elite`: only federated or professional athletes race; for everyone else it is an event to watch.

```yaml
entry: license
```

**When information is missing.** A month-only `date` can be combined with `status: tentative` when the organizer has not announced the edition and the month comes from last year's edition; in that case do not guess the day. At least the month must be known: an event with no month yet cannot be added.

```yaml
date: 2027-04
status: tentative
```

## Common mistakes

- Text with a colon followed by a space (`: `) goes in double quotes: `name: "Vuelta Ciclista: Etapa 1"`.
- Text with ` #` goes in double quotes too: `name: "Fecha #3 Campeonato Nacional"`. Without quotes, everything after the `#` is lost.
- Links are full addresses that start with `https://`, not a handle such as `@corridarambla`.
- The description has the language on its own indented line: `es: Texto en español.`
- Distances go in square brackets, separated by commas: `[10K, 5K]`.
- `entry` uses the American spelling: `license`, not `licence`.

## Rules

- Only events with an official organizer link.
- We do not copy events from aggregator calendars: the source must be the organizer.
- Write the description in your own words, or use the organizer's text if you are part of the organization.
- By submitting an event you agree to publish it under the [CC BY 4.0](../LICENSE) license.

## What happens next

1. The automatic check reviews the format.
2. A team member checks the links and approves the pull request.
3. The event appears on events.uy a few minutes after approval.

## Fixing or cancelling an event

Open the event file on GitHub and click the pencil. If GitHub asks, click **Fork this repository**. Change what is needed and send a pull request. If the event was cancelled, send a pull request that deletes the file.
