package object

import (
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/xorm"
)

type SafeAdapter struct {
	*xormadapter.Adapter
	engine    *xorm.Engine
	tableName string
}

func NewSafeAdapter(a *Adapter) *SafeAdapter {
	if a == nil || a.Adapter == nil || a.engine == nil {
		return nil
	}

	return &SafeAdapter{
		Adapter:   a.Adapter,
		engine:    a.engine,
		tableName: a.Table,
	}
}

func (a *SafeAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := a.buildCasbinRule(ptype, rule)

	session := a.engine.NewSession()
	defer session.Close()

	if a.tableName != "" {
		session = session.Table(a.tableName)
	}

	_, err := session.
		MustCols("ptype", "v0", "v1", "v2", "v3", "v4", "v5").
		Delete(line)

	return err
}

func (a *SafeAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	_, err := a.engine.Transaction(func(tx *xorm.Session) (interface{}, error) {
		for _, rule := range rules {
			line := a.buildCasbinRule(ptype, rule)

			var session *xorm.Session
			if a.tableName != "" {
				session = tx.Table(a.tableName)
			} else {
				session = tx
			}

			_, err := session.
				MustCols("ptype", "v0", "v1", "v2", "v3", "v4", "v5").
				Delete(line)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (a *SafeAdapter) buildCasbinRule(ptype string, rule []string) *xormadapter.CasbinRule {
	line := xormadapter.CasbinRule{Ptype: ptype}

	l := len(rule)
	if l > 0 {
		line.V0 = rule[0]
	}
	if l > 1 {
		line.V1 = rule[1]
	}
	if l > 2 {
		line.V2 = rule[2]
	}
	if l > 3 {
		line.V3 = rule[3]
	}
	if l > 4 {
		line.V4 = rule[4]
	}
	if l > 5 {
		line.V5 = rule[5]
	}

	return &line
}
