package object

import (
	"fmt"
	"log"
	"testing"

	"github.com/xorm-io/core"
)

func Test_update_complex_option_field(t *testing.T) {
	// Set up database connection information
	InitConfig()
	InitAdapter()
	engine := adapter.Engine

	defer engine.Close()

	err := engine.Sync2(new(Organization))
	if err != nil {
		log.Fatal(err)
	}

	pageSize := 100
	page := 1
	const defaultOption = "AtLeast6"
	const batchSize = 100

	for {
		// Paginate the query
		var organizations []*Organization
		err = engine.Limit(pageSize, (page-1)*pageSize).Find(&organizations)
		if err != nil {
			log.Fatal(err)
		}

		if len(organizations) == 0 {
			// All data has been processed
			break
		}

		// Update the password_complex_options field for the current page
		session := engine.NewSession()
		defer session.Close()
		err = session.Begin()
		if err != nil {
			log.Fatal(err)
		}

		for _, org := range organizations {
			if org.PasswordComplexOptions == nil {
				org.PasswordComplexOptions = []string{defaultOption}
				// Accumulate the changes in the session
				_, err = session.ID(core.PK{org.Owner, org.Name}).Cols("password_complex_options").Update(org)
				if err != nil {
					session.Rollback()
					log.Fatal(err)
				}
			}

			// Commit the changes in batches
			if len(organizations) >= batchSize {
				err = session.Commit()
				if err != nil {
					log.Fatal(err)
				}
				// Start a new batch
				session = engine.NewSession()
				defer session.Close()
				err = session.Begin()
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		// Commit any remaining changes

		err = session.Commit()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Processed page %d\n", page)

		page++
	}

	fmt.Println("Password complex options updated successfully!")
}
