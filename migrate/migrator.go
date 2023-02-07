package migrate

import "github.com/casdoor/casdoor/object"

type Migrator interface {
	IsMigrationNeeded(adapter *object.Adapter) bool
	DoMigration(adapter *object.Adapter)
}

func DoMigration(adapter *object.Adapter) {
	migrators := []Migrator{
		&Migrator_1_101_0_PR_1083{},
		&Migrator_1_229_0_PR_1494{},
		// more migrators add here...
	}
	for _, migrator := range migrators {
		if migrator.IsMigrationNeeded(adapter) {
			migrator.DoMigration(adapter)
		}
	}
}
