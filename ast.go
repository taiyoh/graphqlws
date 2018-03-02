package graphqlws

import (
	"sort"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// FieldWithArgsSerializer interface provides getter for serialized field and args of query.
// Its building logic is depends on user's implementation.
// Below is reference implementation.
type FieldWithArgsSerializer interface {
	Field() string
	Serialized() string
}

type fieldWithArgsSerializer struct {
	FieldWithArgsSerializer
	field      string
	serialized string
	args       map[string]string
}

func (fa *fieldWithArgsSerializer) init() {
	sargs := []string{}
	for k := range fa.args {
		sargs = append(sargs, k)
	}
	sort.Slice(sargs, func(i, j int) bool {
		return sargs[i] <= sargs[j]
	})
	strList := []string{fa.field}
	for i := range sargs {
		strList = append(strList, fa.args[sargs[i]])
	}
	fa.serialized = strings.Join(strList, ":")
}

func (fa *fieldWithArgsSerializer) Field() string {
	return fa.field
}

func (fa *fieldWithArgsSerializer) Serialized() string {
	return fa.serialized
}

type newFieldWithArgsSerializerFunc func(f string, a map[string]string) FieldWithArgsSerializer

func (fn newFieldWithArgsSerializerFunc) Generate(f string, a map[string]string) FieldWithArgsSerializer {
	return fn(f, a)
}

// FieldWithArgsFactory interface provides Generate method for getting FieldArgs from field name and args
type FieldWithArgsFactory interface {
	Generate(field string, args map[string]string) FieldWithArgsSerializer
}

func getNewFieldWithArgsSerializerFunc() newFieldWithArgsSerializerFunc {
	return func(f string, a map[string]string) FieldWithArgsSerializer {
		fa := &fieldWithArgsSerializer{
			field: f,
			args:  a,
		}
		fa.init()
		return fa
	}
}

func operationDefinitionsWithOperation(
	doc *ast.Document,
	op string,
) []*ast.OperationDefinition {
	defs := []*ast.OperationDefinition{}
	for _, node := range doc.Definitions {
		if node.GetKind() == "OperationDefinition" {
			if def, ok := node.(*ast.OperationDefinition); ok {
				if def.Operation == op {
					defs = append(defs, def)
				}
			}
		}
	}
	return defs
}

func selectionSetsForOperationDefinitions(
	defs []*ast.OperationDefinition,
) []*ast.SelectionSet {
	sets := []*ast.SelectionSet{}
	for _, def := range defs {
		if set := def.GetSelectionSet(); set != nil {
			sets = append(sets, set)
		}
	}
	return sets
}

func ifToStr(d interface{}) string {
	if v, ok := d.(string); ok {
		return v
	}
	if v, ok := d.(int); ok {
		return strconv.Itoa(v)
	}
	// TODO: more complex rules...(eg: input type)
	return ""
}

func getArgKeyValueFromValue(variables map[string]interface{}, arg *ast.Argument) (string, string, bool) {
	var k, v string
	val := arg.Value
	if val.GetKind() == kinds.Variable {
		n := val.GetValue().(*ast.Name)
		k = n.Value
		vv, ok := variables[n.Value]
		if !ok {
			return "", "", false
		}
		v = ifToStr(vv)
	} else {
		k = arg.Name.Value
		vv := arg.Value.GetValue().(string)
		v = ifToStr(vv)
	}
	if v != "" {
		return k, v, true
	}
	return "", "", false
}

func getArgsFromFieldArguments(variables map[string]interface{}, field *ast.Field) map[string]string {
	args := map[string]string{}
	for _, arg := range field.Arguments {
		if k, v, ok := getArgKeyValueFromValue(variables, arg); ok {
			args[k] = v
		}
	}
	return args
}

func getFieldWithArgs(variables map[string]interface{}, set *ast.SelectionSet) (string, map[string]string) {
	if len(set.Selections) < 1 {
		return "", nil
	}
	field, ok := set.Selections[0].(*ast.Field)
	if !ok {
		return "", nil
	}
	if args := getArgsFromFieldArguments(variables, field); args != nil {
		return field.Name.Value, args
	}
	return "", nil
}

func subscriptionFieldNamesFromDocument(doc *ast.Document, variables map[string]interface{}, fn FieldWithArgsFactory) []FieldWithArgsSerializer {
	defs := operationDefinitionsWithOperation(doc, "subscription")
	sets := selectionSetsForOperationDefinitions(defs)
	fieldList := []FieldWithArgsSerializer{}
	for _, set := range sets {
		if f, args := getFieldWithArgs(variables, set); f != "" {
			fieldList = append(fieldList, fn.Generate(f, args))
		}
	}
	return fieldList
}
