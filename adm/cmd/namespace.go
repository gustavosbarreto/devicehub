package cmd

import (
	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/adm/services"
	"github.com/spf13/cobra"
)

// NamespaceCommands a factory function that creates and returns a new command with
// create and delete subcommands dedicated to namespaces management. It receives a service
// for handling business logic.
func NamespaceCommands(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "Manage namespaces",
		Long:  `Provides an interface for managing namespaces within the system, such as creating new namespaces or deleting existing ones.`,
	}

	cmd.AddCommand(namespaceCreate(service))
	cmd.AddCommand(namespaceDelete(service))
	cmd.AddCommand(memberCommands(service))

	return cmd
}

func namespaceCreate(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "create <namespace> <owner> [tenant]",
		Short: "Create a namespace",
		Long: `Creates a new namespace in the system using the provided namespace name, associated owner's username, and an optional tenant ID.
The owner must be a valid username within the system. If a tenant ID is provided, it should be in UUID format.`,
		Example: `adm namespace create dev john_doe`,
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Avoid panic when TenantID isn't provided.
			if len(args) == 2 {
				args = append(args, "")
			}

			var input inputs.NamespaceCreate

			if err := bind(args, &input); err != nil {
				return err
			}

			namespace, err := service.NamespaceCreate(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("Namespace created successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Owner:", namespace.Owner)

			return nil
		},
	}
}

func namespaceDelete(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <namespace>",
		Short:   "Delete a namespace",
		Long:    `Deletes a namespace and all associated data from the system based on the provided name.`,
		Example: `adm namespace delete dev`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.NamespaceDelete

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := service.NamespaceDelete(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("Namespace deleted successfully")
			cmd.Println("Namespace:", input.Namespace)

			return nil
		},
	}
}

// memberCommands factory function that creates and returns a new command with
// add and remove subcommands dedicated to members management. It receives a service
// for handling business logic.
func memberCommands(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage members",
		Long:  `Manage members`,
	}

	cmd.AddCommand(memberAdd(service))
	cmd.AddCommand(memberRemove(service))

	return cmd
}

func memberAdd(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "add <username> <namespace> <role>",
		Short: "Add a member",
		Long: `Adds a new member to the specified namespace with the given role. 
The username identifies the member to be added, the namespace specifies where the member should be added, 
and the role indicates the permissions that the member will have within that namespace.`,
		Example: `adm member add myuser mynamespace observer`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.MemberAdd

			if err := bind(args, &input); err != nil {
				return err
			}

			namespace, err := service.NamespaceAddMember(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("Member added successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", input.Username)
			cmd.Println("Role:", input.Role)

			return nil
		},
	}
}

func memberRemove(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <username> <namespace>",
		Short: "Remove a member",
		Long: `Removes an existing member from the specified namespace. 
The username identifies the member to be removed, and the namespace specifies where the member is currently located.`,
		Example: `adm member remove john_doe dev`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.MemberRemove

			if err := bind(args, &input); err != nil {
				return err
			}

			namespace, err := service.NamespaceRemoveMember(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("Member removed successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", input.Username)

			return nil
		},
	}
}
