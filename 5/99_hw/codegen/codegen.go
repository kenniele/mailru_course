package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

// код писать тут

type tpl struct {
	Params      string
	PackageName string
}

// Fun - объект функции
type Fun struct {
	Header string       `json:"-"`
	Name   string       `json:"-"`
	Owner  string       `json:"-"`
	Params ParsedStruct `json:"-"`
	Url    string       `json:"url"`
	Auth   bool         `json:"auth"`
	Method string       `json:"method"`
}

// StructElements - элемент, свойство структуры
type StructElements struct {
	Name string
	Typ  string
	Tag  string
}

// ParsedStruct - объект структуры
type ParsedStruct struct {
	Name     string
	Elements []StructElements
	Funcs    []Fun
}

// GetFuncInfo - парсит информацию о функции из *ast.FuncDecl
func GetFuncInfo(fun *ast.FuncDecl) Fun {
	var currFunc = Fun{
		Name: fun.Name.Name,
	}
	if fun.Doc != nil && fun.Recv != nil {
		text := strings.Split(fun.Doc.List[0].Text[3:], " ")
		owner := GetReceiver(fun)
		currFunc.Header = text[0]
		currFunc.Owner = owner
		err := json.Unmarshal([]byte(strings.Join(text[1:], " ")), &currFunc)
		if err != nil {
			log.Fatal(err)
		}
		currFunc.Params = ParsedStruct{
			Name: GetParamsName(fun),
		}
	}
	return currFunc
}

// GetReceiver - парсит "владельца" метода из *ast.FuncDecl
func GetReceiver(fun *ast.FuncDecl) string {
	recv := fun.Recv.List[0]
	switch expr := recv.Type.(type) {
	case *ast.StarExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return expr.Name
	}
	return ""
}

// GetParamsName - парсит тип второго параметра из *ast.FuncDecl
func GetParamsName(fun *ast.FuncDecl) string {
	if fun.Type.Params != nil && len(fun.Type.Params.List) > 1 {
		secondArg := fun.Type.Params.List[1]
		switch t := secondArg.Type.(type) {
		case *ast.Ident:
			return t.Name
		case *ast.StarExpr:
			if xIdent, ok := t.X.(*ast.Ident); ok {
				return xIdent.Name
			}
		}
	}
	return ""
}

// GetGenInfo - парсит информацию о структурах из *ast.GenDecl
func GetGenInfo(gen *ast.GenDecl) ParsedStruct {
	var result ParsedStruct

	for _, spec := range gen.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.TypeSpec\n", spec)
			continue
		}

		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.StructType\n", spec)
			continue
		}
		result = ParsedStruct{
			Name: currType.Name.Name,
		}

		if len(currStruct.Fields.List) == 0 {
			return result
		}

		for _, field := range currStruct.Fields.List {
			var fieldType string
			switch t := field.Type.(type) {
			case *ast.Ident:
				fieldType = t.Name
			case *ast.StarExpr:
				if xIdent, ok := t.X.(*ast.Ident); ok {
					fieldType = "*" + xIdent.Name
				}
			case *ast.ArrayType:
				if elt, ok := t.Elt.(*ast.Ident); ok {
					fieldType = "[]" + elt.Name
				}
			case *ast.MapType:
				keyType := ""
				valueType := ""
				if key, ok := t.Key.(*ast.Ident); ok {
					keyType = key.Name
				}
				if value, ok := t.Value.(*ast.Ident); ok {
					valueType = value.Name
				}
				fieldType = fmt.Sprintf("map[%s]%s", keyType, valueType)
			}

			element := StructElements{
				Name: field.Names[0].Name,
				Typ:  fieldType,
			}

			if field.Tag != nil {
				tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
				element.Tag = string(tag)
			}

			result.Elements = append(result.Elements, element)
		}
	}

	return result
}

// GetSliceFieldName - получает список полей определенной структуры
func GetSliceFieldName(elements []StructElements) []string {
	var result []string
	for _, v := range elements {
		if temp := GetFromTag(v.Tag, "paramname"); temp != "" {
			result = append(result, v.Name+"_"+v.Typ)
		} else {
			result = append(result, v.Name+"_"+v.Typ)
		}
	}
	return result
}

// GetFromTag - получает значение по необходимому тэгу
func GetFromTag(tag string, need string) string {
	if tag == "required" {
		return "true"
	}
	split := strings.Split(strings.Replace(strings.TrimLeft(tag, "apivalidator:"), "\"", "", -1), ",")
	for _, v := range split {
		if strings.HasPrefix(v, need) {
			return strings.Split(v, "=")[1]
		}
	}
	return ""
}

// k - значение поля, v - тэг поля, fieldName, fieldValue - распаршенный тэг поля
// ParseTag(v) - возвращает мапу из элементов
var (
	imports = template.Must(template.New("imports").Parse(`package {{.PackageName}}

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
)

type Params struct {
	Name string
	Typ string
	Tags map[string]string
}

`))

	intTpl = template.Must(template.New("intTpl").Parse(`
			for fieldName, fieldValue := range v.Tags {
				switch fieldName {
				case "required":
					if k == 0 {
						return ApiError(500, errors.New("Field is required"))
					}
				case "enum":
					cases := strings.Split(fieldValue[5:len(v)-1], "|")
					if !slices.Contains(cases, k) {
						return ApiError(500, errors.New("Field doesn't contain required value")
					}
				case "default" && k == 0:
					k = fieldValue
				case "min":
					if k < strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too small")
					}
				case "max":
					if k > strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too big"))
					}
				default:
					panic("Undefined field: " + fieldName)
				}
			}
		}
`))

	strTpl = template.Must(template.New("strTpl").Parse(`
			for fieldName, fieldValue := range v.Tags {
				switch fieldName {
				case "required":
					if k == 0 {
						return ApiError(500, errors.New("Field is required")
					}
				case "enum":
					cases := strings.Split(fieldValue[5:len(v)-1], "|")
					if !slices.Contains(cases, k) {
						return ApiError(500, errors.New("Field doesn't contain required value")
					}
				case "default" && k == 0:
					k = fieldValue
				case "min":
					if len(k) < strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too small")
					}
				case "max":
					if len(k) > strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too big"))
					}
				default:
					panic("Undefined field: " + fieldName)
				}
			}
		}
`))

	paramsParseTpl = template.Must(template.New("paramsParseTpl").Parse(`
func ParamsParse(r *http.Request, paramNames []string) []Params {
	result := []Params{}
	for _, v := range paramNames {
		sp := strings.Split(v, "_")
		result = append(result, Params{
			Name: sp[0],
			Typ: sp[1],
			Tags: ParseTag(r.URL.Query().Get(v)),
		})
	}
	return result
}
`))

	getFromTagTpl = template.Must(template.New("getFromTagTpl").Parse(`
func ParseTag(tag string) map[string]string {
	result := map[string]string{}
	split := strings.Split(strings.Replace(strings.TrimLeft(tag, "apivalidator:"), "\"", "", -1), ",")
	for _, v := range split {
		if v == "required" {
			result[v] = "true"
		} else {
			sp := strings.Split(v, "=")
			result[sp[0]] = sp[1]
		}
	}
	return result
}
`))
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	defer out.Close()

	imports.Execute(out, tpl{PackageName: "main"})

	paramsParseTpl.Execute(out, "")
	getFromTagTpl.Execute(out, "")

	var structs []ParsedStruct
	mapFuncs := make(map[string][]Fun, 10)
	funcs := make([]Fun, 0, 10)

	// Парсинг файла
	for _, f := range node.Decls {
		switch f.(type) {
		case *ast.FuncDecl:
			info := GetFuncInfo(f.(*ast.FuncDecl))
			//fmt.Printf("GET %+v AS TYPE=%T; \nINFO - %+v\n\n", f, f, info)
			funcs = append(funcs, info)
		case *ast.GenDecl:
			info := GetGenInfo(f.(*ast.GenDecl))
			if info.Name == "" {
				//fmt.Println("Skipped info -", info.Name)
				continue
			}
			//fmt.Printf("GET %+v AS TYPE=%T; \nINFO - %+v\n\n", f, f, info)
			structs = append(structs, info)
		default:
			//fmt.Printf("UNIQUE GET %+v AS TYPE=%T\n\n", f, f)
		}
	}

	// Определение владельцев функций
	for _, fn := range funcs {
		for i := range structs {
			if fn.Params.Name == structs[i].Name {
				fn.Params = structs[i]
			}
		}
		mapFuncs[fn.Owner] = append(mapFuncs[fn.Owner], fn)
	}

	// Добавление в структуры функций
	for i := range structs {
		v := &structs[i]
		v.Funcs = append(v.Funcs, mapFuncs[v.Name]...)
	}

	// Определение API структур (с методами)
	var usefulStrs []ParsedStruct
	for _, v := range structs {
		if len(v.Funcs) > 0 {
			usefulStrs = append(usefulStrs, v)
		}
		fmt.Println(v.Name, v.Funcs)
	}

	// Кодогенерация
	for _, v := range usefulStrs {
		fmt.Fprintf(out, `
func (h *%v) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {`, v.Name)
		for _, met := range v.Funcs {
			fmt.Fprintf(out, `
	case "%v":
		h.wrapperDoSomeJob(w, r, %#v, "%#v")
`, met.Url, GetSliceFieldName(met.Params.Elements), met.Name)
		}
		fmt.Fprintln(out, `default:
		w.WriteHeader(404)
	}
}`)

		fmt.Fprintf(out, `
func (h *%v) wrapperDoSomeJob(w http.ResponseWriter, r *http.Request, paramNames []string) {
	params := ParamsParse(r, paramNames)
	for _, v := range params {
		switch v.Typ {
		case "int":`, v.Name)
		intTpl.Execute(out, "")
		fmt.Fprint(out, `
		case "string":`)
		strTpl.Execute(out, "")
		fmt.Fprintf(out, `
		default:
			panic("Undefined fieldType")
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	%v := nil
}`, v.Name)
	}
}
