package graphqlws

import (
	"sort"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// ChannelDescriptionRule provides calculation for fields and channels
type ChannelNameBuilder interface {
	GetFieldsAndArgs(doc *ast.Document, variables map[string]interface{}) []FieldWithArgs
}

// ChannelInfo is interface for channel description
// channel consists of field name and args
type FieldWithArgs interface {
	Field() string
	String() string
}

type fieldWithArgs struct {
	FieldWithArgs
	field string
	args  []*astArgs
}

func (f *fieldWithArgs) Field() string {
	return f.field
}

func (f *fieldWithArgs) String() string {
	sort.Slice(f.args, func(i, j int) bool {
		return f.args[i].Key <= f.args[j].Key
	})
	strList := []string{f.field}
	for _, arg := range f.args {
		strList = append(strList, arg.Val)
	}
	return strings.Join(strList, ":")
}

type astArgs struct {
	Key string
	Val string
}

type channelNameBuilder struct {
	ChannelNameBuilder
}

func NewChannelNameBuilder() *channelNameBuilder {
	return &channelNameBuilder{}
}

func (b *channelNameBuilder) GetSelectionSets(doc *ast.Document) []*ast.SelectionSet {
	// these functions brings from ast.go
	defs := operationDefinitionsWithOperation(doc, "subscription")
	return selectionSetsForOperationDefinitions(defs)
}

func ifToStr(d interface{}) string {
	if v, ok := d.(string); ok {
		return v
	}
	if v, ok := d.(int); ok {
		return strconv.Itoa(v)
	}
	// TODO: more complex rules...
	return ""
}

func getArgKeyValueFromValue(variables map[string]interface{}, arg *ast.Argument) *astArgs {
	var k, v string
	val := arg.Value
	if val.GetKind() == kinds.Variable {
		n := val.GetValue().(*ast.Name)
		k = n.Value
		vv, ok := variables[n.Value]
		if !ok {
			return nil
		}
		v = ifToStr(vv)
	} else {
		k = arg.Name.Value
		vv := arg.Value.GetValue().(string)
		v = ifToStr(vv)
	}
	if v != "" {
		return &astArgs{Key: k, Val: v}
	}
	return nil
}

func getArgsFromFieldArguments(variables map[string]interface{}, field *ast.Field) []*astArgs {
	args := []*astArgs{}
	for _, arg := range field.Arguments {
		a := getArgKeyValueFromValue(variables, arg)
		if a == nil {
			return nil
		}
		args = append(args, a)
	}
	return args
}

func (b *channelNameBuilder) GetFieldWithArgs(variables map[string]interface{}, set *ast.SelectionSet) FieldWithArgs {
	if len(set.Selections) < 1 {
		return nil
	}
	field, ok := set.Selections[0].(*ast.Field)
	if !ok {
		return nil
	}
	if args := getArgsFromFieldArguments(variables, field); args != nil {
		return &fieldWithArgs{field: field.Name.Value, args: args}
	}
	return nil
}

func (b *channelNameBuilder) GetFieldsAndArgs(doc *ast.Document, variables map[string]interface{}) []FieldWithArgs {
	sets := b.GetSelectionSets(doc)
	fieldList := []FieldWithArgs{}
	for _, set := range sets {
		if f := b.GetFieldWithArgs(variables, set); f != nil {
			fieldList = append(fieldList, f)
		}
	}
	return fieldList
}
