package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/hackycy/hackycy-cli"

func TestActiveArchitecture(t *testing.T) {
	root := repositoryRoot(t)
	var violations []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			switch relative {
			case ".git", "legacy", ".scratch", "mock", "node_modules", "web/node_modules", "web/dist", "tools":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		violations = append(violations, inspectGoFile(root, path)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk active Go architecture: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("active architecture violations:\n%s", strings.Join(violations, "\n"))
	}
}

func inspectGoFile(root, path string) []string {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return []string{path + ": parse error: " + err.Error()}
	}
	relativeDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return []string{path + ": resolve package path: " + err.Error()}
	}
	relativeDir = filepath.ToSlash(relativeDir)
	var violations []string
	for _, segment := range strings.Split(relativeDir, "/") {
		switch segment {
		case "utils", "services", "interfaces", "adapters", "common", "pkg":
			violations = append(violations, path+": forbidden generic package segment "+segment)
		}
	}
	for _, imported := range file.Imports {
		pathValue, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			violations = append(violations, path+": invalid import literal")
			continue
		}
		if pathValue == "github.com/spf13/cobra" || pathValue == "github.com/spf13/pflag" {
			if relativeDir != "internal/cliapp" {
				violations = append(violations, path+": Cobra/pflag may be imported only by internal/cliapp")
			}
		}
		if strings.Contains(pathValue, "/legacy/"+"b"+"un") {
			violations = append(violations, path+": active code imports frozen legacy")
		}
		if strings.HasPrefix(pathValue, modulePath+"/internal/commands/") {
			violations = append(violations, validateCommandImport(relativeDir, pathValue)...)
		}
		if pathValue == modulePath+"/web" && !allowedWebassetsConsumer(relativeDir) {
			violations = append(violations, path+": webassets consumer is not an owning command or composition root")
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if ok && literal.Kind == token.STRING && relativeDir != "internal/appconfig" {
			value, err := strconv.Unquote(literal.Value)
			if err == nil && (value == "config.json" || value == ".config.lock") {
				violations = append(violations, path+": config persistence is owned only by internal/appconfig")
			}
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Exit" {
			return true
		}
		identifier, ok := selector.X.(*ast.Ident)
		if ok && identifier.Name == "os" && relativeDir != "cmd/ycy" {
			violations = append(violations, path+": os.Exit is owned only by cmd/ycy")
		}
		return true
	})
	return violations
}

func validateCommandImport(source, imported string) []string {
	if source == "internal/cliapp" || source == "cmd/ycy" {
		return nil
	}
	if !strings.HasPrefix(source, "internal/commands/") {
		return []string{source + ": shared module imports a command package"}
	}
	importedRelative := strings.TrimPrefix(imported, modulePath+"/")
	if commandOwner(source) != commandOwner(importedRelative) {
		return []string{source + ": command module imports sibling " + importedRelative}
	}
	return nil
}

func commandOwner(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return path
	}
	if parts[2] == "config" || parts[2] == "git" {
		if len(parts) >= 4 {
			return strings.Join(parts[:4], "/")
		}
	}
	return strings.Join(parts[:3], "/")
}

func allowedWebassetsConsumer(path string) bool {
	return path == "cmd/ycy" || path == "internal/commands/diff" || path == "internal/commands/fs" || path == "internal/commands/tunnel"
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("determine architecture test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
