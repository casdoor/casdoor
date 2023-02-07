package migrate

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"

	"xorm.io/core"
)

type Migrator_1_101_0_PR_1083 struct{}

type oldModel struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	ModelText string `xorm:"mediumtext" json:"modelText"`
	IsEnabled bool   `json:"isEnabled"`
}

type oldPermission struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Users   []string `xorm:"mediumtext" json:"users"`
	Roles   []string `xorm:"mediumtext" json:"roles"`
	Domains []string `xorm:"mediumtext" json:"domains"`

	Model        string   `xorm:"varchar(100)" json:"model"`
	Adapter      string   `xorm:"varchar(100)" json:"adapter"`
	ResourceType string   `xorm:"varchar(100)" json:"resourceType"`
	Resources    []string `xorm:"mediumtext" json:"resources"`
	Actions      []string `xorm:"mediumtext" json:"actions"`
	Effect       string   `xorm:"varchar(100)" json:"effect"`
	IsEnabled    bool     `json:"isEnabled"`

	Submitter   string `xorm:"varchar(100)" json:"submitter"`
	Approver    string `xorm:"varchar(100)" json:"approver"`
	ApproveTime string `xorm:"varchar(100)" json:"approveTime"`
	State       string `xorm:"varchar(100)" json:"state"`
}

func (*Migrator_1_101_0_PR_1083) IsMigrationNeeded(adapter *object.Adapter) bool {
	if exist1, _ := adapter.Engine.IsTableExist("model"); exist1 {
		if exist2, _ := adapter.Engine.IsTableExist("permission"); exist2 {
			if exist3, _ := adapter.Engine.IsTableExist("permission_rule"); exist3 {
				return true
			}
		}
	}
	return false
}

func (*Migrator_1_101_0_PR_1083) DoMigration(adapter *object.Adapter) {
	// MigratePermissionRule
	models := []*oldModel{}
	err := adapter.Engine.Table("model").Find(&models, &oldModel{})
	if err != nil {
		panic(err)
	}

	isHit := false
	for _, model := range models {
		if strings.Contains(model.ModelText, "permission") {
			// update model table
			model.ModelText = strings.Replace(model.ModelText, "permission,", "", -1)
			updateModel(adapter, model.getId(), model)
			isHit = true
		}
	}

	if isHit {
		// update permission_rule table
		sql := "UPDATE `permission_rule`SET V0 = V1, V1 = V2, V2 = V3, V3 = V4, V4 = V5 WHERE V0 IN (SELECT CONCAT(owner, '/', name) AS permission_id FROM `permission`)"
		_, err = adapter.Engine.Exec(sql)
		if err != nil {
			return
		}
	}
}

func (oldModel *oldModel) getId() string {
	return fmt.Sprintf("%s/%s", oldModel.Owner, oldModel.Name)
}

func updateModel(adapter *object.Adapter, id string, modelObj *oldModel) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getModel(adapter, owner, name) == nil {
		return false
	}

	if name != modelObj.Name {
		err := modelChangeTrigger(adapter, name, modelObj.Name)
		if err != nil {
			return false
		}
	}
	// check model grammar
	_, err := model.NewModelFromString(modelObj.ModelText)
	if err != nil {
		panic(err)
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(modelObj)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func getModel(adapter *object.Adapter, owner string, name string) *oldModel {
	if owner == "" || name == "" {
		return nil
	}

	m := oldModel{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&m)
	if err != nil {
		panic(err)
	}

	if existed {
		return &m
	} else {
		return nil
	}
}

func modelChangeTrigger(adapter *object.Adapter, oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	permission := new(oldPermission)
	permission.Model = newName
	_, err = session.Where("model=?", oldName).Update(permission)
	if err != nil {
		return err
	}

	return session.Commit()
}
