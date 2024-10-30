package cmd

import (
	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/adm/services"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/spf13/cobra"
)

// DeprecatedCommands writes a list of deprecated commands to a parend cmd.
// These commands will be romved in a future release.
func DeprecatedCommands(cmd *cobra.Command, service services.Services) {
	cmd.AddCommand(&cobra.Command{
		Deprecated: "This command is deprecated and will be removed in a future release. Please use 'adm user create' instead.",
		Use:        "add-user",
		Short:      "Usage: <username> <password> <email>",
		Args:       cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := &inputs.UserCreate{
				Username: args[0],
				Password: args[1],
				Email:    args[2],
			}

			user, err := service.UserCreate(cmd.Context(), input)
			if err != nil {
				return err
			}
			cmd.Println("User added!")
			cmd.Println("name:", user.Name)
			cmd.Println("username:", user.Username)
			cmd.Println("email:", user.Email)

			return nil
		},
	},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release. Please use 'adm user delete' instead.",
			Use:        "del-user",
			Short:      "Usage: <username>",
			Args:       cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				input := &inputs.UserDelete{Username: args[0]}

				if err := service.UserDelete(cmd.Context(), input); err != nil {
					return err
				}

				cmd.Println("User deleted")

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release. Please use 'adm user password' instead.",
			Use:        "reset-user-password",
			Short:      "Usage: <username> <password>",
			Args:       cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				input := &inputs.UserUpdate{Username: args[0], Password: args[1]}

				if err := service.UserUpdate(cmd.Context(), input); err != nil {
					return err
				}

				cmd.Println("Password changed")

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release. Please use 'adm namespace create' instead.",
			Use:        "add-namespace",
			Short:      "Usage: <namespace> <owner>",
			Args:       cobra.RangeArgs(2, 3),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Avoid panic when TenantID isn't provided.
				if len(args) == 2 {
					args = append(args, "")
				}

				input := &inputs.NamespaceCreate{
					Namespace: args[0],
					Owner:     args[1],
					TenantID:  args[2],
				}
				namespace, err := service.NamespaceCreate(cmd.Context(), input)
				if err != nil {
					return err
				}

				cmd.Println("Namespace added:", namespace.Name)
				cmd.Println("Owner:", namespace.Owner)
				cmd.Println("Tenant ID:", namespace.TenantID)

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release. Please use 'cli namespace member add' instead.",
			Use:        "add-user-namespace",
			Short:      "Usage: <username> <namespace> <role>",
			Args:       cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				input := &inputs.MemberAdd{
					Username:  args[0],
					Namespace: args[1],
					Role:      authorizer.RoleFromString(args[2]),
				}
				ns, err := service.NamespaceAddMember(cmd.Context(), input)
				if err != nil {
					return err
				}

				cmd.Println("User:", ns.Owner)
				cmd.Println("added to namespace:", ns.Name)
				cmd.Println("with role:", args[2])

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release. Please use 'cli user create remove' instead.",
			Use:        "del-user-namespace",
			Short:      "Usage <username> <namespace>",
			Args:       cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				input := &inputs.MemberRemove{Username: args[0], Namespace: args[1]}
				ns, err := service.NamespaceRemoveMember(cmd.Context(), input)
				if err != nil {
					return err
				}

				cmd.Println("User:", ns.Owner)
				cmd.Println("removed from namespace:", ns.Name)

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release. Please use 'cli namespace delete' instead.",
			Use:        "del-namespace",
			Short:      "Usage: <namespace>",
			Args:       cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				input := &inputs.NamespaceDelete{Namespace: args[0]}
				if err := service.NamespaceDelete(cmd.Context(), input); err != nil {
					return err
				}

				cmd.Println("Namespace deleted")

				return nil
			},
		})
}
