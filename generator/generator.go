package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"

	"github.com/dsnidr/encgen-go/parser"
)

//go:embed template.gotmpl
var encoderTemplate string

type TemplateData struct {
	Version string
	*parser.StructInfo
}

func GenerateEncoder(structInfo *parser.StructInfo, outPath, version string) error {
	tmpl := template.New("encoder").Funcs(template.FuncMap{
		"nextFieldType":     nextFieldType,
		"nextFieldTypeName": nextFieldTypeName,
		"isLastField":       isLastField,
		"isBatchableNext":   isBatchableNext,
		"sub":               func(a, b int) int { return a - b },
	})

	tmpl = template.Must(tmpl.Parse(encoderTemplate))

	outFilePath := filepath.Join(outPath, fmt.Sprintf("%s_encoder.go", strings.ToLower(toSnakeCase(structInfo.Name))))
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := tmpl.Execute(outFile, &TemplateData{
		Version:    version,
		StructInfo: structInfo,
	}); err != nil {
		return err
	}

	fmt.Printf("Generated %s\n", outFilePath)
	return nil
}

func toSnakeCase(input string) string {
	result := ""

	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}

	return result
}

func nextFieldType(structName string, fields []parser.StructField, i int) string {
	if i == len(fields)-1 {
		return structName + "FinishEncoder"
	}

	nextField := fields[i+1]
	if nextField.Batchable {
		return structName + nextField.Name + "EncoderStarter"
	}

	return structName + nextField.Name + "Encoder"
}

func nextFieldTypeName(structName string, fields []parser.StructField, i int) string {
	return nextFieldType(structName, fields, i)
}

func isBatchableNext(fields []parser.StructField, i int) bool {
	if i == len(fields)-1 {
		return false
	}
	return fields[i+1].Batchable
}

func isLastField(fields []parser.StructField, i int) bool {
	return i == len(fields)-1
}
