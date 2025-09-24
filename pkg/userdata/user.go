package userdata

type User struct {
	Phone        string `gorm:"type:varchar(40);column:phone_number;primaryKey"`
	PasswordHash string `gorm:"type:varchar;column:password_hash;type:varchar;not null"`
}

type UserRepo interface {
	Authorize(phone, password string) (*User, error)
	Register(phone, password string) (*User, error)
	DeleteByPhone(phone string) error
	GetAll(limit, offset int) ([]*User, int, error)
}
