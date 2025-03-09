package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
)

type StructInfo struct {
	Package string
	Name    string
	Fields  []StructField
}

type StructField struct {
	// The field's raw name
	Name string

	// The value of the json tag if provided, defaults to field name otherwise
	JSONKey string

	// Whether or not this field has been tagged as batchable
	Batchable bool

	// If this field is a struct, it will have nested fields inside. SubFields holds those nested fields.
	SubFields []StructField

	TypeInfo TypeInfo
}

type TypeInfo struct {
	// The field's type, for example "[]*Item"
	TypeString string

	// If this type represents a slice, this stores the internal element type. For example "*Item"
	ElemType string

	// True if this type is a scalar (e.g. not a struct).
	IsScalar bool
}

// ParseStruct walks the pkgPath to gather all type declarations and then compares type names to structName
// to find the target struct, returning it's fields (and subfields if any exist) as a StructInfo.
func ParseStruct(pkgPath, structName string) (*StructInfo, error) {
	fset := token.NewFileSet()
	typeSpecs := make(map[string]*ast.TypeSpec)
	var pkgName string

	if err := filepath.Walk(pkgPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".go" || info.IsDir() {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		if pkgName == "" {
			// capture package name once
			pkgName = node.Name.Name
		}

		// Iterate through all declarations and gather all struct type definitions
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				typeSpecs[typeSpec.Name.Name] = typeSpec
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if pkgName == "" {
		return nil, fmt.Errorf("no Go files found in %s or package name could not be identified", pkgPath)
	}

	// Check if the target struct was found
	spec, ok := typeSpecs[structName]
	if !ok {
		return nil, fmt.Errorf("struct %q was not found in package at path %q", structName, pkgPath)
	}

	// Make sure it's a struct
	target, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("%q is not a struct", structName)
	}

	// Parse fields belonging to the target struct
	info := &StructInfo{
		Package: pkgName,
		Name:    structName,
	}

	visited := map[string]bool{}
	info.Fields = parseStruct(target, typeSpecs, visited)

	if len(info.Fields) == 0 {
		return nil, fmt.Errorf("struct %q has no fields", structName)
	}

	return info, nil
}

// parseStruct parses an *ast.StructType and returns a slice containing it's fields
func parseStruct(st *ast.StructType, typeSpecs map[string]*ast.TypeSpec, visited map[string]bool) []StructField {
	var fields []StructField

	for _, field := range st.Fields.List {
		var fieldName string
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		}

		// Default jsonKey to the field's name and replace with tag value if available
		jsonKey := fieldName
		batchable := false

		if field.Tag != nil {
			tagString := strings.TrimSpace(strings.Trim(field.Tag.Value, "`"))
			tag := reflect.StructTag(tagString)

			if jsonTag := tag.Get("json"); jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if len(parts) > 0 && parts[0] != "" {
					jsonKey = parts[0]
				}
			}

			if encTag := tag.Get("enc"); encTag == "batch" {
				batchable = true
			}
		}

		typeInfo := parseTypeInfo(field.Type)
		unwrappedType := unwrapExpr(field.Type)

		subFields := []StructField{}

		if named, ok := getUnderlyingNamedType(unwrappedType); ok {
			// Named type
			if typeSpec, ok := typeSpecs[named]; ok && !visited[named] {
				visited[named] = true

				if subStruct, ok := typeSpec.Type.(*ast.StructType); ok {
					subFields = parseStruct(subStruct, typeSpecs, visited)
				}
			}
		} else {
			// Unnamed type: could be an inline struct, a pointer to an inline struct, etc
			switch ft := unwrappedType.(type) {
			case *ast.StructType:
				subFields = parseStruct(ft, typeSpecs, visited)
			case *ast.StarExpr:
				if star, ok := ft.X.(*ast.StructType); ok {
					subFields = parseStruct(star, typeSpecs, visited)
				}
			}
		}

		fields = append(fields, StructField{
			Name:      fieldName,
			JSONKey:   jsonKey,
			Batchable: batchable,
			SubFields: subFields,
			TypeInfo:  typeInfo,
		})
	}

	return fields
}

func parseTypeInfo(t ast.Expr) TypeInfo {
	return TypeInfo{
		TypeString: types.ExprString(t),
		ElemType:   getSliceElementType(t),
		IsScalar:   isScalarType(t),
	}
}

// getUnderlyingNamedType checks if the field references a named type, and returns the
// type's name if found.
func getUnderlyingNamedType(expr ast.Expr) (string, bool) {
	name, valid := "", false

	switch t := expr.(type) {
	case *ast.Ident:
		name, valid = t.Name, true
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			name, valid = id.Name, true
		}
	case *ast.ArrayType:
		switch el := t.Elt.(type) {
		case *ast.Ident:
			name, valid = el.Name, true
		case *ast.StarExpr:
			if id, ok := el.X.(*ast.Ident); ok {
				name, valid = id.Name, true
			}
		}
	}

	if valid && isBuiltinType(name) {
		// If builtin, it's not a user-defined named type
		name, valid = "", false
	}

	return name, valid
}

// getSliceElementType returns the element type from a slice type expression.
func getSliceElementType(expr ast.Expr) string {
	switch slice := expr.(type) {
	case *ast.ArrayType:
		return types.ExprString(slice.Elt)
	default:
		return types.ExprString(expr)
	}
}

// isScalarType returns true if the provided expression represents a scalar
// and false if it represents a composite value (e.g. a struct).
func isScalarType(expr ast.Expr) bool {
	expr = unwrapExpr(expr)

	switch t := expr.(type) {
	case *ast.StructType:
		return false
	case *ast.StarExpr:
		if _, ok := t.X.(*ast.StructType); ok {
			return false
		}
	}

	return true
}

// unwrapExpr unwraps nested pointers and slice types.
// For example, passing in a []*[]**Item expression would return the underlying Item expr
func unwrapExpr(expr ast.Expr) ast.Expr {
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.ArrayType:
			expr = t.Elt
		default:
			return expr
		}
	}
}

func isBuiltinType(name string) bool {
	obj := types.Universe.Lookup(name)
	if obj == nil {
		return false
	}

	_, ok := obj.(*types.Builtin)
	return ok
}
