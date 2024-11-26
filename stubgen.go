package main

import (
	"bytes"
	"fmt"
	"klogproc/servicelog"
	"klogproc/servicelog/apiguard"
	"klogproc/servicelog/kontext013"
	"klogproc/servicelog/kontext015"
	"klogproc/servicelog/kontext018"
	"klogproc/servicelog/korpusdb"
	"klogproc/servicelog/kwords"
	"klogproc/servicelog/kwords2"
	"klogproc/servicelog/mapka"
	"klogproc/servicelog/mapka2"
	"klogproc/servicelog/mapka3"
	"klogproc/servicelog/masm"
	"klogproc/servicelog/morfio"
	"klogproc/servicelog/mquery"
	"klogproc/servicelog/mquerysru"
	"klogproc/servicelog/shiny"
	"klogproc/servicelog/ske"
	"klogproc/servicelog/syd"
	"klogproc/servicelog/treq"
	"klogproc/servicelog/vlo"
	"klogproc/servicelog/wag06"
	"klogproc/servicelog/wag07"
	"klogproc/servicelog/wsserver"
	"os"
	"reflect"
	"text/template"
)

type FieldInfo struct {
	Name        string
	Type        string
	IsContainer bool
	ContentType string
	Nested      []FieldInfo // For nested structs
	Indent      string      // For formatting nested fields
}

func getTypeString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number (integer)"
	case reflect.Float32, reflect.Float64:
		return "number (float)"
	case reflect.Bool:
		return "boolean"
	case reflect.Struct:
		return "struct"
	default:
		return t.String()
	}
}

func analyzeStruct(t reflect.Type, indent string) []FieldInfo {
	fields := make([]FieldInfo, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldInfo := FieldInfo{
			Name:   field.Name,
			Indent: indent,
		}

		// Handle different types including nested structs
		switch field.Type.Kind() {
		case reflect.Struct:
			fieldInfo.Type = "table (map)"
			fieldInfo.Nested = analyzeStruct(field.Type, indent+"  ")
		case reflect.String:
			fieldInfo.Type = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldInfo.Type = "number (integer)"
		case reflect.Float32, reflect.Float64:
			fieldInfo.Type = "number (float)"
		case reflect.Bool:
			fieldInfo.Type = "boolean"
		case reflect.Map:
			fieldInfo.Type = "table (map)"
			fieldInfo.IsContainer = true
			if field.Type.Elem().Kind() == reflect.Struct {
				fieldInfo.ContentType = fmt.Sprintf("key=%v, value=table", getTypeString(field.Type.Key()))
				fieldInfo.Nested = analyzeStruct(field.Type.Elem(), indent+"  ")
			} else {
				fieldInfo.ContentType = fmt.Sprintf("(%v, %v)",
					getTypeString(field.Type.Key()),
					getTypeString(field.Type.Elem()))
			}
		case reflect.Slice, reflect.Array:
			fieldInfo.Type = "table (seq)"
			fieldInfo.IsContainer = true
			if field.Type.Elem().Kind() == reflect.Struct {
				fieldInfo.ContentType = "table"
				fieldInfo.Nested = analyzeStruct(field.Type.Elem(), indent+"  ")
			} else {
				fieldInfo.ContentType = getTypeString(field.Type.Elem())
			}
		default:
			fieldInfo.Type = field.Type.String()
		}

		fields = append(fields, fieldInfo)
	}

	return fields
}

func generateLuaStubForType(inputRec servicelog.InputRecord, outputRec servicelog.OutputRecord) (string, error) {
	t1 := reflect.TypeOf(inputRec)
	if t1.Kind() == reflect.Ptr {
		t1 = t1.Elem()
	}
	if t1.Kind() != reflect.Struct {
		return "", fmt.Errorf("input must be a struct, got %v", t1.Kind())
	}
	t2 := reflect.TypeOf(outputRec)
	if t2.Kind() == reflect.Ptr {
		t2 = t2.Elem()
	}
	if t2.Kind() != reflect.Struct {
		return "", fmt.Errorf("input must be a struct, got %v", t2.Kind())
	}

	// Template for the Lua stub
	const stubTemplate = `--[[
Input record:
{{range .Fields1}}
{{.Indent}}{{.Name}} {{.Type}} {{if .IsContainer}}of {{.ContentType}}{{end}}{{if .Nested}}{{range .Nested}}
{{.Indent}}  {{.Name}} {{.Type}} {{if .IsContainer}}of {{.ContentType}}{{end}}{{end}}{{end}}{{end}}

Output record:
{{range .Fields2}}
{{.Indent}}{{.Name}} {{.Type}} {{if .IsContainer}}of {{.ContentType}}{{end}}{{if .Nested}}{{range .Nested}}
{{.Indent}}  {{.Name}} {{.Type}} {{if .IsContainer}}of {{.ContentType}}{{end}}{{end}}{{end}}{{end}}



-- transform function processes the input record and returns an output record
function transform(inputRec)
    local ans = outputRecord.new()
	-- setting an output field:
	set_out(ans, "{{.FirstFieldName}}", string.format("%s[modified]", logRec.{{.FirstFieldName}}))
    -- TODO: Transform input record to output format
    -- Available fields are documented above
    return ans
end
`

	fields1 := analyzeStruct(t1, "")
	fields2 := analyzeStruct(t2, "")

	// Execute template
	tmpl, err := template.New("luaStub").Parse(stubTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(
		&buf,
		struct {
			Fields1        []FieldInfo
			Fields2        []FieldInfo
			FirstFieldName string
		}{
			fields1,
			fields2,
			fields2[0].Name,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

func GenerateLuaStub(appType, version string) {
	var src string
	var err error
	switch appType {
	case servicelog.AppTypeAkalex:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shiny.OutputRecord{})
	case servicelog.AppTypeAPIGuard:
		src, err = generateLuaStubForType(&apiguard.InputRecord{}, &apiguard.OutputRecord{})
	case servicelog.AppTypeCalc:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shiny.OutputRecord{})
	case servicelog.AppTypeGramatikat:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shiny.OutputRecord{})
	case servicelog.AppTypeKontext:
		switch version {
		case "013":
			src, err = generateLuaStubForType(&kontext013.InputRecord{}, &kontext013.OutputRecord{})
		case "015":
			src, err = generateLuaStubForType(&kontext015.InputRecord{}, &kontext013.OutputRecord{})
		case "018":
			src, err = generateLuaStubForType(&kontext018.QueryInputRecord{}, &kontext018.OutputRecord{})
		default:
			panic("unknown kontext version") // TODO
		}
	case servicelog.AppTypeKontextAPI:
		// TODO
	case servicelog.AppTypeKorpusDB:
		src, err = generateLuaStubForType(&korpusdb.InputRecord{}, &korpusdb.OutputRecord{})
	case servicelog.AppTypeKwords:
		switch version {
		case "1":
			src, err = generateLuaStubForType(&kwords.InputRecord{}, &kwords.OutputRecord{})
		case "2":
			src, err = generateLuaStubForType(&kwords2.InputRecord{}, &kwords2.OutputRecord{})
		default:
			panic("unknown kwords version") // TODO
		}
	case servicelog.AppTypeLists:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shiny.OutputRecord{})
	case servicelog.AppTypeMapka:
		switch version {
		case "1":
			src, err = generateLuaStubForType(&mapka.InputRecord{}, &mapka.OutputRecord{})
		case "2":
			src, err = generateLuaStubForType(&mapka2.InputRecord{}, &mapka2.OutputRecord{})
		case "3":
			src, err = generateLuaStubForType(&mapka3.InputRecord{}, &mapka3.OutputRecord{})
		default:
			panic("unknown mapka version") // TODO
		}
	case servicelog.AppTypeMorfio:
		src, err = generateLuaStubForType(&morfio.InputRecord{}, &morfio.OutputRecord{})
	case servicelog.AppTypeQuitaUp:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shiny.OutputRecord{})
	case servicelog.AppTypeSke:
		src, err = generateLuaStubForType(&ske.InputRecord{}, &ske.OutputRecord{})
	case servicelog.AppTypeSyd:
		src, err = generateLuaStubForType(&syd.InputRecord{}, &syd.OutputRecord{})
	case servicelog.AppTypeTreq:
		src, err = generateLuaStubForType(&treq.InputRecord{}, &treq.OutputRecord{})
	case servicelog.AppTypeWag:
		switch version {
		case "0.6":
			src, err = generateLuaStubForType(&wag06.InputRecord{}, &wag06.OutputRecord{})
		case "0.7":
			src, err = generateLuaStubForType(&wag07.InputRecord{}, &wag06.OutputRecord{})
		default:
			panic("unknown wag version") // TODO
		}
	case servicelog.AppTypeWsserver:
		src, err = generateLuaStubForType(&wsserver.InputRecord{}, &wsserver.OutputRecord{})
	case servicelog.AppTypeMasm:
		src, err = generateLuaStubForType(&masm.InputRecord{}, &masm.OutputRecord{})
	case servicelog.AppTypeMquery:
		src, err = generateLuaStubForType(&mquery.InputRecord{}, &mquery.OutputRecord{})
	case servicelog.AppTypeMquerySRU:
		src, err = generateLuaStubForType(&mquerysru.InputRecord{}, &mquerysru.OutputRecord{})
	case servicelog.AppTypeVLO:
		src, err = generateLuaStubForType(&vlo.InputRecord{}, &vlo.OutputRecord{})
	default:
		fmt.Fprintf(os.Stderr, "error generating Lua script stub - unknown app: %s\n", appType)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating Lua script stub: %s\n", err)
		os.Exit(1)

	} else {
		fmt.Println(src)
	}
}
