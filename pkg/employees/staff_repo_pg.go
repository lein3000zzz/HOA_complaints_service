package employees

type StaffMemberPg StaffMember

func (StaffMemberPg) TableName() string {
	return "staff_members"
}
