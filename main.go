package gen_grpc

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"text/template"
)

func GenerateGRPCForInterface(grpcInterface any, targetPath string) {
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		panic(err)
	}
	reflectIf := reflect.TypeOf(grpcInterface).Elem()
	if reflectIf.Kind() != reflect.Interface {
		panic("Must pass an interface")
	}
	var interfaceDesc = interfaceDescription{
		Name:        reflectIf.Name(),
		PackageName: filepath.Base(targetPath),
		Package:     reflectIf.PkgPath(),
		PackageBase: path.Base(reflectIf.PkgPath()),
	}
	var grpcMethods []grpcMethod
	for i := 0; i < reflectIf.NumMethod(); i++ {
		method := reflectIf.Method(i).Type
		name := reflectIf.Method(i).Name
		var params []goType
		var returnValues []goType
		var hasErrorReturn, hasContext bool
		for i := 0; i < method.NumIn(); i++ {
			param := method.In(i)
			if i == 0 && param == contextType {
				hasContext = true
				continue
			}
			params = append(params, parseType(param))
		}
		for i := 0; i < method.NumOut(); i++ {
			param := method.Out(i)
			if i == method.NumOut()-1 && param == errorType {
				hasErrorReturn = true
				continue
			}
			returnValues = append(returnValues, parseType(param))
		}
		interfaceDesc.Methods = append(interfaceDesc.Methods, interfaceMethod{
			Name:        name,
			Args:        params,
			Results:     returnValues,
			UsesError:   hasErrorReturn,
			UsesContext: hasContext,
		})
		paramStruct := grpcStruct{
			Name: fmt.Sprintf("%sRequest", name),
		}
		for j, param := range params {
			paramStruct.Fields = append(paramStruct.Fields, grpcField{
				Name: fmt.Sprintf("Param%d", j),
				Type: param.GrpcProtoName,
			})
		}
		grpcTypes = append(grpcTypes, paramStruct)
		resultStruct := grpcStruct{
			Name: fmt.Sprintf("%sResponse", name),
		}
		for j, returnValue := range returnValues {
			resultStruct.Fields = append(resultStruct.Fields, grpcField{
				Name: fmt.Sprintf("Result%d", j),
				Type: returnValue.GrpcProtoName,
			})
		}
		grpcTypes = append(grpcTypes, resultStruct)
		grpcMethods = append(grpcMethods, grpcMethod{
			Name:   name,
			Param:  paramStruct.Name,
			Result: resultStruct.Name,
		})
	}
	protoPath := filepath.Join(targetPath, "grpc.proto")
	protoFile, err := os.Create(protoPath)
	if err != nil {
		panic(err)
	}
	service := grpcService{
		Name:        reflectIf.Name(),
		Methods:     grpcMethods,
		Types:       grpcTypes,
		PackagePath: targetPath,
		PackageName: filepath.Base(targetPath),
	}
	if err := templates.ExecuteTemplate(protoFile, "proto.gotemplate", service); err != nil {
		panic(err)
	}
	protoFile.Close()

	protoCmd := exec.Command("protoc", "--go_out=plugins=grpc:.", protoPath)
	protoCmd.Stderr = os.Stderr
	if err := protoCmd.Run(); err != nil {
		panic(err)
	}

	interfaceDesc.RequiredPackages = requiredPackages
	interfaceDesc.RequiredTypes = goTypes

	serverFile, err := os.Create(filepath.Join(targetPath, "server.go"))
	if err != nil {
		panic(err)
	}
	if err := templates.ExecuteTemplate(serverFile, "server.go.gotemplate", interfaceDesc); err != nil {
		panic(err)
	}

	clientFile, err := os.Create(filepath.Join(targetPath, "client.go"))
	if err != nil {
		panic(err)
	}
	if err := templates.ExecuteTemplate(clientFile, "client.go.gotemplate", interfaceDesc); err != nil {
		panic(err)
	}

	convertFile, err := os.Create(filepath.Join(targetPath, "convert.go"))
	if err != nil {
		panic(err)
	}
	if err := templates.ExecuteTemplate(convertFile, "convert.go.gotemplate", interfaceDesc); err != nil {
		panic(err)
	}

}

//go:embed templates/*
var templateFS embed.FS

var templates = template.Must(template.New("").Funcs(map[string]any{
	"increment": func(i int) int { return i + 1 },
	"dereference": func(t string) string {
		if t[0] != '*' {
			panic(fmt.Sprintf("%s is not a pointer type", t))
		}
		return t[1:]
	},
}).ParseFS(templateFS, "templates/*"))

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

var requiredPackages = map[string]string{}

func findImportAlias(pkg string) string {
	if alias := requiredPackages[pkg]; alias != "" {
		return alias
	}
	alias := path.Base(pkg)
	var index int
	for {
		inUse := false
		for _, inUseAlias := range requiredPackages {
			if alias == inUseAlias {
				inUse = true
				break
			}
		}
		if !inUse {
			break
		}
		index++
		alias = path.Base(pkg) + strconv.Itoa(index)
	}
	requiredPackages[pkg] = alias
	return alias
}

var grpcTypes []grpcStruct

type grpcField struct {
	Name string
	Type string
}

type grpcStruct struct {
	Name   string
	Fields []grpcField
}

var goTypeMap = map[reflect.Type]goType{}
var goTypes []goType

func parseType(t reflect.Type) goType {
	if parsedStruct, alreadyParsed := goTypeMap[t]; alreadyParsed {
		return parsedStruct
	}
	var goTyp goType
	goTyp.BaseName = t.Name()
	pkg := t.PkgPath()
	if pkg != "" {
		goTyp.GoName = findImportAlias(pkg) + "." + goTyp.BaseName
	} else {
		goTyp.GoName = goTyp.BaseName
	}
	switch t.Kind() {
	case reflect.String:
		goTyp.GrpcName = "string"
		goTyp.GrpcProtoName = "string"
		goTyp.Kind = kindPrimitive
	case reflect.Int64, reflect.Int:
		goTyp.GrpcName = "int64"
		goTyp.GrpcProtoName = "int64"
		goTyp.Kind = kindPrimitive
	case reflect.Bool:
		goTyp.GrpcName = "bool"
		goTyp.GrpcProtoName = "bool"
		goTyp.Kind = kindPrimitive
	case reflect.Slice:
		goTyp.Kind = kindSlice
		subtype := t.Elem()
		if subtype.Kind() == reflect.Uint8 {
			goTyp.Elem = &goType{
				GoName:   "byte",
				BaseName: "byte",
				GrpcName: "byte",
				Kind:     kindPrimitive,
			}
			goTyp.GrpcName = "[]byte"
			goTyp.GrpcProtoName = "bytes"
			if goTyp.GoName == "" {
				goTyp.BaseName = "Bytes"
				goTyp.GoName = "[]byte"
			}
		} else {
			subtype := parseType(t.Elem())
			goTyp.Elem = &subtype
			goTyp.GrpcProtoName = "repeated " + subtype.GrpcProtoName
			goTyp.GrpcName = "[]" + subtype.GrpcName
			if goTyp.GoName == "" {
				goTyp.BaseName = subtype.BaseName + "Slice"
				goTyp.GoName = "[]" + subtype.GoName
			}
		}
		goTypeMap[t] = goTyp
		goTypes = append(goTypes, goTyp)
	case reflect.Array:
		goTyp.Kind = kindArray
		subtype := t.Elem()
		if subtype.Kind() == reflect.Uint8 {
			goTyp.Elem = &goType{
				GoName:   "byte",
				BaseName: "byte",
				GrpcName: "byte",
				Kind:     kindPrimitive,
			}
			goTyp.GrpcName = "[]byte"
			goTyp.GrpcProtoName = "bytes"
			if goTyp.GoName == "" {
				goTyp.BaseName = "ByteArray" + strconv.Itoa(t.Len())
				goTyp.GoName = fmt.Sprintf("[%d]byte", t.Len())
			}
		} else {
			subtype := parseType(t.Elem())
			goTyp.Elem = &subtype
			goTyp.GrpcProtoName = "repeated " + subtype.GrpcProtoName
			goTyp.GrpcName = "[]" + subtype.GrpcName
			if goTyp.GoName == "" {
				goTyp.BaseName = subtype.BaseName + "Array" + strconv.Itoa(t.Len())
				goTyp.GoName = fmt.Sprintf("[%d]%s", t.Len(), subtype.GoName)
			}
		}
		goTypeMap[t] = goTyp
		goTypes = append(goTypes, goTyp)
	case reflect.Struct:
		name := toGrpcName(t.Name())
		var grpcType = grpcStruct{
			Name: name,
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldtype := parseType(field.Type)
			goTyp.Fields = append(goTyp.Fields, goField{
				Name: field.Name,
				Type: fieldtype,
			})
			grpcType.Fields = append(grpcType.Fields, grpcField{
				Name: field.Name,
				Type: fieldtype.GrpcProtoName,
			})
		}
		goTyp.GrpcName = "*" + name
		goTyp.GrpcProtoName = name
		goTyp.Kind = kindStruct
		goTypeMap[t] = goTyp
		goTypes = append(goTypes, goTyp)
		grpcTypes = append(grpcTypes, grpcType)
	default:
		panic("unsupported kind")
	}
	return goTyp
}

func toGrpcName(fieldName string) string {
	return fieldName
}

type grpcMethod struct {
	Name   string
	Param  string
	Result string
}

type grpcService struct {
	Name    string
	Methods []grpcMethod
	Types   []grpcStruct

	PackageName string
	PackagePath string
}

type interfaceDescription struct {
	Name    string
	Methods []interfaceMethod

	Package     string
	PackageBase string

	PackageName string

	RequiredPackages map[string]string

	RequiredTypes []goType
}

type interfaceMethod struct {
	Name    string
	Args    []goType
	Results []goType

	UsesError   bool
	UsesContext bool // TODO: Use / set
}

type goType struct {
	GoName   string
	BaseName string

	GrpcProtoName string
	GrpcName      string

	Kind string

	Fields []goField

	Elem *goType // Subtype for array / slice
}

const (
	kindPrimitive = "primitive"
	kindStruct    = "struct"
	kindArray     = "array"
	kindSlice     = "slice"
)

type goField struct {
	Name string
	Type goType
}
