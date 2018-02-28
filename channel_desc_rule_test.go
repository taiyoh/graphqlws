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

func TestTestTest(t *testing.T) {
	rule := NewChannelDescriptionRule()

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
		sub.Channels = []string{}
		sub.Fields = []string{}
	}

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
		rule.Calculate(sub)

		if len(sub.Channels) != 1 {
			t.Error("filled channels count should be 1")
		}
		if sub.Channels[0] != "hello:fuu:1" {
			t.Error("filled channel must be 'hello:fuu:1', actualy: ", sub.Channels[0])
		}

		if len(sub.Fields) != 1 {
			t.Error("filled fields count should be 1")
		}
		if sub.Fields[0] != "hello" {
			t.Error("filled field must be 'hello', actualy: ", sub.Fields[0])
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

		rule.Calculate(sub)

		if len(sub.Channels) != 1 {
			t.Error("filled channels count should be 1")
		}
		if sub.Channels[0] != "hello:bbb:2" {
			t.Error("filled channel must be 'hello:bbb:2', actualy: ", sub.Channels[0])
		}

		if len(sub.Fields) != 1 {
			t.Error("filled fields count should be 1")
		}
		if sub.Fields[0] != "hello" {
			t.Error("filled field must be 'hello', actualy: ", sub.Fields[0])
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

		rule.Calculate(sub)

		if len(sub.Channels) != 0 {
			t.Error("filled channels count should be 0, actualy: ", sub.Channels[0])
		}
		if len(sub.Fields) != 0 {
			t.Error("filled fields count should be 0, actualy: ", sub.Fields[0])
		}
	})

}
