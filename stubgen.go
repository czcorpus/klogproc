package main

import (
	"bytes"
	"fmt"
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
	"reflect"
	"text/template"

	apiguardCore "github.com/czcorpus/klogproc-core/storage/apiguard"
	k013Core "github.com/czcorpus/klogproc-core/storage/kontext013"
	k015Core "github.com/czcorpus/klogproc-core/storage/kontext015"
	kdbCore "github.com/czcorpus/klogproc-core/storage/korpusdb"
	kwordsCore "github.com/czcorpus/klogproc-core/storage/kwords"
	kwords2Core "github.com/czcorpus/klogproc-core/storage/kwords2"
	mapkaCore "github.com/czcorpus/klogproc-core/storage/mapka"
	mapka2Core "github.com/czcorpus/klogproc-core/storage/mapka2"
	mapka3Core "github.com/czcorpus/klogproc-core/storage/mapka3"
	masmCore "github.com/czcorpus/klogproc-core/storage/masm"
	morfioCore "github.com/czcorpus/klogproc-core/storage/morfio"
	mqueryCore "github.com/czcorpus/klogproc-core/storage/mquery"
	mquerySRUCore "github.com/czcorpus/klogproc-core/storage/mquerysru"
	shinyCore "github.com/czcorpus/klogproc-core/storage/shiny"
	skeCore "github.com/czcorpus/klogproc-core/storage/ske"
	sydCore "github.com/czcorpus/klogproc-core/storage/syd"
	treqCore "github.com/czcorpus/klogproc-core/storage/treq"
	vloCore "github.com/czcorpus/klogproc-core/storage/vlo"
	wag06Core "github.com/czcorpus/klogproc-core/storage/wag06"
	wsserverCore "github.com/czcorpus/klogproc-core/storage/wsserver"

	"github.com/czcorpus/klogproc-core/storage"
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

func generateLuaStubForType(inputRec storage.InputRecord, outputRec storage.OutputRecord) (string, error) {
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

To get a record time in a normalized (RFC3339) format:

input_rec.GetTime()

Output record:
{{range .Fields2}}
{{.Indent}}{{.Name}} {{.Type}} {{if .IsContainer}}of {{.ContentType}}{{end}}{{if .Nested}}{{range .Nested}}
{{.Indent}}  {{.Name}} {{.Type}} {{if .IsContainer}}of {{.ContentType}}{{end}}{{end}}{{end}}{{end}}

To create an output record using default (hardcoded)
transformation (suitable if you want to perform just
a slight modification with your Lua script):

out = transform_default(input_rec, tz_shift_mins)

To create a new output record from scratch:

out = new_out_record()

To set a property in output record:

set_out_prop(rec, name, value)

To test if a record (input or output) has a property:

if record_prop_exists(any_rec, name, value)

For debugging, use:

dump = require('dump')
print(dump(my_value))

To generate a deterministic ID:

local id = out_rec_deterministic_id(out_rec)

For logging:

log.info("message", map_with_args)

(also: warn, debug, error)
The second arg. is optional

Global variables:

app_type, app_version, anonymous_users

]]--

-- preprocess function gives a way how to
-- decide whether the input_rec should be processed
-- (just return {} to ignore the record) or it can
-- be transformed into a set of records with help of
-- the buffer (this is e.g. used with mapka3 to generate
-- "fake" records representing clusters of user activity).
function preprocess(input_rec, buffer)
	-- if you want to make sure a possible hardcoded
    -- preprocess is always applied, use:
	-- local out = preprocess_default(input_rec, buffer)
	-- otherwise just return a table of size 1 with the orig. record:
    return {input_rec}
end

-- transform function processes the input record and returns an output record
function transform(input_rec)
    local out = transform_default(input_rec)
	-- setting an output field:
	set_out_prop(out, "{{.FirstFieldName}}", string.format("%s[modified]", input_rec.{{.FirstFieldName}}))
    -- TODO: Transform input record to output format
    -- Available fields and functions are documented above
    return out
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

func generateLuaStub(appType, version string) error {
	var src string
	var err error
	switch appType {
	case storage.AppTypeAkalex:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shinyCore.OutputRecord{})
	case storage.AppTypeAPIGuard:
		src, err = generateLuaStubForType(&apiguard.InputRecord{}, &apiguardCore.OutputRecord{})
	case storage.AppTypeCalc:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shinyCore.OutputRecord{})
	case storage.AppTypeGramatikat:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shinyCore.OutputRecord{})
	case storage.AppTypeKontext:
		switch version {
		case storage.AppVersionKontext013,
			storage.AppVersionKontext014:
			src, err = generateLuaStubForType(&kontext013.InputRecord{}, &k013Core.OutputRecord{})
		case storage.AppVersionKontext015,
			storage.AppVersionKontext016,
			storage.AppVersionKontext017,
			storage.AppVersionKontext017API,
			storage.AppVersionKontext018:
			src, err = generateLuaStubForType(&kontext015.InputRecord{}, &k015Core.OutputRecord{})
		case "018":
			src, err = generateLuaStubForType(&kontext018.InputRecord{}, &k015Core.OutputRecord{})
		default:
			return fmt.Errorf("failed to create Lua script stub for 'kontext': unknown version '%s'", version)
		}
	case storage.AppTypeKorpusDB:
		src, err = generateLuaStubForType(&korpusdb.InputRecord{}, &kdbCore.OutputRecord{})
	case storage.AppTypeKwords:
		switch version {
		case "1":
			src, err = generateLuaStubForType(&kwords.InputRecord{}, &kwordsCore.OutputRecord{})
		case "2":
			src, err = generateLuaStubForType(&kwords2.InputRecord{}, &kwords2Core.OutputRecord{})
		default:
			return fmt.Errorf("failed to create Lua script stub for 'kwords': unknown version '%s'", version)
		}
	case storage.AppTypeLists:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shinyCore.OutputRecord{})
	case storage.AppTypeMapka:
		switch version {
		case "1":
			src, err = generateLuaStubForType(&mapka.InputRecord{}, &mapkaCore.OutputRecord{})
		case "2":
			src, err = generateLuaStubForType(&mapka2.InputRecord{}, &mapka2Core.OutputRecord{})
		case "3":
			src, err = generateLuaStubForType(&mapka3.InputRecord{}, &mapka3Core.OutputRecord{})
		default:
			return fmt.Errorf("failed to create Lua script stub for 'mapka': unknown version '%s'", version)
		}
	case storage.AppTypeMorfio:
		src, err = generateLuaStubForType(&morfio.InputRecord{}, &morfioCore.OutputRecord{})
	case storage.AppTypeQuitaUp:
		src, err = generateLuaStubForType(&shiny.InputRecord{}, &shinyCore.OutputRecord{})
	case storage.AppTypeSke:
		src, err = generateLuaStubForType(&ske.InputRecord{}, &skeCore.OutputRecord{})
	case storage.AppTypeSyd:
		src, err = generateLuaStubForType(&syd.InputRecord{}, &sydCore.OutputRecord{})
	case storage.AppTypeTreq:
		src, err = generateLuaStubForType(&treq.InputRecord{}, &treqCore.OutputRecord{})
	case storage.AppTypeWag:
		switch version {
		case "0.6":
			src, err = generateLuaStubForType(&wag06.InputRecord{}, &wag06Core.OutputRecord{})
		case "0.7":
			src, err = generateLuaStubForType(&wag07.InputRecord{}, &wag06Core.OutputRecord{})
		default:
			return fmt.Errorf("failed to create Lua script stub for 'wag': unknown version '%s'", version)
		}
	case storage.AppTypeWsserver:
		src, err = generateLuaStubForType(&wsserver.InputRecord{}, &wsserverCore.OutputRecord{})
	case storage.AppTypeMasm:
		src, err = generateLuaStubForType(&masm.InputRecord{}, &masmCore.OutputRecord{})
	case storage.AppTypeMquery:
		src, err = generateLuaStubForType(&mquery.InputRecord{}, &mqueryCore.OutputRecord{})
	case storage.AppTypeMquerySRU:
		src, err = generateLuaStubForType(&mquerysru.InputRecord{}, &mquerySRUCore.OutputRecord{})
	case storage.AppTypeVLO:
		src, err = generateLuaStubForType(&vlo.InputRecord{}, &vloCore.OutputRecord{})
	default:
		return fmt.Errorf("failed to create Lua script stub: unknown application '%s'", version)
	}
	if err != nil {
		return fmt.Errorf("failed to create Lua script stub for %s: %w", appType, err)

	} else {
		fmt.Println(src)
	}
	return nil
}
