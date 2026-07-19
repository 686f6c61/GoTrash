# Changelog

Todos los cambios relevantes de este proyecto se documentan aquí.

Este proyecto sigue una convención simple de versiones semánticas.

## [0.2.0] - 2026-07-19

Release centrada en endurecimiento técnico, seguridad operativa y calidad del proyecto.

### Added

- Nuevo paquete interno para centralizar la política de rutas protegidas y reutilizarla entre escaneo y borrado.
- Nuevos tests para guardrails de borrado, rutas protegidas y utilidades de UI.
- Cobertura básica añadida para paquetes que antes no tenían tests propios.
- Informe de auditoría técnica y de seguridad documentado en el repositorio.

### Changed

- La confirmación de borrado ahora calcula solo las carpetas realmente eliminables y separa con más claridad las rutas que se van a omitir.
- La selección por proyectos refleja mejor cuándo un grupo incluye rutas protegidas.
- El estado del proyecto en `README` se actualiza para reflejar una línea `0.2.x` más madura que la release inicial.
- El proceso de release usa `gh release` en lugar de una acción third-party para publicar assets.

### Security

- La política de protección de rutas sensibles se unifica entre escaneo y borrado para evitar inconsistencias operativas.
- El proyecto sube su toolchain declarada a `go 1.25.8`, corrigiendo la vulnerabilidad alcanzable detectada durante la auditoría.
- Se integra `govulncheck` en CI y en el workflow de release.
- Las GitHub Actions críticas quedan pinneadas por SHA para reducir riesgo de supply chain.

### Quality

- `go test ./...` pasa con la nueva base de tests.
- `go test -cover ./...` pasa y deja cobertura explícita en `cmd`, `internal/scan`, `internal/safety` e `internal/ui`.
- `go vet ./...`, `go build ./...` y `govulncheck ./...` quedan validados sobre esta versión.

## [0.1.0] - 2026-04-02

Primera versión pública funcional de GoTrah.

### Added

- CLI en Go para macOS orientado a localizar carpetas pesadas de desarrollo.
- Escaneo de una o varias rutas con detección de directorios típicos como `node_modules`, `.venv`, `dist`, `build`, `Pods`, `DerivedData` y otros similares.
- Detección del proyecto padre usando marcadores como `.git`, `package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml` o `Podfile`.
- Ordenación de coincidencias por tamaño para priorizar el espacio recuperable.
- Menú guiado al ejecutar `go run .` sin argumentos.
- TUI interactiva con navegación por flechas, checkboxes y confirmación desde terminal.
- Búsqueda en vivo dentro de la TUI para filtrar por proyecto, tipo o ruta.
- Resumen por proyecto y por tipo de carpeta para facilitar la revisión cuando hay muchas coincidencias.
- Exportación completa a CSV.
- Selección por carpetas concretas o por proyectos completos.

### Changed

- El flujo guiado vuelve al menú principal después de completar un borrado, en lugar de cerrar el programa.
- La TUI muestra más contexto por entrada: tipo, tamaño, proyecto y ruta completa.
- La presentación visual se reforzó con badges, resaltado de la fila activa e iconos por tipo de carpeta.
- El branding visible del CLI pasó a mostrarse como `GoTrah`.

### Safety

- Confirmación explícita de borrado escribiendo `BORRAR`.
- Omisión automática de rutas protegidas del sistema para evitar borrados peligrosos.
- Tratamiento más claro de rutas omitidas en el resumen final.
- Tolerancia a errores de permisos durante el escaneo, con opción de mostrarlos cuando el usuario lo solicite.

### Notes

- La distribución inicial está pensada para ejecutarse desde código fuente o binario local.
- La integración con Homebrew queda como siguiente paso natural después de consolidar esta base.
