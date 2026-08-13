package ldap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	ber "github.com/go-asn1-ber/asn1-ber"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/lor00x/goldap/message"
	"github.com/xorm-io/builder"

	"github.com/casdoor/casdoor/object"
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
		{"Should be SQL for country attribute c", "(c=US)", "region=?", args("US")},
		{"Should be SQL for country attribute co", "(co=United States)", "region=?", args("United States")},
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

			q, err := buildUserFilter(req.Filter())
			if err != nil {
				assert.FailNow(t, "Unable to build condition", err)
			}
			expr, args, err := builder.ToSQL(q.condition)
			if err != nil {
				assert.FailNow(t, "Unable to build sql", err)
			}

			assert.Equal(t, scenery.expectedExpr, expr)
			assert.Equal(t, scenery.expectedArgs, args)
		})
	}
}

func TestLdapFilterOnSyntheticAttributes(t *testing.T) {
	alice := &object.User{Name: "alice", UidNumber: 10001}
	bob := &object.User{Name: "bob"}
	users := []*object.User{alice, bob}

	unassigned := "uid_number=? OR uid_number IS NULL"
	byNumber := "uid_number=? OR " + unassigned

	scenarios := []struct {
		description   string
		input         string
		expectedExpr  string
		expectedArgs  []interface{}
		expectedUsers []*object.User
	}{
		{
			"Should match a user by an assigned uidNumber",
			"(&(objectClass=posixAccount)(uidNumber=10001))",
			"(1 = 1) AND (" + byNumber + ")",
			args(10001, 0),
			[]*object.User{alice},
		},
		{
			"Should match a user by a name-derived uidNumber",
			"(&(objectClass=posixAccount)(uidNumber=" + getUidNumber(bob) + "))",
			"(1 = 1) AND (" + byNumber + ")",
			args(int(hash(bob.Name)), 0),
			[]*object.User{bob},
		},
		{
			"Should match a user by gidNumber",
			"(gidNumber=10001)",
			byNumber,
			args(10001, 0),
			[]*object.User{alice},
		},
		{
			"Should match a user by homeDirectory",
			"(homeDirectory=/home/alice)",
			"1 = 1",
			nil,
			[]*object.User{alice},
		},
		{
			"Should match no user for an unknown uidNumber",
			"(uidNumber=99)",
			byNumber,
			args(99, 0),
			[]*object.User{},
		},
		{
			"Should only keep unassigned users for a non-numeric uidNumber",
			"(uidNumber=abc)",
			unassigned,
			args(0),
			[]*object.User{},
		},
		{
			"Should combine uidNumber with a condition pushed down to SQL",
			"(&(uid=alice)(uidNumber=10001))",
			"name=? AND (" + byNumber + ")",
			args("alice", 10001, 0),
			[]*object.User{alice},
		},
		{
			"Should treat synthetic attributes as always present",
			"(uidNumber=*)",
			"1 = 1",
			nil,
			users,
		},
	}

	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			q, err := buildUserFilterFromString(scenery.input)
			if err != nil {
				assert.FailNow(t, "Unable to build condition", err)
			}
			expr, args, err := builder.ToSQL(q.condition)
			if err != nil {
				assert.FailNow(t, "Unable to build sql", err)
			}

			assert.Equal(t, scenery.expectedExpr, expr)
			assert.Equal(t, scenery.expectedArgs, args)
			assert.Equal(t, scenery.expectedUsers, q.apply(users))
		})
	}
}

func TestGetUidNumber(t *testing.T) {
	assert.Equal(t, "10001", getUidNumber(&object.User{Name: "alice", UidNumber: 10001}))
	assert.Equal(t, fmt.Sprintf("%v", hash("bob")), getUidNumber(&object.User{Name: "bob"}))
}

func TestGetGidNumber(t *testing.T) {
	assert.Equal(t, "500", getGidNumber(&object.Group{Name: "devs", GidNumber: 500}))
	assert.Equal(t, fmt.Sprintf("%v", hash("ops")), getGidNumber(&object.Group{Name: "ops"}))
}

func TestLdapFilterOnSyntheticAttributesOutsideAnd(t *testing.T) {
	for _, input := range []string{"(|(uid=alice)(uidNumber=1))", "(!(uidNumber=1))"} {
		t.Run(input, func(t *testing.T) {
			_, err := buildUserFilterFromString(input)
			assert.ErrorContains(t, err, "only supported in an AND filter")
		})
	}
}

func TestMatchGroupFilter(t *testing.T) {
	group := &object.Group{Owner: "org", Name: "devs", GidNumber: 500}
	attributes := getGroupAttributes(group, []string{"alice", "bob"})

	scenarios := []struct {
		input    string
		expected bool
	}{
		{"(objectClass=posixGroup)", true},
		{"(&(objectClass=posixGroup)(gidNumber=" + getGidNumber(group) + "))", true},
		{"(&(objectClass=posixGroup)(gidNumber=0))", false},
		{"(&(objectClass=posixGroup)(cn=devs))", true},
		{"(&(objectClass=posixGroup)(cn=ops))", false},
		{"(&(objectClass=posixGroup)(memberUid=bob))", true},
		{"(&(objectClass=posixGroup)(memberUid=carol))", false},
		{"(cn=de*)", true},
		{"(cn=op*)", false},
		{"(memberUid=*)", true},
		{"(description=*)", false},
		{"(!(cn=ops))", true},
		{"(|(cn=ops)(cn=devs))", true},
	}

	for _, scenery := range scenarios {
		t.Run(scenery.input, func(t *testing.T) {
			filter, err := parseLdapFilter(scenery.input)
			if err != nil {
				assert.FailNow(t, "Unable to parse filter", err)
			}
			assert.Equal(t, scenery.expected, matchGroupFilter(filter, attributes))
		})
	}
}

func TestIsPosixGroupFilter(t *testing.T) {
	scenarios := []struct {
		input    string
		expected bool
	}{
		{"(objectClass=posixGroup)", true},
		{"(&(objectClass=posixGroup)(gidNumber=1))", true},
		{"(&(objectClass=posixAccount)(uidNumber=1))", false},
		{"(cn=devs)", false},
	}

	for _, scenery := range scenarios {
		t.Run(scenery.input, func(t *testing.T) {
			filter, err := parseLdapFilter(scenery.input)
			if err != nil {
				assert.FailNow(t, "Unable to parse filter", err)
			}
			assert.Equal(t, scenery.expected, isPosixGroupFilter(filter))
		})
	}
}

func buildUserFilterFromString(filter string) (*userSearchFilter, error) {
	parsed, err := parseLdapFilter(filter)
	if err != nil {
		return nil, err
	}
	return buildUserFilter(parsed)
}

func parseLdapFilter(filter string) (message.Filter, error) {
	searchRequest, err := buildLdapSearchRequest(filter)
	if err != nil {
		return nil, err
	}
	m, err := message.ReadLDAPMessage(message.NewBytes(0, searchRequest.Bytes()))
	if err != nil {
		return nil, err
	}
	req := m.ProtocolOp().(message.SearchRequest)
	return req.Filter(), nil
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
