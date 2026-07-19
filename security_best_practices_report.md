# GoTrah Technical and Security Audit

Date: 2026-07-18

## Remediation Status

This document captures the original audit findings detected on 2026-07-18.

Status after remediation work in this workspace on 2026-07-18:

- `GT-SEC-001`: addressed
- `GT-SEC-002`: addressed
- `GT-SEC-003`: addressed
- `GT-TECH-001`: addressed
- `GT-TECH-002`: addressed

## Executive Summary

GoTrah tiene una base razonable para un CLI destructivo local, pero ahora mismo arrastra un problema de seguridad operativo en sus barreras de borrado, una vulnerabilidad real y alcanzable del runtime de Go, y varios riesgos de supply chain en GitHub Actions. Además, el pipeline de CI no está ejecutando ningún escaneo de vulnerabilidades, lo que ha permitido que una vulnerabilidad alcanzable llegue a release sin ser detectada.

El riesgo más importante no está en la TUI ni en la exportación CSV, sino en la consistencia de las protecciones alrededor del borrado. La política de "rutas protegidas" no es única ni se aplica igual en escaneo y borrado, lo que deja huecos cuando se usan raíces explícitas o nombres personalizados.

## High Severity Findings

### GT-SEC-001: Las rutas protegidas no están protegidas de forma consistente en escaneo y borrado

- Rule ID: GT-SEC-001
- Severity: High
- Location:
  - `internal/scan/scanner.go:169-176`
  - `cmd/root.go:633-652`
  - `cmd/root.go:56-63`
  - `cmd/menu.go:323-336`
- Evidence:

```go
if root == string(filepath.Separator) {
    cleanPath := filepath.Clean(path)
    for _, prefix := range protectedSystemPrefixes {
```

```go
var protectedDeletionPrefixes = []string{
    "/opt/homebrew",
    "/usr/local",
    "/opt/local",
    "/Library",
    "/System",
    "/Applications",
    "/usr",
    "/bin",
    "/sbin",
}
```

La protección de escaneo de `protectedSystemPrefixes` solo se activa cuando la raíz elegida es exactamente `/`. En cambio, la protección de borrado usa otra lista distinta y más corta, que no cubre prefijos sensibles como `/private/preboot`, `/private/var/db`, `/private/var/vm` o `/dev`.

Además, el CLI permite nombres arbitrarios mediante `--names` y también desde el perfil personalizado del menú, con lo que deja de ser un “limpiador de carpetas típicas” y pasa a poder borrar cualquier directorio cuyo nombre coincida con el patrón elegido.

- Impact:

Un usuario o una automatización puede escanear subárboles sensibles directamente, por ejemplo bajo `/private/...`, y encontrar directorios elegidos con `--names` o desde el modo personalizado. Como el borrado no bloquea todos esos prefijos, el programa puede eliminar directorios fuera del perímetro que aparenta proteger.

En una herramienta destructiva, esta inconsistencia rompe la garantía de seguridad operativa más importante del producto.

- Fix:

Crear una única política de rutas protegidas y reutilizarla tanto en escaneo como en borrado.

Cambios recomendados:

1. Centralizar la lista de prefijos protegidos en un único paquete o función compartida.
2. Aplicar esa política independientemente de la raíz elegida, no solo cuando `root == "/"`.
3. Añadir a la política de borrado los prefijos que hoy solo existen en el escaneo:
   - `/private/preboot`
   - `/private/var/db`
   - `/private/var/vm`
   - `/dev`
4. Considerar un modo explícito `--unsafe` o confirmación extra para permitir escaneos dentro de rutas protegidas.
5. Filtrar también nombres personalizados si la raíz solicitada cae dentro de zonas sensibles.

- Mitigation:

Hasta corregirlo, evita usar raíces como `/private`, `/dev` o similares y no utilices `--names` con directorios sensibles del sistema.

- False positive notes:

El impacto práctico es mayor cuando se usan nombres personalizados o cuando existen directorios coincidentes dentro de esos prefijos. Aun así, el hueco en la política de protección es real aunque no siempre sea trivial explotarlo de forma accidental.

## Medium Severity Findings

### GT-SEC-002: El binario se compila con una versión vulnerable del runtime de Go

- Rule ID: GT-SEC-002
- Severity: Medium
- Location:
  - `go.mod:3`
  - `internal/scan/scanner.go:84-127`
- Evidence:

```go
go 1.25.5
```

```go
err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
```

`govulncheck` reportó una vulnerabilidad alcanzable en la librería estándar:

- `GO-2026-4602`
- Título: `FileInfo can escape from a Root in os`
- Afecta a: `os@go1.25.5`
- Corregida en: `os@go1.25.8`

Referencia oficial:

- <https://pkg.go.dev/vuln/GO-2026-4602>

- Impact:

GoTrah trabaja precisamente recorriendo árboles de ficheros arbitrarios. Si se ejecuta sobre directorios no fiables o manipulados por terceros, esta vulnerabilidad puede permitir que el recorrido de metadatos escape de la raíz esperada.

Aunque el hallazgo no implica por sí solo lectura o borrado arbitrario fuera del root, sí debilita el aislamiento que el escaneo asume sobre el sistema de ficheros.

- Fix:

Subir el toolchain a una versión corregida, como mínimo `go1.25.8`, y regenerar los binarios y releases con esa versión.

- Mitigation:

Mientras no se actualice, evita escanear árboles de ficheros no confiables o controlados por terceros.

- False positive notes:

Aquí no lo estoy infiriendo por patrón: lo he validado con `govulncheck`, que encontró una traza alcanzable desde `scan.Scan` hacia `filepath.WalkDir` y `os.ReadDir`.

### GT-SEC-003: Los workflows de GitHub Actions dependen de tags flotantes, no de SHAs inmutables

- Rule ID: GT-SEC-003
- Severity: Medium
- Location:
  - `.github/workflows/ci.yml:27-30`
  - `.github/workflows/release.yml:25-29`
  - `.github/workflows/release.yml:53-71`
- Evidence:

```yaml
uses: actions/checkout@v4
uses: actions/setup-go@v5
uses: actions/upload-artifact@v4
uses: softprops/action-gh-release@v2
```

Referencia oficial:

- <https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions>

- Impact:

Si un tag upstream es comprometido o movido, el workflow puede ejecutar código distinto del auditado durante CI o durante la publicación de releases. En este proyecto el riesgo es especialmente sensible porque el workflow de release genera y publica binarios.

- Fix:

Pinear cada action a un commit SHA completo y documentar el proceso de actualización periódica.

- Mitigation:

Como medida transitoria, revisa periódicamente las versiones de actions utilizadas y restringe al mínimo los permisos de los jobs de release.

- False positive notes:

No es un fallo explotable dentro del binario local, pero sí un riesgo real de supply chain del pipeline.

### GT-TECH-001: La CI no ejecuta `govulncheck`, así que vulnerabilidades alcanzables pueden llegar a release sin alarma

- Rule ID: GT-TECH-001
- Severity: Medium
- Location:
  - `.github/workflows/ci.yml:35-39`
  - `.github/workflows/release.yml:34-50`
- Evidence:

```yaml
- name: Run tests
  run: go test ./...

- name: Build binary
  run: go build ./...
```

El pipeline solo ejecuta `go test` y `go build`. No hay ningún paso de `govulncheck`, ni en CI ni antes de publicar la release.

- Impact:

El ejemplo más claro es esta misma auditoría: `govulncheck` ha encontrado una vulnerabilidad alcanzable en el runtime, pero el pipeline la habría seguido publicando sin fricción.

- Fix:

Añadir `govulncheck ./...` a CI y al workflow de release, idealmente fallando el job cuando detecte vulnerabilidades alcanzables.

- Mitigation:

Hasta integrarlo en CI, ejecútalo manualmente antes de cada tag de release.

- False positive notes:

Esto no introduce una vulnerabilidad por sí solo, pero sí reduce de forma clara la capacidad del proyecto para detectar y bloquear releases inseguras.

## Low Severity Findings

### GT-TECH-002: La confirmación de borrado exagera lo que realmente se va a borrar

- Rule ID: GT-TECH-002
- Severity: Low
- Location:
  - `cmd/root.go:539-557`
  - `cmd/root.go:570-593`
- Evidence:

```go
for _, candidate := range candidates {
    totalBytes += candidate.SizeBytes
    if reason := deletionGuardReason(candidate.Path); reason != "" {
        protectedCount++
    }
}
```

```go
if reason := deletionGuardReason(candidate.Path); reason != "" {
    results = append(results, deleteResult{
        Candidate: candidate,
        Skipped:   true,
        Message:   reason,
    })
    continue
}
```

La confirmación suma tamaño y cuenta todas las rutas seleccionadas antes de excluir las que luego se van a omitir por protección.

- Impact:

El usuario puede confirmar pensando que va a borrar más carpetas o liberar más espacio del que realmente será posible. En un CLI destructivo, esa diferencia reduce confianza y puede inducir decisiones operativas erróneas.

- Fix:

Filtrar primero las rutas protegidas y construir el mensaje de confirmación con la lista efectiva de borrado. Si quieres mantener visibilidad, muestra algo como “3 carpetas se borrarán, 1 se omitirá”.

- Mitigation:

Ninguna relevante más allá de revisar el resumen final.

- False positive notes:

No es una vulnerabilidad de seguridad directa, pero sí un defecto de seguridad operativa y precisión del flujo destructivo.

## Residual Risks and Testing Gaps

- La cobertura de `cmd` es baja para una herramienta con operaciones destructivas: `21.3%`.
- No he encontrado tests dedicados a:
  - `deletionGuardReason`
  - `deleteCandidates`
  - consistencia entre protección de escaneo y de borrado
  - flujos con `--names` personalizados sobre raíces sensibles
- `internal/ui` no tiene cobertura, lo que no es grave por seguridad, pero sí deja sin pruebas parte importante del flujo interactivo.

## Commands Run During Audit

```text
go vet ./...
go test ./... -cover
GOWORK=off go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Recommended Remediation Order

1. Corregir la política de rutas protegidas y unificar escaneo + borrado.
2. Subir el toolchain de Go a una versión corregida y regenerar releases.
3. Añadir `govulncheck` a CI y release.
4. Pinear GitHub Actions por SHA.
5. Añadir tests para los guardrails destructivos.
