#!/bin/bash

# Base URL
BASE_URL="http://localhost:8008"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Room API Test Script"
echo "==================="

# 1. Register a new user
echo -e "\n${GREEN}1. Registering new user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser123",
    "password": "123456"
  }')
echo "Register Response: $REGISTER_RESPONSE"

# 2. Login to get JWT token
echo -e "\n${GREEN}2. Logging in...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser123",
    "password": "123456"
  }')
echo "Login Response: $LOGIN_RESPONSE"

# Extract JWT token from login response
JWT_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}Failed to get JWT token${NC}"
    exit 1
fi

# 3. Create a room (Simple version)
echo -e "\n${GREEN}3. Creating room (Simple version)...${NC}"
CREATE_ROOM_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/rooms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "hms_room_id": "",
    "topics": []
  }')
echo "Create Room Response: $CREATE_ROOM_RESPONSE"

# 4. Create a room (Detailed version)
echo -e "\n${GREEN}4. Creating room (Detailed version)...${NC}"
CREATE_ROOM_DETAILED_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/rooms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "hms_room_id": "",
    "topics": [],
    "queue": {
      "id": 0,
      "name": "",
      "description": "",
      "is_closed": false,
      "participants": []
    }
  }')
echo "Create Room Detailed Response: $CREATE_ROOM_DETAILED_RESPONSE"

echo -e "\n${GREEN}Script completed${NC}" 