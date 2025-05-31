package sync

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/bi"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/gws"
	"github.com/sirupsen/logrus"
)

// Mock clients for testing
type mockGWSClient struct {
	groups      map[string]*gws.Group
	members     map[string][]*gws.GroupMember
	shouldError bool
}

func (m *mockGWSClient) GetGroup(email string) (*gws.Group, error) {
	if m.shouldError {
		return nil, errors.New("mock GWS error")
	}
	if group, exists := m.groups[email]; exists {
		return group, nil
	}
	return nil, fmt.Errorf("group not found: %s", email)
}

func (m *mockGWSClient) GetGroupMembers(email string) ([]*gws.GroupMember, error) {
	if m.shouldError {
		return nil, errors.New("mock GWS members error")
	}
	if members, exists := m.members[email]; exists {
		return members, nil
	}
	return []*gws.GroupMember{}, nil
}

type mockBIClient struct {
	groups      map[string]*bi.Group
	users       map[string]*bi.User
	shouldError bool
}

func (m *mockBIClient) FindGroupByDisplayName(name string) (*bi.Group, error) {
	if m.shouldError {
		return nil, errors.New("mock BI group search error")
	}
	for _, group := range m.groups {
		if group.DisplayName == name {
			return group, nil
		}
	}
	return nil, nil
}

func (m *mockBIClient) CreateGroup(group *bi.Group) (*bi.Group, error) {
	if m.shouldError {
		return nil, errors.New("mock BI group creation error")
	}
	newGroup := &bi.Group{
		ID:          fmt.Sprintf("group-%d", len(m.groups)+1),
		DisplayName: group.DisplayName,
	}
	m.groups[newGroup.ID] = newGroup
	return newGroup, nil
}

func (m *mockBIClient) FindUserByEmail(email string) (*bi.User, error) {
	if m.shouldError {
		return nil, errors.New("mock BI user search error")
	}
	for _, user := range m.users {
		if len(user.Emails) > 0 && user.Emails[0].Value == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockBIClient) CreateUser(user *bi.User) (*bi.User, error) {
	if m.shouldError {
		return nil, errors.New("mock BI user creation error")
	}
	newUser := &bi.User{
		ID:          fmt.Sprintf("user-%d", len(m.users)+1),
		UserName:    user.UserName,
		DisplayName: user.DisplayName,
		Emails:      user.Emails,
		Active:      user.Active,
	}
	m.users[newUser.ID] = newUser
	return newUser, nil
}

func (m *mockBIClient) UpdateGroupMembers(groupID string, membersToAdd []bi.GroupMember, membersToRemove []bi.GroupMember) error {
	if m.shouldError {
		return errors.New("mock BI group update error")
	}
	return nil
}

func TestNewEngine(t *testing.T) {
	gwsClient := &mockGWSClient{}
	biClient := &mockBIClient{}
	cfg := &config.Config{}
	logger := logrus.New()

	engine := NewEngine(gwsClient, biClient, cfg, logger)

	if engine == nil {
		t.Error("Expected engine to be created, got nil")
		return
	}

	if engine.gwsClient != gwsClient {
		t.Error("Expected GWS client to match input")
	}

	if engine.biClient != biClient {
		t.Error("Expected BI client to match input")
	}

	if engine.config != cfg {
		t.Error("Expected config to match input")
	}

	if engine.logger != logger {
		t.Error("Expected logger to match input")
	}
}

func TestSync(t *testing.T) {
	tests := []struct {
		name          string
		setupClients  func() (*mockGWSClient, *mockBIClient)
		config        *config.Config
		expectError   bool
		expectedStats func(*SyncResult) error
	}{
		{
			name: "successful sync with new group and users",
			setupClients: func() (*mockGWSClient, *mockBIClient) {
				gwsClient := &mockGWSClient{
					groups: map[string]*gws.Group{
						"test@example.com": {
							Name:        "TestGroup",
							Description: "Test group description",
						},
					},
					members: map[string][]*gws.GroupMember{
						"test@example.com": {
							{Email: "user1@example.com", Type: "USER", Status: "ACTIVE"},
							{Email: "user2@example.com", Type: "USER", Status: "ACTIVE"},
						},
					},
				}
				biClient := &mockBIClient{
					groups: make(map[string]*bi.Group),
					users:  make(map[string]*bi.User),
				}
				return gwsClient, biClient
			},
			config: &config.Config{
				Sync: config.SyncConfig{
					Groups: []string{"test@example.com"},
				},
				BeyondIdentity: config.BeyondIdentityConfig{
					GroupPrefix: "GWS_",
				},
			},
			expectError: false,
			expectedStats: func(result *SyncResult) error {
				if result.GroupsProcessed != 1 {
					return fmt.Errorf("expected 1 group processed, got %d", result.GroupsProcessed)
				}
				if result.GroupsCreated != 1 {
					return fmt.Errorf("expected 1 group created, got %d", result.GroupsCreated)
				}
				if result.UsersCreated != 2 {
					return fmt.Errorf("expected 2 users created, got %d", result.UsersCreated)
				}
				if result.MembershipsAdded != 2 {
					return fmt.Errorf("expected 2 memberships added, got %d", result.MembershipsAdded)
				}
				return nil
			},
		},
		{
			name: "sync with existing group and users",
			setupClients: func() (*mockGWSClient, *mockBIClient) {
				gwsClient := &mockGWSClient{
					groups: map[string]*gws.Group{
						"existing@example.com": {
							Name:        "ExistingGroup",
							Description: "Existing group",
						},
					},
					members: map[string][]*gws.GroupMember{
						"existing@example.com": {
							{Email: "existing@example.com", Type: "USER", Status: "ACTIVE"},
						},
					},
				}
				biClient := &mockBIClient{
					groups: map[string]*bi.Group{
						"group-1": {
							ID:          "group-1",
							DisplayName: "GWS_ExistingGroup",
						},
					},
					users: map[string]*bi.User{
						"user-1": {
							ID:       "user-1",
							UserName: "existing@example.com",
							Emails: []bi.Email{
								{Value: "existing@example.com", Type: "work", Primary: true},
							},
						},
					},
				}
				return gwsClient, biClient
			},
			config: &config.Config{
				Sync: config.SyncConfig{
					Groups: []string{"existing@example.com"},
				},
				BeyondIdentity: config.BeyondIdentityConfig{
					GroupPrefix: "GWS_",
				},
			},
			expectError: false,
			expectedStats: func(result *SyncResult) error {
				if result.GroupsProcessed != 1 {
					return fmt.Errorf("expected 1 group processed, got %d", result.GroupsProcessed)
				}
				if result.GroupsCreated != 0 {
					return fmt.Errorf("expected 0 groups created, got %d", result.GroupsCreated)
				}
				if result.UsersCreated != 0 {
					return fmt.Errorf("expected 0 users created, got %d", result.UsersCreated)
				}
				return nil
			},
		},
		{
			name: "sync with suspended and non-user members",
			setupClients: func() (*mockGWSClient, *mockBIClient) {
				gwsClient := &mockGWSClient{
					groups: map[string]*gws.Group{
						"mixed@example.com": {
							Name:        "MixedGroup",
							Description: "Group with mixed members",
						},
					},
					members: map[string][]*gws.GroupMember{
						"mixed@example.com": {
							{Email: "active@example.com", Type: "USER", Status: "ACTIVE"},
							{Email: "suspended@example.com", Type: "USER", Status: "SUSPENDED"},
							{Email: "group@example.com", Type: "GROUP", Status: "ACTIVE"},
						},
					},
				}
				biClient := &mockBIClient{
					groups: make(map[string]*bi.Group),
					users:  make(map[string]*bi.User),
				}
				return gwsClient, biClient
			},
			config: &config.Config{
				Sync: config.SyncConfig{
					Groups: []string{"mixed@example.com"},
				},
				BeyondIdentity: config.BeyondIdentityConfig{
					GroupPrefix: "GWS_",
				},
			},
			expectError: false,
			expectedStats: func(result *SyncResult) error {
				if result.UsersCreated != 1 {
					return fmt.Errorf("expected 1 user created (only active user), got %d", result.UsersCreated)
				}
				if result.MembershipsAdded != 1 {
					return fmt.Errorf("expected 1 membership added, got %d", result.MembershipsAdded)
				}
				return nil
			},
		},
		{
			name: "test mode sync",
			setupClients: func() (*mockGWSClient, *mockBIClient) {
				gwsClient := &mockGWSClient{
					groups: map[string]*gws.Group{
						"test@example.com": {
							Name:        "TestGroup",
							Description: "Test group",
						},
					},
					members: map[string][]*gws.GroupMember{
						"test@example.com": {
							{Email: "user@example.com", Type: "USER", Status: "ACTIVE"},
						},
					},
				}
				biClient := &mockBIClient{
					groups: make(map[string]*bi.Group),
					users:  make(map[string]*bi.User),
				}
				return gwsClient, biClient
			},
			config: &config.Config{
				App: config.AppConfig{
					TestMode: true,
				},
				Sync: config.SyncConfig{
					Groups: []string{"test@example.com"},
				},
				BeyondIdentity: config.BeyondIdentityConfig{
					GroupPrefix: "GWS_",
				},
			},
			expectError: false,
			expectedStats: func(result *SyncResult) error {
				if result.GroupsProcessed != 1 {
					return fmt.Errorf("expected 1 group processed, got %d", result.GroupsProcessed)
				}
				// In test mode, we create fake groups/users but don't track them in stats
				if result.GroupsCreated != 0 {
					return fmt.Errorf("expected 0 groups created in test mode, got %d", result.GroupsCreated)
				}
				if result.UsersCreated != 0 {
					return fmt.Errorf("expected 0 users created in test mode, got %d", result.UsersCreated)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gwsClient, biClient := tt.setupClients()
			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel) // Reduce log noise during tests

			engine := NewEngine(gwsClient, biClient, tt.config, logger)
			result, err := engine.Sync()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result, got nil")
				return
			}

			if tt.expectedStats != nil {
				if err := tt.expectedStats(result); err != nil {
					t.Errorf("Stats validation failed: %v", err)
				}
			}
		})
	}
}

func TestExtractDisplayName(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{
			email:    "john.doe@example.com",
			expected: "John Doe",
		},
		{
			email:    "jane_smith@example.com",
			expected: "Jane Smith",
		},
		{
			email:    "bob-wilson@example.com",
			expected: "Bob Wilson",
		},
		{
			email:    "alice.mary.jones@example.com",
			expected: "Alice Mary Jones",
		},
		{
			email:    "simple@example.com",
			expected: "Simple",
		},
		{
			email:    "test.user_name-final@example.com",
			expected: "Test User Name Final",
		},
		{
			email:    "@example.com",
			expected: "@example.com", // Fallback to email
		},
		{
			email:    "noemail",
			expected: "Noemail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := extractDisplayName(tt.email)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name        string
		operation   func() error
		maxAttempts int
		expectError bool
		expectCalls int
	}{
		{
			name: "success on first try",
			operation: func() error {
				return nil
			},
			maxAttempts: 3,
			expectError: false,
			expectCalls: 1,
		},
		{
			name: "success on second try",
			operation: func() func() error {
				calls := 0
				return func() error {
					calls++
					if calls == 1 {
						return errors.New("first attempt fails")
					}
					return nil
				}
			}(),
			maxAttempts: 3,
			expectError: false,
			expectCalls: 2,
		},
		{
			name: "fail all attempts",
			operation: func() error {
				return errors.New("always fails")
			},
			maxAttempts: 2,
			expectError: true,
			expectCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &Engine{
				logger: logrus.New(),
			}
			engine.logger.SetLevel(logrus.FatalLevel) // Reduce log noise

			calls := 0
			operation := func() error {
				calls++
				return tt.operation()
			}

			err := engine.RetryWithBackoff(operation, tt.maxAttempts, 1*time.Millisecond)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if calls != tt.expectCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectCalls, calls)
			}
		})
	}
}

func TestSyncResult(t *testing.T) {
	result := &SyncResult{
		GroupsProcessed:    5,
		UsersCreated:       10,
		UsersUpdated:       3,
		GroupsCreated:      2,
		MembershipsAdded:   15,
		MembershipsRemoved: 1,
		Errors:             []error{fmt.Errorf("test error")},
	}

	if result.GroupsProcessed != 5 {
		t.Errorf("Expected GroupsProcessed 5, got %d", result.GroupsProcessed)
	}

	if result.UsersCreated != 10 {
		t.Errorf("Expected UsersCreated 10, got %d", result.UsersCreated)
	}

	if result.UsersUpdated != 3 {
		t.Errorf("Expected UsersUpdated 3, got %d", result.UsersUpdated)
	}

	if result.GroupsCreated != 2 {
		t.Errorf("Expected GroupsCreated 2, got %d", result.GroupsCreated)
	}

	if result.MembershipsAdded != 15 {
		t.Errorf("Expected MembershipsAdded 15, got %d", result.MembershipsAdded)
	}

	if result.MembershipsRemoved != 1 {
		t.Errorf("Expected MembershipsRemoved 1, got %d", result.MembershipsRemoved)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}
