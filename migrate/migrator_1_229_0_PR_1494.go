package migrate

import (
	"errors"

	"github.com/casdoor/casdoor/object"
)

type Migrator_1_229_0_PR_1494 struct{}

type newSession struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	Application string `xorm:"varchar(100) notnull pk" json:"application"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	SessionId []string `json:"sessionId"`
}

type oldSession struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	SessionId []string `json:"sessionId"`
}

func (*Migrator_1_229_0_PR_1494) IsMigrationNeeded(adapter *object.Adapter) bool {
	if exist, _ := adapter.Engine.IsTableExist("session"); exist {
		if colErr := adapter.Engine.Table("session").Find(&[]*newSession{}); colErr != nil {
			return true
		}
	}
	return false
}

func (*Migrator_1_229_0_PR_1494) DoMigration(adapter *object.Adapter) {
	// Create a new field called 'application' and add it to the primary key for table `session`
	var err error
	tx := adapter.Engine.NewSession()

	if alreadyCreated, _ := adapter.Engine.IsTableExist("session_tmp"); alreadyCreated {
		panic(errors.New("there is already a table called 'session_tmp', please rename or delete it for casdoor version migration and restart"))
	}

	tx.Table("session_tmp").CreateTable(&newSession{})

	oldSessions := []*oldSession{}
	newSessions := []*newSession{}

	tx.Table("session").Find(&oldSessions)

	for _, oldSession := range oldSessions {
		newApplication := "null"
		if oldSession.Owner == "built-in" {
			newApplication = "app-built-in"
		}
		newSessions = append(newSessions, &newSession{
			Owner:       oldSession.Owner,
			Name:        oldSession.Name,
			Application: newApplication,
			CreatedTime: oldSession.CreatedTime,
			SessionId:   oldSession.SessionId,
		})
	}

	rollbackFlag := false
	_, err = tx.Table("session_tmp").Insert(newSessions)
	count1, _ := tx.Table("session_tmp").Count()
	count2, _ := tx.Table("session").Count()

	if err != nil || count1 != count2 {
		rollbackFlag = true
	}

	delete := &newSession{
		Application: "null",
	}
	_, err = tx.Table("session_tmp").Delete(*delete)
	if err != nil {
		rollbackFlag = true
	}

	if rollbackFlag {
		tx.DropTable("session_tmp")
		panic(errors.New("there is something wrong with data migration for table `session`, if there is a table called `session_tmp` not created by you in casdoor, please drop it, then restart anyhow"))
	}

	err = tx.DropTable("session")
	if err != nil {
		panic(errors.New("fail to drop table `session` for casdoor, please drop it and rename the table `session_tmp` to `session` manually and restart"))
	}

	// Already drop table `session`
	// Can't find an api from xorm for altering table name
	err = tx.Table("session").CreateTable(&newSession{})
	if err != nil {
		panic(errors.New("there is something wrong with data migration for table `session`, please restart"))
	}

	sessions := []*newSession{}
	tx.Table("session_tmp").Find(&sessions)
	_, err = tx.Table("session").Insert(sessions)
	if err != nil {
		panic(errors.New("there is something wrong with data migration for table `session`, please drop table `session` and rename table `session_tmp` to `session` and restart"))
	}

	err = tx.DropTable("session_tmp")
	if err != nil {
		panic(errors.New("fail to drop table `session_tmp` for casdoor, please drop it manually and restart"))
	}

	tx.Close()
}
