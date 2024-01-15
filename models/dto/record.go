package dto

type Record struct {
	ID         int64  `json:"-" sql.field:"id"`
	Name       string `json:"name,omitempty" sql.field:"name"`
	LastName   string `json:"last_name,omitempty" sql.field:"last_name"`
	MiddleName string `json:"middle_name,omitempty" sql.field:"middle_name"`
	Address    string `json:"address,omitempty" sql.field:"address"`
	Phone      string `json:"phone,omitempty" sql.field:"phone"`
}
