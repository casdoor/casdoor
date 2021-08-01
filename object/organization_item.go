package object

type AccountItem struct {
	Name     string `json:"name"`
	Visible  bool   `json:"visible"`
	Required bool   `json:"required"`
	Editable bool   `json:"editable"`
	Public   bool `json:"public"`
}

func (org *Organization) getAccountItem(itemName string) *AccountItem {
	for _, accountItem := range org.AccountItems {
		if accountItem.Name == itemName {
			return accountItem
		}
	}
	return nil
}

func (org *Organization) IsAccountItemEnabled(itemName string) bool {
	return org.getAccountItem(itemName) != nil
}

func (org *Organization) IsAccountItemVisible(itemName string) bool {
	accountItem := org.getAccountItem(itemName)
	if accountItem == nil {
		return false
	}

	return accountItem.Visible
}

func (org *Organization) IsAccountItemRequired(itemName string) bool {
	accountItem := org.getAccountItem(itemName)
	if accountItem == nil {
		return false
	}

	return accountItem.Required
}

func (org *Organization) IsAccountItemEditable(itemName string) bool {
	accountItem := org.getAccountItem(itemName)
	if accountItem == nil {
		return false
	}

	return accountItem.Editable
}

func (org *Organization) GetAccountItemPublic(itemName string) bool {
	accountItem := org.getAccountItem(itemName)
	if accountItem == nil {
		return false
	}

	return accountItem.Public
}
