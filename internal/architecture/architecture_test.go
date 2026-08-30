package architecture

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
	assertThinBinaryEntry(t, root)
	assertNoRootHandlerTransition(t, root)
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

func assertNoRootHandlerTransition(t *testing.T, root string) {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filepath.Join(root, "pkg", "cmd", "root", "app.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse root app: %v", err)
	}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "registerUpgrade" {
			t.Fatal("root retains Upgrade registration transition")
		}
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.Name.Name == "Dependencies" || typeSpec.Name.Name == "UpgradeHandler" {
				t.Fatalf("root retains transitional type %q", typeSpec.Name.Name)
			}
			if typeSpec.Name.Name != "App" {
				continue
			}
			structure, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					if name.Name == "upgrade" {
						t.Fatal("root App retains Upgrade handler field")
					}
				}
			}
		}
	}
}

func assertThinBinaryEntry(t *testing.T, root string) {
	t.Helper()
	directory := filepath.Join(root, "cmd", "ycy")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read cmd/ycy: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "main.go" || entries[0].IsDir() {
		t.Fatalf("cmd/ycy inventory = %#v, want only main.go", entries)
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filepath.Join(directory, "main.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse thin binary entry: %v", err)
	}
	for _, imported := range file.Imports {
		value, unquoteErr := strconv.Unquote(imported.Path.Value)
		if unquoteErr != nil || (value != "os" && value != modulePath+"/internal/ycycmd") {
			t.Fatalf("thin binary import = %q, want os and internal/ycycmd only", value)
		}
	}
	var mainFunction *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "main" {
			mainFunction = function
			break
		}
	}
	if mainFunction == nil || mainFunction.Body == nil || len(mainFunction.Body.List) != 1 {
		t.Fatal("thin binary entry must contain one main statement")
	}
	expression, ok := mainFunction.Body.List[0].(*ast.ExprStmt)
	if !ok {
		t.Fatal("thin binary entry main statement is not a call")
	}
	exitCall, ok := expression.X.(*ast.CallExpr)
	if !ok || !isSelectorCall(exitCall, "os", "Exit") || len(exitCall.Args) != 1 {
		t.Fatal("thin binary entry must call os.Exit once")
	}
	mainCall, ok := exitCall.Args[0].(*ast.CallExpr)
	if !ok || !isSelectorCall(mainCall, "ycycmd", "Main") || len(mainCall.Args) != 1 {
		t.Fatal("thin binary entry must pass version to ycycmd.Main")
	}
	argument, ok := mainCall.Args[0].(*ast.Ident)
	if !ok || argument.Name != "version" {
		t.Fatal("thin binary entry must pass the injected version")
	}
}

func isSelectorCall(call *ast.CallExpr, packageName, functionName string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != functionName {
		return false
	}
	identifier, ok := selector.X.(*ast.Ident)
	return ok && identifier.Name == packageName
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
		case "utils", "services", "interfaces", "adapters", "common":
			violations = append(violations, path+": forbidden generic package segment "+segment)
		}
	}
	for _, imported := range file.Imports {
		pathValue, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			violations = append(violations, path+": invalid import literal")
			continue
		}
		violations = append(violations, validateImport(relativeDir, path, pathValue)...)
	}
	violations = append(violations, inspectCalls(relativeDir, path, file)...)
	return violations
}

func validateImport(source, filePath, imported string) []string {
	var violations []string
	if imported == "github.com/spf13/cobra" || imported == "github.com/spf13/pflag" {
		if !cobraOwner(source) {
			violations = append(violations, filePath+": Cobra/pflag may be imported only by command packages")
		}
	}
	if strings.Contains(imported, "/legacy/"+"b"+"un") {
		violations = append(violations, filePath+": active code imports frozen legacy")
	}
	if strings.HasPrefix(imported, modulePath+"/") {
		violations = append(violations, validateInternalImport(source, filePath, strings.TrimPrefix(imported, modulePath+"/"))...)
	}
	return violations
}

func validateInternalImport(source, filePath, imported string) []string {
	var violations []string
	if strings.HasPrefix(imported, "internal/commands/") {
		violations = append(violations, source+": old command package import is forbidden: "+imported)
	}

	if isInternalPackage(source) && strings.HasPrefix(imported, "pkg/cmd") {
		if source != "internal/ycycmd" {
			violations = append(violations, filePath+": internal package imports command package "+imported)
		}
	}
	if source == "pkg/cmdutil" && strings.HasPrefix(imported, "pkg/cmd") {
		violations = append(violations, filePath+": pkg/cmdutil must not import "+imported)
	}
	if source == "pkg/cmd/factory" && strings.HasPrefix(imported, "pkg/cmd/") && imported != "pkg/cmdutil" {
		violations = append(violations, filePath+": command factory imports command package "+imported)
	}
	if strings.HasPrefix(source, "pkg/cmd/") {
		violations = append(violations, validateTargetCommandImport(source, filePath, imported)...)
	}
	if source == "cmd/ycy" && !allowedCurrentBinaryImport(imported) {
		violations = append(violations, filePath+": cmd/ycy imports outside its transition boundary: "+imported)
	}
	if source == "internal/ycycmd" && !allowedYcycmdImport(imported) {
		violations = append(violations, filePath+": internal/ycycmd imports outside composition boundary: "+imported)
	}
	if isInternalPackage(source) && imported == "cmd/ycy" {
		violations = append(violations, filePath+": internal package imports cmd/ycy")
	}
	if isInternalPackage(source) && imported == "legacy/bun" {
		violations = append(violations, filePath+": internal package imports legacy/bun")
	}
	if imported == "web" && !allowedWebassetsConsumer(source) {
		violations = append(violations, filePath+": webassets consumer is not an owning command or composition root")
	}
	return violations
}

func inspectCalls(relativeDir, path string, file *ast.File) []string {
	var violations []string
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

func validateTargetCommandImport(source, filePath, imported string) []string {
	if !strings.HasPrefix(imported, "pkg/cmd/") {
		return nil
	}
	// A parent may import only its direct child groups/leaves. A leaf may not
	// import any sibling or parent; shared behavior belongs under internal.
	if source == "pkg/cmd/root" {
		return nil
	}
	if source == "pkg/cmd/factory" {
		return []string{filePath + ": factory imports command package " + imported}
	}
	if targetCommandParent(source, imported) {
		return nil
	}
	return []string{filePath + ": command package imports sibling or parent " + imported}
}

func targetCommandParent(source, imported string) bool {
	sourceParts := strings.Split(source, "/")
	importedParts := strings.Split(imported, "/")
	if len(importedParts) <= len(sourceParts) {
		return false
	}
	for index := range sourceParts {
		if sourceParts[index] != importedParts[index] {
			return false
		}
	}
	// A command parent may register only its direct child package. Leaves have
	// no children in the command package graph, so they cannot cross owners.
	return len(importedParts) == len(sourceParts)+1
}

func cobraOwner(path string) bool {
	return path == "pkg/cmd/root" || strings.HasPrefix(path, "pkg/cmd/")
}

func isInternalPackage(path string) bool {
	return strings.HasPrefix(path, "internal/")
}

func allowedCurrentBinaryImport(imported string) bool {
	return imported == "internal/ycycmd"
}

func allowedYcycmdImport(imported string) bool {
	return strings.HasPrefix(imported, "pkg/cmd/") || imported == "pkg/cmdutil" || strings.HasPrefix(imported, "internal/") || imported == "web"
}

func allowedWebassetsConsumer(path string) bool {
	return path == "internal/ycycmd" || path == "pkg/cmd/diff" || path == "pkg/cmd/fs" || path == "pkg/cmd/tunnel/server" || path == "tools/web-browser-harness"
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("determine architecture test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
