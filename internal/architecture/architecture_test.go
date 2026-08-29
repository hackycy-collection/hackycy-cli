package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	rootcommand "github.com/hackycy/hackycy-cli/pkg/cmd/root"
)

const modulePath = "github.com/hackycy/hackycy-cli"

// transitionAllowlist is intentionally short and explicit. Each entry is a
// temporary ownership root that Slice 0 permits while the migration proceeds;
// the owning slice must delete its entry before its checkpoint passes. This is
// not a package-wide exemption and must never be widened to include pkg.
var transitionAllowlist = map[string]string{
	"cmd/ycy":           "temporary process composition root and adapters",
	"internal/commands": "unmigrated command Modules (removed by their leaf slices)",
}

// rootCommandImportAllowlist records the handler Modules still owned by the
// lifted root. Each migrated leaf removes its corresponding entry instead of
// broadening the command-package transition.
var rootCommandImportAllowlist = map[string]string{
	"internal/commands/diff":      "unmigrated Diff handler",
	"internal/commands/fs":        "unmigrated FS handler",
	"internal/commands/git/cm":    "unmigrated Git CM handler",
	"internal/commands/git/fork":  "unmigrated Git Fork handler",
	"internal/commands/git/heat":  "unmigrated Git Heat handler",
	"internal/commands/git/pulse": "unmigrated Git Pulse handler",
	"internal/commands/tunnel":    "unmigrated Tunnel handlers",
}

// rootHandlerAllowlist is the complete, shrinking set of temporary handler
// capabilities. Process facts must remain in cmdutil.Factory.
var rootHandlerAllowlist = map[string]string{
	"GitHeat":        "unmigrated Git Heat handler",
	"GitPulse":       "unmigrated Git Pulse handler",
	"GitFork":        "unmigrated Git Fork handler",
	"GitCM":          "unmigrated Git CM handler",
	"Diff":           "unmigrated Diff handler",
	"FS":             "unmigrated FS handler",
	"TunnelServer":   "unmigrated Tunnel server handler",
	"TunnelConnect":  "unmigrated Tunnel connect handler",
	"Upgrade":        "unmigrated Upgrade handler",
}

func TestActiveArchitecture(t *testing.T) {
	root := repositoryRoot(t)
	assertTransitionAllowlist(t, root)
	assertRootHandlerAllowlist(t)
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

func assertRootHandlerAllowlist(t *testing.T) {
	t.Helper()
	dependencies := reflect.TypeFor[rootcommand.Dependencies]()
	if dependencies.NumField() != len(rootHandlerAllowlist) {
		t.Fatalf("root Dependencies fields = %d, allowlist = %d", dependencies.NumField(), len(rootHandlerAllowlist))
	}
	for index := 0; index < dependencies.NumField(); index++ {
		field := dependencies.Field(index)
		if _, ok := rootHandlerAllowlist[field.Name]; !ok {
			t.Fatalf("root Dependencies field %q is outside the transition allowlist", field.Name)
		}
	}
}

func assertTransitionAllowlist(t *testing.T, root string) {
	t.Helper()
	want := map[string]string{
		"cmd/ycy":           "temporary process composition root and adapters",
		"internal/commands": "unmigrated command Modules (removed by their leaf slices)",
	}
	if len(transitionAllowlist) != len(want) {
		t.Fatalf("transition allowlist has %d entries, want %d", len(transitionAllowlist), len(want))
	}
	for path, purpose := range want {
		if transitionAllowlist[path] != purpose {
			t.Fatalf("transition allowlist %q = %q, want %q", path, transitionAllowlist[path], purpose)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("transition allowlist path %q: %v", path, err)
		}
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
		violations = append(violations, validateCommandImport(source, imported)...)
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

func validateCommandImport(source, imported string) []string {
	if source == "cmd/ycy" {
		return nil
	}
	if source == "pkg/cmd/root" {
		if _, ok := rootCommandImportAllowlist[imported]; ok {
			return nil
		}
		return []string{source + ": root transition does not allow command module " + imported}
	}
	if !strings.HasPrefix(source, "internal/commands/") {
		return []string{source + ": shared module imports a command package " + imported}
	}
	if commandOwner(source) != commandOwner(imported) {
		return []string{source + ": command module imports sibling " + imported}
	}
	return nil
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

func commandOwner(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return path
	}
	if parts[0] == "pkg" && parts[1] == "cmd" {
		if len(parts) >= 4 && (parts[2] == "config" || parts[2] == "git" || parts[2] == "tunnel") {
			return strings.Join(parts[:4], "/")
		}
		return strings.Join(parts[:3], "/")
	}
	if parts[0] == "internal" && parts[1] == "commands" {
		if len(parts) >= 4 && (parts[2] == "config" || parts[2] == "git") {
			return strings.Join(parts[:4], "/")
		}
		return strings.Join(parts[:3], "/")
	}
	return strings.Join(parts[:3], "/")
}

func cobraOwner(path string) bool {
	return path == "cmd/ycy" || path == "pkg/cmd/root" || strings.HasPrefix(path, "pkg/cmd/")
}

func isInternalPackage(path string) bool {
	return strings.HasPrefix(path, "internal/")
}

func allowedCurrentBinaryImport(imported string) bool {
	if strings.HasPrefix(imported, "internal/") || imported == "pkg/cmd/factory" || imported == "pkg/cmd/root" || imported == "pkg/cmdutil" || imported == "web" {
		return true
	}
	return false
}

func allowedYcycmdImport(imported string) bool {
	return strings.HasPrefix(imported, "pkg/cmd/") || imported == "pkg/cmdutil" || strings.HasPrefix(imported, "internal/") || imported == "web"
}

func allowedWebassetsConsumer(path string) bool {
	return path == "cmd/ycy" || path == "internal/ycycmd" || path == "internal/commands/diff" || path == "internal/commands/fs" || path == "internal/commands/tunnel" || path == "pkg/cmd/diff" || path == "pkg/cmd/fs" || path == "pkg/cmd/tunnel" || path == "tools/web-browser-harness"
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("determine architecture test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
