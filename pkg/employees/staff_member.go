package employees

type StaffMember struct {
	ID       int               `gorm:"type:bigint;primaryKey"`
	FullName string            `gorm:"not null"`
	Phone    string            `gorm:"not null;unique"`
	Status   StaffMemberStatus `gorm:"not null"`

	Specializations []Specialization `gorm:"many2many:employee_specializations;"`
}

type Specialization struct {
	ID    string `gorm:"type:char(40);primaryKey"`
	Title string `gorm:"not null"`

	Employees []StaffMemberSpecialization `gorm:"foreignKey:SpecializationID"`
}

type StaffMemberSpecialization struct {
	EmployeeID       string `gorm:"type:bigint;primaryKey"`
	SpecializationID string `gorm:"type:char(40);primaryKey"`
	IsActive         bool   `gorm:"not null;default:true"`

	Employee       StaffMember    `gorm:"foreignKey:EmployeeID;references:ID"`
	Specialization Specialization `gorm:"foreignKey:SpecializationID;references:ID"`
}

type StaffMemberRepo interface {
	ChangeStaffMemberStatus(id string, status string) error
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
