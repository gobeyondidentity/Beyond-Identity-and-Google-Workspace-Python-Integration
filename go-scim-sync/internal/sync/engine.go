package sync

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobeyondidentity/go-scim-sync/internal/bi"
	"github.com/gobeyondidentity/go-scim-sync/internal/config"
	"github.com/gobeyondidentity/go-scim-sync/internal/gws"
	"github.com/sirupsen/logrus"
)

// Engine orchestrates the synchronization between Google Workspace and Beyond Identity
type Engine struct {
	gwsClient *gws.Client
	biClient  *bi.Client
	config    *config.Config
	logger    *logrus.Logger
}

// SyncResult contains the results of a synchronization operation
type SyncResult struct {
	GroupsProcessed   int
	UsersCreated      int
	UsersUpdated      int
	GroupsCreated     int
	MembershipsAdded  int
	MembershipsRemoved int
	Errors            []error
}

// NewEngine creates a new sync engine
func NewEngine(gwsClient *gws.Client, biClient *bi.Client, cfg *config.Config, logger *logrus.Logger) *Engine {
	return &Engine{
		gwsClient: gwsClient,
		biClient:  biClient,
		config:    cfg,
		logger:    logger,
	}
}

// Sync performs the complete synchronization process
func (e *Engine) Sync() (*SyncResult, error) {
	result := &SyncResult{}
	
	e.logger.Info("Starting sync process...")
	
	for _, groupEmail := range e.config.Sync.Groups {
		e.logger.Infof("Processing group: %s", groupEmail)
		
		if err := e.syncGroup(groupEmail, result); err != nil {
			e.logger.Errorf("Failed to sync group %s: %v", groupEmail, err)
			result.Errors = append(result.Errors, fmt.Errorf("group %s: %w", groupEmail, err))
			continue
		}
		
		result.GroupsProcessed++
	}
	
	e.logger.Infof("Sync completed. Groups: %d, Users created: %d, Users updated: %d, Groups created: %d, Memberships added: %d, Memberships removed: %d, Errors: %d",
		result.GroupsProcessed, result.UsersCreated, result.UsersUpdated, result.GroupsCreated,
		result.MembershipsAdded, result.MembershipsRemoved, len(result.Errors))
	
	return result, nil
}

// syncGroup synchronizes a single Google Workspace group to Beyond Identity
func (e *Engine) syncGroup(groupEmail string, result *SyncResult) error {
	// Get the Google Workspace group
	gwsGroup, err := e.gwsClient.GetGroup(groupEmail)
	if err != nil {
		return fmt.Errorf("failed to get GWS group: %w", err)
	}
	
	// Get group members from Google Workspace
	gwsMembers, err := e.gwsClient.GetGroupMembers(groupEmail)
	if err != nil {
		return fmt.Errorf("failed to get GWS group members: %w", err)
	}
	
	e.logger.Infof("Found %d members in Google Workspace group %s", len(gwsMembers), groupEmail)
	
	// Create or get the Beyond Identity group
	biGroupName := e.config.BeyondIdentity.GroupPrefix + gwsGroup.Name
	biGroup, err := e.ensureBIGroup(biGroupName, gwsGroup.Description, result)
	if err != nil {
		return fmt.Errorf("failed to ensure BI group: %w", err)
	}
	
	// Sync users and collect their IDs
	userIDs, err := e.syncUsers(gwsMembers, result)
	if err != nil {
		return fmt.Errorf("failed to sync users: %w", err)
	}
	
	// Update group membership
	if err := e.updateGroupMembership(biGroup.ID, userIDs, result); err != nil {
		return fmt.Errorf("failed to update group membership: %w", err)
	}
	
	return nil
}

// ensureBIGroup creates or retrieves a Beyond Identity group
func (e *Engine) ensureBIGroup(groupName, description string, result *SyncResult) (*bi.Group, error) {
	// Try to find existing group
	existingGroup, err := e.biClient.FindGroupByDisplayName(groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for group: %w", err)
	}
	
	if existingGroup != nil {
		e.logger.Debugf("Using existing group: %s (ID: %s)", groupName, existingGroup.ID)
		return existingGroup, nil
	}
	
	// Create new group
	if e.config.App.TestMode {
		e.logger.Infof("TEST MODE: Would create group '%s' with description '%s'", groupName, description)
		// Return a fake group for test mode
		return &bi.Group{
			ID:          "test-group-id",
			DisplayName: groupName,
		}, nil
	}
	
	e.logger.Infof("Creating new group: %s", groupName)
	newGroup := &bi.Group{
		DisplayName: groupName,
	}
	
	if description != "" {
		// Note: SCIM 2.0 Group schema doesn't have description field in core schema
		// We'll just log it for now
		e.logger.Debugf("Group description (not stored in SCIM): %s", description)
	}
	
	createdGroup, err := e.biClient.CreateGroup(newGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	
	result.GroupsCreated++
	e.logger.Infof("Created group: %s (ID: %s)", groupName, createdGroup.ID)
	
	return createdGroup, nil
}

// syncUsers ensures all users exist in Beyond Identity and returns their IDs
func (e *Engine) syncUsers(gwsMembers []*gws.GroupMember, result *SyncResult) ([]string, error) {
	var userIDs []string
	
	for _, member := range gwsMembers {
		// Skip non-user members (groups, etc.)
		if member.Type != "USER" {
			e.logger.Debugf("Skipping non-user member: %s (type: %s)", member.Email, member.Type)
			continue
		}
		
		// Skip suspended members
		if member.Status == "SUSPENDED" {
			e.logger.Debugf("Skipping suspended member: %s", member.Email)
			continue
		}
		
		userID, err := e.ensureBIUser(member.Email, result)
		if err != nil {
			e.logger.Errorf("Failed to ensure user %s: %v", member.Email, err)
			result.Errors = append(result.Errors, fmt.Errorf("user %s: %w", member.Email, err))
			continue
		}
		
		if userID != "" {
			userIDs = append(userIDs, userID)
		}
	}
	
	return userIDs, nil
}

// ensureBIUser creates or updates a user in Beyond Identity
func (e *Engine) ensureBIUser(email string, result *SyncResult) (string, error) {
	// Try to find existing user
	existingUser, err := e.biClient.FindUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to search for user: %w", err)
	}
	
	if existingUser != nil {
		e.logger.Debugf("Found existing user: %s (ID: %s)", email, existingUser.ID)
		// Check if user needs updating (could add logic here to update displayName, etc.)
		return existingUser.ID, nil
	}
	
	// Create new user
	if e.config.App.TestMode {
		e.logger.Infof("TEST MODE: Would create user '%s'", email)
		return "test-user-id", nil
	}
	
	e.logger.Infof("Creating new user: %s", email)
	
	// Extract display name from email
	displayName := extractDisplayName(email)
	
	newUser := &bi.User{
		UserName:    email,
		DisplayName: displayName,
		Emails: []bi.Email{
			{
				Value:   email,
				Type:    "work",
				Primary: true,
			},
		},
		Active: true,
	}
	
	createdUser, err := e.biClient.CreateUser(newUser)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	
	result.UsersCreated++
	e.logger.Infof("Created user: %s (ID: %s)", email, createdUser.ID)
	
	return createdUser.ID, nil
}

// updateGroupMembership updates the membership of a Beyond Identity group
func (e *Engine) updateGroupMembership(groupID string, userIDs []string, result *SyncResult) error {
	if e.config.App.TestMode {
		e.logger.Infof("TEST MODE: Would update group %s with %d members", groupID, len(userIDs))
		return nil
	}
	
	// Convert user IDs to group members
	var newMembers []bi.GroupMember
	for _, userID := range userIDs {
		newMembers = append(newMembers, bi.GroupMember{
			Value: userID,
		})
	}
	
	// For simplicity, we'll replace all members (remove all, then add all)
	// In a production system, you might want to be more surgical about this
	e.logger.Infof("Updating group membership for group %s with %d members", groupID, len(newMembers))
	
	// Note: This is a simplified approach. A more sophisticated implementation
	// would compare existing members and only add/remove the differences.
	err := e.biClient.UpdateGroupMembers(groupID, newMembers, []bi.GroupMember{})
	if err != nil {
		return fmt.Errorf("failed to update group members: %w", err)
	}
	
	result.MembershipsAdded += len(newMembers)
	e.logger.Infof("Updated group membership: added %d members", len(newMembers))
	
	return nil
}

// extractDisplayName extracts a display name from an email address
func extractDisplayName(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return email
	}
	
	localPart := parts[0]
	
	// Try to make it more readable
	// Replace common separators with spaces
	displayName := strings.ReplaceAll(localPart, ".", " ")
	displayName = strings.ReplaceAll(displayName, "_", " ")
	displayName = strings.ReplaceAll(displayName, "-", " ")
	
	// Title case each word
	words := strings.Fields(displayName)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	
	result := strings.Join(words, " ")
	if result == "" {
		return email // Fallback to email if we can't parse it
	}
	
	return result
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func (e *Engine) RetryWithBackoff(operation func() error, maxAttempts int, baseDelay time.Duration) error {
	var lastErr error
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := operation(); err != nil {
			lastErr = err
			
			if attempt == maxAttempts {
				break
			}
			
			delay := time.Duration(attempt) * baseDelay
			e.logger.Warnf("Operation failed (attempt %d/%d), retrying in %v: %v", 
				attempt, maxAttempts, delay, err)
			time.Sleep(delay)
			continue
		}
		
		return nil // Success
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}