package userdata

type User struct {
	Phone        string `gorm:"column:phone_number;primaryKey"`
	PasswordHash string `gorm:"column:password_hash;type:varchar;not null"`
}

type UserRepo interface {
	Authorize(phone, password string) (*User, error)
	Register(phone, password string) (*User, error)
}
