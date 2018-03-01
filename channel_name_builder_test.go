package graphqlws

import (
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type connForTest struct {
	Connection
}

func (c *connForTest) ID() string {
	return ""
}

func (c *connForTest) User() interface{} {
	return struct{}{}
}

func (c *connForTest) SendData(opID string, d *DataMessagePayload) {}

func (c *connForTest) SendError(e error) {}

func TestChannelDescriptionRule(t *testing.T) {
	b := NewChannelNameBuilder()

	conn := &connForTest{}

	sub := &Subscription{
		ID:            "foo",
		OperationName: "",
		Connection:    conn,
		SendData:      func(d *DataMessagePayload) {},
	}

	initSubscription := func(query string, doc *ast.Document) {
		sub.Query = query
		sub.Document = doc
		sub.Fields = []FieldWithArgs{}
	}

	t.Run("query without args", func(t *testing.T) {
		query := `
			subscription {
				hello {
					foo
					bar
				}
			}
		`

		document, _ := parser.Parse(parser.ParseParams{
			Source: query,
		})

		initSubscription(query, document)
		fields := b.GetFieldsAndArgs(document, sub.Variables)

		if len(fields) != 1 {
			t.Error("filled fields count should be 1")
		}
		if fields[0].Field() != "hello" {
			t.Error("filled field must be 'hello', actually: ", fields[0].Field())
		}

		if fields[0].String() != "hello" {
			t.Error("filled channel must be 'hello', actually: ", fields[0].String())
		}

	})

	t.Run("simple query", func(t *testing.T) {
		query := `
			subscription {
				hello(id: 1, aaa: "fuu") {
					foo
					bar
				}
			}
		`

		document, _ := parser.Parse(parser.ParseParams{
			Source: query,
		})

		initSubscription(query, document)
		fields := b.GetFieldsAndArgs(document, sub.Variables)

		if len(fields) != 1 {
			t.Error("filled fields count should be 1")
		}
		if fields[0].Field() != "hello" {
			t.Error("filled field must be 'hello', actually: ", fields[0].Field())
		}

		if fields[0].String() != "hello:fuu:1" {
			t.Error("filled channel must be 'hello:fuu:1', actually: ", fields[0].String())
		}

	})

	t.Run("query with variables", func(t *testing.T) {
		query := `
			subscription mySubscribe($id: ID!, $aaa: String!) {
				hello(id: $id, aaa: $aaa) {
					foo
					bar
				}
			}
		`

		document, _ := parser.Parse(parser.ParseParams{
			Source: query,
		})

		initSubscription(query, document)
		sub.Variables = map[string]interface{}{
			"id":  2,
			"aaa": "bbb",
		}

		fields := b.GetFieldsAndArgs(document, sub.Variables)

		if len(fields) != 1 {
			t.Error("filled fields count should be 1")
		}
		if fields[0].Field() != "hello" {
			t.Error("filled field must be 'hello', actually: ", fields[0].Field())
		}

		if fields[0].String() != "hello:bbb:2" {
			t.Error("filled channel must be 'hello:bbb:2', actually: ", fields[0].String())
		}

	})

	t.Run("query with variables, but no assign", func(t *testing.T) {
		query := `
			subscription mySubscribe($id: ID!, $aaa: String!) {
				hello(id: $id, aaa: $aaa) {
					foo
					bar
				}
			}
		`

		document, _ := parser.Parse(parser.ParseParams{
			Source: query,
		})

		initSubscription(query, document)
		sub.Variables = map[string]interface{}{}

		fields := b.GetFieldsAndArgs(document, sub.Variables)

		if len(fields) != 0 {
			t.Error("filled fields count should be 0, actually: ", fields[0])
		}
	})

}
