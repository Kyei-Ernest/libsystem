# Redis Token Blacklisting - Quick Test Guide

## Testing Immediate Logout with Redis

**1. Start Redis locally:**
```bash
redis-server
```

**2. Start user service with Redis:**
```bash
cd ~/Documents/libsystem
./run-services.sh
```

**3. Register and login:**
```bash
# Register
curl -X POST http://localhost:8086/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"Test@1234","first_name":"Test","last_name":"User"}'

# Login (save the token)
TOKEN=$(curl -s -X POST http://localhost:8086/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_username":"test@example.com","password":"Test@1234"}' | jq -r '.data.token')

echo "Token: $TOKEN"
```

**4. Test the token works:**
```bash
curl http://localhost:8086/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

**5. Logout (blacklist the token):**
```bash
curl -X POST http://localhost:8086/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

**6. Try using the token again (should fail):**
```bash
curl http://localhost:8086/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

Expected result: "Token has been revoked" error

## What's Implemented

✅ **Immediate Logout** - Tokens blacklisted in Redis instantly  
✅ **Token Validation** - Checks blacklist before accepting tokens  
✅ **TTL Management** - Blacklisted tokens auto-expire with token expiration  
✅ **Graceful Fallback** - Works without Redis (soft logout only)  
✅ **User Revocation** - Can revoke all tokens for a user (password change, etc.)

## Environment Variables

```bash
# Redis configuration (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=  # leave empty for local dev
```

Without Redis, the service works normally but logout doesn't immediately invalidate tokens - they just expire naturally after 24 hours.
