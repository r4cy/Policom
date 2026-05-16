package core

type UpdateStatus string

const (
	StatusUpdateUnknown UpdateStatus = "unknown"
	StatusUpdateIdle    UpdateStatus = "idle"
	StatusUpdateRunning UpdateStatus = "running"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type contextUser string

const UserKey contextUser = "auth_user"

type User struct {
	ID       string
	Username string
	Email    string
	Role     UserRole
}

type UserLog struct {
	ID            string
	Password_hash string
	Email         string
	Role          UserRole
}

type UpdateStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
}

type Comics struct {
	ID             int
	Title          string
	URL            string
	Description    string
	ImgDescription string
}
