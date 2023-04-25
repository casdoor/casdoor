package forms

type EmailForm struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Sender    string   `json:"sender"`
	Receivers []string `json:"receivers"`
	Provider  string   `json:"provider"`
}

type SmsForm struct {
	Content   string   `json:"content"`
	Receivers []string `json:"receivers"`
	OrgId     string   `json:"organizationId"` // e.g. "admin/built-in"
}
