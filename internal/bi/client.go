package bi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles Beyond Identity SCIM API operations
type Client struct {
	apiToken    string
	scimBaseURL string
	httpClient  *http.Client
}

// User represents a Beyond Identity SCIM user
type User struct {
	ID          string      `json:"id,omitempty"`
	UserName    string      `json:"userName"`
	DisplayName string      `json:"displayName"`
	Emails      []Email     `json:"emails"`
	Active      bool        `json:"active"`
	Schemas     []string    `json:"schemas"`
	Groups      []UserGroup `json:"groups,omitempty"`
}

// Email represents a user's email address
type Email struct {
	Value   string `json:"value"`
	Type    string `json:"type"`
	Primary bool   `json:"primary"`
}

// UserGroup represents a group membership for a user
type UserGroup struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
}

// Group represents a Beyond Identity SCIM group
type Group struct {
	ID          string        `json:"id,omitempty"`
	DisplayName string        `json:"displayName"`
	Members     []GroupMember `json:"members,omitempty"`
	Schemas     []string      `json:"schemas"`
}

// GroupMember represents a member of a group
type GroupMember struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
}

// PatchOperation represents a SCIM PATCH operation
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// PatchRequest represents a SCIM PATCH request
type PatchRequest struct {
	Schemas    []string         `json:"schemas"`
	Operations []PatchOperation `json:"Operations"`
}

// SCIMError represents a SCIM API error response
type SCIMError struct {
	Schemas []string `json:"schemas"`
	Detail  string   `json:"detail"`
	Status  string   `json:"status"`
}

func (e *SCIMError) Error() string {
	return fmt.Sprintf("SCIM API error (status %s): %s", e.Status, e.Detail)
}

// NewClient creates a new Beyond Identity SCIM client
func NewClient(apiToken, scimBaseURL string) *Client {
	return &Client{
		apiToken:    apiToken,
		scimBaseURL: strings.TrimSuffix(scimBaseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest performs an HTTP request with proper authentication and error handling
func (c *Client) makeRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	// Handle SCIM errors
	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		bodyBytes, _ := io.ReadAll(resp.Body)

		var scimErr SCIMError
		if err := json.Unmarshal(bodyBytes, &scimErr); err == nil {
			return resp, &scimErr
		}

		return resp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// CreateUser creates a new user in Beyond Identity
func (c *Client) CreateUser(user *User) (*User, error) {
	user.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	user.Active = true

	resp, err := c.makeRequest("POST", c.scimBaseURL+"/Users", user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var createdUser User
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		return nil, fmt.Errorf("failed to decode created user: %w", err)
	}

	return &createdUser, nil
}

// UpdateUser updates an existing user in Beyond Identity
func (c *Client) UpdateUser(userID string, user *User) (*User, error) {
	user.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	resp, err := c.makeRequest("PUT", c.scimBaseURL+"/Users/"+userID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var updatedUser User
	if err := json.NewDecoder(resp.Body).Decode(&updatedUser); err != nil {
		return nil, fmt.Errorf("failed to decode updated user: %w", err)
	}

	return &updatedUser, nil
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(userID string) (*User, error) {
	resp, err := c.makeRequest("GET", c.scimBaseURL+"/Users/"+userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// FindUserByEmail searches for a user by email address
func (c *Client) FindUserByEmail(email string) (*User, error) {
	filter := fmt.Sprintf(`userName eq "%s"`, email)
	url := fmt.Sprintf("%s/Users?filter=%s", c.scimBaseURL, filter)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var searchResult struct {
		TotalResults int    `json:"totalResults"`
		Resources    []User `json:"Resources"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	if searchResult.TotalResults == 0 {
		return nil, nil // User not found
	}

	return &searchResult.Resources[0], nil
}

// CreateGroup creates a new group in Beyond Identity
func (c *Client) CreateGroup(group *Group) (*Group, error) {
	group.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:Group"}

	resp, err := c.makeRequest("POST", c.scimBaseURL+"/Groups", group)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var createdGroup Group
	if err := json.NewDecoder(resp.Body).Decode(&createdGroup); err != nil {
		return nil, fmt.Errorf("failed to decode created group: %w", err)
	}

	return &createdGroup, nil
}

// FindGroupByDisplayName searches for a group by display name
func (c *Client) FindGroupByDisplayName(displayName string) (*Group, error) {
	filter := fmt.Sprintf(`displayName eq "%s"`, displayName)
	url := fmt.Sprintf("%s/Groups?filter=%s", c.scimBaseURL, filter)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var searchResult struct {
		TotalResults int     `json:"totalResults"`
		Resources    []Group `json:"Resources"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	if searchResult.TotalResults == 0 {
		return nil, nil // Group not found
	}

	return &searchResult.Resources[0], nil
}

// UpdateGroupMembers updates group membership using PATCH operations
func (c *Client) UpdateGroupMembers(groupID string, addMembers, removeMembers []GroupMember) error {
	var operations []PatchOperation

	// Add remove operations first
	for _, member := range removeMembers {
		operations = append(operations, PatchOperation{
			Op:   "remove",
			Path: fmt.Sprintf("members[value eq \"%s\"]", member.Value),
		})
	}

	// Add add operations
	for _, member := range addMembers {
		operations = append(operations, PatchOperation{
			Op:    "add",
			Path:  "members",
			Value: member,
		})
	}

	if len(operations) == 0 {
		return nil // No changes needed
	}

	patchRequest := PatchRequest{
		Schemas:    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		Operations: operations,
	}

	resp, err := c.makeRequest("PATCH", c.scimBaseURL+"/Groups/"+groupID, patchRequest)
	if err != nil {
		return fmt.Errorf("failed to update group members: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}
