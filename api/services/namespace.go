package services

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type NamespaceService interface {
	ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *requests.NamespaceCreate) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, tenantID string) error

	// EditNamespace updates a namespace for the specified requests.NamespaceEdit#Tenant.
	// It returns the namespace with the updated fields and an error, if any.
	EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error)
	// AddNamespaceMember adds a member to a namespace. The member's role cannot have more authority than the user who is
	// adding the member; owners cannot be created. It returns the namespace and an error, if any.
	AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error)
	// UpdateNamespaceMember updates a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is updating the member; owners cannot be created. It returns an error, if any.
	UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error
	// RemoveNamespaceMember removes a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is removing the member; owners cannot be removed. It returns the namespace
	// and an error, if any.
	RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error)

	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
	GetSessionRecord(ctx context.Context, tenantID string) (bool, error)
}

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error) {
	namespaces, count, err := s.store.NamespaceList(ctx, req.Paginator, req.Filters, false)
	if err != nil {
		return nil, 0, NewErrNamespaceList(err)
	}

	for index, namespace := range namespaces {
		members, err := s.fillMembersData(ctx, namespace.Members)
		if err != nil {
			return nil, 0, NewErrNamespaceMemberFillData(err)
		}

		namespaces[index].Members = members
	}

	return namespaces, count, nil
}

// CreateNamespace creates a new namespace.
func (s *service) CreateNamespace(ctx context.Context, req *requests.NamespaceCreate) (*models.Namespace, error) {
	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	// When MaxNamespaces is less than or equal to zero, it means that the user has no limit
	// of namespaces.
	if user.MaxNamespaces > 0 {
		info, err := s.store.UserGetInfo(ctx, req.UserID)
		switch {
		case err != nil:
			return nil, err
		case len(info.OwnedNamespaces) >= user.MaxNamespaces:
			return nil, NewErrNamespaceLimitReached(user.MaxNamespaces, nil)
		}
	}

	if dup, err := s.store.NamespaceGetByName(ctx, strings.ToLower(req.Name)); dup != nil || (err != nil && err != store.ErrNoDocuments) {
		return nil, NewErrNamespaceDuplicated(err)
	}

	ns := &models.Namespace{
		Name:  strings.ToLower(req.Name),
		Owner: user.ID,
		Members: []models.Member{
			{
				ID:      user.ID,
				Role:    authorizer.RoleOwner,
				Status:  models.MemberStatusAccepted,
				AddedAt: clock.Now(),
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: "",
		},
		TenantID: req.TenantID,
	}

	if envs.IsCommunity() {
		ns.Settings.ConnectionAnnouncement = models.DefaultAnnouncementMessage
	}

	if req.TenantID == "" {
		ns.TenantID = uuid.Generate()
	}

	// Set limits according to ShellHub instance type
	if envs.IsCloud() {
		// cloud free plan is limited only by the max of devices
		ns.MaxDevices = 3
	} else {
		// we don't set limits on enterprise and community instances
		ns.MaxDevices = -1
	}

	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, NewErrNamespaceCreateStore(err)
	}

	return ns, nil
}

// GetNamespace gets a namespace.
//
// It receives a context, used to "control" the request flow and the tenant ID from models.Namespace.
//
// GetNamespace returns a models.Namespace and an error. When error is not nil, the models.Namespace is nil.
func (s *service) GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenantID, true)
	if err != nil || namespace == nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	members, err := s.fillMembersData(ctx, namespace.Members)
	if err != nil {
		return nil, NewErrNamespaceMemberFillData(err)
	}

	namespace.Members = members

	return namespace, nil
}

// DeleteNamespace deletes a namespace.
//
// It receives a context, used to "control" the request flow and the tenant ID from models.Namespace.
//
// When cloud and billing is enabled, it will try to delete the namespace's billing information from the billing
// service if it exists.
func (s *service) DeleteNamespace(ctx context.Context, tenantID string) error {
	ns, err := s.store.NamespaceGet(ctx, tenantID, true)
	if err != nil {
		return NewErrNamespaceNotFound(tenantID, err)
	}

	ableToReportDeleteNamespace := func(ns *models.Namespace) bool {
		return !ns.Billing.IsNil() && ns.Billing.HasCutomer() && ns.Billing.HasSubscription()
	}

	if envs.IsCloud() && envs.HasBilling() && ableToReportDeleteNamespace(ns) {
		if err := s.BillingReport(s.client, tenantID, ReportNamespaceDelete); err != nil {
			return NewErrBillingReportNamespaceDelete(err)
		}
	}

	return s.store.NamespaceDelete(ctx, tenantID)
}

// fillMembersData fill the member data with the user data.
//
// This method exist because the namespace stores only the user ID and the role from its member as a list of models.Member.
// To avoid unnecessary calls to store for member information, member username, this "conversion" is ony made when
// required by the service.
//
// It receives a context, used to "control" the request flow and a slice of models.Member with just ID and return an
// other slice with ID, username and role set.
//
// fillMembersData returns a slice of models.Member and an error. When error is not nil, the slice of models.Member is nil.
func (s *service) fillMembersData(ctx context.Context, members []models.Member) ([]models.Member, error) {
	for index, member := range members {
		user, _, err := s.store.UserGetByID(ctx, member.ID, false)
		if err != nil || user == nil {
			return nil, NewErrUserNotFound(member.ID, err)
		}

		members[index] = models.Member{
			ID:       user.ID,
			AddedAt:  member.AddedAt,
			Username: user.Username, // TODO: aggregate this in a query
			Role:     member.Role,
			Status:   member.Status,
		}
	}

	return members, nil
}

func (s *service) EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error) {
	changes := &models.NamespaceChanges{
		Name:                   strings.ToLower(req.Name),
		SessionRecord:          req.Settings.SessionRecord,
		ConnectionAnnouncement: req.Settings.ConnectionAnnouncement,
	}

	if envs.IsEnterprise() {
		changes.VPNEnable = req.VPN.Enable

		if req.VPN.Address != nil {
			address := *req.VPN.Address
			ip := net.IPv4(address[0], address[1], address[2], address[3])

			if ip.IsLoopback() || ip.IsUnspecified() {
				return nil, NewErrNamespaceIPInvalid()
			}

			if !ip.IsPrivate() {
				return nil, NewErrNamespaceIPNotPrivate()
			}

			changes.VPNAddress = &address
		}

		changes.VPNMask = req.VPN.Mask
	}

	if err := s.store.NamespaceEdit(ctx, req.Tenant, changes); err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return nil, NewErrNamespaceNotFound(req.Tenant, err)
		default:
			return nil, err
		}
	}

	if envs.IsEnterprise() {
		cli := s.client.(internalclient.Client)

		if err := cli.VPNStopRouter(req.Tenant); err != nil {
			return nil, err
		}
	}

	return s.store.NamespaceGet(ctx, req.Tenant, true)
}

func (s *service) AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID, true)
	if err != nil || namespace == nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	// checks if the active member is in the namespace. user is the active member.
	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(user.ID, err)
	}

	if !active.Role.HasAuthority(req.MemberRole) {
		return nil, NewErrRoleInvalid()
	}

	passiveUser, err := s.store.UserGetByEmail(ctx, strings.ToLower(req.MemberEmail))
	if err != nil {
		return nil, NewErrUserNotFound(req.MemberEmail, err)
	}

	// Currently, the member's status is always "accepted".
	member := &models.Member{
		ID:      passiveUser.ID,
		AddedAt: clock.Now(),
		Role:    req.MemberRole,
		Status:  models.MemberStatusAccepted,
	}
	if err := s.store.NamespaceAddMember(ctx, req.TenantID, member); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNamespaceDuplicatedMember):
			return nil, NewErrNamespaceMemberDuplicated(passiveUser.ID, err)
		default:
			return nil, err
		}
	}

	return s.store.NamespaceGet(ctx, req.TenantID, true)
}

func (s *service) UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID, true)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil {
		return NewErrUserNotFound(req.UserID, err)
	}

	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return NewErrNamespaceMemberNotFound(user.ID, err)
	}

	if _, ok := namespace.FindMember(req.MemberID); !ok {
		return NewErrNamespaceMemberNotFound(req.MemberID, err)
	}

	changes := &models.MemberChanges{Role: req.MemberRole}

	if changes.Role != authorizer.RoleInvalid {
		if !active.Role.HasAuthority(req.MemberRole) {
			return NewErrRoleInvalid()
		}
	}

	if err := s.store.NamespaceUpdateMember(ctx, req.TenantID, req.MemberID, changes); err != nil {
		return err
	}

	s.AuthUncacheToken(ctx, namespace.TenantID, req.MemberID) // nolint: errcheck

	return nil
}

func (s *service) RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID, true)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(user.ID, err)
	}

	member, _, err := s.store.UserGetByID(ctx, req.MemberID, false)
	if err != nil {
		return nil, NewErrUserNotFound(req.MemberID, err)
	}

	passive, ok := namespace.FindMember(member.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(member.ID, err)
	}

	if !active.Role.HasAuthority(passive.Role) {
		return nil, NewErrRoleInvalid()
	}

	if err := s.store.NamespaceRemoveMember(ctx, req.TenantID, req.MemberID); err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return nil, NewErrNamespaceNotFound(req.TenantID, err)
		case errors.Is(err, mongo.ErrUserNotFound):
			return nil, NewErrNamespaceMemberNotFound(req.MemberID, err)
		default:
			return nil, err
		}
	}

	s.AuthUncacheToken(ctx, namespace.TenantID, member.ID) // nolint: errcheck

	return s.store.NamespaceGet(ctx, req.TenantID, true)
}

// EditSessionRecordStatus defines if the sessions will be recorded.
//
// It receives a context, used to "control" the request flow, a boolean to define if the sessions will be recorded and
// the tenant ID from models.Namespace.
func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	return s.store.NamespaceSetSessionRecord(ctx, sessionRecord, tenantID)
}

// GetSessionRecord gets the session record data.
//
// It receives a context, used to "control" the request flow, the tenant ID from models.Namespace.
//
// GetSessionRecord returns a boolean indicating the session record status and an error. When error is not nil,
// the boolean is false.
func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	if _, err := s.store.NamespaceGet(ctx, tenantID, false); err != nil {
		return false, NewErrNamespaceNotFound(tenantID, err)
	}

	return s.store.NamespaceGetSessionRecord(ctx, tenantID)
}
