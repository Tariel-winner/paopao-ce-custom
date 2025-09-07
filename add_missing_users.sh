#!/bin/bash

# Base variables
BASE_URL="http://localhost:8008"
AVATAR_FILE="/Users/taibov/Desktop/Screenshot 2025-04-07 at 02.08.24.png"

# Create temporary files to store user data
TOKENS_FILE=$(mktemp)
IDS_FILE=$(mktemp)

# Function to decode JWT token and extract user ID
decode_jwt() {
    local token=$1
    local payload=$(echo $token | cut -d'.' -f2)
    # Add padding if needed for base64 decoding
    local padded_payload=$(printf '%s' "$payload" | sed 's/-/+/g' | sed 's/_/\//g')
    local mod=$((${#padded_payload} % 4))
    if [ $mod -eq 2 ]; then padded_payload="${padded_payload}=="; fi
    if [ $mod -eq 3 ]; then padded_payload="${padded_payload}="; fi
    local decoded=$(echo $padded_payload | base64 -d 2>/dev/null)
    echo $decoded | jq -r '.uid'
}

# Create users from 50 to 120 (68 users total)
echo "Creating users from 50 to 120..."
for i in {50..120}; do
    echo "Creating user testuser$i..."
    
    # Register user
    REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"testuser$i\",
            \"password\": \"password123\",
            \"nickname\": \"Test User $i\"
        }")
    
    # Login to get token
    echo "Logging in testuser$i..."
    LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"testuser$i\",
            \"password\": \"password123\"
        }")
    
    TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
    echo "Token for testuser$i: $TOKEN"
    echo "testuser$i:$TOKEN" >> "$TOKENS_FILE"

    # Get user ID from profile
    USER_ID=$(curl -s -X GET "$BASE_URL/v1/user/profile?username=testuser$i" \
        -H "Content-Type: application/json" | jq -r '.data.id')
    echo "User ID for testuser$i: $USER_ID"
    echo "testuser$i:$USER_ID" >> "$IDS_FILE"

    # Upload avatar
    echo "Uploading avatar for testuser$i..."
    AVATAR_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/attachment" \
        -H "Authorization: Bearer $TOKEN" \
        -F "type=public/image" \
        -F "file=@$AVATAR_FILE")
    
    echo "Avatar upload response for testuser$i: $AVATAR_RESPONSE"
    echo "----------------------------------------"
    
    # Add a small delay between user creation
    sleep 1
done

# Function to get token for a user
get_token() {
    local username=$1
    grep "^$username:" "$TOKENS_FILE" | cut -d':' -f2
}

# Function to get ID for a user
get_id() {
    local username=$1
    grep "^$username:" "$IDS_FILE" | cut -d':' -f2
}

# Get sdsds user ID
echo "Getting sdsds user ID..."
SDSDS_ID=$(curl -s -X GET "$BASE_URL/v1/user/profile?username=sdsds" \
    -H "Content-Type: application/json" | jq -r '.data.id')
echo "sdsds user ID: $SDSDS_ID"

# Add reactions to sdsds
echo "Adding reactions to sdsds (ID: $SDSDS_ID)..."

# First 60 users give "like" reaction (reaction_type_id = 1)
echo "Adding 'like' reactions from first 60 users..."
for i in {1..60}; do
    CURRENT_USER="testuser$i"
    CURRENT_TOKEN=$(get_token "$CURRENT_USER")
    
    if [ ! -z "$CURRENT_TOKEN" ]; then
        echo "  $CURRENT_USER giving 'like' reaction to sdsds..."
        
        REACTION_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/user/reaction" \
            -H "Authorization: Bearer $CURRENT_TOKEN" \
            -H "Content-Type: application/json" \
            -d "{
                \"target_user_id\": $SDSDS_ID,
                \"reaction_type_id\": 1
            }")
        
        echo "  Reaction response: $REACTION_RESPONSE"
        sleep 0.5
    fi
done

# Next 58 users give "love" reaction (reaction_type_id = 2)
echo "Adding 'love' reactions from next 58 users..."
for i in {61..118}; do
    CURRENT_USER="testuser$i"
    CURRENT_TOKEN=$(get_token "$CURRENT_USER")
    
    if [ ! -z "$CURRENT_TOKEN" ]; then
        echo "  $CURRENT_USER giving 'love' reaction to sdsds..."
        
        REACTION_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/user/reaction" \
            -H "Authorization: Bearer $CURRENT_TOKEN" \
            -H "Content-Type: application/json" \
            -d "{
                \"target_user_id\": $SDSDS_ID,
                \"reaction_type_id\": 2
            }")
        
        echo "  Reaction response: $REACTION_RESPONSE"
        sleep 0.5
    fi
done

# Test the profile to see reaction counts
echo "Testing sdsds profile with reactions..."
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/user/profile?username=sdsds" \
    -H "Content-Type: application/json")
echo "Profile response: $PROFILE_RESPONSE"

# Clean up temporary files
rm "$TOKENS_FILE" "$IDS_FILE"

echo "Done! Added 68 users and reactions to sdsds." 