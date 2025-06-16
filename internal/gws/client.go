package gws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// Client handles Google Workspace Admin SDK operations
type Client struct {
	service         *admin.Service
	domain          string
	superAdminEmail string
}

// User represents a Google Workspace user
type User struct {
	ID           string   `json:"id"`
	PrimaryEmail string   `json:"primaryEmail"`
	Name         UserName `json:"name"`
	Suspended    bool     `json:"suspended"`
	Archived     bool     `json:"archived"`
}

// UserName represents a user's name components
type UserName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	FullName   string `json:"fullName"`
}

// Group represents a Google Workspace group
type Group struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GroupMember represents a member of a Google Workspace group
type GroupMember struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// NewClient creates a new Google Workspace client
func NewClient(serviceAccountKeyPath, domain, superAdminEmail string) (*Client, error) {
	ctx := context.Background()

	// Read service account credentials
	credentialsJSON, err := os.ReadFile(serviceAccountKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file: %w", err)
	}

	// Parse credentials to get client email
	var creds struct {
		ClientEmail string `json:"client_email"`
	}
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
	}

	// Create JWT config for domain-wide delegation
	config, err := google.JWTConfigFromJSON(
		credentialsJSON,
		admin.AdminDirectoryUserScope,
		admin.AdminDirectoryGroupScope,
		admin.AdminDirectoryGroupMemberScope,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT config: %w", err)
	}

	// Set the subject for domain-wide delegation
	config.Subject = superAdminEmail

	// Create HTTP client
	httpClient := config.Client(ctx)

	// Create Admin SDK service
	service, err := admin.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Admin SDK service: %w", err)
	}

	return &Client{
		service:         service,
		domain:          domain,
		superAdminEmail: superAdminEmail,
	}, nil
}

// GetUsers retrieves all users in the domain
func (c *Client) GetUsers() ([]*User, error) {
	var allUsers []*User
	pageToken := ""

	for {
		call := c.service.Users.List().Domain(c.domain).MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		for _, user := range resp.Users {
			allUsers = append(allUsers, &User{
				ID:           user.Id,
				PrimaryEmail: user.PrimaryEmail,
				Name: UserName{
					GivenName:  user.Name.GivenName,
					FamilyName: user.Name.FamilyName,
					FullName:   user.Name.FullName,
				},
				Suspended: user.Suspended,
				Archived:  user.Archived,
			})
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allUsers, nil
}

// GetGroup retrieves a specific group by email
func (c *Client) GetGroup(groupEmail string) (*Group, error) {
	group, err := c.service.Groups.Get(groupEmail).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get group %s: %w", groupEmail, err)
	}

	return &Group{
		ID:          group.Id,
		Email:       group.Email,
		Name:        group.Name,
		Description: group.Description,
	}, nil
}

// GetGroupMembers retrieves all members of a group
func (c *Client) GetGroupMembers(groupEmail string) ([]*GroupMember, error) {
	var allMembers []*GroupMember
	pageToken := ""

	for {
		call := c.service.Members.List(groupEmail).MaxResults(200)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			// Handle case where group has no members
			if isNotFoundError(err) {
				return allMembers, nil
			}
			return nil, fmt.Errorf("failed to list members for group %s: %w", groupEmail, err)
		}

		for _, member := range resp.Members {
			allMembers = append(allMembers, &GroupMember{
				ID:     member.Id,
				Email:  member.Email,
				Role:   member.Role,
				Type:   member.Type,
				Status: member.Status,
			})
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allMembers, nil
}

// AddMemberToGroup adds a user to a Google Workspace group
func (c *Client) AddMemberToGroup(groupEmail, userEmail string) error {
	member := &admin.Member{
		Email: userEmail,
		Role:  "MEMBER",
		Type:  "USER",
	}

	_, err := c.service.Members.Insert(groupEmail, member).Do()
	if err != nil {
		// Check if user is already a member
		if googleErr, ok := err.(*googleapi.Error); ok && googleErr.Code == http.StatusConflict {
			return nil // User already in group, no error
		}
		return fmt.Errorf("failed to add member %s to group %s: %w", userEmail, groupEmail, err)
	}

	return nil
}

// RemoveMemberFromGroup removes a user from a Google Workspace group
func (c *Client) RemoveMemberFromGroup(groupEmail, userEmail string) error {
	err := c.service.Members.Delete(groupEmail, userEmail).Do()
	if err != nil {
		// Check if user is not a member (404 error)
		if isNotFoundError(err) {
			return nil // User not in group, no error
		}
		return fmt.Errorf("failed to remove member %s from group %s: %w", userEmail, groupEmail, err)
	}

	return nil
}

// CreateGroup creates a new Google Workspace group
func (c *Client) CreateGroup(groupEmail, groupName, description string) (*Group, error) {
	group := &admin.Group{
		Email:       groupEmail,
		Name:        groupName,
		Description: description,
	}

	createdGroup, err := c.service.Groups.Insert(group).Do()
	if err != nil {
		// Check if group already exists
		if googleErr, ok := err.(*googleapi.Error); ok && googleErr.Code == http.StatusConflict {
			// Group exists, fetch and return it
			return c.GetGroup(groupEmail)
		}
		return nil, fmt.Errorf("failed to create group %s: %w", groupEmail, err)
	}

	return &Group{
		ID:          createdGroup.Id,
		Email:       createdGroup.Email,
		Name:        createdGroup.Name,
		Description: createdGroup.Description,
	}, nil
}

// EnsureGroup ensures a group exists, creating it if necessary
func (c *Client) EnsureGroup(groupEmail, groupName, description string) (*Group, error) {
	// Try to get existing group
	group, err := c.GetGroup(groupEmail)
	if err != nil {
		// If not found, create it
		if isNotFoundError(err) {
			return c.CreateGroup(groupEmail, groupName, description)
		}
		return nil, fmt.Errorf("failed to check for existing group: %w", err)
	}
	return group, nil
}

// isNotFoundError checks if the error is a 404 not found error
func isNotFoundError(err error) bool {
	if googleErr, ok := err.(*googleapi.Error); ok {
		return googleErr.Code == http.StatusNotFound
	}
	// Also check for string patterns that indicate not found
	errorStr := err.Error()
	return strings.Contains(errorStr, "404") || strings.Contains(errorStr, "notFound") || strings.Contains(errorStr, "Resource Not Found")
}
