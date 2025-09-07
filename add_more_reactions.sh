#!/bin/bash

# Base variables
BASE_URL="http://localhost:8008"

# Create temporary files to store user data
TOKENS_FILE=$(mktemp)
IDS_FILE=$(mktemp)

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

# Check current reaction counts
echo -e "\n=== Current Reaction Counts ==="
CURRENT_COUNTS=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -c "
SELECT ur.reaction_type_id, r.name as reaction_name, r.icon as reaction_icon, COUNT(*) as count
FROM p_user_reactions ur
JOIN p_reactions r ON ur.reaction_type_id = r.id
WHERE ur.target_user_id = $SDSDS_ID 
  AND ur.is_del = 0
GROUP BY ur.reaction_type_id, r.name, r.icon
ORDER BY ur.reaction_type_id;
")
echo "$CURRENT_COUNTS"

# Get users who currently have love reactions (we'll convert some to like)
echo -e "\n=== Getting users with love reactions to convert ==="
LOVE_USERS=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT ur.reactor_user_id, u.username
FROM p_user_reactions ur
JOIN p_user u ON ur.reactor_user_id = u.id
WHERE ur.target_user_id = $SDSDS_ID 
  AND ur.reaction_type_id = 2 
  AND ur.is_del = 0
ORDER BY ur.created_on
LIMIT 49;
")

echo "Users to convert from love to like:"
echo "$LOVE_USERS"

# Login to get tokens for users we need to convert
echo -e "\n=== Getting tokens for users to convert ==="
while read -r line; do
    if [ ! -z "$line" ] && [[ "$line" =~ [0-9]+[[:space:]]+\|[[:space:]]+[a-zA-Z0-9]+ ]]; then
        USER_ID=$(echo "$line" | sed 's/^[[:space:]]*\([0-9]*\)[[:space:]]*|[[:space:]]*\([a-zA-Z0-9]*\).*/\1/')
        USERNAME=$(echo "$line" | sed 's/^[[:space:]]*\([0-9]*\)[[:space:]]*|[[:space:]]*\([a-zA-Z0-9]*\).*/\2/')
        
        echo "Getting token for $USERNAME (ID: $USER_ID)..."
        
        LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/login" \
            -H "Content-Type: application/json" \
            -d "{
                \"username\": \"$USERNAME\",
                \"password\": \"password123\"
            }")
        
        TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
        if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
            echo "$USERNAME:$TOKEN" >> "$TOKENS_FILE"
            echo "  Token obtained for $USERNAME"
        else
            echo "  Failed to get token for $USERNAME"
        fi
    fi
done <<< "$LOVE_USERS"

# Convert love reactions to like reactions
echo -e "\n=== Converting 49 love reactions to like reactions ==="
while read -r line; do
    if [ ! -z "$line" ] && [[ "$line" =~ [0-9]+[[:space:]]+\|[[:space:]]+[a-zA-Z0-9]+ ]]; then
        USER_ID=$(echo "$line" | sed 's/^[[:space:]]*\([0-9]*\)[[:space:]]*|[[:space:]]*\([a-zA-Z0-9]*\).*/\1/')
        USERNAME=$(echo "$line" | sed 's/^[[:space:]]*\([0-9]*\)[[:space:]]*|[[:space:]]*\([a-zA-Z0-9]*\).*/\2/')
        
        CURRENT_TOKEN=$(get_token "$USERNAME")
        
        if [ ! -z "$CURRENT_TOKEN" ]; then
            echo "Converting $USERNAME from love to like..."
            
            # Update reaction from love (2) to like (1)
            REACTION_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/user/reaction" \
                -H "Authorization: Bearer $CURRENT_TOKEN" \
                -H "Content-Type: application/json" \
                -d "{
                    \"target_user_id\": $SDSDS_ID,
                    \"reaction_type_id\": 1
                }")
            
            echo "  Reaction response: $REACTION_RESPONSE"
            sleep 0.5
        else
            echo "  No token available for $USERNAME, skipping..."
        fi
    fi
done <<< "$LOVE_USERS"

# Check reaction counts after conversion
echo -e "\n=== Reaction Counts After Conversion ==="
AFTER_CONVERSION_COUNTS=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -c "
SELECT ur.reaction_type_id, r.name as reaction_name, r.icon as reaction_icon, COUNT(*) as count
FROM p_user_reactions ur
JOIN p_reactions r ON ur.reaction_type_id = r.id
WHERE ur.target_user_id = $SDSDS_ID 
  AND ur.is_del = 0
GROUP BY ur.reaction_type_id, r.name, r.icon
ORDER BY ur.reaction_type_id;
")
echo "$AFTER_CONVERSION_COUNTS"

# Test the profile to see reaction counts
echo -e "\n=== Profile Response ==="
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/user/profile?username=sdsds" \
    -H "Content-Type: application/json")
echo "Profile response: $PROFILE_RESPONSE"

# Clean up temporary files
rm "$TOKENS_FILE" "$IDS_FILE"

echo -e "\n=== Fix Complete ==="
echo "Converted 49 love reactions to like reactions."
echo "Expected final counts: ~60 like reactions, ~69 love reactions" 