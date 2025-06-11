package sync

import (
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/bi"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/gws"
)

// GWSClient interface for Google Workspace operations
type GWSClient interface {
	GetGroup(email string) (*gws.Group, error)
	GetGroupMembers(email string) ([]*gws.GroupMember, error)
	AddMemberToGroup(groupEmail, userEmail string) error
	RemoveMemberFromGroup(groupEmail, userEmail string) error
	EnsureGroup(groupEmail, groupName, description string) (*gws.Group, error)
}

// BIClient interface for Beyond Identity operations
type BIClient interface {
	FindGroupByDisplayName(name string) (*bi.Group, error)
	CreateGroup(group *bi.Group) (*bi.Group, error)
	FindUserByEmail(email string) (*bi.User, error)
	CreateUser(user *bi.User) (*bi.User, error)
	UpdateGroupMembers(groupID string, membersToAdd []bi.GroupMember, membersToRemove []bi.GroupMember) error
	GetUserStatus(userEmail string) (bool, error)
	GetGroupWithMembers(groupID string) (*bi.Group, error)
}
