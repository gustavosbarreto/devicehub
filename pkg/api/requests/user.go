// Package requests defines structures to represent requests' bodies from API.
package requests

import "github.com/shellhub-io/shellhub/pkg/models"

type UserParam struct {
	ID string `param:"id" validate:"required"`
}

// UserDataUpdate is the structure to represent the request body of the update user data endpoint.
type UserDataUpdate struct {
	UserParam
	Name     string `json:"name" validate:"required,name"`
	Username string `json:"username" validate:"required,username"`
	Email    string `json:"email" validate:"required,email"`
}

// UserPasswordUpdate is the structure to represent the request body for the update user password endpoint.
type UserPasswordUpdate struct {
	UserParam
	CurrentPassword string `json:"current_password" validate:"required,min=5,max=32,nefield=NewPassword"`
	NewPassword     string `json:"new_password" validate:"required,password,nefield=CurrentPassword"`
}

// UserAuth is the structure to represent the request body for the user auth endpoint.
type UserAuth struct {
	// Identifier represents an username or email.
	//
	// TODO: change json tag from username to identifier and update the OpenAPI.
	Identifier models.UserAuthIdentifier `json:"username" validate:"required"`
	Password   string                    `json:"password" validate:"required"`
}
