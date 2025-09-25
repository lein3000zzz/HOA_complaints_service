package staffdata

type StaffMember struct {
	ID              int                         `gorm:"type:bigint;primaryKey"`
	FullName        string                      `gorm:"type:varchar(40);not null"`
	Phone           string                      `gorm:"type:varchar(40);column:phone_number;not null;unique"`
	Status          StaffMemberStatus           `gorm:"not null"`
	Specializations []StaffMemberSpecialization `gorm:"foreignKey:MemberID;references:ID"`
}

type Specialization struct {
	ID    string `gorm:"type:char(40);primaryKey"`
	Title string `gorm:"type:varchar(40);column:name;not null"`

	Employees []StaffMemberSpecialization `gorm:"foreignKey:SpecializationID"`
}

type StaffMemberSpecialization struct {
	MemberID         int    `gorm:"column:id_member;type:bigint;primaryKey"`
	SpecializationID string `gorm:"column:id_specialization;type:char(40);primaryKey"`
	IsActive         bool   `gorm:"column:is_active;not null;default:true"`

	Employee       StaffMember    `gorm:"foreignKey:MemberID;references:ID"`
	Specialization Specialization `gorm:"foreignKey:SpecializationID;references:ID"`
}

type StaffRepo interface {
	RegisterNewMember(phone, fullName string) (*StaffMember, error)
	AddStaffMemberSpecializationAssoc(staffMemberID int, specializationID string) error
	RegisterNewSpecialization(jobTitle string) (*Specialization, error)
	GetStaffMemberByPhoneNumber(phoneNumber string) (*StaffMember, error)
	DeleteByPhone(phoneNumber string) error
	FindLeastBusyByJobID(jobID string) (*StaffMember, error)
}

type StaffMemberStatus string

const (
	StatusActive    StaffMemberStatus = "работает"
	StatusInactive  StaffMemberStatus = "уволился"
	StatusSuspended StaffMemberStatus = "недоступен"
)

func (s StaffMemberStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusSuspended:
		return true
	default:
		return false
	}
}
