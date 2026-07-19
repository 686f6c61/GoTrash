package cmd

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"basura/internal/safety"
	"basura/internal/scan"
	"basura/internal/ui"

	"github.com/spf13/cobra"
)

type cliOptions struct {
	names       string
	minSize     string
	interactive bool
	deleteAll   bool
	yes         bool
	showErrors  bool
	csvPath     string
}

type deleteResult struct {
	Candidate scan.Candidate
	Err       error
	Skipped   bool
	Message   string
}

type deletionPlan struct {
	Actionable      []scan.Candidate
	SkippedResults  []deleteResult
	ActionableBytes int64
}

var opts cliOptions

var rootCmd = &cobra.Command{
	Use:           "gotrah [path ...]",
	Short:         "Encuentra carpetas pesadas que puedes limpiar",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ArbitraryArgs,
	RunE:          runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&opts.names, "names", "", "Lista de nombres separados por comas para buscar")
	rootCmd.Flags().StringVar(&opts.minSize, "min-size", "", "Tamano minimo por carpeta: 500MB, 2GB, 120M...")
	rootCmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Elegir interactivamente que borrar")
	rootCmd.Flags().BoolVar(&opts.deleteAll, "delete-all", false, "Borrar todos los resultados tras pedir confirmacion")
	rootCmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "Saltarse la confirmacion final al borrar")
	rootCmd.Flags().BoolVar(&opts.showErrors, "show-errors", false, "Mostrar algunos errores de permisos y acceso")
	rootCmd.Flags().StringVar(&opts.csvPath, "csv", "", "Exportar todas las coincidencias a un CSV")
}

func runRoot(cmd *cobra.Command, args []string) error {
	if opts.interactive && opts.deleteAll {
		return errors.New("usa --interactive o --delete-all, no ambos a la vez")
	}

	styles := ui.NewStyles()

	reader := newInputReader()

	if shouldUseInteractiveMenu(cmd, args) {
		return runInteractiveMenu(styles)
	}

	fmt.Println(styles.RenderBanner())

	roots, err := resolveRoots(args)
	if err != nil {
		return err
	}

	names := parseNames(opts.names)
	minSize, err := parseByteSize(opts.minSize)
	if err != nil {
		return err
	}

	result, err := scanAndRender(scanRequest{
		Roots:      roots,
		Names:      names,
		MinSize:    minSize,
		ShowErrors: opts.showErrors,
		CSVPath:    opts.csvPath,
	}, styles, reader)
	if err != nil {
		return err
	}

	if !opts.interactive && !opts.deleteAll {
		if len(result.Candidates) > 0 {
			fmt.Println(styles.Hint("Usa --interactive para elegir que borrar o --delete-all para borrarlo todo con confirmacion."))
		}
		return nil
	}

	selected, err := selectCandidates(result.Candidates, styles, reader)
	if err != nil {
		return err
	}
	if opts.deleteAll {
		selected = append([]scan.Candidate(nil), result.Candidates...)
	}

	if len(selected) == 0 {
		fmt.Println(styles.Hint("No se ha borrado nada."))
		return nil
	}

	plan := buildDeletionPlan(selected)
	if len(plan.Actionable) == 0 {
		fmt.Println(styles.Warn("Todas las rutas seleccionadas estan protegidas. No se ha borrado nada."))
		return nil
	}

	if !opts.yes {
		if err := confirmDeletion(plan, styles, reader); err != nil {
			if errors.Is(err, errDeletionCancelled) {
				fmt.Println(styles.Hint("Operacion cancelada."))
				return nil
			}
			return err
		}
	}

	results, freed := deleteCandidates(plan)
	printDeleteResults(results, freed, styles)
	return nil
}

type scanRequest struct {
	Roots      []string
	Names      []string
	MinSize    int64
	ShowErrors bool
	CSVPath    string
}

func scanAndRender(request scanRequest, styles ui.Styles, reader *bufio.Reader) (scan.Result, error) {
	result, err := runProgressScan(scan.Options{
		Roots:   request.Roots,
		Names:   request.Names,
		MinSize: request.MinSize,
	}, styles)
	if err != nil {
		return scan.Result{}, err
	}

	fmt.Println()
	fmt.Println(styles.RenderSummary(result))

	if len(result.Candidates) == 0 {
		fmt.Println(styles.Info("No encontré carpetas candidatas con esos filtros."))
		if request.ShowErrors && len(result.Errors) > 0 {
			fmt.Println(styles.RenderErrors(result.Errors, 8))
		}
		if request.CSVPath != "" {
			path, err := exportCandidatesCSV(result.Candidates, request.CSVPath)
			if err != nil {
				return scan.Result{}, err
			}
			fmt.Println(styles.Info("CSV exportado en: " + path))
		}
		return result, nil
	}

	projectGroups := scan.GroupByProject(result.Candidates)
	typeGroups := scan.GroupByType(result.Candidates)
	fmt.Println(styles.RenderProjectTable(projectGroups, 12))
	fmt.Println(styles.RenderTypeTable(typeGroups, 8))
	if err := renderCandidates(result.Candidates, styles, reader); err != nil {
		return scan.Result{}, err
	}

	if request.ShowErrors && len(result.Errors) > 0 {
		fmt.Println(styles.RenderErrors(result.Errors, 8))
	}

	if request.CSVPath != "" {
		path, err := exportCandidatesCSV(result.Candidates, request.CSVPath)
		if err != nil {
			return scan.Result{}, err
		}
		fmt.Println(styles.Info("CSV exportado en: " + path))
	}

	return result, nil
}

func renderCandidates(candidates []scan.Candidate, styles ui.Styles, reader *bufio.Reader) error {
	const pageSize = 200

	if len(candidates) <= pageSize || !isInteractiveTerminal() {
		fmt.Println(styles.RenderTable(candidates))
		return nil
	}

	for start := 0; start < len(candidates); start += pageSize {
		fmt.Println(styles.RenderTablePage(candidates, start, pageSize))

		if start+pageSize >= len(candidates) {
			break
		}

		fmt.Print(styles.Prompt("Enter = siguiente pagina, a = ver todo, q = seguir sin listar mas: "))
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		switch strings.ToLower(strings.TrimSpace(line)) {
		case "", "n":
			continue
		case "a", "all":
			fmt.Println(styles.RenderTablePage(candidates, start+pageSize, 0))
			return nil
		case "q", "quit":
			fmt.Println(styles.Hint(fmt.Sprintf("Se han omitido %d coincidencias del listado en pantalla. Puedes exportarlas a CSV.", len(candidates)-(start+pageSize))))
			return nil
		default:
			fmt.Println(styles.Warn("Opcion no valida, sigo con la siguiente pagina."))
		}
	}

	return nil
}

func newInputReader() *bufio.Reader {
	return bufio.NewReader(os.Stdin)
}

func isInteractiveTerminal() bool {
	stdinInfo, stdinErr := os.Stdin.Stat()
	stdoutInfo, stdoutErr := os.Stdout.Stat()
	if stdinErr != nil || stdoutErr != nil {
		return false
	}
	return (stdinInfo.Mode()&os.ModeCharDevice) != 0 && (stdoutInfo.Mode()&os.ModeCharDevice) != 0
}

func exportCandidatesCSV(candidates []scan.Candidate, requestedPath string) (string, error) {
	path := strings.TrimSpace(requestedPath)
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(cwd, fmt.Sprintf("gotrah-report-%s.csv", time.Now().Format("20060102-150405")))
	}

	if !strings.HasPrefix(path, "~") {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	rows := [][]string{{"index", "type", "size_bytes", "size_human", "project", "path"}}
	for index, candidate := range candidates {
		rows = append(rows, []string{
			strconv.Itoa(index + 1),
			candidate.Name,
			strconv.FormatInt(candidate.SizeBytes, 10),
			ui.FormatBytes(candidate.SizeBytes),
			candidate.Project,
			candidate.Path,
		})
	}

	if err := writer.WriteAll(rows); err != nil {
		return "", err
	}
	if err := writer.Error(); err != nil {
		return "", err
	}

	return path, nil
}

func resolveRoots(args []string) ([]string, error) {
	if len(args) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("no pude resolver tu HOME: %w", err)
		}
		return []string{home}, nil
	}

	roots := make([]string, 0, len(args))
	seen := map[string]struct{}{}
	for _, raw := range args {
		if raw == "" {
			continue
		}
		path := raw
		if strings.HasPrefix(raw, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("no pude resolver %q: %w", raw, err)
			}
			path = filepath.Join(home, strings.TrimPrefix(raw, "~/"))
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("no pude resolver %q: %w", raw, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("no pude acceder a %q: %w", abs, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%q no es una carpeta", abs)
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		roots = append(roots, abs)
	}
	return roots, nil
}

func parseNames(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return append([]string(nil), scan.DefaultNames...)
	}
	parts := strings.Split(raw, ",")
	names := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func parseByteSize(raw string) (int64, error) {
	value := strings.TrimSpace(strings.ToUpper(raw))
	if value == "" {
		return 0, nil
	}

	units := map[string]float64{
		"B":  1,
		"K":  1024,
		"KB": 1024,
		"M":  1024 * 1024,
		"MB": 1024 * 1024,
		"G":  1024 * 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"T":  1024 * 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	numberPart := strings.TrimRightFunc(value, func(r rune) bool {
		return (r >= 'A' && r <= 'Z')
	})
	unitPart := strings.TrimSpace(strings.TrimPrefix(value, numberPart))
	if unitPart == "" {
		unitPart = "B"
	}

	number, err := strconv.ParseFloat(strings.TrimSpace(numberPart), 64)
	if err != nil {
		return 0, fmt.Errorf("no entiendo el valor de --min-size %q", raw)
	}

	multiplier, ok := units[unitPart]
	if !ok {
		return 0, fmt.Errorf("unidad no soportada en --min-size: %q", unitPart)
	}
	if number < 0 {
		return 0, errors.New("--min-size no puede ser negativo")
	}

	return int64(number * multiplier), nil
}

func selectCandidates(candidates []scan.Candidate, styles ui.Styles, reader *bufio.Reader) ([]scan.Candidate, error) {
	if isInteractiveTerminal() {
		items := make([]tuiChecklistItem, 0, len(candidates))
		for index, candidate := range candidates {
			warning := ""
			if deletionGuardReason(candidate.Path) != "" {
				warning = "protegida"
			}
			items = append(items, tuiChecklistItem{
				Label:       fmt.Sprintf("%s #%d  %s", candidateTypeIcon(candidate.Name), index+1, candidate.Name),
				ProjectName: shortProjectName(candidate.Project),
				SizeText:    ui.FormatBytes(candidate.SizeBytes),
				SizeBytes:   candidate.SizeBytes,
				Detail:      fmt.Sprintf("%s  ->  %s", candidate.Project, candidate.Path),
				Warning:     warning,
			})
		}

		indexes, err := runChecklistPromptTUI("Selecciona carpetas para borrar", items, styles)
		if err != nil {
			return nil, err
		}

		selected := make([]scan.Candidate, 0, len(indexes))
		for _, index := range indexes {
			selected = append(selected, candidates[index])
		}
		return selected, nil
	}

	for {
		fmt.Print(styles.Prompt("Selecciona indices para borrar (ej: 1,3-5), 'all' o 'none': "))
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		indexes, err := parseSelection(strings.TrimSpace(line), len(candidates))
		if err != nil {
			fmt.Println(styles.Warn(err.Error()))
			continue
		}

		selected := make([]scan.Candidate, 0, len(indexes))
		for _, index := range indexes {
			selected = append(selected, candidates[index])
		}
		return selected, nil
	}
}

func parseSelection(input string, total int) ([]int, error) {
	if total == 0 {
		return nil, nil
	}

	value := strings.TrimSpace(strings.ToLower(input))
	switch value {
	case "", "none", "n":
		return nil, nil
	case "all", "a":
		indexes := make([]int, total)
		for i := range total {
			indexes[i] = i
		}
		return indexes, nil
	}

	seen := map[int]struct{}{}
	indexes := make([]int, 0)
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			chunks := strings.Split(part, "-")
			if len(chunks) != 2 {
				return nil, fmt.Errorf("rango no valido: %q", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(chunks[0]))
			if err != nil {
				return nil, fmt.Errorf("indice no valido: %q", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(chunks[1]))
			if err != nil {
				return nil, fmt.Errorf("indice no valido: %q", part)
			}
			if start < 1 || end < 1 || start > total || end > total || end < start {
				return nil, fmt.Errorf("rango fuera de limite: %q", part)
			}
			for current := start; current <= end; current++ {
				index := current - 1
				if _, ok := seen[index]; ok {
					continue
				}
				seen[index] = struct{}{}
				indexes = append(indexes, index)
			}
			continue
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("indice no valido: %q", part)
		}
		if value < 1 || value > total {
			return nil, fmt.Errorf("indice fuera de limite: %d", value)
		}
		index := value - 1
		if _, ok := seen[index]; ok {
			continue
		}
		seen[index] = struct{}{}
		indexes = append(indexes, index)
	}

	sort.Ints(indexes)
	return indexes, nil
}

var errDeletionCancelled = errors.New("deletion cancelled")

func buildDeletionPlan(candidates []scan.Candidate) deletionPlan {
	plan := deletionPlan{
		Actionable: make([]scan.Candidate, 0, len(candidates)),
	}

	for _, candidate := range candidates {
		if reason := deletionGuardReason(candidate.Path); reason != "" {
			plan.SkippedResults = append(plan.SkippedResults, deleteResult{
				Candidate: candidate,
				Skipped:   true,
				Message:   reason,
			})
			continue
		}

		plan.Actionable = append(plan.Actionable, candidate)
		plan.ActionableBytes += candidate.SizeBytes
	}

	return plan
}

func confirmDeletion(plan deletionPlan, styles ui.Styles, reader *bufio.Reader) error {
	if len(plan.SkippedResults) > 0 {
		fmt.Println(styles.Warn(fmt.Sprintf("%d rutas seleccionadas estan protegidas y se omitiran.", len(plan.SkippedResults))))
	}

	message := fmt.Sprintf(
		"Vas a borrar %d carpetas y liberar aprox. %s. Escribe BORRAR para continuar: ",
		len(plan.Actionable),
		ui.FormatBytes(plan.ActionableBytes),
	)

	fmt.Print(styles.Prompt(message))
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToUpper(line)) != "BORRAR" {
		return errDeletionCancelled
	}
	return nil
}

func deleteCandidates(plan deletionPlan) ([]deleteResult, int64) {
	results := append([]deleteResult(nil), plan.SkippedResults...)
	var freed int64
	for _, candidate := range plan.Actionable {
		err := os.RemoveAll(candidate.Path)
		results = append(results, deleteResult{
			Candidate: candidate,
			Err:       err,
			Message:   explainDeletionErr(err),
		})
		if err == nil {
			freed += candidate.SizeBytes
		}
	}
	return results, freed
}

func printDeleteResults(results []deleteResult, freed int64, styles ui.Styles) {
	okCount := 0
	failCount := 0
	skipCount := 0
	lines := make([]string, 0, len(results))

	for _, result := range results {
		if result.Skipped {
			skipCount++
			lines = append(lines, styles.Warn(fmt.Sprintf("SKIP %s -> %s", result.Candidate.Path, result.Message)))
			continue
		}
		if result.Err != nil {
			failCount++
			message := result.Message
			if message == "" {
				message = result.Err.Error()
			}
			lines = append(lines, styles.Danger(fmt.Sprintf("ERR  %s -> %s", result.Candidate.Path, message)))
			continue
		}
		okCount++
		lines = append(lines, styles.Success(fmt.Sprintf("OK   %s", result.Candidate.Path)))
	}

	fmt.Println(styles.RenderDeleteSummary(okCount, failCount, skipCount, freed))
	fmt.Println(strings.Join(lines, "\n"))
}

func shortProjectName(path string) string {
	base := filepath.Base(filepath.Clean(path))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return path
	}
	return base
}

func deletionGuardReason(path string) string {
	return safety.GuardReason(path)
}

func explainDeletionErr(err error) string {
	if err == nil {
		return ""
	}
	if os.IsPermission(err) {
		return "permiso denegado"
	}
	return err.Error()
}
