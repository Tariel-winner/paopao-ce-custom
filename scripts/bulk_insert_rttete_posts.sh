#!/bin/bash
set -e

# Function to insert a batch of posts
insert_posts() {
  local start_id=$1
  local end_id=$2
  local start_seq=$3
  local end_seq=$4
  
  echo "Inserting posts ${start_seq}-${end_seq} for user rttete (user_id 45)..."
  docker-compose exec db psql -U paopao -d paopao -c "INSERT INTO p_post (id, user_id, comment_count, collection_count, upvote_count, is_top, is_essence, is_lock, latest_replied_on, tags, attachment_price, ip, ip_loc, created_on, modified_on, deleted_on, is_del, visibility, room_id, share_count, session_id) VALUES
$(for ((i=start_id; i<=end_id; i++)); do
    seq=$((start_seq + i - start_id))
    echo "($i, '[45, 44]', 0, 0, 0, 0, 0, 0, 1750794780, '$seq', 0, '192.168.65.1', '局域网', 1750794780, 1750794780, 0, 0, 90, '685b021873fc384c53e4033a', 0, NULL)"
    if [ $i -ne $end_id ]; then echo ","; fi
  done)
ON CONFLICT (id) DO NOTHING;"
}

# Function to insert post content
insert_post_content() {
  local start_id=$1
  local end_id=$2
  
  echo "Inserting post_content for posts ${start_id}-${end_id}..."
  # No ON CONFLICT here because p_post_content.post_id is not unique in schema
  docker-compose exec db psql -U paopao -d paopao -c "INSERT INTO p_post_content (post_id, user_id, content, type, sort, created_on, modified_on, deleted_on, is_del, duration, size, room_id) VALUES
$(for ((i=start_id; i<=end_id; i++)); do
    echo "($i, '[45, 44]', '45:https://conversations.fc9bb64b8e9130de6c3dd1a617f62a9b.r2.cloudflarestorage.com/spaces/685b021873fc384c53e4033a/20250720/687ce170d0fd99456f64ee53/65290982-EE01-49B1-84F0-9E1DAB134084/42266AB8-B054-4B50-BC33-2ED474CDDAEB/42266AB8-B054-4B50-BC33-2ED474CDDAEB.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=90f510507ebb799c81183619756d6cde%2F20250720%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250720T123728Z&X-Amz-Expires=259200&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=8b1afa348e6af1726c289dbeb1b0c217e4fb6bb3afcddfc2175828320378f41f|44:https://conversations.fc9bb64b8e9130de6c3dd1a617f62a9b.r2.cloudflarestorage.com/spaces/685b021873fc384c53e4033a/20250720/687ce170d0fd99456f64ee53/C2C6CE97-1D5B-4380-B0F1-8343F0ED75FD/627F8CEE-A796-44C1-96C4-A403A96E68ED/627F8CEE-A796-44C1-96C4-A403A96E68ED.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=90f510507ebb799c81183619756d6cde%2F20250720%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250720T123728Z&X-Amz-Expires=259200&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=3660e124fdd650225342987e02f12dafb66bddd6ed8e724d74c9e5e47082c4c2', 5, 0, 1750794780, 1753015077, 0, 0, '51.0', 791459, '685b021873fc384c53e4033a')"
    if [ $i -ne $end_id ]; then echo ","; fi
  done)
;"
}

# Process batches of 10 records each to avoid command length issues
for batch in {0..3}; do
  start_seq=$((26 + batch*10))
  end_seq=$((35 + batch*10))
  
  # Cap at 60
  if [ $end_seq -gt 60 ]; then
    end_seq=60
  fi
  
  start_id=$((1080018093 + batch*10))
  end_id=$((1080018093 + (end_seq - 26)))
  
  insert_posts $start_id $end_id $start_seq $end_seq
  insert_post_content $start_id $end_id
  
  # Break if we've reached 60
  if [ $end_seq -eq 60 ]; then
    break
  fi

done

echo "All posts up to 60 inserted successfully."

echo "Checking for missing post_content rows for user rttete (user_id 45)..."

# Get missing post_ids (those in p_post but not in p_post_content)
missing_ids=$(docker-compose exec db psql -U paopao -d paopao -t -A -c "SELECT id FROM p_post WHERE (user_id->>0)::bigint = 45 AND id NOT IN (SELECT DISTINCT post_id FROM p_post_content WHERE (user_id->>0)::bigint = 45);")

for post_id in $missing_ids; do
  if [[ -n "$post_id" ]]; then
    echo "Inserting missing post_content for post_id $post_id..."
    docker-compose exec db psql -U paopao -d paopao -c "INSERT INTO p_post_content (post_id, user_id, content, type, sort, created_on, modified_on, deleted_on, is_del, duration, size, room_id) VALUES
($post_id, '[45, 44]', '45:https://conversations.fc9bb64b8e9130de6c3dd1a617f62a9b.r2.cloudflarestorage.com/spaces/685b021873fc384c53e4033a/20250720/687ce170d0fd99456f64ee53/65290982-EE01-49B1-84F0-9E1DAB134084/42266AB8-B054-4B50-BC33-2ED474CDDAEB/42266AB8-B054-4B50-BC33-2ED474CDDAEB.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=90f510507ebb799c81183619756d6cde%2F20250720%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250720T123728Z&X-Amz-Expires=259200&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=8b1afa348e6af1726c289dbeb1b0c217e4fb6bb3afcddfc2175828320378f41f|44:https://conversations.fc9bb64b8e9130de6c3dd1a617f62a9b.r2.cloudflarestorage.com/spaces/685b021873fc384c53e4033a/20250720/687ce170d0fd99456f64ee53/C2C6CE97-1D5B-4380-B0F1-8343F0ED75FD/627F8CEE-A796-44C1-96C4-A403A96E68ED/627F8CEE-A796-44C1-96C4-A403A96E68ED.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=90f510507ebb799c81183619756d6cde%2F20250720%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250720T123728Z&X-Amz-Expires=259200&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=3660e124fdd650225342987e02f12dafb66bddd6ed8e724d74c9e5e47082c4c2', 5, 0, 1750794780, 1753015077, 0, 0, '51.0', 791459, '685b021873fc384c53e4033a');"
  fi
done

echo "All missing post_content rows inserted."

# Update created_on and latest_replied_on for posts with tags 17 to 60 for user_id 45
update_post_times() {
  echo "Updating created_on and latest_replied_on for posts with tags 17 to 60 (user_id 45)..."
  # Start from 2025-06-24 22:53:00 (UNIX: 1750794780), increment 1 day per tag
  local base_ts=1750794780
  for tag in $(seq 60 -1 17); do
    docker-compose exec db psql -U paopao -d paopao -c "UPDATE p_post SET created_on = $base_ts, latest_replied_on = $base_ts WHERE (user_id->>0)::bigint = 45 AND tags = '$tag';"
    base_ts=$((base_ts + 86400))
  done
  echo "Timestamps updated."
}

# Call the update function at the end
update_post_times