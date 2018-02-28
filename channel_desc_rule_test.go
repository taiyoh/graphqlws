package graphqlws_test

import (
	"testing"

	"github.com/functionalfoundry/graphqlws"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type connForTest struct {
	graphqlws.Connection
	id              string
	user            interface{}
	ReceivedOpID    string
	ReceivedPayload *graphqlws.DataMessagePayload
	ReceivedError   error
}

func (c *connForTest) ID() string {
	return c.id
}

func (c *connForTest) User() interface{} {
	return c.user
}

func (c *connForTest) SendData(opID string, d *graphqlws.DataMessagePayload) {
	c.ReceivedOpID = opID
	c.ReceivedPayload = d
}

func (c *connForTest) SendError(e error) {
	c.ReceivedError = e
}
func TestTestTest(t *testing.T) {
	rule := graphqlws.NewChannelDescriptionRule()

	conn := &connForTest{
		id: "hoge",
		user: map[string]interface{}{
			"id": "fuga",
		},
	}

	sub := &graphqlws.Subscription{
		ID:            "foo",
		OperationName: "",
		Connection:    conn,
		SendData:      func(d *graphqlws.DataMessagePayload) {},
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

}
