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

echo "=== Setting up mutual follows between special users only ==="

# Define the 21 special users (non-testusers)
SPECIAL_USERS=(
    "abbbb"
    "adasd"
    "asas"
    "asasdsa"
    "asdasd"
    "asdasdasd"
    "awwwwww"
    "beyonce123"
    "cfsffsdf"
    "dsaddss"
    "dsdsd"
    "gggfgfgg"
    "rffffff"
    "rttete"
    "sads"
    "sadsad"
    "sdfdsf"
    "sdsdd"
    "sdsdsdsds"
    "sdsdsdsss"
    "sdsf"
    "sgdggdhd"
    "shsg"
    "ssgundk"
    "ssssssssds"
    "tamm"
    "tammsdsd"
    "tarel"
    "TTaaae"
    "sdsds"
)

echo "Special users to process: ${#SPECIAL_USERS[@]} users"
echo "Users: ${SPECIAL_USERS[*]}"

# Login to get tokens for all special users
echo -e "\n=== Getting tokens for special users ==="
for username in "${SPECIAL_USERS[@]}"; do
    echo "Getting token for $username..."
    
    LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$username\",
            \"password\": \"password123\"
        }")
    
    TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
    if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
        echo "$username:$TOKEN" >> "$TOKENS_FILE"
        echo "  ✅ Token obtained for $username"
        
        # Get user ID from profile
        USER_ID=$(curl -s -X GET "$BASE_URL/v1/user/profile?username=$username" \
            -H "Content-Type: application/json" | jq -r '.data.id')
        echo "$username:$USER_ID" >> "$IDS_FILE"
        echo "  ✅ User ID: $USER_ID"
    else
        echo "  ❌ Failed to get token for $username"
    fi
done

# Count total users with valid tokens
TOTAL_USERS=$(wc -l < "$TOKENS_FILE")
echo -e "\n=== Total special users with valid tokens: $TOTAL_USERS ==="

# Create follows matrix (each user follows every other special user)
echo -e "\n=== Creating mutual follows between special users ==="

# Read tokens file to get all usernames
USERNAMES=()
while IFS=':' read -r username token; do
    USERNAMES+=("$username")
done < "$TOKENS_FILE"

# Create follows matrix
FOLLOW_COUNT=0
TOTAL_FOLLOWS=$((TOTAL_USERS * (TOTAL_USERS - 1)))  # Each user follows (n-1) others

echo "Total follows to create: $TOTAL_FOLLOWS"

for ((i=0; i<${#USERNAMES[@]}; i++)); do
    CURRENT_USER="${USERNAMES[$i]}"
    CURRENT_TOKEN=$(get_token "$CURRENT_USER")
    
    if [ ! -z "$CURRENT_TOKEN" ]; then
        echo "User $CURRENT_USER is following other special users..."
        
        for ((j=0; j<${#USERNAMES[@]}; j++)); do
            # Skip following yourself
            if [ $i -ne $j ]; then
                TARGET_USER="${USERNAMES[$j]}"
                TARGET_ID=$(grep "^$TARGET_USER:" "$IDS_FILE" | cut -d':' -f2)
                
                if [ ! -z "$TARGET_ID" ]; then
                    echo "  $CURRENT_USER → $TARGET_USER (ID: $TARGET_ID)..."
                    
                    FOLLOW_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/user/follow" \
                        -H "Authorization: Bearer $CURRENT_TOKEN" \
                        -H "Content-Type: application/json" \
                        -d "{
                            \"user_id\": $TARGET_ID
                        }")
                    
                    # Check if follow was successful
                    if echo "$FOLLOW_RESPONSE" | jq -e '.code == 0' > /dev/null 2>&1; then
                        echo "    ✅ Follow successful"
                        FOLLOW_COUNT=$((FOLLOW_COUNT + 1))
                    else
                        echo "    ❌ Follow failed: $FOLLOW_RESPONSE"
                    fi
                    
                    # Add small delay to avoid overwhelming the server
                    sleep 0.2
                fi
            fi
        done
        echo "  Completed follows for $CURRENT_USER"
        echo "  ---"
    fi
done

# Verify the follows were created
echo -e "\n=== Verifying follows were created ==="

# Check follow counts for a few sample users
echo "Checking follow counts for sample special users..."
for username in "${USERNAMES[@]:0:3}"; do  # Check first 3 users
    USER_ID=$(grep "^$username:" "$IDS_FILE" | cut -d':' -f2)
    if [ ! -z "$USER_ID" ]; then
        FOLLOW_COUNT=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT COUNT(*) 
FROM p_following 
WHERE user_id = $USER_ID AND is_del = 0;
" | tr -d ' ')
        
        FOLLOWING_COUNT=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT COUNT(*) 
FROM p_following 
WHERE follow_id = $USER_ID AND is_del = 0;
" | tr -d ' ')
        
        echo "  $username: follows $FOLLOW_COUNT users, followed by $FOLLOWING_COUNT users"
    fi
done

# Check total follows in database for special users
TOTAL_FOLLOWS_IN_DB=$(docker exec paopao-ce-db-1 psql -U paopao -d paopao -t -c "
SELECT COUNT(*) 
FROM p_following f
JOIN p_user u1 ON f.user_id = u1.id
JOIN p_user u2 ON f.follow_id = u2.id
WHERE f.is_del = 0 
  AND u1.username NOT LIKE 'testuser%'
  AND u2.username NOT LIKE 'testuser%';
" | tr -d ' ')

echo -e "\n=== Summary ==="
echo "Total follows created: $FOLLOW_COUNT"
echo "Total follows between special users in database: $TOTAL_FOLLOWS_IN_DB"
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
echo "All special users should now be following each other!"
echo "Check the database to verify the follows were created correctly."
