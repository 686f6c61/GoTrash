package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"basura/internal/scan"
	"basura/internal/ui"

	"github.com/spf13/cobra"
)

type menuChoice struct {
	Label string
	Hint  string
}

type menuConfig struct {
	Roots      []string
	Names      []string
	MinSize    int64
	ShowErrors bool
	CSVPath    string
}

const (
	postActionPick    = "pick"
	postActionProject = "project"
	postActionExport  = "export"
	postActionAll     = "all"
	postActionBack    = "back"
	postActionExit    = "exit"
)

func shouldUseInteractiveMenu(cmd *cobra.Command, args []string) bool {
	if len(args) > 0 {
		return false
	}
	if !isInteractiveTerminal() {
		return false
	}

	flagNames := []string{"names", "min-size", "interactive", "delete-all", "yes", "show-errors", "csv"}
	for _, name := range flagNames {
		if cmd.Flags().Changed(name) {
			return false
		}
	}

	return true
}

func runInteractiveMenu(styles ui.Styles) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println(styles.Header.Render("Menu inicial"))

		config, exit, err := promptScanConfig(reader, styles)
		if err != nil {
			return err
		}
		if exit {
			fmt.Println(styles.Hint("Hasta luego."))
			return nil
		}

		result, err := scanAndRender(scanRequest(config), styles, reader)
		if err != nil {
			return err
		}

		if len(result.Candidates) == 0 {
			next, err := promptChoice(
				reader,
				styles,
				"Que quieres hacer ahora?",
				[]menuChoice{
					{Label: "↺ Volver al menu principal", Hint: "Cambiar carpeta o filtros"},
					{Label: "⎋ Salir", Hint: "Cerrar el programa"},
				},
			)
			if err != nil {
				return err
			}
			if next == 0 {
				continue
			}
			fmt.Println(styles.Hint("Hasta luego."))
			return nil
		}

		action, err := promptPostScanAction(reader, styles)
		if err != nil {
			return err
		}

		switch action {
		case postActionBack:
			continue
		case postActionExit:
			fmt.Println(styles.Hint("Hasta luego."))
			return nil
		case postActionPick:
			selected, err := selectCandidates(result.Candidates, styles, reader)
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				fmt.Println(styles.Hint("No se ha borrado nada."))
				continue
			}
			plan := buildDeletionPlan(selected)
			if len(plan.Actionable) == 0 {
				fmt.Println(styles.Warn("Todas las rutas seleccionadas estan protegidas. No se ha borrado nada."))
				continue
			}
			if err := confirmDeletion(plan, styles, reader); err != nil {
				if err == errDeletionCancelled {
					fmt.Println(styles.Hint("Operacion cancelada."))
					continue
				}
				return err
			}
			results, freed := deleteCandidates(plan)
			printDeleteResults(results, freed, styles)
			if err := waitForMainMenu(reader, styles); err != nil {
				return err
			}
			continue
		case postActionProject:
			selected, err := selectProjects(result.Candidates, styles, reader)
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				fmt.Println(styles.Hint("No se ha borrado nada."))
				continue
			}
			plan := buildDeletionPlan(selected)
			if len(plan.Actionable) == 0 {
				fmt.Println(styles.Warn("Todas las rutas seleccionadas estan protegidas. No se ha borrado nada."))
				continue
			}
			if err := confirmDeletion(plan, styles, reader); err != nil {
				if err == errDeletionCancelled {
					fmt.Println(styles.Hint("Operacion cancelada."))
					continue
				}
				return err
			}
			results, freed := deleteCandidates(plan)
			printDeleteResults(results, freed, styles)
			if err := waitForMainMenu(reader, styles); err != nil {
				return err
			}
			continue
		case postActionExport:
			path, err := promptCSVPath(reader, styles)
			if err != nil {
				return err
			}
			savedPath, err := exportCandidatesCSV(result.Candidates, path)
			if err != nil {
				return err
			}
			fmt.Println(styles.Info("CSV exportado en: " + savedPath))
			continue
		case postActionAll:
			plan := buildDeletionPlan(result.Candidates)
			if len(plan.Actionable) == 0 {
				fmt.Println(styles.Warn("Todas las rutas encontradas estan protegidas. No se ha borrado nada."))
				continue
			}
			if err := confirmDeletion(plan, styles, reader); err != nil {
				if err == errDeletionCancelled {
					fmt.Println(styles.Hint("Operacion cancelada."))
					continue
				}
				return err
			}
			results, freed := deleteCandidates(plan)
			printDeleteResults(results, freed, styles)
			if err := waitForMainMenu(reader, styles); err != nil {
				return err
			}
			continue
		}
	}
}

func waitForMainMenu(reader *bufio.Reader, styles ui.Styles) error {
	fmt.Print(styles.Prompt("Pulsa Enter para volver al menu principal: "))
	_, err := reader.ReadString('\n')
	return err
}

func promptScanConfig(reader *bufio.Reader, styles ui.Styles) (menuConfig, bool, error) {
	scopeIndex, err := promptChoice(
		reader,
		styles,
		"Que quieres escanear?",
		[]menuChoice{
			{Label: "⌂ Mi carpeta personal", Hint: "HOME completa"},
			{Label: "▣ Una carpeta concreta", Hint: "Tu eliges la ruta"},
			{Label: "▦ Varias carpetas", Hint: "Separadas por comas"},
			{Label: "◉ Todo el disco", Hint: "Puede tardar y dar avisos de permisos"},
			{Label: "⎋ Salir", Hint: "No hacer nada"},
		},
	)
	if err != nil {
		return menuConfig{}, false, err
	}
	if scopeIndex == 4 {
		return menuConfig{}, true, nil
	}

	roots, showErrors, err := rootsForScope(scopeIndex, reader, styles)
	if err != nil {
		return menuConfig{}, false, err
	}

	profileIndex, err := promptChoice(
		reader,
		styles,
		"Que tipo de carpetas quieres buscar?",
		[]menuChoice{
			{Label: "◇ Basura tipica de desarrollo", Hint: "node_modules, venv, build, Pods..."},
			{Label: "⬢ Solo JavaScript / frontend", Hint: "node_modules, .next, dist, build..."},
			{Label: "◌ Solo Python", Hint: ".venv, venv, env, __pycache__..."},
			{Label: "⌘ Solo Apple / Xcode", Hint: "DerivedData, Pods, build"},
			{Label: "✎ Personalizado", Hint: "Tu lista de nombres"},
		},
	)
	if err != nil {
		return menuConfig{}, false, err
	}

	names, err := namesForProfile(profileIndex, reader, styles)
	if err != nil {
		return menuConfig{}, false, err
	}

	minSize, err := promptMinSize(reader, styles)
	if err != nil {
		return menuConfig{}, false, err
	}

	return menuConfig{
		Roots:      roots,
		Names:      names,
		MinSize:    minSize,
		ShowErrors: showErrors,
	}, false, nil
}

func rootsForScope(scopeIndex int, reader *bufio.Reader, styles ui.Styles) ([]string, bool, error) {
	switch scopeIndex {
	case 0:
		roots, err := resolveRoots(nil)
		return roots, false, err
	case 1:
		for {
			fmt.Print(styles.Prompt("Escribe la ruta a escanear: "))
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, false, err
			}
			roots, err := resolveRoots([]string{strings.TrimSpace(line)})
			if err != nil {
				fmt.Println(styles.Warn(err.Error()))
				continue
			}
			return roots, false, nil
		}
	case 2:
		for {
			fmt.Print(styles.Prompt("Escribe varias rutas separadas por comas: "))
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, false, err
			}
			parts := splitCommaList(line)
			roots, err := resolveRoots(parts)
			if err != nil {
				fmt.Println(styles.Warn(err.Error()))
				continue
			}
			return roots, false, nil
		}
	case 3:
		return []string{string(os.PathSeparator)}, true, nil
	default:
		return nil, false, fmt.Errorf("opcion de alcance no soportada: %d", scopeIndex)
	}
}

func namesForProfile(profileIndex int, reader *bufio.Reader, styles ui.Styles) ([]string, error) {
	switch profileIndex {
	case 0:
		return append([]string(nil), scan.DefaultNames...), nil
	case 1:
		return []string{
			"node_modules",
			".next",
			".nuxt",
			".svelte-kit",
			".turbo",
			".cache",
			".parcel-cache",
			"dist",
			"build",
		}, nil
	case 2:
		return []string{
			".venv",
			"venv",
			"env",
			"__pycache__",
			".pytest_cache",
			".mypy_cache",
			".ruff_cache",
			".tox",
			"build",
			"dist",
		}, nil
	case 3:
		return []string{
			"DerivedData",
			"Pods",
			"build",
		}, nil
	case 4:
		for {
			fmt.Print(styles.Prompt("Escribe nombres de carpetas separados por comas: "))
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			names := parseNames(strings.Join(splitCommaList(line), ","))
			if len(names) == 0 {
				fmt.Println(styles.Warn("Necesito al menos un nombre de carpeta."))
				continue
			}
			return names, nil
		}
	default:
		return nil, fmt.Errorf("perfil no soportado: %d", profileIndex)
	}
}

func promptMinSize(reader *bufio.Reader, styles ui.Styles) (int64, error) {
	index, err := promptChoice(
		reader,
		styles,
		"Tamano minimo por carpeta",
		[]menuChoice{
			{Label: "○ Sin minimo", Hint: "Muestra todo lo encontrado"},
			{Label: "◔ 100 MB", Hint: "Ideal para ver ruido pequeno"},
			{Label: "◑ 500 MB", Hint: "Buen filtro general"},
			{Label: "◕ 1 GB", Hint: "Solo lo realmente pesado"},
			{Label: "✎ Personalizado", Hint: "Ej: 250MB o 2GB"},
		},
	)
	if err != nil {
		return 0, err
	}

	switch index {
	case 0:
		return 0, nil
	case 1:
		return 100 * 1024 * 1024, nil
	case 2:
		return 500 * 1024 * 1024, nil
	case 3:
		return 1024 * 1024 * 1024, nil
	case 4:
		for {
			fmt.Print(styles.Prompt("Escribe el tamano minimo (ej: 250MB, 2GB): "))
			line, err := reader.ReadString('\n')
			if err != nil {
				return 0, err
			}
			size, err := parseByteSize(strings.TrimSpace(line))
			if err != nil {
				fmt.Println(styles.Warn(err.Error()))
				continue
			}
			return size, nil
		}
	default:
		return 0, fmt.Errorf("tamano minimo no soportado: %d", index)
	}
}

func promptPostScanAction(reader *bufio.Reader, styles ui.Styles) (string, error) {
	index, err := promptChoice(
		reader,
		styles,
		"Que quieres hacer con los resultados?",
		[]menuChoice{
			{Label: "☑ Elegir carpetas concretas", Hint: "Seleccion visual con checkboxes"},
			{Label: "▤ Elegir proyectos completos", Hint: "Borrar por bloques, no una a una"},
			{Label: "⇩ Exportar CSV", Hint: "Guardar todo el listado completo"},
			{Label: "✖ Borrar todo lo encontrado", Hint: "Con confirmacion final"},
			{Label: "↺ Volver al menu principal", Hint: "Cambiar carpeta o filtros"},
			{Label: "⎋ Salir sin borrar", Hint: "Cerrar el programa"},
		},
	)
	if err != nil {
		return "", err
	}

	switch index {
	case 0:
		return postActionPick, nil
	case 1:
		return postActionProject, nil
	case 2:
		return postActionExport, nil
	case 3:
		return postActionAll, nil
	case 4:
		return postActionBack, nil
	default:
		return postActionExit, nil
	}
}

func selectProjects(candidates []scan.Candidate, styles ui.Styles, reader *bufio.Reader) ([]scan.Candidate, error) {
	if isInteractiveTerminal() {
		items := make([]tuiChecklistItem, 0)
		groups := scan.GroupByProject(candidates)
		for index, group := range groups {
			warning := ""
			for _, candidate := range group.Candidates {
				if deletionGuardReason(candidate.Path) != "" {
					warning = "incluye rutas protegidas"
					break
				}
			}
			items = append(items, tuiChecklistItem{
				Label:       fmt.Sprintf("▤ #%d  %d carpetas", index+1, group.Count),
				ProjectName: filepath.Base(filepath.Clean(group.Project)),
				SizeText:    ui.FormatBytes(group.TotalBytes),
				SizeBytes:   group.TotalBytes,
				Detail:      fmt.Sprintf("%s  |  Tipos: %s", group.Project, strings.Join(group.Types, ", ")),
				Warning:     warning,
			})
		}

		indexes, err := runChecklistPromptTUI("Selecciona proyectos completos", items, styles)
		if err != nil {
			return nil, err
		}

		selected := make([]scan.Candidate, 0)
		for _, index := range indexes {
			selected = append(selected, groups[index].Candidates...)
		}
		return selected, nil
	}

	groups := scan.GroupByProject(candidates)
	fmt.Println(styles.RenderProjectTable(groups, 0))

	for {
		fmt.Print(styles.Prompt("Selecciona proyectos por indice (ej: 1,3-5), 'all' o 'none': "))
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		indexes, err := parseSelection(strings.TrimSpace(line), len(groups))
		if err != nil {
			fmt.Println(styles.Warn(err.Error()))
			continue
		}

		selected := make([]scan.Candidate, 0)
		for _, index := range indexes {
			selected = append(selected, groups[index].Candidates...)
		}
		return selected, nil
	}
}

func promptCSVPath(reader *bufio.Reader, styles ui.Styles) (string, error) {
	fmt.Print(styles.Prompt("Ruta del CSV [Enter = generar nombre automatico]: "))
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptChoice(reader *bufio.Reader, styles ui.Styles, title string, choices []menuChoice) (int, error) {
	if isInteractiveTerminal() {
		return runChoicePromptTUI(title, choices, styles)
	}

	for {
		fmt.Println()
		fmt.Println(styles.Header.Render(title))
		for index, choice := range choices {
			line := fmt.Sprintf("  %d. %s", index+1, choice.Label)
			if choice.Hint != "" {
				line += "  " + styles.Muted.Render(choice.Hint)
			}
			fmt.Println(line)
		}

		fmt.Print(styles.Prompt("Elige una opcion [1]: "))
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		value := strings.TrimSpace(line)
		if value == "" {
			return 0, nil
		}

		number, err := strconv.Atoi(value)
		if err == nil && number >= 1 && number <= len(choices) {
			return number - 1, nil
		}

		fmt.Println(styles.Warn("Opcion no valida."))
	}
}

func splitCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}
