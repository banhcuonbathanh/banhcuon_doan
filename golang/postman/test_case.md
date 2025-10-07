1. CreateUser Test Cases:

   Password too short ❌
   Password too long ❌
   Password contains user's email ❌
   Email too long (exceeds limit) ❌
   Name contains special characters/numbers ❌
   Empty/whitespace-only name ❌
   Invalid role value ❌
   Invalid branch ID ❌
   Database connection error ❌
   SQL injection attempt ❌
   Request timeout ❌
   Password hashing failure ❌
   Large payload attack ❌
   Concurrent duplicate email creation ❌
   Invalid character encoding ❌
   Email duplication ❌
   Wrong email format ❌
   Weak password ❌
   Missing required fields ❌
   Invalid JSON ❌
   Success case ✅

2. Login Test Cases:

   Invalid email format ❌
   Empty/missing email ❌
   Empty/missing password ❌
   Non-existent user ❌
   Incorrect password ❌
   User account deactivated/disabled ❌
   Database connection error ❌
   Token generation failure ❌
   Invalid JSON payload ❌
   Missing request body ❌
   SQL injection attempt ❌
   Rate limiting exceeded ❌
   Success case with valid credentials ✅
3. Daily Token Cleanup Test Cases Structure

   cases:1. Success Cases ✅

   Successful cleanup with expired tokens removed
   Successful cleanup with no expired tokens
   Successful cleanup with revoked tokens removed
   2. Error Cases ❌

   Database connection error
   Context timeout/cancellation
   Service unavailable
   Invalid request method (should be POST/GET based on your implementation)
   Concurrent cleanup requests
   Nil response from service
   Cleanup service panic/crash
   Transaction rollback failure
   3. Edge Cases ❌

   Large number of tokens to cleanup
   Cleanup during high system load
   Partial cleanup failure