package main

import (
	"database/sql"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dsn      = flag.String("dsn", "", "MySQL DSN, e.g. root:root@tcp(192.168.1.45:3306)/sdim_v1")
	tables   = flag.String("tables", "", "Comma-separated table names, or '_all' for all tables")
	excludes = flag.String("excludes", "", "Comma-separated exclude patterns (use * for wildcard)")
	tags     = flag.String("tags", "db", "Struct tags to generate, e.g. 'db,json=omitempty,gorm'")
	prefix   = flag.String("prefix", "t_", "Table name prefix to strip (default: 't_')")
	outDir   = flag.String("out", "./models", "Output directory (required)")
)

// TagSpec represents a tag like "json=omitempty"
type TagSpec struct {
	Name    string
	Options []string
}

// parseTagSpec parses "json=omitempty" into TagSpec
func parseTagSpec(spec string) TagSpec {
	spec = strings.TrimSpace(spec)
	if idx := strings.Index(spec, "="); idx != -1 {
		name := spec[:idx]
		opts := strings.Split(spec[idx+1:], ",")
		var cleaned []string
		for _, opt := range opts {
			opt = strings.TrimSpace(opt)
			if opt != "" {
				cleaned = append(cleaned, opt)
			}
		}
		return TagSpec{Name: name, Options: cleaned}
	}
	return TagSpec{Name: spec}
}

// buildStructTags builds `db:"id" json:"id,omitempty"`
func buildStructTags(colName string, tagSpecs []TagSpec) string {
	if len(tagSpecs) == 0 {
		return "`db:\"" + colName + "\"`"
	}
	var parts []string
	for _, ts := range tagSpecs {
		if ts.Name == "" {
			continue
		}
		value := colName
		if len(ts.Options) > 0 {
			value += "," + strings.Join(ts.Options, ",")
		}
		parts = append(parts, fmt.Sprintf(`%s:"%s"`, ts.Name, value))
	}
	return "`" + strings.Join(parts, " ") + "`"
}

// sanitizeFileName ensures valid file name
func sanitizeFileName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	name = reg.ReplaceAllString(name, "_")
	if name != "" && unicode.IsDigit(rune(name[0])) {
		name = "_" + name
	}
	return strings.Trim(name, "_")
}

func camelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "")
}

func stripPrefix(tableName, prefix string) string {
	if prefix != "" && strings.HasPrefix(strings.ToLower(tableName), strings.ToLower(prefix)) {
		return tableName[len(prefix):]
	}
	return tableName
}

func inferPackageName(outDir string) string {
	packageName := filepath.Base(filepath.Clean(outDir))
	if packageName == "." || packageName == "/" || packageName == "\\" {
		return "models"
	}
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	packageName = reg.ReplaceAllString(packageName, "_")
	if packageName == "" || unicode.IsDigit(rune(packageName[0])) {
		return "models"
	}
	return packageName
}

func goTypeFromMySQL(colName, dataType string, nullable bool) string {
	if colName == "id" && !nullable {
		return "int64"
	}

	if nullable {
		switch dataType {
		case "tinyint":
			return "sql.NullInt16"
		case "smallint":
			return "sql.NullInt16" // smallint: -32768~32767 → fits in int32
		case "mediumint", "int", "integer", "bigint":
			return "sql.NullInt64"
		case "varchar", "char", "text", "longtext", "enum", "set":
			return "sql.NullString"
		case "datetime", "timestamp", "date", "time":
			return "sql.NullTime"
		case "float", "double", "decimal":
			return "sql.NullFloat64"
		case "bit", "bool", "boolean":
			return "sql.NullBool"
		default:
			return "sql.NullString"
		}
	} else {
		switch dataType {
		case "tinyint":
			return "int8"
		case "smallint":
			return "int16"
		case "mediumint", "int", "integer":
			return "int32"
		case "bigint":
			return "int64"
		case "varchar", "char", "text", "longtext", "enum", "set":
			return "string"
		case "datetime", "timestamp", "date", "time":
			return "time.Time"
		case "float", "double", "decimal":
			// float → float32, double/decimal → float64
			if dataType == "float" {
				return "float32"
			}
			return "float64"
		case "bit", "bool", "boolean":
			return "bool"
		default:
			return "string"
		}
	}
}

func fetchAllTableNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func compileExcludePatterns(patterns string) ([]*regexp.Regexp, error) {
	if patterns == "" {
		return nil, nil
	}
	var regexps []*regexp.Regexp
	for _, p := range strings.Split(patterns, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var pattern string
		if strings.Contains(p, "*") {
			pattern = "^" + regexp.QuoteMeta(p)
			pattern = strings.ReplaceAll(pattern, "\\*", ".*")
			pattern += "$"
		} else {
			pattern = "^" + regexp.QuoteMeta(p) + "$"
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern %q: %w", p, err)
		}
		regexps = append(regexps, re)
	}
	return regexps, nil
}

func shouldExclude(tableName string, excludes []*regexp.Regexp) bool {
	for _, re := range excludes {
		if re.MatchString(tableName) {
			return true
		}
	}
	return false
}

func generateModelForTable(db *sql.DB, tableName, prefix, outDir, packageName string, tagSpecs []TagSpec) error {
	var tableComment string
	err := db.QueryRow(`
		SELECT TABLE_COMMENT 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`, tableName).Scan(&tableComment)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get table comment for %s: %w", tableName, err)
	}

	tableComment = strings.TrimSpace(tableComment)
	if tableComment == "" {
		tableComment = fmt.Sprintf("represents a row from table %s", tableName)
	}

	rows, err := db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	var fields []string
	var usesTime, usesSQL bool

	for rows.Next() {
		var colName, dataType, isNullable, comment string
		if err := rows.Scan(&colName, &dataType, &isNullable, &comment); err != nil {
			return err
		}

		nullable := isNullable == "YES"
		goType := goTypeFromMySQL(colName, dataType, nullable)

		if strings.Contains(goType, "time.Time") {
			usesTime = true
		}
		if strings.Contains(goType, "sql.") {
			usesSQL = true
		}

		fieldName := camelCase(colName)
		tagStr := buildStructTags(colName, tagSpecs)
		line := fmt.Sprintf("\t%s %s %s", fieldName, goType, tagStr)
		if comment != "" {
			line += " // " + comment
		}
		fields = append(fields, line)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	baseName := stripPrefix(tableName, prefix)
	typeName := camelCase(baseName)
	receiver := strings.ToLower(typeName[:1])

	fileBase := sanitizeFileName(baseName)
	if fileBase == "" {
		fileBase = sanitizeFileName(tableName)
	}
	fileName := fmt.Sprintf("%s_mod_gen.go", fileBase)
	filePath := filepath.Join(outDir, fileName)

	src := fmt.Sprintf(`// Code generated by genmodel. DO NOT EDIT.

package %s

`, packageName)

	if usesSQL || usesTime {
		src += "import (\n"
		if usesSQL {
			src += "\t\"database/sql\"\n"
		}
		if usesTime {
			src += "\t\"time\"\n"
		}
		src += ")\n\n"
	}

	src += fmt.Sprintf(`// %s %s
type %s struct {
%s
}

func (%s %s) TableName() string {
	return "%s"
}
`, typeName, tableComment, typeName, strings.Join(fields, "\n"), receiver, typeName, tableName)

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format error for table %s: %w", tableName, err)
	}

	if err := os.WriteFile(filePath, formatted, 0644); err != nil {
		return err
	}

	fmt.Printf("✅ %s\n", filePath)
	return nil
}

func main() {
	flag.Parse()

	if *dsn == "" || *tables == "" {
		log.Fatal("Usage:\n" +
			"  go run genmodel.go -dsn <dsn> -tables t_user -tags db,json=omitempty -out ./models\n" +
			"  go run genmodel.go -dsn <dsn> -tables _all -excludes t_log_* -tags db,gorm,json=omitempty -out ./models")
	}

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatal(err)
	}

	packageName := inferPackageName(*outDir)

	var tableList []string
	if *tables == "_all" {
		tableList, err = fetchAllTableNames(db)
		if err != nil {
			log.Fatalf("Failed to fetch all tables: %v", err)
		}
		if len(tableList) == 0 {
			log.Fatal("No tables found in current database")
		}
		fmt.Printf("Found %d tables.\n", len(tableList))
	} else {
		tableList = strings.Split(*tables, ",")
		for i := range tableList {
			tableList[i] = strings.TrimSpace(tableList[i])
		}
	}

	// Apply excludes
	excludeRegexps, err := compileExcludePatterns(*excludes)
	if err != nil {
		log.Fatalf("Invalid excludes: %v", err)
	}

	var filteredTables []string
	for _, table := range tableList {
		if table == "" {
			continue
		}
		if shouldExclude(table, excludeRegexps) {
			fmt.Printf("⏭️  Excluded table: %s\n", table)
			continue
		}
		filteredTables = append(filteredTables, table)
	}
	tableList = filteredTables

	// Parse tags
	var tagSpecs []TagSpec
	if *tags != "" {
		for _, spec := range strings.Split(*tags, ",") {
			tagSpecs = append(tagSpecs, parseTagSpec(spec))
		}
	}

	// Generate
	fmt.Printf("Generating models for %d table(s) into package '%s'...\n", len(tableList), packageName)
	for _, table := range tableList {
		if err := generateModelForTable(db, table, *prefix, *outDir, packageName, tagSpecs); err != nil {
			log.Printf("⚠️  Skipping table %s: %v", table, err)
		}
	}

	fmt.Printf("🎉 Done! Generated %d model file(s) in %s\n", len(tableList), *outDir)
}
