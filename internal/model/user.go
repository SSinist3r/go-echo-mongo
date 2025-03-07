package model

// Role constants for easier access and consistency
const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleEditor    = "editor"
	RoleViewer    = "viewer"
	RoleManager   = "manager"
	RoleModerator = "moderator"
)

// User represents the user model in the system
type User struct {
	BaseModel `bson:",inline"`
	Name      string   `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Email     string   `json:"email" bson:"email" validate:"required,email"`
	Password  string   `json:"password,omitempty" bson:"password" validate:"required,min=6"`
	ApiKey    string   `json:"api_key,omitempty" bson:"api_key"`
	Roles     []string `json:"roles" bson:"roles"`
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *User) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles
func (u *User) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !u.HasRole(role) {
			return false
		}
	}
	return true
}

// IsAdmin is a convenience method to check if user has admin role
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// Ensure User implements BaseModel interface
var _ Model = (*User)(nil)
