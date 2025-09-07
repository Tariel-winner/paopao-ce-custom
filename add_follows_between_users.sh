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

# Function to get why ID for a user
get_id() {
    local username=$1
    grep "^$USERNAME:" "$IDS_FILE" | cut -d':' -f2
}

echo "=== Setting up mutual follows between all users ==="

# First, get all existing users and their tokens
echo "Getting all existing users and their tokens..."

# Get all users from database (including all active users)
echo "Fetching user list from database..."
USER_LIST=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT id, username 
FROM p_user 
WHERE is_del = 0 
ORDER BY username;
")

echo "Found users:"
echo "$USER_LIST"

# Login to get tokens for all users
echo -e "\n=== Getting tokens for all users ==="
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
            echo "$USERNAME:$USER_ID" >> "$IDS_FILE"
            echo "  ✅ Token obtained for $USERNAME"
        else
            echo "  ❌ Failed to get token for $USERNAME"
        fi
    fi
done <<< "$USER_LIST"

# Count total users
TOTAL_USERS=$(wc -l < "$TOKENS_FILE")
echo -e "\n=== Total users with valid tokens: $TOTAL_USERS ==="

# Make all testusers follow the "aw" user
echo -e "\n=== Making all testusers follow the 'aw' user ==="

# Find the "aw" user (awwwwww) - get ID directly from database since login fails
AW_USER="awwwwww"
AW_USER_ID=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT id FROM p_user WHERE username = '$AW_USER' AND is_del = 0;
" | tr -d ' ')

if [ ! -z "$AW_USER_ID" ]; then
    echo "Target user: $AW_USER (ID: $AW_USER_ID)"
    
    # Create follows matrix
    FOLLOW_COUNT=0
    TOTAL_FOLLOWS=$((TOTAL_USERS - 1))  # All users except 'aw' will follow 'aw'
    
    echo "Total follows to create: $TOTAL_FOLLOWS"
    
    # Read tokens file to get all usernames
    USERNAMES=()
    while IFS=':' read -r username token; do
        USERNAMES+=("$username")
    done < "$TOKENS_FILE"
    
    for ((i=0; i<${#USERNAMES[@]}; i++)); do
        CURRENT_USER="${USERNAMES[$i]}"
        CURRENT_TOKEN=$(get_token "$CURRENT_USER")
        
        # Skip if this is the 'aw' user (don't follow yourself)
        if [ "$CURRENT_USER" != "$AW_USER" ] && [ ! -z "$CURRENT_TOKEN" ]; then
            echo "User $CURRENT_USER is following $AW_USER..."
            
            FOLLOW_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/user/follow" \
                -H "Authorization: Bearer $CURRENT_TOKEN" \
                -H "Content-Type: application/json" \
                -d "{
                    \"user_id\": $AW_USER_ID
                }")
            
            # Check if follow was successful
            if echo "$FOLLOW_RESPONSE" | jq -e '.code == 0' > /dev/null 2>&1; then
                echo "  ✅ Follow successful: $CURRENT_USER → $AW_USER"
                FOLLOW_COUNT=$((FOLLOW_COUNT + 1))
            else
                echo "  ❌ Follow failed: $FOLLOW_RESPONSE"
            fi
            
            # Add small delay to avoid overwhelming the server
            sleep 0.2
        fi
    done
else
    echo "❌ Could not find user $AW_USER in the database"
    exit 1
fi

# Verify the follows were created
echo -e "\n=== Verifying follows were created ==="

# Check follow counts for the "aw" user
echo "Checking follow counts for $AW_USER..."
AW_USER_ID=$(grep "^$AW_USER:" "$IDS_FILE" | cut -d':' -f2)
if [ ! -z "$AW_USER_ID" ]; then
    FOLLOWERS_COUNT=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT COUNT(*) 
FROM p_following 
WHERE follow_id = $AW_USER_ID AND is_del = 0;
" | tr -d ' ')
    
    echo "  $AW_USER: followed by $FOLLOWERS_COUNT users"
fi

# Check a few sample users to see if they're following the "aw" user
echo "Checking if sample users are following $AW_USER..."
for username in "${USERNAMES[@]:0:5}"; do  # Check first 5 users
    if [ "$username" != "$AW_USER" ]; then
        USER_ID=$(grep "^$username:" "$IDS_FILE" | cut -d':' -f2)
        if [ ! -z "$USER_ID" ]; then
            IS_FOLLOWING=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT COUNT(*) 
FROM p_following 
WHERE user_id = $USER_ID AND follow_id = $AW_USER_ID AND is_del = 0;
" | tr -d ' ')
            
            if [ "$IS_FOLLOWING" -gt 0 ]; then
                echo "  ✅ $username is following $AW_USER"
            else
                echo "  ❌ $username is NOT following $AW_USER"
            fi
        fi
    fi
done

# Check total follows in database
TOTAL_FOLLOWS_IN_DB=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT COUNT(*) 
FROM p_following 
WHERE is_del = 0;
" | tr -d ' ')

echo -e "\n=== Summary ==="
echo "Total follows created: $FOLLOW_COUNT"
echo "Total follows in database: $TOTAL_FOLLOWS_IN_DB"
echo "Expected follows: $TOTAL_FOLLOWS"

# Test a user profile to see follow counts
if [ ${#USERNAMES[@]} -gt 0 ]; then
    TEST_USER="${USERNAMES[0]}"
    echo -e "\n=== Testing profile for $TEST_USER ==="
    PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/user/profile?username=$TEST_USER" \
        -H "Content-Type: application/json")
    echo "Profile response: $PROFILE_RESPONSE"
fi

# Clean up temporary files
rm "$TOKENS_FILE" "$IDS_FILE"

echo -e "\n=== Script Complete ==="
echo "All testusers should now be following $AW_USER!"
echo "Check the database to verify the follows were created correctly."
