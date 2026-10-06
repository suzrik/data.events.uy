# events.uy — datos abiertos / open data

Calendario abierto de eventos deportivos de Uruguay: ciclismo, running, trail, triatlón y patinaje. Estos archivos alimentan [events.uy](https://events.uy).

Open calendar of sports events in Uruguay: cycling, running, trail running, triathlon and skating. These files feed [events.uy](https://events.uy).

## Agregar un evento / Add an event

- Español: [docs/add-event.es.md](docs/add-event.es.md)
- English: [docs/add-event.en.md](docs/add-event.en.md)

## Estructura / Layout

| Ruta / Path | Contenido / Contents |
|---|---|
| `events/<año>/<evento>.yaml` | Un archivo por evento / One file per event |
| `cities.yaml` | Ciudades permitidas / Allowed cities |
| `templates/event.yaml` | Plantilla / Template |
| `cmd/validate` | Verificador de formato / Format checker |

## Verificar / Validate

    go run ./cmd/validate

## Licencia / License

- Datos y textos / Data and texts (`events/`, `cities.yaml`, `docs/`, `templates/`): [CC BY 4.0](LICENSE). Atribución / Attribution: «events.uy contributors», con enlace a este repositorio / with a link to this repository.
- Código / Code (`cmd/`, `internal/`): [MIT](LICENSE-CODE).
