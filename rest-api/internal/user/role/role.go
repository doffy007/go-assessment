package role

const (
	Writter    = "writter"
	SuperAdmin = "super_admin"
)

type AssignRoleRequest struct {
	Role string `json:"role"`
}
