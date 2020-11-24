package object

type Application struct {
	Id   string `xorm:"notnull pk" json:"id"`
	Name string `xorm:"notnull pk" json:"name"`

	CreatedTime string `json:"createdTime"`
	CreatedBy   string `xorm:"notnull" json:"createdBy"`
}
