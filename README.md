# GoTrah

[![CI](https://github.com/686f6c61/GoTrash/actions/workflows/ci.yml/badge.svg)](https://github.com/686f6c61/GoTrash/actions/workflows/ci.yml)
[![Release](https://github.com/686f6c61/GoTrash/actions/workflows/release.yml/badge.svg)](https://github.com/686f6c61/GoTrash/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/v/release/686f6c61/GoTrash)](https://github.com/686f6c61/GoTrash/releases)

> Un CLI nativo para macOS que localiza residuos habituales de desarrollo, calcula cuánto ocupan, detecta a qué proyecto pertenecen y te permite revisarlos o eliminarlos con una interfaz guiada, visual y segura.

`GoTrah` es el nombre del CLI y de la interfaz.  
El repositorio vive en [686f6c61/GoTrash](https://github.com/686f6c61/GoTrash).

GoTrah está pensado para ese momento en el que sabes que el disco se está llenando, pero no quieres perder tiempo entrando carpeta por carpeta, improvisando comandos o borrando cosas a ciegas. En lugar de eso, te ofrece un flujo claro: escanear, entender qué ocupa espacio, relacionarlo con su proyecto y decidir con calma qué merece la pena limpiar.

## Qué hace

GoTrah recorre una o varias rutas, detecta carpetas típicas de "basura de desarrollo" y las presenta con contexto suficiente para que puedas tomar decisiones informadas. La idea no es solo listar carpetas grandes, sino ayudarte a entender por qué están ahí, de qué proyecto vienen y cuánto impacto real tiene borrarlas.

Entre otras, busca carpetas como:

- `node_modules`
- `.venv`
- `venv`
- `env`
- `__pycache__`
- `.pytest_cache`
- `.mypy_cache`
- `.ruff_cache`
- `.tox`
- `dist`
- `build`
- `.next`
- `.nuxt`
- `.svelte-kit`
- `.turbo`
- `.cache`
- `.parcel-cache`
- `Pods`
- `DerivedData`
- `.gradle`

Además, GoTrah no se limita a encontrar nombres conocidos. También intenta convertir ese escaneo en una revisión útil para alguien que trabaja todos los días con proyectos reales, tooling variado y muchas rutas repartidas por el disco.

- Calcula el tamaño aproximado de cada coincidencia para que puedas priorizar primero lo que realmente libera espacio.
- Intenta inferir el proyecto padre con marcadores como `.git`, `package.json`, `pyproject.toml` o `go.mod`, así no ves solo una carpeta suelta, sino su contexto.
- Ordena los resultados por tamaño para que la primera decisión sea casi siempre la más rentable.
- Muestra resúmenes por proyecto y por tipo de carpeta, algo especialmente útil cuando aparecen cientos de coincidencias.
- Permite seleccionar con flechas, checkboxes y búsqueda instantánea, de forma que no dependas de memorizar flags o escribir índices largos.
- Exporta todo a CSV para revisar o auditar resultados fuera de la terminal.
- Protege rutas sensibles del sistema y evita borrados peligrosos en zonas donde un fallo sería caro.

## Pensado para

GoTrah está especialmente orientado a perfiles técnicos que trabajan con muchos proyectos, varios stacks y bastante tooling local. Cuanto más heterogéneo sea tu entorno de desarrollo, más sentido tiene una herramienta que unifique la limpieza en un solo flujo.

- Está pensado para desarrolladores que acumulan `node_modules`, `venv`, builds y cachés sin darse cuenta con el paso de las semanas.
- Encaja bien en equipos o freelancers que trabajan en varias carpetas de proyectos y necesitan entender rápido qué pertenece a qué.
- Tiene mucho valor en máquinas macOS donde `DerivedData`, `Pods` o cachés de tooling terminan comiéndose una parte importante del SSD.

## Estado

La base del proyecto ya está operativa y usable. La versión `0.2.0` ya no es solo un primer prototipo funcional: además del flujo principal de limpieza, incorpora una capa más seria de hardening técnico, validación y seguridad operativa.

Versión actual: `0.2.0`

Plataforma objetivo:

- `macOS`, que es donde el flujo de uso y la experiencia están más cuidados ahora mismo.

Estado actual:

- Funciona desde código fuente y ya se puede usar como herramienta real para revisar y borrar carpetas pesadas.
- El flujo guiado ya está operativo y es la forma recomendada de empezar a usarlo.
- Las protecciones de borrado y escaneo ya están más endurecidas y cubren mejor rutas sensibles del sistema.
- La CI y el proceso de release ya incluyen validaciones de seguridad para reducir regresiones antes de publicar nuevas versiones.
- La distribución por Homebrew se puede añadir después, cuando la línea `0.2.x` quede más asentada.

## Instalación

Por ahora, la forma recomendada es usarlo directamente desde el repositorio o compilar un binario local. Eso permite iterar rápido sobre el proyecto mientras se estabiliza la `0.2.x`, sin depender todavía de una distribución externa.

### Ejecutarlo directamente

Esta opción es la más cómoda si quieres probar el flujo sin instalar nada de forma permanente en tu sistema.

```bash
git clone https://github.com/686f6c61/GoTrash.git
cd GoTrash
go run .
```

### Compilar binario

Si prefieres ejecutarlo como un comando local, puedes compilar un binario y lanzarlo directamente. Es una buena opción si vas a usarlo varias veces desde la misma máquina.

```bash
git clone https://github.com/686f6c61/GoTrash.git
cd GoTrash
go build -o gotrah .
./gotrah
```

### Descargar desde GitHub Releases

Cuando publiques una versión etiquetada, GoTrah también puede distribuirse desde la sección de releases del repositorio. Eso hace más fácil descargar un binario preparado sin tener que compilarlo a mano y encaja muy bien con el flujo típico de una herramienta CLI para macOS.

Las releases públicas estarán aquí:

- [GitHub Releases](https://github.com/686f6c61/GoTrash/releases)

## Uso rápido

### Arranque guiado

El arranque guiado es el punto de entrada principal. Está pensado para que puedas empezar sin aprender flags, sin recordar sintaxis y sin miedo a equivocarte en la primera pasada.

```bash
go run .
```

Si no pasas rutas ni flags, GoTrah abre un menú guiado para:

- Elegir qué escanear: `HOME`, una carpeta, varias carpetas o todo el disco completo.
- Elegir el perfil de búsqueda: general, JavaScript, Python, Apple/Xcode o personalizado.
- Definir un tamaño mínimo para quedarte solo con carpetas que realmente merezcan atención.
- Revisar los resultados antes de borrar nada, con más contexto que una simple lista de rutas.

### Ejemplos útiles

Cuando ya sabes lo que quieres hacer, los flags te permiten ir más rápido o automatizar parte del flujo. Estos son algunos ejemplos que cubren los casos más habituales.

```bash
go run . ~/Code ~/Projects
go run . --interactive
go run . / --min-size 500MB --show-errors
go run . --names node_modules,.venv,DerivedData
go run . --csv ./reporte.csv
go run . --delete-all
```

### Opciones principales

Estas son las banderas más importantes del comando actual. No hace falta usarlas todas desde el principio, pero conviene conocerlas porque cambian mucho la velocidad con la que puedes revisar resultados.

```text
gotrah [path ...]
```

- `--names` te permite pasar una lista de nombres separados por comas para buscar solo los tipos de carpeta que te interesan.
- `--min-size` fija un tamaño mínimo por carpeta, por ejemplo `500MB`, `2GB` o `120M`, para no perder tiempo con coincidencias pequeñas.
- `--interactive`, `-i` fuerza la selección interactiva de qué borrar incluso cuando no has arrancado desde el menú guiado.
- `--delete-all` borra todos los resultados detectados tras pedir confirmación final.
- `--yes`, `-y` salta la confirmación final al borrar, pensado para flujos más directos y conscientes.
- `--show-errors` muestra errores de permisos y acceso detectados durante el escaneo, algo útil cuando revisas rutas amplias o todo el disco.
- `--csv` exporta todas las coincidencias a un fichero CSV para revisarlas después fuera de la terminal.

## Controles interactivos

La TUI está pensada para revisar muchos resultados sin agobio. En lugar de obligarte a recordar números o a pasar por una salida interminable, te deja moverte, filtrar y marcar con un flujo más visual.

En las pantallas TUI:

- `↑` `↓` o `j` `k` mueven la selección actual por la lista.
- `espacio` marca o desmarca la entrada activa.
- `/` o `tab` activa la búsqueda para filtrar resultados sin salir de la pantalla.
- `esc` o `enter` dentro de la búsqueda salen del modo de edición y te devuelven a la lista.
- `a` marca todos los resultados visibles con el filtro actual.
- `n` limpia la selección visible para rehacerla rápido.
- `enter` confirma la selección y continúa con el siguiente paso del flujo.

La búsqueda filtra por:

- El tipo de carpeta, para localizar rápido cosas como `node_modules` o `DerivedData`.
- El nombre corto del proyecto, para limpiar por contexto y no solo por tamaño.
- La ruta completa, útil cuando recuerdas una zona del disco pero no el nombre exacto del proyecto.

## Exportación CSV

La exportación a CSV sirve para dos cosas: revisar los resultados con más calma y tener una base auditable cuando el volumen de coincidencias es grande. En vez de decidir todo dentro de la terminal, puedes abrir el fichero en Numbers, Excel o cualquier hoja de cálculo y ordenar por tamaño, proyecto o ruta.

Puedes guardar el listado completo así:

```bash
go run . --csv ./reporte.csv
```

Columnas exportadas:

- `index`, para conservar el orden del listado exportado.
- `type`, para saber qué tipo de carpeta se ha detectado.
- `size_bytes`, para cálculos o filtros precisos.
- `size_human`, para lectura rápida.
- `project`, para relacionar la coincidencia con su proyecto probable.
- `path`, para tener la ruta completa exacta.

Si no pasas una ruta, GoTrah genera un fichero con nombre parecido a este en el directorio actual:

```text
gotrah-report-20260402-153000.csv
```

## Cómo decide a qué proyecto pertenece una carpeta

Uno de los objetivos del proyecto es que no veas simplemente "`/ruta/larga/node_modules`", sino que puedas entender a qué proyecto pertenece esa carpeta. Para conseguirlo, GoTrah sube por los directorios padre hasta encontrar una señal razonable de raíz de proyecto.

Los marcadores más habituales son:

- `.git`
- `package.json`
- `pnpm-workspace.yaml`
- `pyproject.toml`
- `requirements.txt`
- `go.mod`
- `Cargo.toml`
- `Gemfile`
- `Podfile`

Si no encuentra ninguno, usa como referencia la carpeta inmediatamente superior. No es una inferencia perfecta en todos los casos, pero en la práctica mejora mucho la capacidad de decidir qué borrar con seguridad.

## Seguridad y límites

GoTrah intenta ser útil sin animarte a borrar a ciegas. Por eso incorpora varias salvaguardas desde esta primera versión, especialmente pensadas para macOS y para rutas donde un borrado accidental sería mala idea.

Hay varias protecciones ya activas:

- Pide confirmación explícita escribiendo `BORRAR`, para evitar un Enter accidental en un momento delicado.
- Omite rutas protegidas del sistema cuando detecta que no deberían formar parte de una limpieza normal de desarrollo.
- Si eliges una ruta protegida, la marca y la salta en el borrado, en lugar de tratarla como un error confuso.
- Al escanear `/`, tolera errores de permisos y puede mostrarlos si usas `--show-errors`, algo normal en macOS moderno.

Desde `0.2.0`, esas protecciones están más alineadas entre sí. La política de rutas protegidas ya no depende tanto del flujo concreto que uses, y el resumen previo al borrado refleja mejor qué se va a eliminar realmente y qué se va a omitir.

También conviene tener claros algunos límites de la versión actual:

- El borrado actual es real y no mueve nada a la papelera, así que conviene revisar bien la selección.
- Algunas rutas de herramientas del sistema o de paquetes globales pueden requerir permisos elevados o estar bloqueadas por diseño.
- El cálculo de espacio liberado es aproximado, porque se basa en el tamaño estimado durante el escaneo.

## Flujo recomendado

Si es la primera vez que lo usas, este es el recorrido que mejor equilibrio da entre seguridad, velocidad y claridad:

1. Lanza `go run .`
2. Escanea `HOME` o tus carpetas de desarrollo principales antes de pasar a algo más agresivo como `/`.
3. Revisa primero el resumen por proyecto para detectar rápido qué zonas del disco concentran más peso.
4. Filtra o busca en la TUI para quedarte solo con lo que reconoces bien.
5. Borra únicamente lo que entiendas con claridad y te conste que es regenerable.
6. Exporta a CSV si quieres revisar con más calma o guardar un informe antes de limpiar.

## Changelog

El historial de versiones está en [CHANGELOG.md](./CHANGELOG.md), incluyendo la release actual `0.2.0` y la release inicial `0.1.0`.
