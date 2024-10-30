package inputs

// UserCreate defines the structure for inputs when creating a user.
type UserCreate struct {
	Username string `validate:"required,username"`
	Password string `validate:"required,password"`
	Email    string `validate:"required,email"`
}

// UserUpdate defines the structure for inputs when updating a user.
type UserUpdate struct {
	Username string `validate:"required,username"`
	Password string
}

// UserDelete defines the structure for inputs when deleting a user.
type UserDelete struct {
	Username string `validate:"required,username"`
}

// UserPassword the structure for validate passowrd.
type UserPassword struct {
	Password string `validate:"required,password"`
}
