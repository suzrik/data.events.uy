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
| `date` | sí | Fecha en formato año-mes-día. Si el organizador solo anunció el mes, año-mes. En un evento de varios días, el primer día. | `2026-10-11` o `2027-04` |
| `end_date` | no | Último día de un evento de varios días, en formato año-mes-día. Tiene que ser posterior a `date`, a lo sumo 31 días después, y `date` tiene que incluir el día. | `2026-10-12` |
| `status` | no | `tentative` si el organizador todavía no confirmó la fecha. El otro valor, `confirmed`, se entiende si no escribís la línea. | `tentative` |
| `entry` | no | Quién puede participar: `open`, `license` o `elite`. Sin esta línea, el evento se considera abierto. | `license` |
| `sport` | sí | `road` (ciclismo de ruta), `mtb` (mountain bike), `gravel`, `run` (running de calle), `trail` (trail running), `tri` (triatlón y duatlón) o `roll` (patinaje). Si una carrera de ciclismo tiene dos disciplinas, van las dos entre corchetes. | `run` o `[mtb, gravel]` |
| `city` | sí | Ciudad, escrita igual que en [cities.yaml](../cities.yaml). Si falta, agregala a esa lista en el mismo pull request. | `Montevideo` |
| `venue` | no | Lugar de largada. Hasta 120 caracteres. | `Rambla de Pocitos` |
| `distances` | no | Hasta 6 distancias de hasta 20 caracteres cada una, entre corchetes y separadas por comas. | `[10K, 5K]` |
| `links.site` | uno de los tres | Sitio oficial del organizador. Tiene que empezar con `https://`. | `https://example.org` |
| `links.instagram` | uno de los tres | Instagram oficial del organizador. Tiene que empezar con `https://`. | `https://instagram.com/example` |
| `links.facebook` | uno de los tres | Facebook oficial del organizador. Tiene que empezar con `https://`. | `https://facebook.com/example` |
| `links.register` | no | Página de inscripción. Tiene que empezar con `https://`. | `https://example.org/inscripcion` |
| `description` | sí | Descripción breve, hasta 500 caracteres por idioma. Es obligatorio `es` o `en`; el otro idioma es opcional. | ver la plantilla |

## Fechas y participación

Las líneas `end_date`, `status` y `entry` son opcionales: si no las necesitás, no las pongas. En la plantilla vienen comentadas, escritas así: `#end_date: 2026-10-12`. Para usar una, borrá el `#` que tiene adelante.

**Solo se anunció el mes.** Si el organizador todavía no dijo el día, escribí solo el año y el mes:

```yaml
date: 2027-04
```

Cuando el organizador anuncie el día, corregí la fecha con un pull request nuevo (ver «Corregir o cancelar un evento»).

**El evento dura varios días.** En `date` va el primer día y en `end_date` el último (hasta 31 días después). Los dos llevan día, así que no se combina con una fecha de solo mes:

```yaml
date: 2026-10-10
end_date: 2026-10-12
```

**El organizador todavía no confirmó la fecha.** Por ejemplo, si la tomaste de la edición del año pasado o de un anuncio preliminar, agregá `status: tentative`. En la descripción, aclará de dónde sale la fecha:

```yaml
date: 2027-04-18
status: tentative
```

**La carrera tiene MTB y gravel.** Las tres disciplinas de ciclismo, `road`, `mtb` y `gravel`, se pueden combinar: si el evento tiene un recorrido o una categoría de cada una, escribí las dos entre corchetes y va a aparecer en los dos filtros. Los demás deportes van de a uno.

```yaml
sport: [mtb, gravel]
```

**Quién puede participar.** Con `entry` indicás el tipo de participación:

- `open`: puede participar cualquiera que se inscriba. Es lo que se entiende si no escribís la línea.
- `license`: pueden participar aficionados que cumplan una condición, por ejemplo una licencia por el día o por la temporada, o ser socio de un club.
- `elite`: corren solo deportistas federados o profesionales; para el resto es un evento para ir a ver.

```yaml
entry: license
```

**Cuando falta información.** Un `date` de solo mes se puede combinar con `status: tentative` si el organizador todavía no anunció la edición y el mes sale de la edición del año pasado; en ese caso no adivines el día. Como mínimo hay que saber el mes: un evento que todavía no tiene ni mes no se puede agregar.

```yaml
date: 2027-04
status: tentative
```

## Errores frecuentes

- Un texto con dos puntos seguidos de un espacio (`: `) va entre comillas dobles: `name: "Vuelta Ciclista: Etapa 1"`.
- Un texto con ` #` también va entre comillas dobles: `name: "Fecha #3 Campeonato Nacional"`. Sin comillas, todo lo que sigue al `#` se pierde.
- Los enlaces son direcciones completas que empiezan con `https://`, no un usuario como `@corridarambla`.
- La descripción lleva el idioma en una línea aparte, con sangría: `es: Texto en español.`
- Las distancias van entre corchetes, separadas por comas: `[10K, 5K]`.
- `sport: bike` ya no existe: el ciclismo se indica como `road`, `mtb` o `gravel`.
- `entry` se escribe en inglés, con `s`: `license`, no `licence`.

## Salidas grupales

Una salida grupal es una salida en bici o a correr que se repite: todos los sábados, los martes y jueves, el tercer domingo de cada mes. No tiene fecha, tiene días. Va en un archivo aparte, en la carpeta `rides/`, y aparece en la página «Salidas» del sitio, no en el calendario.

1. Copiá la [plantilla de salida](../templates/ride.yaml) a `rides/<nombre-de-la-salida>.yaml`.
2. Completá los campos. `name`, `sport`, `city`, `venue`, `links` y `description` son los mismos que en un evento. En lugar de `date` van estos:

| Campo | Obligatorio | Qué poner | Ejemplo |
|---|---|---|---|
| `days` | sí | Los días en que sale, entre corchetes: `mon` (lunes), `tue` (martes), `wed` (miércoles), `thu` (jueves), `fri` (viernes), `sat` (sábado), `sun` (domingo). | `[tue, thu]` |
| `week` | no | Solo si sale una vez por mes: `1`, `2`, `3`, `4` o `last`. Con `days: [sun]` y `week: 3` es el tercer domingo de cada mes. | `3` |
| `time` | no | Hora de salida, en 24 horas y entre comillas. | `"18:45"` |

3. Enviá un pull request, igual que con un evento.

En la descripción contá lo que alguien necesita saber antes de ir: ritmo o nivel, distancia o recorridos habituales, si es gratis, si hay que avisar antes, qué llevar. Si la salida deja de hacerse, enviá un pull request que elimine el archivo.

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
