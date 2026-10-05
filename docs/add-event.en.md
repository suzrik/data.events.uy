# How to add an event

events.uy is an open calendar of sports events in Uruguay. Anyone can propose an event: organizers, clubs and participants.

## What you need

- A free [GitHub](https://github.com/signup) account.
- At least one official organizer link: website, Instagram or Facebook.

## From the browser, nothing to install

1. Open the new-event form. On events.uy, use the **Create event on GitHub** button at the top of this page: it opens the form with the template already filled in. On GitHub, open the `events/` folder, then the folder for the event's year, choose **Add file → Create new file** and copy the contents of the [template](../templates/event.yaml).
2. Name the file: the event name in lowercase, without accents or spaces, with hyphens, ending in `.yaml`. Example: `corrida-rambla-10k.yaml`.
3. Fill in the fields. The table below explains each one.
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
| `date` | yes | Date as year-month-day. | `2026-10-11` |
| `sport` | yes | `bike` (cycling), `run` (running) or `roll` (skating). | `run` |
| `city` | yes | City, spelled exactly as in [cities.yaml](../cities.yaml). If it is missing, add it to that list in the same pull request. | `Montevideo` |
| `venue` | no | Start venue. Up to 120 characters. | `Rambla de Pocitos` |
| `distances` | no | Up to 6 distances, in square brackets, separated by commas. | `[10K, 5K]` |
| `links.site` | one of the three | Official organizer website. Must start with `https://`. | `https://example.org` |
| `links.instagram` | one of the three | Official organizer Instagram. | `https://instagram.com/example` |
| `links.facebook` | one of the three | Official organizer Facebook. | `https://facebook.com/example` |
| `links.register` | no | Registration page. | `https://example.org/inscripcion` |
| `description` | yes | Short description, up to 500 characters per language. Either `es` or `en` is required; the other language is optional. | see the template |

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

Open the event file on GitHub, click the pencil, change what is needed and send a pull request. If the event was cancelled, send a pull request that deletes the file.
