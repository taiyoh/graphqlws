package graphqlws

import (
	"sort"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// ChannelDescriptionRule provides calculation for fields and channels
type ChannelDescriptionRule interface {
	FillFieldsAndChannels(subscription *Subscription)
}

// ChannelInfo is interface for channel description
// channel consists of field name and args
type ChannelInfo interface {
	Field() string
	Describe() string
}

type channelInfo struct {
	ChannelInfo
	field string
	args  []*astArgs
}

func (chi *channelInfo) Field() string {
	return chi.field
}

func (chi *channelInfo) Describe() string {
	sort.Slice(chi.args, func(i, j int) bool {
		return chi.args[i].Key <= chi.args[j].Key
	})
	strList := []string{chi.field}
	for _, arg := range chi.args {
		strList = append(strList, arg.Val)
	}
	return strings.Join(strList, ":")
}

type astArgs struct {
	Key string
	Val string
}

type channelDescriptionRule struct {
	ChannelDescriptionRule
}

func NewChannelDescriptionRule() *channelDescriptionRule {
	return &channelDescriptionRule{}
}

func (r *channelDescriptionRule) GetSelectionSets(doc *ast.Document) []*ast.SelectionSet {
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

func (r *channelDescriptionRule) GetChannelInfoList(variables map[string]interface{}, set *ast.SelectionSet) []ChannelInfo {
	chList := []ChannelInfo{}
	if len(set.Selections) < 1 {
		return chList
	}
	field, ok := set.Selections[0].(*ast.Field)
	if !ok {
		return chList
	}
	if args := getArgsFromFieldArguments(variables, field); args != nil {
		chList = append(chList, &channelInfo{
			field: field.Name.Value,
			args:  args,
		})
	}
	return chList
}

func (r *channelDescriptionRule) FillFieldsAndChannels(subscription *Subscription) {
	sets := r.GetSelectionSets(subscription.Document)
	fieldList := []string{}
	channelList := []string{}
	for _, set := range sets {
		if chList := r.GetChannelInfoList(subscription.Variables, set); len(chList) > 0 {
			for _, ch := range chList {
				channelList = append(channelList, ch.Describe())
				fieldList = append(fieldList, ch.Field())
			}
		}
	}
	subscription.Channels = channelList
	subscription.Fields = fieldList
}
