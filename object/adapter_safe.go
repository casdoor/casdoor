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

func (a *SafeAdapter) UpdatePolicy(sec string, ptype string, oldRule []string, newRule []string) error {
	oldLine := a.buildCasbinRule(ptype, oldRule)
	newLine := a.buildCasbinRule(ptype, newRule)

	session := a.engine.NewSession()
	defer session.Close()

	if a.tableName != "" {
		session = session.Table(a.tableName)
	}

	_, err := session.
		Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
			oldLine.Ptype, oldLine.V0, oldLine.V1, oldLine.V2, oldLine.V3, oldLine.V4, oldLine.V5).
		MustCols("ptype", "v0", "v1", "v2", "v3", "v4", "v5").
		Update(newLine)

	return err
}

func (a *SafeAdapter) UpdatePolicies(sec string, ptype string, oldRules [][]string, newRules [][]string) error {
	_, err := a.engine.Transaction(func(tx *xorm.Session) (interface{}, error) {
		for i, oldRule := range oldRules {
			oldLine := a.buildCasbinRule(ptype, oldRule)
			newLine := a.buildCasbinRule(ptype, newRules[i])

			var session *xorm.Session
			if a.tableName != "" {
				session = tx.Table(a.tableName)
			} else {
				session = tx
			}

			_, err := session.
				Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
					oldLine.Ptype, oldLine.V0, oldLine.V1, oldLine.V2, oldLine.V3, oldLine.V4, oldLine.V5).
				MustCols("ptype", "v0", "v1", "v2", "v3", "v4", "v5").
				Update(newLine)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (a *SafeAdapter) GetRules() ([]*xormadapter.CasbinRule, error) {
	rules := []*xormadapter.CasbinRule{}
	session := a.engine.NewSession()
	defer session.Close()

	if a.tableName != "" {
		session = session.Table(a.tableName)
	}

	err := session.Find(&rules)
	return rules, err
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
