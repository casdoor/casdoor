package object

import (
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/xorm"
)

type SafeAdapter struct {
	*xormadapter.Adapter
	engine    *xorm.Engine
	tableName string
}

type safeCasbinRule struct {
	Id    int64  `xorm:"pk autoincr"`
	Ptype string `xorm:"varchar(100) index not null default ''"`
	V0    string `xorm:"varchar(100) index not null default ''"`
	V1    string `xorm:"varchar(100) index not null default ''"`
	V2    string `xorm:"varchar(100) index not null default ''"`
	V3    string `xorm:"varchar(100) index not null default ''"`
	V4    string `xorm:"varchar(100) index not null default ''"`
	V5    string `xorm:"varchar(100) index not null default ''"`

	tableName string `xorm:"-"`
}

func (rule *safeCasbinRule) TableName() string {
	if rule.tableName == "" {
		return "casbin_rule"
	}
	return rule.tableName
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

func (a *SafeAdapter) SavePolicy(model model.Model) error {
	lines := make([]*safeCasbinRule, 0, 64)

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			lines = append(lines, a.buildCasbinRule(ptype, rule))
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			lines = append(lines, a.buildCasbinRule(ptype, rule))
		}
	}

	_, err := a.engine.Transaction(func(tx *xorm.Session) (interface{}, error) {
		_, err := tx.Where("ptype IS NOT NULL").Delete(&safeCasbinRule{tableName: a.policyTableName()})
		if err != nil {
			return nil, err
		}

		for _, line := range lines {
			_, err = tx.InsertOne(line)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
	return err
}

func (a *SafeAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := a.buildCasbinRule(ptype, rule)

	session := a.engine.NewSession()
	defer session.Close()

	_, err := session.
		MustCols("ptype", "v0", "v1", "v2", "v3", "v4", "v5").
		Delete(line)

	return err
}

func (a *SafeAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	_, err := a.engine.Transaction(func(tx *xorm.Session) (interface{}, error) {
		for _, rule := range rules {
			line := a.buildCasbinRule(ptype, rule)

			_, err := tx.
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

			_, err := tx.
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
	rules := []*safeCasbinRule{}
	session := a.engine.NewSession()
	defer session.Close()

	err := session.Table(&safeCasbinRule{tableName: a.policyTableName()}).Find(&rules)
	if err != nil {
		return nil, err
	}

	res := make([]*xormadapter.CasbinRule, 0, len(rules))
	for _, rule := range rules {
		res = append(res, rule.toXormCasbinRule())
	}
	return res, nil
}

func (a *SafeAdapter) policyTableName() string {
	if a.tableName == "" {
		return "casbin_rule"
	}
	return a.tableName
}

func (a *SafeAdapter) buildCasbinRule(ptype string, rule []string) *safeCasbinRule {
	line := safeCasbinRule{Ptype: ptype, tableName: a.policyTableName()}

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

func (rule *safeCasbinRule) toXormCasbinRule() *xormadapter.CasbinRule {
	return &xormadapter.CasbinRule{
		Id:    rule.Id,
		Ptype: rule.Ptype,
		V0:    rule.V0,
		V1:    rule.V1,
		V2:    rule.V2,
		V3:    rule.V3,
		V4:    rule.V4,
		V5:    rule.V5,
	}
}
