package cmd

import (
	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/adm/services"
	"github.com/spf13/cobra"
)

// UserCommands is a factory function that creates and returns a new command with
// create, update and delete subcommands dedicated to user management. It receives a service
// for handling business logic.
func UserCommands(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `Provides an interface for managing users in the system, such as creating, updating, and deleting user accounts.`,
	}

	cmd.AddCommand(userCreate(service))
	cmd.AddCommand(userResetPassword(service))
	cmd.AddCommand(userDelete(service))

	return cmd
}

func userCreate(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "create <username> <password> <email>",
		Args:  cobra.ExactArgs(3),
		Short: "Create a user.",
		Long: `Creates a new user in the system using the provided username, password, and email.
The username must be unique, and the password should meet the system's security requirements.`,
		Example: `cli user create john_doe Secret123!- john.doe@test.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.UserCreate

			if err := bind(args, &input); err != nil {
				return err
			}

			user, err := service.UserCreate(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("User created successfully")
			cmd.Println("Username:", user.Username)
			cmd.Println("Email:", user.Email)

			return nil
		},
	}
}

func userResetPassword(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "password <username> <password>",
		Args:    cobra.ExactArgs(2),
		Short:   "Change user's password",
		Long:    `Updates the password for an existing user identified by the given username.`,
		Example: `cli user password john_doe Secret123!-`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.UserUpdate

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := service.UserUpdate(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("User password changed successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	}
}

func userDelete(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <username>",
		Args:    cobra.ExactArgs(1),
		Short:   "Delete a user",
		Long:    `Deletes a user and all associated data from the system based on the provided username.`,
		Example: `cli user delete john_doe`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.UserDelete

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := service.UserDelete(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("User deleted successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	}
}
