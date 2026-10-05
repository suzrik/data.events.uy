# Cómo agregar un evento

events.uy es un calendario abierto de eventos deportivos de Uruguay. Cualquier persona puede proponer un evento: organizadores, clubes y participantes.

## Qué necesitás

- Una cuenta gratuita en [GitHub](https://github.com/signup).
- Al menos un enlace oficial del organizador: sitio web, Instagram o Facebook.

## Desde el navegador, sin instalar nada

1. Abrí el formulario de alta. En events.uy, usá el botón **Crear evento en GitHub** que está arriba de esta página: abre el formulario con la plantilla ya cargada. En GitHub, entrá a la carpeta `events/`, elegí **Add file → Create new file** y copiá el contenido de la [plantilla](../templates/event.yaml).
2. Poné el nombre del archivo: el año del evento, una barra, y el nombre del evento en minúsculas, sin tildes ni espacios, con guiones, terminado en `.yaml`. Ejemplo: `2026/corrida-rambla-10k.yaml`. Al escribir la barra, GitHub crea la carpeta del año si todavía no existe. El año de la carpeta tiene que ser el de la fecha del evento.
3. Completá los campos y borrá las líneas opcionales que no uses. La tabla de abajo explica cada campo.
4. Hacé clic en **Commit changes…** y después en **Propose changes**.
5. En la pantalla siguiente, hacé clic en **Create pull request**.

Listo. Una verificación automática revisa el formato en un par de minutos. Si encuentra un error, lo vas a ver en la misma página, con el nombre del campo a corregir.

## Con Git

1. Hacé un fork del repositorio y clonalo.
2. Copiá `templates/event.yaml` a `events/<año>/<nombre-del-evento>.yaml` y completalo.
3. Verificá el formato con `go run ./cmd/validate` (requiere [Go](https://go.dev/dl/)).
4. Abrí un pull request.

## Campos

| Campo | Obligatorio | Qué poner | Ejemplo |
|---|---|---|---|
| `name` | sí | Nombre del evento, tal como lo publica el organizador. Hasta 120 caracteres. | `Corrida Rambla 10K` |
| `date` | sí | Fecha en formato año-mes-día. | `2026-10-11` |
| `sport` | sí | `bike` (ciclismo), `run` (running) o `roll` (patinaje). | `run` |
| `city` | sí | Ciudad, escrita igual que en [cities.yaml](../cities.yaml). Si falta, agregala a esa lista en el mismo pull request. | `Montevideo` |
| `venue` | no | Lugar de largada. Hasta 120 caracteres. | `Rambla de Pocitos` |
| `distances` | no | Hasta 6 distancias de hasta 20 caracteres cada una, entre corchetes y separadas por comas. | `[10K, 5K]` |
| `links.site` | uno de los tres | Sitio oficial del organizador. Tiene que empezar con `https://`. | `https://example.org` |
| `links.instagram` | uno de los tres | Instagram oficial del organizador. Tiene que empezar con `https://`. | `https://instagram.com/example` |
| `links.facebook` | uno de los tres | Facebook oficial del organizador. Tiene que empezar con `https://`. | `https://facebook.com/example` |
| `links.register` | no | Página de inscripción. Tiene que empezar con `https://`. | `https://example.org/inscripcion` |
| `description` | sí | Descripción breve, hasta 500 caracteres por idioma. Es obligatorio `es` o `en`; el otro idioma es opcional. | ver la plantilla |

## Errores frecuentes

- Un texto con dos puntos seguidos de un espacio (`: `) va entre comillas dobles: `name: "Vuelta Ciclista: Etapa 1"`.
- Un texto con ` #` también va entre comillas dobles: `name: "Fecha #3 Campeonato Nacional"`. Sin comillas, todo lo que sigue al `#` se pierde.
- Los enlaces son direcciones completas que empiezan con `https://`, no un usuario como `@corridarambla`.
- La descripción lleva el idioma en una línea aparte, con sangría: `es: Texto en español.`
- Las distancias van entre corchetes, separadas por comas: `[10K, 5K]`.

## Reglas

- Solo eventos con un enlace oficial del organizador.
- No copiamos eventos de calendarios agregadores: la fuente tiene que ser el organizador.
- Escribí la descripción con tus palabras, o usá el texto del organizador si sos parte de la organización.
- Al enviar un evento aceptás que se publique bajo la licencia [CC BY 4.0](../LICENSE).

## Qué pasa después

1. La verificación automática revisa el formato.
2. Una persona del equipo revisa los enlaces y aprueba el pull request.
3. El evento aparece en events.uy unos minutos después de la aprobación.

## Corregir o cancelar un evento

Abrí el archivo del evento en GitHub y hacé clic en el lápiz. Si GitHub te lo pide, hacé clic en **Fork this repository**. Cambiá lo necesario y enviá un pull request. Si el evento se canceló, enviá un pull request que elimine el archivo.
