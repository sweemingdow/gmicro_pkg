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
	Name        string
	OriginalReq string // e.g., "AuthReq", "rpccall.RpcReqWrapper[AuthReq]", or ""
	ReqType     string // core type or ""
	IsWrapper   bool
	RespType    string
	Path        string
	Timeout     time.Duration
	HasParam    bool
}

type Interface struct {
	Name       string
	ServerName string
	Methods    []Method
	PkgName    string
	FileName   string
	PkgPrefix  string
}

func main() {
	var (
		typeName  string
		pkgPrefix string
	)
	flag.StringVar(&typeName, "type", "", "interface name to generate")
	flag.StringVar(&pkgPrefix, "pkg_prefix", "github.com/sweemingdow/gmicro_pkg", "-pkg_prefix")
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
			iface.PkgPrefix = pkgPrefix
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
	if i.ServerName == "" {
		base := strings.TrimSuffix(typeName, "RpcProvider")
		if base == typeName {
			base = typeName
		}
		i.ServerName = strings.ToLower(base) + "_service"
	}

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		paramCount := len(funcType.Params.List)
		if paramCount > 1 {
			log.Fatalf("method %s must have 0 or 1 parameter", methodName)
		}
		if len(funcType.Results.List) != 2 {
			log.Fatalf("method %s must return (T, error)", methodName)
		}

		var originalReqStr, coreReq string
		var isWrapper bool
		hasParam := paramCount == 1

		if hasParam {
			originalReqStr = astString(fset, funcType.Params.List[0].Type)
			originalReqStr, coreReq, isWrapper = parseReqType(originalReqStr)
		} else {
			originalReqStr = ""
			coreReq = ""
			isWrapper = false
		}

		respTypeStr := astString(fset, funcType.Results.List[0].Type)

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
			Name:        methodName,
			OriginalReq: originalReqStr,
			ReqType:     coreReq,
			IsWrapper:   isWrapper,
			RespType:    respTypeStr,
			Path:        path,
			Timeout:     methodTimeout,
			HasParam:    hasParam,
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

func parseReqType(reqExpr string) (original, core string, isWrapper bool) {
	s := strings.TrimSpace(reqExpr)
	if strings.HasPrefix(s, "rpccall.RpcReqWrapper[") && strings.HasSuffix(s, "]") {
		core = s[len("rpccall.RpcReqWrapper[") : len(s)-1]
		return s, core, true
	}
	return s, s, false
}

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
	"{{.PkgPrefix}}/pkg/server/srpc/rclient"
	"{{.PkgPrefix}}/pkg/server/srpc/rclient/rcfactory"
	"{{.PkgPrefix}}/pkg/server/srpc/rpccall"
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
func (p *{{$.ImplName}}) {{.Name}}({{if .HasParam}}req {{.OriginalReq}}{{end}}) ({{.RespType}}, error) {
	cp := p.acquireClientProxy()
	var resp {{.RespType}}
	var reqForCall interface{}
	{{if .HasParam}}
	reqForCall = &req
	{{else}}
	reqForCall = nil
	{{end}}
	if err := cp.Call("{{.Path}}", reqForCall, &resp, {{durationToGoExpr .Timeout}}); err != nil {
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
	data := struct {
		*Interface
		ImplName string
	}{
		Interface: i,
		ImplName:  i.ImplName(),
	}

	funcMap := template.FuncMap{
		"durationToGoExpr": durationToGoExpr,
	}
	t := template.Must(template.New("rpc").Funcs(funcMap).Parse(tmpl))

	outputFile := toSnakeCase(i.Name) + "_gen.go"

	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("create %s: %v", outputFile, err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		log.Fatalf("execute template: %v", err)
	}

	fmt.Printf("Generated %s\n", outputFile)
}
