package sync

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/bi"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/gws"
	"github.com/sirupsen/logrus"
)

// Engine orchestrates the synchronization between Google Workspace and Beyond Identity
type Engine struct {
	gwsClient GWSClient
	biClient  BIClient
	config    *config.Config
	logger    *logrus.Logger
}

// SyncResult contains the results of a synchronization operation
type SyncResult struct {
	GroupsProcessed    int
	UsersCreated       int
	UsersUpdated       int
	GroupsCreated      int
	MembershipsAdded   int
	MembershipsRemoved int
	Errors             []error
}

// NewEngine creates a new sync engine
func NewEngine(gwsClient GWSClient, biClient BIClient, cfg *config.Config, logger *logrus.Logger) *Engine {
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

	// Sync enrollment status to Google Workspace
	e.logger.Infof("Starting enrollment status sync for %d members", len(gwsMembers))
	if err := e.syncEnrollmentStatus(gwsMembers, result); err != nil {
		e.logger.Errorf("Failed to sync enrollment status: %v", err)
		result.Errors = append(result.Errors, fmt.Errorf("enrollment sync: %w", err))
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
		// Return a mock group for test mode (no actual API call made)
		return &bi.Group{
			ID:          "mock-group-id-for-testing",
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
		return "mock-user-id-for-testing", nil
	}

	e.logger.Infof("Creating new user: %s", email)

	// Extract display name from email
	displayName := extractDisplayName(email)

	newUser := &bi.User{
		ExternalID:  email,
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
func (e *Engine) updateGroupMembership(groupID string, desiredUserIDs []string, result *SyncResult) error {
	if e.config.App.TestMode {
		e.logger.Infof("TEST MODE: Would update group %s with %d members", groupID, len(desiredUserIDs))
		return nil
	}

	// Get current group members from BI to calculate what needs to change
	e.logger.Debugf("Getting current members for group %s", groupID)
	currentGroup, err := e.biClient.GetGroupWithMembers(groupID)
	if err != nil {
		return fmt.Errorf("failed to get current group members: %w", err)
	}

	// Create sets for easier comparison
	currentMemberIDs := make(map[string]bool)
	for _, member := range currentGroup.Members {
		currentMemberIDs[member.Value] = true
	}

	desiredMemberIDs := make(map[string]bool)
	for _, userID := range desiredUserIDs {
		desiredMemberIDs[userID] = true
	}

	// Calculate members to add (in desired but not in current)
	var membersToAdd []bi.GroupMember
	for userID := range desiredMemberIDs {
		if !currentMemberIDs[userID] {
			membersToAdd = append(membersToAdd, bi.GroupMember{
				Value: userID,
			})
		}
	}

	// Calculate members to remove (in current but not in desired)
	var membersToRemove []bi.GroupMember
	for _, member := range currentGroup.Members {
		if !desiredMemberIDs[member.Value] {
			membersToRemove = append(membersToRemove, bi.GroupMember{
				Value: member.Value,
			})
		}
	}

	// Only make API call if there are changes needed
	if len(membersToAdd) == 0 && len(membersToRemove) == 0 {
		e.logger.Infof("Group %s membership is already up to date (%d members)", groupID, len(currentGroup.Members))
		return nil
	}

	e.logger.Infof("Updating group membership for group %s: +%d members, -%d members", 
		groupID, len(membersToAdd), len(membersToRemove))

	// Update group membership with proper add/remove operations
	err = e.biClient.UpdateGroupMembers(groupID, membersToAdd, membersToRemove)
	if err != nil {
		return fmt.Errorf("failed to update group members: %w", err)
	}

	result.MembershipsAdded += len(membersToAdd)
	result.MembershipsRemoved += len(membersToRemove)
	
	e.logger.Infof("Successfully updated group membership: added %d, removed %d members", 
		len(membersToAdd), len(membersToRemove))

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

// syncEnrollmentStatus manages the BYID_Enrolled Google group based on Beyond Identity user enrollment status (active + has active passkey)
func (e *Engine) syncEnrollmentStatus(gwsMembers []*gws.GroupMember, result *SyncResult) error {
	e.logger.Infof("Managing enrollment group: %s (%s)", e.config.Sync.EnrollmentGroupName, e.config.Sync.EnrollmentGroupEmail)

	// Ensure the enrollment group exists
	enrollmentGroup, err := e.gwsClient.EnsureGroup(
		e.config.Sync.EnrollmentGroupEmail,
		e.config.Sync.EnrollmentGroupName,
		"Users who have successfully enrolled with Beyond Identity",
	)
	if err != nil {
		return fmt.Errorf("failed to ensure enrollment group: %w", err)
	}

	e.logger.Debugf("Managing enrollment group: %s", enrollmentGroup.Email)

	// Get current members of the enrollment group
	currentMembers, err := e.gwsClient.GetGroupMembers(enrollmentGroup.Email)
	if err != nil {
		return fmt.Errorf("failed to get enrollment group members: %w", err)
	}

	// Create a map of current members for quick lookup
	currentMemberMap := make(map[string]bool)
	for _, member := range currentMembers {
		currentMemberMap[member.Email] = true
	}

	// Process each user in the sync scope
	for _, member := range gwsMembers {
		// Skip non-user members
		if member.Type != "USER" {
			continue
		}

		// Skip suspended members
		if member.Status == "SUSPENDED" {
			continue
		}

		// Check Beyond Identity enrollment status (active AND has active passkey)
		isEnrolled, err := e.biClient.GetUserStatus(member.Email)
		if err != nil {
			e.logger.Warnf("Failed to get BI enrollment status for %s: %v", member.Email, err)
			continue
		}

		isCurrentlyInGroup := currentMemberMap[member.Email]

		if isEnrolled && !isCurrentlyInGroup {
			// User is enrolled in BI (active + has passkey) but not in enrollment group - add them
			if e.config.App.TestMode {
				e.logger.Infof("TEST MODE: Would add %s to enrollment group (active with passkey)", member.Email)
			} else {
				e.logger.Infof("Adding %s to enrollment group (active with passkey)", member.Email)
				if err := e.gwsClient.AddMemberToGroup(enrollmentGroup.Email, member.Email); err != nil {
					e.logger.Errorf("Failed to add %s to enrollment group: %v", member.Email, err)
					continue
				}
			}
			result.MembershipsAdded++
		} else if !isEnrolled && isCurrentlyInGroup {
			// User is not enrolled in BI (inactive or no passkey) but still in enrollment group - remove them
			if e.config.App.TestMode {
				e.logger.Infof("TEST MODE: Would remove %s from enrollment group (not enrolled or no passkey)", member.Email)
			} else {
				e.logger.Infof("Removing %s from enrollment group (not enrolled or no passkey)", member.Email)
				if err := e.gwsClient.RemoveMemberFromGroup(enrollmentGroup.Email, member.Email); err != nil {
					e.logger.Errorf("Failed to remove %s from enrollment group: %v", member.Email, err)
					continue
				}
			}
			result.MembershipsRemoved++
		}
	}

	return nil
}
