#!/usr/bin/env python3

import sys
import requests
import json
import logging
from datetime import datetime
from pathlib import Path
from google.oauth2 import service_account
from googleapiclient.discovery import build
from config import (
    GWS_SUPER_ADMIN_EMAIL,
    SERVICE_ACCOUNT_KEY_PATH,
    BI_TENANT_API_TOKEN,
    GWS_GROUPS,
    BI_GROUP_PREFIX,
    GWS_DOMAIN_NAME,
    LOGGING_ENABLED,
    LOGGING_LEVEL,
    LOGGING_DIR,
    LOGGING_FILE_FORMAT,
    TEST_MODE
)

# Set up logging
handlers = [logging.StreamHandler(sys.stdout)]  # Always log to stdout

if LOGGING_ENABLED:
    log_dir = Path(LOGGING_DIR)
    log_dir.mkdir(exist_ok=True)
    log_file = log_dir / datetime.now().strftime(LOGGING_FILE_FORMAT)
    handlers.append(logging.FileHandler(log_file))

logging.basicConfig(
    level=getattr(logging, LOGGING_LEVEL.upper()),
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=handlers
)

logger = logging.getLogger(__name__)

# Log start of sync with configuration info
logger.info("Starting Google Workspace to Beyond Identity sync process")
logger.info(f"Configured to sync the following groups: {', '.join(GWS_GROUPS)}")
logger.info(f"Logging {'enabled' if LOGGING_ENABLED else 'disabled'} at {LOGGING_LEVEL} level")
if TEST_MODE:
    logger.info("TEST MODE ENABLED - No actual changes will be made")

# Google API Scopes
SCOPES = [
    'https://www.googleapis.com/auth/admin.directory.user',
    'https://www.googleapis.com/auth/admin.directory.group',
    'https://www.googleapis.com/auth/admin.directory.group.member'
]

# API endpoints
BI_SCIM_USERS_URL = "https://api.byndid.com/scim/v2/Users"
BI_SCIM_GROUPS_URL = "https://api.byndid.com/scim/v2/Groups"
BI_NATIVE_USERS_URL = "https://api.byndid.com/v2/users"  # Native API endpoint

# Headers for API calls
headers = {
    "Content-Type": "application/scim+json",
    "Authorization": f"Bearer {BI_TENANT_API_TOKEN}"
}

def get_google_service():
    """Create and return a Google Workspace service instance"""
    credentials = service_account.Credentials.from_service_account_file(
        SERVICE_ACCOUNT_KEY_PATH, scopes=SCOPES
    )
    delegated_credentials = credentials.with_subject(GWS_SUPER_ADMIN_EMAIL)
    service = build('admin', 'directory_v1', credentials=delegated_credentials)
    if TEST_MODE:
        logger.info("[TEST MODE] Google Workspace service initialized - will show changes but not make them")
    return service

def get_group_members(service, group_email):
    """Get all members of a Google Workspace group"""
    try:
        members = []
        page_token = None
        while True:
            results = service.members().list(
                groupKey=group_email,
                pageToken=page_token
            ).execute()
            members.extend(results.get('members', []))
            page_token = results.get('nextPageToken')
            if not page_token:
                break
        if TEST_MODE:
            logger.info(f"[TEST MODE] Found {len(members)} members in group {group_email}")
        return members
    except Exception as e:
        logger.error(f"Error getting group members: {str(e)}")
        return []

def get_user_details(service, user_id):
    """Get detailed user information from Google Workspace"""
    if TEST_MODE:
        logger.info(f"[TEST MODE] Would get details for user {user_id}")
        return {
            'id': user_id,
            'primaryEmail': f"test_{user_id}@example.com",
            'name': {'givenName': 'Test', 'familyName': 'User'},
            'suspended': False
        }
    
    try:
        return service.users().get(userKey=user_id).execute()
    except Exception as e:
        logger.error(f"Error getting user details: {str(e)}")
        return None

def create_bi_group(group_email):
    """Create a group in Beyond Identity or get existing one"""
    try:
        # Check if group exists
        params = {'filter': f'displayName eq "{BI_GROUP_PREFIX}{group_email}"'}
        response = requests.get(BI_SCIM_GROUPS_URL, headers=headers, params=params)
        
        if response.status_code == 200:
            groups = response.json().get('Resources', [])
            if groups:
                if TEST_MODE:
                    logger.info(f"[TEST MODE] Group {BI_GROUP_PREFIX}{group_email} already exists with ID: {groups[0]['id']}")
                return groups[0]['id']
        
        if TEST_MODE:
            logger.info(f"[TEST MODE] Would create new group {BI_GROUP_PREFIX}{group_email}")
            return "test-group-id"
        
        # Create new group
        group_data = {
            "displayName": f"{BI_GROUP_PREFIX}{group_email}",
            "members": []
        }
        response = requests.post(BI_SCIM_GROUPS_URL, headers=headers, json=group_data)
        
        if response.status_code == 201:
            group_id = response.json()['id']
            logger.info(f"Created new group {BI_GROUP_PREFIX}{group_email} with ID: {group_id}")
            return group_id
        else:
            logger.error(f"Failed to create group: {response.text}")
            return None
    except Exception as e:
        logger.error(f"Error creating/getting group: {str(e)}")
        return None

def create_or_update_bi_user(user_data):
    """Create or update a user in Beyond Identity"""
    try:
        # Check if user exists
        params = {'filter': f'externalId eq "{user_data["id"]}"'}
        response = requests.get(BI_SCIM_USERS_URL, headers=headers, params=params)
        
        if response.status_code == 200:
            users = response.json().get('Resources', [])
            if users:
                user_id = users[0]['id']
                if TEST_MODE:
                    logger.info(f"[TEST MODE] Would update user {user_data['primaryEmail']} with ID {user_id}")
                    return user_id
                
                # Update existing user
                update_data = {
                    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
                    "userName": user_data['primaryEmail'],
                    "name": {
                        "givenName": user_data['name'].get('givenName', ''),
                        "familyName": user_data['name'].get('familyName', '')
                    },
                    "active": not user_data.get('suspended', False),
                    "externalId": user_data['id']
                }
                response = requests.put(
                    f"{BI_SCIM_USERS_URL}/{user_id}",
                    headers=headers,
                    json=update_data
                )
                if response.status_code == 200:
                    logger.info(f"Updated user {user_data['primaryEmail']}")
                    return user_id
                else:
                    logger.error(f"Failed to update user: {response.text}")
                    return None
        
        if TEST_MODE:
            logger.info(f"[TEST MODE] Would create new user {user_data['primaryEmail']}")
            return "test-user-id"
        
        # Create new user
        user_data = {
            "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
            "userName": user_data['primaryEmail'],
            "name": {
                "givenName": user_data['name'].get('givenName', ''),
                "familyName": user_data['name'].get('familyName', '')
            },
            "active": True,
            "externalId": user_data['id']
        }
        response = requests.post(BI_SCIM_USERS_URL, headers=headers, json=user_data)
        
        if response.status_code == 201:
            user_id = response.json()['id']
            logger.info(f"Created new user {user_data['userName']} with ID: {user_id}")
            return user_id
        else:
            logger.error(f"Failed to create user: {response.text}")
            return None
    except Exception as e:
        logger.error(f"Error creating/updating user: {str(e)}")
        return None

def add_user_to_group(user_id, group_id):
    """Add a user to a Beyond Identity group"""
    try:
        # Get current group members
        response = requests.get(f"{BI_SCIM_GROUPS_URL}/{group_id}", headers=headers)
        if response.status_code != 200:
            logger.error(f"Error getting group: {response.text}")
            return False
        
        group_data = response.json()
        members = group_data.get('members', [])
        
        # Check if user is already a member
        if any(member['value'] == user_id for member in members):
            if TEST_MODE:
                logger.info(f"[TEST MODE] User {user_id} is already a member of group {group_id}")
            return True
        
        if TEST_MODE:
            logger.info(f"[TEST MODE] Would add user {user_id} to group {group_id}")
            return True
        
        # Add user to group
        members.append({"value": user_id})
        group_data['members'] = members
        
        response = requests.put(
            f"{BI_SCIM_GROUPS_URL}/{group_id}",
            headers=headers,
            json=group_data
        )
        
        if response.status_code == 200:
            logger.info(f"Added user {user_id} to group {group_id}")
            return True
        else:
            logger.error(f"Failed to add user to group: {response.text}")
            return False
    except Exception as e:
        logger.error(f"Error adding user to group: {str(e)}")
        return False

def remove_user_from_group(user_id, group_id):
    """Remove a user from a Beyond Identity group"""
    try:
        # Get current group members
        response = requests.get(f"{BI_SCIM_GROUPS_URL}/{group_id}", headers=headers)
        if response.status_code != 200:
            logger.error(f"Error getting group: {response.text}")
            return False
        
        group_data = response.json()
        members = group_data.get('members', [])
        
        # Check if user is a member
        if not any(member['value'] == user_id for member in members):
            if TEST_MODE:
                logger.info(f"[TEST MODE] User {user_id} is not a member of group {group_id}")
            return True
        
        if TEST_MODE:
            logger.info(f"[TEST MODE] Would remove user {user_id} from group {group_id}")
            return True
        
        # Remove user from members list
        members = [member for member in members if member['value'] != user_id]
        group_data['members'] = members
        
        response = requests.put(
            f"{BI_SCIM_GROUPS_URL}/{group_id}",
            headers=headers,
            json=group_data
        )
        
        if response.status_code == 200:
            logger.info(f"Removed user {user_id} from group {group_id}")
            return True
        else:
            logger.error(f"Failed to remove user from group: {response.text}")
            return False
    except Exception as e:
        logger.error(f"Error removing user from group: {str(e)}")
        return False

def suspend_user(user_id):
    """Suspend a user in Beyond Identity"""
    try:
        # Get current user data
        response = requests.get(f"{BI_SCIM_USERS_URL}/{user_id}", headers=headers)
        if response.status_code != 200:
            logger.error(f"Error getting user: {response.text}")
            return False
        
        user_data = response.json()
        if user_data.get('active') == False:
            if TEST_MODE:
                logger.info(f"[TEST MODE] User {user_id} is already suspended")
            return True
        
        if TEST_MODE:
            logger.info(f"[TEST MODE] Would suspend user {user_id}")
            return True
        
        user_data['active'] = False
        
        response = requests.put(
            f"{BI_SCIM_USERS_URL}/{user_id}",
            headers=headers,
            json=user_data
        )
        
        if response.status_code == 200:
            logger.info(f"User {user_id} suspended successfully")
            return True
        else:
            logger.error(f"Failed to suspend user: {response.text}")
            return False
    except Exception as e:
        logger.error(f"Error suspending user: {str(e)}")
        return False

def create_or_get_byid_enrolled_group(service):
    """Create or get the BYID_Enrolled group in Google Workspace"""
    group_email = f"BYID_Enrolled@{GWS_DOMAIN_NAME}"
    
    if TEST_MODE:
        logger.info(f"[TEST MODE] Would create/get BYID_Enrolled group: {group_email}")
        return group_email
    
    try:
        # Check if group exists
        group = service.groups().get(groupKey=group_email).execute()
        logger.info(f"BYID_Enrolled group already exists: {group_email}")
        return group_email
    except Exception as e:
        if 'notFound' in str(e):
            # Create the group if it doesn't exist
            group_body = {
                'email': group_email,
                'name': 'BYID_Enrolled',
                'description': 'Users enrolled in Beyond Identity'
            }
            try:
                group = service.groups().insert(body=group_body).execute()
                logger.info(f"Created BYID_Enrolled group: {group_email}")
                return group_email
            except Exception as create_error:
                logger.error(f"Failed to create BYID_Enrolled group: {str(create_error)}")
                return None
        else:
            logger.error(f"Error checking BYID_Enrolled group: {str(e)}")
            return None

def get_bi_enrollment_status(bi_user_id):
    """Check if a user has any active passkeys using the native API"""
    try:
        # Query the native API endpoint
        response = requests.get(
            f"https://api.byndid.com/v2/users/{bi_user_id}",
            headers={
                "Authorization": f"Bearer {BI_TENANT_API_TOKEN}",
                "Content-Type": "application/json"
            }
        )
        
        if response.status_code == 200:
            user = response.json()
            has_active_passkey = user.get("has_active_passkey", False)
            if TEST_MODE:
                logger.info(f"[TEST MODE] User {user.get('username', 'unknown')} {'has' if has_active_passkey else 'does not have'} active passkeys")
            return has_active_passkey
        else:
            logger.error(f"Error checking enrollment status for user {bi_user_id}: {response.text}")
            return False
    except Exception as e:
        logger.error(f"Error checking enrollment status for user {bi_user_id}: {str(e)}")
        return False

def update_byid_enrolled_group(service, group_email, user_email, should_be_member):
    """Add or remove a user from the BYID_Enrolled group"""
    try:
        if TEST_MODE:
            action = "add" if should_be_member else "remove"
            logger.info(f"[TEST MODE] Would {action} {user_email} to/from BYID_Enrolled group")
            return
        
        if should_be_member:
            # Add user to group
            member_body = {
                'email': user_email,
                'role': 'MEMBER'
            }
            service.members().insert(
                groupKey=group_email,
                body=member_body
            ).execute()
            logger.info(f"Added {user_email} to BYID_Enrolled group")
        else:
            # Remove user from group
            service.members().delete(
                groupKey=group_email,
                memberKey=user_email
            ).execute()
            logger.info(f"Removed {user_email} from BYID_Enrolled group")
    except Exception as e:
        if 'memberKey' in str(e) and 'notFound' in str(e):
            # User is not in group, which is fine if we're trying to remove them
            if not should_be_member:
                logger.info(f"User {user_email} not in BYID_Enrolled group (as expected)")
            else:
                logger.error(f"Error adding {user_email} to BYID_Enrolled group: {str(e)}")
        else:
            logger.error(f"Error updating BYID_Enrolled group for {user_email}: {str(e)}")

def get_detailed_user_info(user_email):
    """Get detailed user information from Beyond Identity"""
    try:
        # First, find the user by email
        params = {'filter': f'userName eq "{user_email}"'}
        response = requests.get(BI_SCIM_USERS_URL, headers=headers, params=params)
        
        if response.status_code == 200:
            users = response.json().get('Resources', [])
            if users:
                user = users[0]
                if TEST_MODE:
                    logger.info(f"[TEST MODE] Found user {user_email} with ID {user.get('id')}")
                    return True
                
                logger.info("\nDetailed User Information:")
                logger.info(f"User ID: {user.get('id')}")
                logger.info(f"Username: {user.get('userName')}")
                logger.info(f"Active: {user.get('active')}")
                logger.info(f"External ID: {user.get('externalId')}")
                
                # Print name information
                name = user.get('name', {})
                logger.info("\nName Information:")
                logger.info(f"Given Name: {name.get('givenName')}")
                logger.info(f"Family Name: {name.get('familyName')}")
                logger.info(f"Display Name: {name.get('displayName')}")
                
                # Print email information
                emails = user.get('emails', [])
                logger.info("\nEmail Information:")
                for email in emails:
                    logger.info(f"Email: {email.get('value')} (Primary: {email.get('primary', False)})")
                
                # Print passkey information
                passkeys = user.get('passkeys', [])
                logger.info("\nPasskey Information:")
                if passkeys:
                    for i, passkey in enumerate(passkeys, 1):
                        logger.info(f"\nPasskey {i}:")
                        logger.info(f"  ID: {passkey.get('id')}")
                        logger.info(f"  Active: {passkey.get('active')}")
                        logger.info(f"  Created: {passkey.get('created')}")
                        logger.info(f"  Last Used: {passkey.get('lastUsed')}")
                else:
                    logger.info("No passkeys found")
                
                # Print group memberships
                groups = user.get('groups', [])
                logger.info("\nGroup Memberships:")
                if groups:
                    for group in groups:
                        logger.info(f"Group ID: {group.get('value')}")
                        logger.info(f"Display: {group.get('display')}")
                else:
                    logger.info("No group memberships found")
                
                return True
            else:
                logger.info(f"No user found with email {user_email}")
                return False
        else:
            logger.error(f"Error querying user: {response.text}")
            return False
    except Exception as e:
        logger.error(f"Error getting user information: {str(e)}")
        return False

def main():
    logger.info("Starting main sync process")
    
    # Initialize Google Workspace service
    service = get_google_service()
    if not TEST_MODE:
        logger.info("Successfully initialized Google Workspace service")
    
    # Create or get the BYID_Enrolled group
    byid_enrolled_group = create_or_get_byid_enrolled_group(service)
    if not byid_enrolled_group:
        logger.error("Failed to create/get BYID_Enrolled group, exiting")
        return
    
    # Get detailed information for Will May
    logger.info("\nQuerying Will May's user information...")
    get_detailed_user_info("will.may@beyondidentity.dev")
    
    # Get all BI users first to track existing SCIM users
    if not TEST_MODE:
        response = requests.get(BI_SCIM_USERS_URL, headers=headers)
        if response.status_code != 200:
            logger.error(f"Error getting BI users: {response.text}")
            return
        
        bi_users = response.json().get('Resources', [])
    else:
        bi_users = []
        logger.info("[TEST MODE] Would get all BI users")
    
    logger.info(f"Found {len(bi_users)} total users in Beyond Identity")
    
    # Track all users and their group memberships
    all_users = {}  # user_id -> {groups: set(), bi_id: str, was_in_group: bool, is_scim_user: bool}
    
    # Initialize all_users with existing BI users
    for bi_user in bi_users:
        external_id = bi_user.get('externalId')
        if external_id:
            # Check if this is a SCIM-sourced user by looking for a Google Workspace ID format
            # Google Workspace IDs are numeric and typically 21 digits long
            is_scim_user = external_id.isdigit() and len(external_id) == 21
            if is_scim_user:
                all_users[external_id] = {
                    'groups': set(),
                    'bi_id': bi_user['id'],
                    'was_in_group': False,  # Will be set to True if found in current groups
                    'is_scim_user': True,
                    'username': bi_user['userName']
                }
            else:
                logger.info(f"Skipping non-SCIM user {bi_user['userName']} during initialization")
    
    # Process each group
    for group_email in GWS_GROUPS:
        logger.info(f"\nProcessing group: {group_email}")
        
        # Create or get BI group
        bi_group_id = create_bi_group(group_email)
        if not bi_group_id:
            logger.error(f"Failed to get/create BI group for {group_email}, skipping")
            continue
        
        # Get group members
        members = get_group_members(service, group_email)
        logger.info(f"Found {len(members)} members in group {group_email}")
        
        # Process each member
        for member in members:
            user_id = member['id']
            
            # Initialize user tracking if needed
            if user_id not in all_users:
                all_users[user_id] = {
                    'groups': set(),
                    'bi_id': None,
                    'was_in_group': True,
                    'is_scim_user': False  # Will be set to True when we create/update in BI
                }
            
            # Add group to user's tracked groups and mark as currently in group
            all_users[user_id]['groups'].add(group_email)
            all_users[user_id]['was_in_group'] = True
            
            # Get user details and create/update in BI
            user_data = get_user_details(service, user_id)
            if not user_data:
                continue
            
            bi_user_id = create_or_update_bi_user(user_data)
            if bi_user_id:
                all_users[user_id]['bi_id'] = bi_user_id
                all_users[user_id]['is_scim_user'] = True  # Mark as SCIM user
                all_users[user_id]['username'] = user_data['primaryEmail']
                add_user_to_group(bi_user_id, bi_group_id)
                
                # Check enrollment status and update BYID_Enrolled group
                is_enrolled = get_bi_enrollment_status(bi_user_id)
                update_byid_enrolled_group(
                    service,
                    byid_enrolled_group,
                    user_data['primaryEmail'],
                    is_enrolled
                )
    
    # Process users who are no longer in any groups
    for user_id, user_info in all_users.items():
        if not user_info['is_scim_user']:
            logger.info(f"\nSkipping non-SCIM user {user_info.get('username', user_id)}")
            continue
            
        if not user_info['was_in_group']:
            logger.info(f"\nUser {user_info['username']} is no longer in any tracked groups")
            
            # Remove from all BI groups
            for bi_user in bi_users:
                if bi_user['id'] == user_info['bi_id']:
                    for group in bi_user.get('groups', []):
                        if group.get('display', '').startswith(BI_GROUP_PREFIX):
                            logger.info(f"Removing user from BI group {group.get('display')}")
                            remove_user_from_group(user_info['bi_id'], group['value'])
            
            # Suspend the user
            suspend_user(user_info['bi_id'])
            
            # Remove from BYID_Enrolled group
            update_byid_enrolled_group(
                service,
                byid_enrolled_group,
                user_info['username'],
                False
            )
        else:
            logger.info(f"\nUser {user_info['username']} is still in tracked groups, checking enrollment")
            # Check enrollment status and update BYID_Enrolled group
            is_enrolled = get_bi_enrollment_status(user_info['bi_id'])
            update_byid_enrolled_group(
                service,
                byid_enrolled_group,
                user_info['username'],
                is_enrolled
            )
    
    logger.info("Sync process completed successfully")

if __name__ == "__main__":
    main() 