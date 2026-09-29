package permissions

type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
	ActionAdmin  Action = "admin"
)

type Permission struct {
	Action        Action
	InstitutionID *int64
}

func (p Permission) Allows(req Permission) bool {
	if p.Action != req.Action && p.Action != ActionAdmin {
		return false
	}

	if p.InstitutionID == nil {
		return true
	}

	if req.InstitutionID == nil {
		return false
	}

	return *p.InstitutionID == *req.InstitutionID
}
