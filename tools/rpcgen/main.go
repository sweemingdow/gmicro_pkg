// tools/rpcgen/main.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

type Method struct {
	Name     string
	ReqType  string
	RespType string
	Path     string
	Timeout  time.Duration
}

type Interface struct {
	Name       string
	ServerName string
	Methods    []Method
	PkgName    string
	FileName   string
}

func main() {
	var typeName string
	flag.StringVar(&typeName, "type", "", "interface name to generate")
	flag.Parse()

	if typeName == "" {
		log.Fatal("-type flag is required")
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("parse dir: %v", err)
	}
	if len(pkgs) != 1 {
		log.Fatal("expected exactly one package in directory")
	}
	var pkg *ast.Package
	for _, p := range pkgs {
		pkg = p
		break
	}

	var iface *Interface
	for fileName, file := range pkg.Files {
		if i := parseInterface(file, typeName, fset); i != nil {
			iface = i
			iface.PkgName = pkg.Name
			iface.FileName = filepath.Base(fileName)
			break
		}
	}
	if iface == nil {
		log.Fatalf("interface %s not found", typeName)
	}

	generate(iface)
}

func parseInterface(file *ast.File, typeName string, fset *token.FileSet) *Interface {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if typeSpec.Name.Name == typeName {
						if ifaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
							return parseInterfaceType(file, ifaceType, typeName, fset, genDecl.Doc)
						}
					}
				}
			}
		}
	}
	return nil
}

func parseInterfaceType(file *ast.File, iface *ast.InterfaceType, typeName string, fset *token.FileSet, doc *ast.CommentGroup) *Interface {
	i := &Interface{
		Name: typeName,
	}

	// Parse @rpc_server from interface comment
	if doc != nil {
		for _, c := range doc.List {
			text := c.Text
			if strings.Contains(text, "@rpc_server") {
				parts := strings.Fields(text)
				for j, part := range parts {
					if part == "@rpc_server" && j+1 < len(parts) {
						i.ServerName = strings.Trim(parts[j+1], "\"'")
						break
					}
				}
			}
		}
	}
	// fallback: OneRpcProvider -> one_service
	if i.ServerName == "" {
		base := strings.TrimSuffix(typeName, "RpcProvider")
		if base == typeName {
			base = typeName
		}
		i.ServerName = strings.ToLower(base) + "_service"
	}

	// Parse each method
	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		if len(funcType.Params.List) != 1 {
			log.Fatalf("method %s must have exactly one parameter", methodName)
		}
		if len(funcType.Results.List) != 2 {
			log.Fatalf("method %s must return (T, error)", methodName)
		}

		param := funcType.Params.List[0]
		reqTypeStr := astString(fset, param.Type)
		respTypeStr := astString(fset, funcType.Results.List[0].Type)

		// Path: @path or /snake_case
		path := "/unknown"
		if method.Doc != nil {
			for _, c := range method.Doc.List {
				text := c.Text
				if strings.Contains(text, "@path") {
					parts := strings.Fields(text)
					for j, part := range parts {
						if part == "@path" && j+1 < len(parts) {
							path = parts[j+1]
							break
						}
					}
					break
				}
			}
		}
		if path == "/unknown" {
			path = "/" + toSnakeCase(methodName)
		}

		// Timeout: @timeout or 1s
		methodTimeout := 1 * time.Second
		if method.Doc != nil {
			for _, c := range method.Doc.List {
				text := c.Text
				if strings.Contains(text, "@timeout") {
					parts := strings.Fields(text)
					for j, part := range parts {
						if part == "@timeout" && j+1 < len(parts) {
							if d, err := time.ParseDuration(parts[j+1]); err == nil {
								methodTimeout = d
							}
							break
						}
					}
					break
				}
			}
		}

		i.Methods = append(i.Methods, Method{
			Name:     methodName,
			ReqType:  extractGenericArg(reqTypeStr, "rpccall.RpcReqWrapper"),
			RespType: respTypeStr,
			Path:     path,
			Timeout:  methodTimeout,
		})
	}

	return i
}

// --- Helpers ---

func astString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return "UNKNOWN"
	}
	return buf.String()
}

func extractGenericArg(s, wrapper string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, wrapper+"[") || !strings.HasSuffix(s, "]") {
		log.Fatalf("expected %s[T], got %s", wrapper, s)
	}
	return s[len(wrapper)+1 : len(s)-1]
}

// toSnakeCase converts CamelCase to snake_case (e.g., OneRpcProvider -> one_rpc_provider)
func toSnakeCase(name string) string {
	if name == "" {
		return ""
	}
	var result []rune
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func durationToGoExpr(d time.Duration) string {
	if d == 0 {
		return "0"
	}
	if d%time.Second == 0 {
		return fmt.Sprintf("%d * time.Second", d/time.Second)
	}
	if d%time.Millisecond == 0 {
		return fmt.Sprintf("%d * time.Millisecond", d/time.Millisecond)
	}
	if d%time.Microsecond == 0 {
		return fmt.Sprintf("%d * time.Microsecond", d/time.Microsecond)
	}
	return fmt.Sprintf("%d * time.Nanosecond", d/time.Nanosecond)
}

// --- Template ---

const tmpl = `// Code generated by rpcgen. DO NOT EDIT.

package {{.PkgName}}

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient/rcfactory"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rpccall"
	"time"
)

type {{.ImplName}} struct {
	clientFactory rcfactory.ArpcClientFactory
}

func New{{.Name}}(clientFactory rcfactory.ArpcClientFactory) {{.Name}} {
	return &{{.ImplName}}{
		clientFactory: clientFactory,
	}
}

{{range .Methods}}
func (p *{{$.ImplName}}) {{.Name}}(req rpccall.RpcReqWrapper[{{.ReqType}}]) ({{.RespType}}, error) {
	cp := p.acquireClientProxy()
	var resp {{.RespType}}
	if err := cp.Call("{{.Path}}", &req, &resp, {{durationToGoExpr .Timeout}}); err != nil {
		return resp, err
	}
	return resp, nil
}
{{end}}

func (p *{{.ImplName}}) acquireClientProxy() rclient.ArpcClientProxy {
	return p.clientFactory.AcquireClient("{{.ServerName}}")
}
`

func (i *Interface) ImplName() string {
	name := i.Name
	if len(name) == 0 {
		return ""
	}
	return strings.ToLower(string(name[0])) + name[1:]
}

func generate(i *Interface) {
	funcMap := template.FuncMap{
		"durationToGoExpr": durationToGoExpr,
	}
	t := template.Must(template.New("rpc").Funcs(funcMap).Parse(tmpl))

	// Generate filename: OneRpcProvider -> one_rpc_provider_gen.go
	outputFile := toSnakeCase(i.Name) + "_gen.go"

	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("create %s: %v", outputFile, err)
	}
	defer f.Close()

	if err := t.Execute(f, i); err != nil {
		log.Fatalf("execute template: %v", err)
	}

	fmt.Printf("Generated %s\n", outputFile)
}
