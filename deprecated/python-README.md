![Beyond Identity Logo](docs/images/bi_logo.png)

# Google Workspace users - BI Tenant SCIM

## Table of Contents

1. [Introduction](#introduction)
2. [Prerequisites](#prerequisites)
3. [GWS configuration](#gws-configuration)
   1. [Create a new project](#create-a-new-project)
   2. [Enable Admin SDK API](#enable-admin-sdk-api)
   3. [Create google service account](#create-google-service-account)
   4. [Create key the service account](#create-key-the-service-account)
   5. [Grant access to service account](#grant-access-to-service-account)
4. [Beyond Identity Configuration](#beyond-identity-configuration)
5. [Installation and Setup](#installation-and-setup)

# Introduction

The `gwbisync.py` script is a powerful synchronization tool that bridges Google Workspace (GWS) and Beyond Identity (BI) platforms. Here's what it does:

1. **User Synchronization**:
   - Automatically syncs users from specified Google Workspace groups to Beyond Identity
   - Creates new users in Beyond Identity for any new Google Workspace users in synchronized groups
   - Updates existing Beyond Identity users when their Google Workspace profiles change
   - Handles user suspension/deletion when users are removed from Google Workspace synchorinzed groups

2. **Group Management**:
   - Creates corresponding groups in Beyond Identity for each selected Google Workspace group
   - Maintains group membership synchronization to these groups
   

3. **Enrollment Tracking**:
   - Creates and manages a special "BYID_Enrolled" group in Google Workspace
   - Tracks which users have enrolled in Beyond Identity
   - Automatically adds/removes users from the enrollment group based on their Beyond Identity enrollment status

4. **Data Consistency**:
   - Maintains a consistent link between Google Workspace and Beyond Identity users using immutable Google Workspace user IDs
   - Ensures user attributes (name, email, status) stay in sync between both platforms
   - Handles edge cases like user suspension, deletion, and group membership changes

5. **Automated Maintenance**:
   - Includes error handling
   - Does not affect users in Beyond Identity created outside of SCIM
   - Will prevent duplicate users from being created
   - Provides detailed user information and status tracking

## Prerequisites

> Ensure that you have the following:

- A Beyond Identity tenant configured for your organization
- A valid Beyond Identity API bearer token with the appropriate scopes
- A GWS tenant and admin credentials to login to Google cloud and admin console

# GWS configuration

## Create a new project

> Step 1: Go to the Google Cloud Console
>
> First, go to the Google Cloud Console by visiting https://console.cloud.google.com/ and signing in with your Google Workspace account.
>
> Step 2: Create a new project.
>
> Once you're in the console, click on the project dropdown in the top navigation bar and select "New Project".

![Create New Project](docs/images/gcp_create_project.png)

> Type in a name for the project. Choose your organization from the drop down. Click "CREATE"

![Project Creation Form](docs/images/gcp_project_form.png)

## Enable Admin SDK API

> Admin SDK API must be enabled at the project level to access google directory user APIs. Select the newly created project.

![Project Selection](docs/images/gcp_project_selection.png)

> Navigate to the API Library by clicking on "APIs & Services" in the left navigation menu, then select "Library".

![API Library Navigation](docs/images/gcp_api_library.png)

> In the search bar, type "Admin API" and select "Admin SDK API". Then click the "Enable" button to enable the API for your project.

![Admin SDK API Search](docs/images/gcp_admin_sdk_search.png)

![Admin SDK API Details](docs/images/gcp_admin_sdk_details.png)

![API Enablement](docs/images/gcp_api_enabled.png)

## Create google service account

> Next, you'll need to create a service account for your project. In the left navigation menu, select "IAM & Admin" and then select "Service Accounts".
>
> Click on the "Create Service Account" button, give your service account a name, and click "Create".

![Service Accounts Page](docs/images/gcp_service_accounts.png)

> Click on "CREATE SERVICE ACCOUNT"

![Create Service Account Form](docs/images/gcp_create_service_account.png)

> Type in a name for the service account, description and click on "CREATE AND CONTINUE"

![Service Account Details](docs/images/gcp_service_account_details.png)

> Click on "CONTINUE" on the next screen shown below.

![Service Account Permissions](docs/images/gcp_service_account_permissions.png)

> Click on "DONE" on the next screen as shown below.

![Service Account Completion](docs/images/gcp_service_account_completion.png)

> You will see the service account created as shown below.

![Service Account Created](docs/images/gcp_service_account_created.png)

## Create key the service account

> After you create your service account, you'll need to grant it permissions to access the Admin API. To do this, click on the service account you just created and then click on the "Add Key" button.
>
> Select "JSON" as the key type and click "Create". This will download a JSON file containing the private key for your service account.

![Service Account Key Creation](docs/images/gcp_key_creation.png)

![Key Type Selection](docs/images/gcp_key_type.png)

![Key Creation](docs/images/gcp_key_created.png)

![Key Download](docs/images/gcp_key_download.png)

## Grant access to service account

> Next, you'll need to grant access to the service account for the specific Google Workspace organization or domain you want to manage.
>
> To do this, log in to your Google Workspace admin console and navigate to "Security" > "Access and data control">"API controls". Under "Domain wide delegation", click on "Manage Domain Wide Delegation".

![API Controls Navigation](docs/images/gws_api_controls.png)

![Domain Wide Delegation](docs/images/gws_domain_delegation.png)

> In "Add a new client ID" screen, type in Client ID: The "Client ID" of your service account, which can be found in the JSON file you downloaded. For OAuth Scopes: Enter the following scopes for the Admin SDK user API:

```
https://www.googleapis.com/auth/admin.directory.user
https://www.googleapis.com/auth/admin.directory.group
https://www.googleapis.com/auth/admin.directory.group.member
```

>  click "AUTHORIZE"

![Client ID Authorization](docs/images/gws_client_id_auth.png)

![Authorization Complete](docs/images/gws_auth_complete.png)

# Beyond Identity Configuration

>Generate Tenant API token from support console with the following scopes:
  - scim:groups:create
  - scim:groups:delete
  - scim:groups:read
  - scim:groups:update
  - scim:users:create
  - scim:users:delete
  - scim:users:read
  - scim:users:update
  - users:read
  

![Beyond Identity Token Creation Interface](docs/images/bi_token_creation.png)

# Installation and Setup

## Quick Start

1. Run the setup script:
   ```bash
   ./setup.sh
   ```
   This will:
   - Create a Python virtual environment
   - Install all required dependencies
   - Set up necessary directories

## Manual Setup

If you prefer to set up manually or the setup script doesn't work for your environment:

1. Create a virtual environment:
   ```bash
   python3 -m venv venv
   ```

2. Activate the virtual environment:
   - On Unix or MacOS:
     ```bash
     source venv/bin/activate
     ```
   - On Windows:
     ```bash
     .\venv\Scripts\activate
     ```

3. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Configuration

1. Open the `config.py` file and update the following variables:

```python
# Google Workspace Configuration
GWS_SUPER_ADMIN_EMAIL = "your.admin@domain.com"  # GWS super admin email
GWS_DOMAIN_NAME = "domain.com"  # Your GWS domain name
SERVICE_ACCOUNT_KEY_PATH = "path/to/your-service-account-key.json"  # Path to your downloaded service account key

# List of Google Workspace groups to sync
GWS_GROUPS = [
    "group1@domain.com",
    "group2@domain.com",
    # Add more groups as needed
]

# Beyond Identity Configuration
BI_GROUP_PREFIX = "GoogleSCIM_"  # Prefix for Beyond Identity groups (can be customized)
BI_TENANT_API_TOKEN = "your-bi-tenant-api-token"  # The API token generated with required scopes

# Logging Configuration
LOGGING_ENABLED = True  # Set to False to disable file logging (console logging always enabled)
LOGGING_LEVEL = "INFO"  # Choose from: DEBUG, INFO, WARNING, ERROR, CRITICAL
LOGGING_DIR = "logs"  # Directory where log files will be stored
LOGGING_FILE_FORMAT = "sync_%Y%m%d_%H%M%S.log"  # Log file name format using strftime
```

### Logging Configuration

The script provides flexible logging options that can be configured in `config.py`:

- **File Logging**: Enable or disable logging to files using `LOGGING_ENABLED`
  - When enabled, logs are written to both console and file
  - When disabled, logs are only written to console
  - Log files are stored in the directory specified by `LOGGING_DIR`

- **Log Levels**:
  - `DEBUG`: Detailed information for debugging
  - `INFO`: General operational information (default)
  - `WARNING`: Warning messages for potential issues
  - `ERROR`: Error messages for failed operations
  - `CRITICAL`: Critical errors that may prevent operation

- **Log File Format**:
  - Default format: `sync_YYYYMMDD_HHMMSS.log`
  - Customizable using standard strftime format codes
  - Each sync operation creates a new log file

- **Log Message Format**:
  - Timestamp
  - Log level
  - Detailed message
  - Example: `2025-04-10 23:57:56,102 - INFO - Starting sync process`

## Test Mode Configuration

The script includes a test mode feature that allows you to safely preview changes without making actual modifications to your systems:

- **Test Mode**: Enable or disable test mode using `TEST_MODE`
  - When enabled (`TEST_MODE = True`):
    - The script makes actual API calls to check current state
    - Shows what changes would be made
    - No actual modifications are performed
    - All operations are logged with "[TEST MODE]" prefix
  - When disabled (`TEST_MODE = False`):
    - The script performs all operations normally
    - Makes actual changes to both Google Workspace and Beyond Identity

- **Test Mode Features**:
  - Real-time status checks of groups and users
  - Preview of all potential changes
  - Safe environment for testing configuration
  - Detailed logging of what would be changed

- **Example Test Mode Output**:
  ```
  [TEST MODE] Found 1 members in group testgroup@example.com
  [TEST MODE] Would update user test@example.com with ID 123456
  [TEST MODE] Would add user 123456 to group 789012
  ```

2. Ensure the service account key JSON file is placed in the location specified in `SERVICE_ACCOUNT_KEY_PATH`

## Running the Script

1. Make sure you're in the virtual environment (you should see `(venv)` in your terminal prompt)
   ```bash
   source venv/bin/activate  # if not already activated
   ```

2. Run the synchronization script:
   ```bash
   python gwbisync.py
   ```

The script will:
- Connect to both Google Workspace and Beyond Identity using the provided credentials
- Create/update groups in Beyond Identity for each specified Google Workspace group
- Synchronize users from the specified Google Workspace groups
- Track enrollment status and maintain the BYID_Enrolled group
- Output progress and status information during the sync

## Project Structure

```
├── README.md
├── config.py
├── gwbisync.py
├── requirements.txt
├── setup.sh
└── logs/
```
