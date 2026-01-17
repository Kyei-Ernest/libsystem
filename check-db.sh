#!/bin/bash
# Direct database check script

echo "=== Checking Database State ==="
echo ""

echo "1. List all users:"
sudo -u postgres psql -d libsystem -t -c "SELECT id, email, username FROM users ORDER BY created_at DESC LIMIT 5;"

echo ""
echo "2. Count users:"
sudo -u postgres psql -d libsystem -t -c "SELECT COUNT(*) FROM users;"

echo ""
echo "3. Check foreign key constraint on collections:"
sudo -u postgres psql -d libsystem -c "\d collections" | grep -A 5 "Foreign-key"

echo ""
echo "4. Check if demo user exists:"
sudo -u postgres psql -d libsystem -t -c "SELECT id, email FROM users WHERE email = 'demo@example.com';"

echo ""
echo "5. Try to select from both tables:"
sudo -u postgres psql -d libsystem -t -c "SELECT u.id as user_id, u.email, c.id as collection_id, c.name FROM users u LEFT JOIN collections c ON u.id = c.owner_id WHERE u.email = 'demo@example.com';"
