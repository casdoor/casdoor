package ldap

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ber "github.com/go-asn1-ber/asn1-ber"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/lor00x/goldap/message"
	"github.com/xorm-io/builder"
)

func args(exp ...interface{}) []interface{} {
	return exp
}

func TestLdapFilterAsQuery(t *testing.T) {
	scenarios := []struct {
		description  string
		input        string
		expectedExpr string
		expectedArgs []interface{}
	}{
		{"Should be SQL for FilterAnd", "(&(mail=2)(email=1))", "email=? AND email=?", args("2", "1")},
		{"Should be SQL for FilterOr", "(|(mail=2)(email=1))", "email=? OR email=?", args("2", "1")},
		{"Should be SQL for FilterNot", "(!(mail=2))", "NOT email=?", args("2")},
		{"Should be SQL for FilterEqualityMatch", "(mail=2)", "email=?", args("2")},
		{"Should be SQL for FilterPresent", "(mail=*)", "email IS NOT NULL", nil},
		{"Should be SQL for FilterGreaterOrEqual", "(mail>=admin)", "email>=?", args("admin")},
		{"Should be SQL for FilterLessOrEqual", "(mail<=admin)", "email<=?", args("admin")},
		{"Should be SQL for FilterSubstrings", "(mail=admin*ex*c*m)", "email LIKE ?", args("admin%ex%c%m")},
	}

	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			searchRequest, err := buildLdapSearchRequest(scenery.input)
			if err != nil {
				assert.FailNow(t, "Unable to create searchRequest", err)
			}
			m, err := message.ReadLDAPMessage(message.NewBytes(0, searchRequest.Bytes()))
			if err != nil {
				assert.FailNow(t, "Unable to create searchRequest", err)
			}
			req := m.ProtocolOp().(message.SearchRequest)

			cond, err := buildUserFilterCondition(req.Filter())
			if err != nil {
				assert.FailNow(t, "Unable to build condition", err)
			}
			expr, args, err := builder.ToSQL(cond)
			if err != nil {
				assert.FailNow(t, "Unable to build sql", err)
			}

			assert.Equal(t, scenery.expectedExpr, expr)
			assert.Equal(t, scenery.expectedArgs, args)
		})
	}
}

func buildLdapSearchRequest(filter string) (*ber.Packet, error) {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, "MessageID"))

	pkt := ber.Encode(ber.ClassApplication, ber.TypeConstructed, goldap.ApplicationSearchRequest, nil, "Search Request")
	pkt.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "Base DN"))
	pkt.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, "Scope"))
	pkt.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, "Deref Aliases"))
	pkt.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 0, "Size Limit"))
	pkt.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 0, "Time Limit"))
	pkt.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, false, "Types Only"))
	// compile and encode filter
	filterPacket, err := goldap.CompileFilter(filter)
	if err != nil {
		return nil, err
	}
	pkt.AppendChild(filterPacket)
	// encode attributes
	attributesPacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attributes")
	attributesPacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "*", "Attribute"))
	pkt.AppendChild(attributesPacket)

	packet.AppendChild(pkt)

	return packet, nil
}
