package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		panic("GOFILE environment variable not set")
	}
	packageName := os.Getenv("GOPACKAGE")
	if goFile == "" {
		panic("GOPACKAGE environment variable not set")
	}
	outputFile, err := os.Open(wd + string(os.PathSeparator) + "repo.sql")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	files := getSqlFileNames(wd + string(os.PathSeparator) + "sql")
	repo := NewRepoFile(packageName, "GetQuery", "sql", files)
	fmt.Printf(repo.Generate())
}

func getSqlFileNames(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "directory read error: %v\n", err)
		os.Exit(1)
	}

	var sqlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, entry.Name())
		}
	}

	return sqlFiles
}

type RepoFile struct {
	Files   []SqlFile
	Package string
}

func NewRepoFile(packageName, prefix, dir string, names []string) *RepoFile {
	r := &RepoFile{Package: packageName}
	for _, name := range names {
		r.Files = append(r.Files, NewSqlFile(name, prefix, dir))
	}

	return r
}

func (r *RepoFile) Generate() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("package %s\n\nimport _ \"embed\"\n\n", r.Package))
	for _, file := range r.Files {
		builder.WriteString(file.GenerateEmbedSql())
	}

	return builder.String()
}

type SqlFile struct {
	Filename string
	Prefix   string
	Dir      string
}

func NewSqlFile(filename, prefix, dir string) SqlFile {
	return SqlFile{Filename: filename, Prefix: prefix, Dir: dir}
}

func (f SqlFile) GetNameWithoutSQLSuffix() string {
	return strings.TrimRight(f.Filename, ".sql")
}

func (f SqlFile) GenerateEmbedSql() string {
	return fmt.Sprintf("//go:embed %s\nvar %s string\nfunc %s%s() string {\n\treturn %s\n}\n\n", f.GetRelativePath(), f.GetNameWithoutSQLSuffix(), f.Prefix, f.GetNameWithoutSQLSuffix(), f.GetNameWithoutSQLSuffix())
}

func (f SqlFile) GetRelativePath() string {
	return f.Dir + string(os.PathSeparator) + f.Filename
}
