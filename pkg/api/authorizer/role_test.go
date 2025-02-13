package authorizer_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/stretchr/testify/require"
)

func TestRoleFromString(t *testing.T) {
	cases := []struct {
		description string
		role        string
		expected    authorizer.Role
	}{
		{
			description: "fails with invalid roles",
			role:        "N/A",
			expected:    authorizer.RoleInvalid,
		},
		{
			description: "succeeds with 'owner'",
			role:        "owner",
			expected:    authorizer.RoleOwner,
		},
		{
			description: "succeeds with 'administrator'",
			role:        "administrator",
			expected:    authorizer.RoleAdministrator,
		},
		{
			description: "succeeds with 'operator'",
			role:        "operator",
			expected:    authorizer.RoleOperator,
		},
		{
			description: "succeeds with 'observer'",
			role:        "observer",
			expected:    authorizer.RoleObserver,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			require.Equal(tt, tc.expected, authorizer.RoleFromString(tc.role))
		})
	}
}

func TestRolePermissions(t *testing.T) {
	cases := []struct {
		description string
		role        authorizer.Role
		expected    []authorizer.Permission
	}{
		{
			description: "fails with invalid roles",
			role:        authorizer.RoleInvalid,
			expected:    []authorizer.Permission{},
		},
		{
			description: "succeeds with 'owner'",
			role:        authorizer.RoleOwner,
			expected: []authorizer.Permission{
				authorizer.DeviceAccept,
				authorizer.DeviceReject,
				authorizer.DeviceRemove,
				authorizer.DeviceConnect,
				authorizer.DeviceRename,
				authorizer.DeviceDetails,
				authorizer.DeviceUpdate,
				authorizer.TagCreate,
				authorizer.TagUpdate,
				authorizer.TagDelete,
				authorizer.SessionPlay,
				authorizer.SessionClose,
				authorizer.SessionRemove,
				authorizer.SessionDetails,
				authorizer.FirewallCreate,
				authorizer.FirewallEdit,
				authorizer.FirewallRemove,
				authorizer.FirewallAddTag,
				authorizer.FirewallRemoveTag,
				authorizer.FirewallUpdateTag,
				authorizer.PublicKeyCreate,
				authorizer.PublicKeyEdit,
				authorizer.PublicKeyRemove,
				authorizer.PublicKeyAddTag,
				authorizer.PublicKeyRemoveTag,
				authorizer.PublicKeyUpdateTag,
				authorizer.NamespaceUpdate,
				authorizer.NamespaceAddMember,
				authorizer.NamespaceRemoveMember,
				authorizer.NamespaceEditMember,
				authorizer.NamespaceEnableSessionRecord,
				authorizer.NamespaceDelete,
				authorizer.BillingCreateCustomer,
				authorizer.BillingChooseDevices,
				authorizer.BillingAddPaymentMethod,
				authorizer.BillingUpdatePaymentMethod,
				authorizer.BillingRemovePaymentMethod,
				authorizer.BillingCancelSubscription,
				authorizer.BillingCreateSubscription,
				authorizer.BillingGetSubscription,
				authorizer.APIKeyCreate,
				authorizer.APIKeyUpdate,
				authorizer.APIKeyDelete,
				authorizer.ConnectorDelete,
				authorizer.ConnectorUpdate,
				authorizer.ConnectorSet,
				authorizer.TunnelsCreate,
				authorizer.TunnelsDelete,
			},
		},
		{
			description: "succeeds with 'administrator'",
			role:        authorizer.RoleAdministrator,
			expected: []authorizer.Permission{
				authorizer.DeviceAccept,
				authorizer.DeviceReject,
				authorizer.DeviceRemove,
				authorizer.DeviceConnect,
				authorizer.DeviceRename,
				authorizer.DeviceDetails,
				authorizer.DeviceUpdate,
				authorizer.TagCreate,
				authorizer.TagUpdate,
				authorizer.TagDelete,
				authorizer.SessionPlay,
				authorizer.SessionClose,
				authorizer.SessionRemove,
				authorizer.SessionDetails,
				authorizer.FirewallCreate,
				authorizer.FirewallEdit,
				authorizer.FirewallRemove,
				authorizer.FirewallAddTag,
				authorizer.FirewallRemoveTag,
				authorizer.FirewallUpdateTag,
				authorizer.PublicKeyCreate,
				authorizer.PublicKeyEdit,
				authorizer.PublicKeyRemove,
				authorizer.PublicKeyAddTag,
				authorizer.PublicKeyRemoveTag,
				authorizer.PublicKeyUpdateTag,
				authorizer.NamespaceUpdate,
				authorizer.NamespaceAddMember,
				authorizer.NamespaceRemoveMember,
				authorizer.NamespaceEditMember,
				authorizer.NamespaceEnableSessionRecord,
				authorizer.APIKeyCreate,
				authorizer.APIKeyUpdate,
				authorizer.APIKeyDelete,
				authorizer.ConnectorDelete,
				authorizer.ConnectorUpdate,
				authorizer.ConnectorSet,
				authorizer.TunnelsCreate,
				authorizer.TunnelsDelete,
			},
		},
		{
			description: "succeeds with 'operator'",
			role:        authorizer.RoleOperator,
			expected: []authorizer.Permission{
				authorizer.DeviceAccept,
				authorizer.DeviceReject,
				authorizer.DeviceConnect,
				authorizer.DeviceRename,
				authorizer.DeviceDetails,
				authorizer.DeviceUpdate,
				authorizer.TagCreate,
				authorizer.TagUpdate,
				authorizer.TagDelete,
				authorizer.SessionDetails,
			},
		},
		{
			description: "succeeds with 'observer'",
			role:        authorizer.RoleObserver,
			expected: []authorizer.Permission{
				authorizer.DeviceConnect,
				authorizer.DeviceDetails,
				authorizer.SessionDetails,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			require.Equal(tt, tc.expected, tc.role.Permissions())
		})
	}
}

func TestRolePreferences(t *testing.T) {
	cases := []struct {
		description string
		role        authorizer.Role
		greater     []authorizer.Role
		less        []authorizer.Role
	}{
		{
			description: authorizer.RoleInvalid.String(),
			role:        authorizer.RoleInvalid,
			greater:     []authorizer.Role{authorizer.RoleOwner, authorizer.RoleAdministrator, authorizer.RoleOperator, authorizer.RoleObserver},
			less:        []authorizer.Role{},
		},
		{
			description: authorizer.RoleOwner.String(),
			role:        authorizer.RoleOwner,
			greater:     []authorizer.Role{},
			less:        []authorizer.Role{authorizer.RoleAdministrator, authorizer.RoleOperator, authorizer.RoleObserver},
		},
		{
			description: authorizer.RoleAdministrator.String(),
			role:        authorizer.RoleAdministrator,
			greater:     []authorizer.Role{authorizer.RoleOwner},
			less:        []authorizer.Role{authorizer.RoleOperator, authorizer.RoleObserver},
		},
		{
			description: authorizer.RoleOperator.String(),
			role:        authorizer.RoleOperator,
			greater:     []authorizer.Role{authorizer.RoleOwner, authorizer.RoleAdministrator},
			less:        []authorizer.Role{authorizer.RoleObserver},
		},
		{
			description: authorizer.RoleObserver.String(),
			role:        authorizer.RoleObserver,
			greater:     []authorizer.Role{authorizer.RoleOwner, authorizer.RoleAdministrator, authorizer.RoleOperator},
			less:        []authorizer.Role{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			for _, r := range tc.greater {
				require.Equal(tt, false, tc.role.HasAuthority(r))
			}

			for _, r := range tc.less {
				require.Equal(tt, true, tc.role.HasAuthority(r))
			}
		})
	}
}
