package errdef

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"flag"

	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy"
)

var (
	// Matches patterns like: e = cosy.NewErrorScope("user") and captures:
	// group(1): variable name (e.g. e)
	// group(2): scope name (e.g. user)
	reScope = regexp.MustCompile(`(\w+)\s*=\s*cosy\.NewErrorScope\s*\(\s*"([^"]+)"\s*\)`)

	// Updated regex templates to support negative error codes
	// Example: -1, -100, etc.
	newTemplate           = `%s\s*\.\s*New\s*\(\s*(-?[0-9]+)\s*,\s*"([^"]*)"\s*\)`
	newWithParamsTemplate = `%s\s*\.\s*NewWithParams\s*\(\s*(-?[0-9]+)\s*,\s*"([^"]*)"\s*,`
)

// parseResult stores mapping of scopeName -> []ErrorInfo for a file
type parseResult map[string][]cosy.Error

func Generate() {
	var (
		projectFolder string
		docType       string
		outDir        string
		wrapper       string
		trailingComma bool
		ignoreDirs    string
	)

	flag.StringVar(&projectFolder, "project", "", "Project folder path (required)")
	flag.StringVar(&docType, "type", "", "Documentation type: md|ts|js (required)")
	flag.StringVar(&outDir, "output", "", "Output directory (required)")
	flag.StringVar(&wrapper, "wrapper", "$gettext", "Wrapper function name")
	flag.BoolVar(&trailingComma, "trailing-comma", true, "Add trailing comma in output")
	flag.StringVar(&ignoreDirs, "ignore-dirs", "", "Comma-separated directories to ignore")
	flag.Parse()

	// Validate required flags
	if projectFolder == "" || docType == "" || outDir == "" {
		log.Fatal("Missing required flags: -project, -type, -output")
	}

	// Process doc type
	docType = strings.ToLower(strings.TrimSpace(docType))
	switch docType {
	case "md", "ts", "js":
		// valid type
	default:
		log.Fatalf("Invalid type: %s. Must be one of: md, ts, js", docType)
	}

	// Process ignore directories
	var ignoreDirsList []string
	if ignoreDirs != "" {
		ignoreDirsList = strings.Split(ignoreDirs, ",")
	}

	globalScopeMap := make(map[string][]cosy.Error)

	// find all .go files
	err := filepath.Walk(projectFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirName := filepath.Base(path)
			for _, d := range ignoreDirsList {
				if strings.EqualFold(dirName, d) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if filepath.Ext(path) == ".go" {
			res, parseErr := parseGoFile(path)
			if parseErr != nil {
				log.Printf("[Error] Parse file %s error: %v\n", path, parseErr)
			}
			for scopeName, errArr := range res {
				globalScopeMap[scopeName] = append(globalScopeMap[scopeName], errArr...)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("[Error] Walk folder error: %v\n", err)
	}

	// If no scope information is found, prompt and exit
	if len(globalScopeMap) == 0 {
		log.Println("No cosy.NewErrorScope(...) definitions or related error information found.")
		return
	}

	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("[Error] Create output directory %s failed: %v\n", outDir, err)
	}

	// Generate documents for each scope
	for scope, errInfos := range globalScopeMap {
		var outFile string
		var writeErr error

		switch docType {
		case "md":
			outFile = filepath.Join(outDir, fmt.Sprintf("%s.md", strings.ToLower(strings.ReplaceAll(scope, " ", "_"))))
			writeErr = generateMarkdown(scope, errInfos, outFile)
		case "ts":
			outFile = filepath.Join(outDir, fmt.Sprintf("%s.ts", strings.ToLower(strings.ReplaceAll(scope, " ", "_"))))
			writeErr = generateTypeScript(errInfos, outFile, wrapper, trailingComma)
		case "js":
			outFile = filepath.Join(outDir, fmt.Sprintf("%s.js", strings.ToLower(strings.ReplaceAll(scope, " ", "_"))))
			writeErr = generateJavaScript(errInfos, outFile, wrapper, trailingComma)
		}

		if writeErr != nil {
			log.Printf("[Error] Write to %s error: %v\n", outFile, writeErr)
		} else {
			log.Printf("[Generated] %s %s\n", docType, outFile)
		}
	}
}

// parseGoFile parses a single .go file and returns a mapping of scopeName -> []ErrorInfo
func parseGoFile(filePath string) (parseResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scopeVarMap := make(map[string]string) // varName -> scopeName
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := reScope.FindStringSubmatch(line)
		if len(matches) == 3 {
			varName := matches[1]
			scopeName := matches[2]
			scopeVarMap[varName] = scopeName
		}
	}

	if len(scopeVarMap) == 0 {
		return parseResult{}, nil
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	// scopeName -> error array
	res := make(parseResult)
	for _, v := range scopeVarMap {
		res[v] = []cosy.Error{}
	}

	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		for varName, scopeName := range scopeVarMap {
			reNew := regexp.MustCompile(fmt.Sprintf(newTemplate, regexp.QuoteMeta(varName)))
			reNewWithParams := regexp.MustCompile(fmt.Sprintf(newWithParamsTemplate, regexp.QuoteMeta(varName)))

			// match .New(...)
			matchesNew := reNew.FindAllStringSubmatch(line, -1)
			for _, m := range matchesNew {
				if len(m) == 3 {
					codeStr := m[1]
					msg := m[2]
					res[scopeName] = append(res[scopeName], cosy.Error{Code: cast.ToInt32(codeStr), Message: msg})
				}
			}

			// match .NewWithParams(...)
			matchesNWP := reNewWithParams.FindAllStringSubmatch(line, -1)
			for _, m := range matchesNWP {
				if len(m) == 3 {
					codeStr := m[1]
					msg := m[2]
					res[scopeName] = append(res[scopeName], cosy.Error{Code: cast.ToInt32(codeStr), Message: msg})
				}
			}
		}
	}

	return res, scanner.Err()
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// generateMarkdown generates Markdown error code documentation for a single scopeName
func generateMarkdown(scopeName string, errInfos []cosy.Error, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write scope title
	title := fmt.Sprintf("# %s\n\n", scopeName)
	_, _ = f.WriteString(title)

	// Write table header
	_, _ = f.WriteString("| Error Code | Error Message |\n")
	_, _ = f.WriteString("| --- | --- |\n")

	// Write each error message
	for _, e := range errInfos {
		line := fmt.Sprintf("| %d | %s |\n", e.Code, capitalizeFirst(e.Message))
		_, _ = f.WriteString(line)
	}

	return nil
}

// escapeString escapes special characters in a string for JavaScript/TypeScript
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// generateTypeScript generates TypeScript error code documentation
func generateTypeScript(errInfos []cosy.Error, outPath string, wrapper string, trailingComma bool) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write export default opening
	_, _ = f.WriteString("export default {\n")

	// Write each error message
	for i, e := range errInfos {
		var codeStr string
		if e.Code < 0 {
			codeStr = fmt.Sprintf("'%d'", e.Code)
		} else {
			codeStr = fmt.Sprintf("%d", e.Code)
		}

		line := fmt.Sprintf("  %s: () => %s('%s')",
			codeStr,
			wrapper,
			escapeString(capitalizeFirst(e.Message)))

		// Add comma if it's not the last item or if trailingComma is true
		if i < len(errInfos)-1 || trailingComma {
			line += ","
		}
		line += "\n"
		_, _ = f.WriteString(line)
	}

	// Write closing brace
	_, _ = f.WriteString("}\n")

	return nil
}

// generateJavaScript generates JavaScript error code documentation
func generateJavaScript(errInfos []cosy.Error, outPath string, wrapper string, trailingComma bool) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write export default opening
	_, _ = f.WriteString("module.exports = {\n")

	// Write each error message
	for i, e := range errInfos {
		var codeStr string
		if e.Code < 0 {
			codeStr = fmt.Sprintf("'%d'", e.Code)
		} else {
			codeStr = fmt.Sprintf("%d", e.Code)
		}

		line := fmt.Sprintf("  %s: () => %s('%s')",
			codeStr,
			wrapper,
			escapeString(capitalizeFirst(e.Message)))

		// Add comma if it's not the last item or if trailingComma is true
		if i < len(errInfos)-1 || trailingComma {
			line += ","
		}
		line += "\n"
		_, _ = f.WriteString(line)
	}

	// Write closing brace
	_, _ = f.WriteString("}\n")

	return nil
}
