package requests

type Request struct {
	ID             string  `gorm:"type:char(40);primaryKey"`
	ResidentID     string  `gorm:"column:id_resident;type:char(40);not null"`
	HouseID        string  `gorm:"column:id_house;type:char(40);not null"`
	RequestType    string  `gorm:"column:type;type:request_type;not null"`
	Complaint      string  `gorm:"type:text;not null"`
	Cost           float64 `gorm:"type:numeric(10,2)"`
	Status         string  `gorm:"type:request_status;default:'создана'"`
	ResponsibleID  string  `gorm:"column:id_responsible;type:char(40)"`
	OrganizationID string  `gorm:"column:id_organization;type:char(40);not null"`
}

type RequestRepo interface {
	CreateRequest(request Request) error
	UpdateRequest(request Request) error
	ChangeRequestStatus(id string, status string) error
	SelectResponsible(request Request)
}

type RequestType string

const (
	TypeApartmentInternal RequestType = "ремон_внутриквартирный"
	TypeHouseCommon       RequestType = "ремонт_общедомового_ищущества"
)

func (s RequestType) IsValid() bool {
	switch s {
	case TypeApartmentInternal, TypeHouseCommon:
		return true
	default:
		return false
	}
}

type RequestStatus string

const (
	StatusCreated     RequestStatus = "создана"
	StatusAssigned    RequestStatus = "назначена_исполнителю"
	StatusCompleted   RequestStatus = "выполнена"
	StatusCancelled   RequestStatus = "отменена"
	StatusSuspended   RequestStatus = "приостановлена"
	StatusTransferred RequestStatus = "передана_организации"
)

func (s RequestStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusAssigned, StatusCompleted, StatusCancelled, StatusSuspended, StatusTransferred:
		return true
	default:
		return false
	}
}
