Feature: Authentication and Token Management (mvp-userflow)

  Background:
    Given the PostgreSQL database is running
    And the "users" table exists
    And the "refresh_tokens" table exists
    And the API server is running on "/api/v1"

  # ─────────────────────────────────────────────────────────────
  # Scenario: POST /auth/register
  # ─────────────────────────────────────────────────────────────

  Scenario: Register a new user successfully
    Given no user exists with email "newuser@example.com"
    When I send a POST request to "/auth/register" with body:
      """
      {
        "name": "Bruno Costa",
        "email": "newuser@example.com",
        "password": "s3cr3tP@ssw0rd"
      }
      """
    Then the response status should be 201
    And the response should have a JSON body:
      """
      {
        "data": {
          "accessToken": "<JWT_TOKEN>",
          "refreshToken": "<OPAQUE_TOKEN>",
          "expiresIn": 3600
        }
      }
      """
    And the "accessToken" should be a valid JWT with claim "sub" containing a valid UUID
    And the "refreshToken" should be a non-empty string
    And a user should exist in the database with email "newuser@example.com"
    And the user's password should be hashed with bcrypt
    And the user's profile_image_url should be "/assets/avatars/default.png"
    And a refresh token should exist in the database for the user
    And the refresh token should be stored as SHA-256 hash
    And the refresh token's "revoked_at" should be NULL

  Scenario: Attempt to register with an email that already exists
    Given a user exists with email "existing@example.com" and password "OldPassword123"
    When I send a POST request to "/auth/register" with body:
      """
      {
        "name": "New User",
        "email": "existing@example.com",
        "password": "NewPassword456"
      }
      """
    Then the response status should be 409
    And the response should have a JSON body:
      """
      {
        "code": "EMAIL_ALREADY_EXISTS",
        "message": "An account with this email already exists."
      }
      """
    And no new user should be created in the database

  Scenario: Attempt to register with missing required fields
    When I send a POST request to "/auth/register" with body:
      """
      {
        "email": "incomplete@example.com"
      }
      """
    Then the response status should be 422
    And the response should have a JSON body with "code" equal to "VALIDATION_ERROR"
    And the response should contain validation details for field "name"
    And the response should contain validation details for field "password"

  Scenario: Attempt to register with invalid email format
    When I send a POST request to "/auth/register" with body:
      """
      {
        "name": "Invalid Email",
        "email": "not-an-email",
        "password": "ValidPassword123"
      }
      """
    Then the response status should be 422
    And the response should have a JSON body with "code" equal to "VALIDATION_ERROR"
    And the response should contain validation details for field "email"

  Scenario: Attempt to register with password shorter than 8 characters
    When I send a POST request to "/auth/register" with body:
      """
      {
        "name": "Short Password User",
        "email": "shortpw@example.com",
        "password": "short"
      }
      """
    Then the response status should be 422
    And the response should have a JSON body with "code" equal to "VALIDATION_ERROR"
    And the response should contain validation details for field "password"

  # ─────────────────────────────────────────────────────────────
  # Scenario: POST /auth/login
  # ─────────────────────────────────────────────────────────────

  Scenario: Login with valid credentials
    Given a user exists with email "user@example.com" and password "MyPassword123"
    When I send a POST request to "/auth/login" with body:
      """
      {
        "email": "user@example.com",
        "password": "MyPassword123"
      }
      """
    Then the response status should be 200
    And the response should have a JSON body:
      """
      {
        "data": {
          "accessToken": "<JWT_TOKEN>",
          "refreshToken": "<OPAQUE_TOKEN>",
          "expiresIn": 3600
        }
      }
      """
    And the "accessToken" should be a valid JWT with claim "sub" equal to the user's ID
    And the "refreshToken" should be a non-empty string
    And a new refresh token should exist in the database for the user
    And the refresh token should be stored as SHA-256 hash

  Scenario: Attempt to login with incorrect password
    Given a user exists with email "user@example.com" and password "CorrectPassword"
    When I send a POST request to "/auth/login" with body:
      """
      {
        "email": "user@example.com",
        "password": "WrongPassword"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "INVALID_CREDENTIALS",
        "message": "Email or password is incorrect."
      }
      """
    And no new refresh token should be created

  Scenario: Attempt to login with non-existent email
    Given no user exists with email "nonexistent@example.com"
    When I send a POST request to "/auth/login" with body:
      """
      {
        "email": "nonexistent@example.com",
        "password": "SomePassword123"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "INVALID_CREDENTIALS",
        "message": "Email or password is incorrect."
      }
      """

  Scenario: Attempt to login with missing fields
    When I send a POST request to "/auth/login" with body:
      """
      {
        "email": "user@example.com"
      }
      """
    Then the response status should be 422
    And the response should have a JSON body with "code" equal to "VALIDATION_ERROR"
    And the response should contain validation details for field "password"

  # ─────────────────────────────────────────────────────────────
  # Scenario: POST /auth/refresh
  # ─────────────────────────────────────────────────────────────

  Scenario: Refresh access token with valid refresh token
    Given a user exists with ID "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And the user has a valid refresh token "valid-refresh-token-xyz" that expires in 30 days
    When I send a POST request to "/auth/refresh" with body:
      """
      {
        "refreshToken": "valid-refresh-token-xyz"
      }
      """
    Then the response status should be 200
    And the response should have a JSON body:
      """
      {
        "data": {
          "accessToken": "<NEW_JWT_TOKEN>",
          "refreshToken": "<NEW_OPAQUE_TOKEN>",
          "expiresIn": 3600
        }
      }
      """
    And the "accessToken" should be a valid JWT with claim "sub" equal to "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And the old refresh token "valid-refresh-token-xyz" should be revoked in the database
    And a new refresh token should exist in the database for the user
    And the new refresh token should not equal "valid-refresh-token-xyz"

  Scenario: Attempt to refresh with an expired refresh token
    Given a user exists with ID "b2c3d4e5-f6a7-8901-bcde-f12345678901"
    And the user has an expired refresh token "expired-token-abc"
    When I send a POST request to "/auth/refresh" with body:
      """
      {
        "refreshToken": "expired-token-abc"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """
    And no new refresh token should be created

  Scenario: Attempt to refresh with a revoked refresh token
    Given a user exists with ID "c3d4e5f6-a7b8-9012-cdef-123456789012"
    And the user has a revoked refresh token "revoked-token-def"
    When I send a POST request to "/auth/refresh" with body:
      """
      {
        "refreshToken": "revoked-token-def"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """

  Scenario: Attempt to refresh with an invalid (non-existent) refresh token
    When I send a POST request to "/auth/refresh" with body:
      """
      {
        "refreshToken": "non-existent-token-ghi"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """

  Scenario: Attempt to refresh with missing refresh token field
    When I send a POST request to "/auth/refresh" with body:
      """
      {}
      """
    Then the response status should be 422
    And the response should have a JSON body with "code" equal to "VALIDATION_ERROR"
    And the response should contain validation details for field "refreshToken"

  # ─────────────────────────────────────────────────────────────
  # Scenario: POST /auth/logout
  # ─────────────────────────────────────────────────────────────

  Scenario: Logout successfully with valid tokens
    Given a user exists with ID "d4e5f6a7-b8c9-0123-defa-234567890123"
    And the user has a valid refresh token "logout-refresh-token-jkl"
    And I have a valid JWT access token for user ID "d4e5f6a7-b8c9-0123-defa-234567890123"
    When I send a POST request to "/auth/logout" with:
      | Header         | Value                       |
      | Authorization  | Bearer <VALID_JWT_TOKEN>    |
    And body:
      """
      {
        "refreshToken": "logout-refresh-token-jkl"
      }
      """
    Then the response status should be 204
    And the response body should be empty
    And the refresh token "logout-refresh-token-jkl" should be revoked in the database

  Scenario: Logout with already revoked refresh token (idempotent)
    Given a user exists with ID "e5f6a7b8-c9d0-1234-efab-345678901234"
    And the user has a revoked refresh token "already-revoked-token-mno"
    And I have a valid JWT access token for user ID "e5f6a7b8-c9d0-1234-efab-345678901234"
    When I send a POST request to "/auth/logout" with:
      | Header         | Value                    |
      | Authorization  | Bearer <VALID_JWT_TOKEN> |
    And body:
      """
      {
        "refreshToken": "already-revoked-token-mno"
      }
      """
    Then the response status should be 204
    And the response body should be empty

  Scenario: Attempt to logout without Authorization header
    When I send a POST request to "/auth/logout" with body:
      """
      {
        "refreshToken": "some-token-pqr"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """

  Scenario: Attempt to logout with invalid JWT token
    When I send a POST request to "/auth/logout" with:
      | Header         | Value                 |
      | Authorization  | Bearer invalid-token  |
    And body:
      """
      {
        "refreshToken": "some-token-stu"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """

  Scenario: Attempt to logout with expired JWT token
    Given a user exists with ID "f6a7b8c9-d0e1-2345-fabc-456789012345"
    And I have an expired JWT access token for user ID "f6a7b8c9-d0e1-2345-fabc-456789012345"
    When I send a POST request to "/auth/logout" with:
      | Header         | Value                    |
      | Authorization  | Bearer <EXPIRED_JWT>     |
    And body:
      """
      {
        "refreshToken": "some-token-vwx"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """

  Scenario: Attempt to logout with missing refresh token field
    Given a user exists with ID "a7b8c9d0-e1f2-3456-bcde-567890123456"
    And I have a valid JWT access token for user ID "a7b8c9d0-e1f2-3456-bcde-567890123456"
    When I send a POST request to "/auth/logout" with:
      | Header         | Value                    |
      | Authorization  | Bearer <VALID_JWT_TOKEN> |
    And body:
      """
      {}
      """
    Then the response status should be 422
    And the response should have a JSON body with "code" equal to "VALIDATION_ERROR"
    And the response should contain validation details for field "refreshToken"

  # ─────────────────────────────────────────────────────────────
  # Edge Cases & Security
  # ─────────────────────────────────────────────────────────────

  Scenario: Refresh token rotation prevents token reuse
    Given a user exists with ID "b8c9d0e1-f2a3-4567-cdef-678901234567"
    And the user has a valid refresh token "rotation-test-token-abc"
    When I send a POST request to "/auth/refresh" with body:
      """
      {
        "refreshToken": "rotation-test-token-abc"
      }
      """
    Then the response status should be 200
    And a new refresh token should be returned
    And the old token "rotation-test-token-abc" should be revoked
    When I send a POST request to "/auth/refresh" with body:
      """
      {
        "refreshToken": "rotation-test-token-abc"
      }
      """
    Then the response status should be 401
    And the response should have a JSON body with "code" equal to "UNAUTHORIZED"

  Scenario: Password is never logged in plain text
    Given logging is enabled
    When I send a POST request to "/auth/register" with body:
      """
      {
        "name": "Security Test User",
        "email": "sectest@example.com",
        "password": "SuperSecretPassword123"
      }
      """
    Then the application logs should NOT contain "SuperSecretPassword123"
    And the application logs should NOT contain the plain text password

  Scenario: Refresh token is stored as hash, not plain text
    Given a user exists with email "hashtest@example.com" and password "Password123"
    When I send a POST request to "/auth/login" with body:
      """
      {
        "email": "hashtest@example.com",
        "password": "Password123"
      }
      """
    Then the response status should be 200
    And I capture the "refreshToken" from the response
    And the database should NOT contain the plain text refresh token
    And the database should contain a SHA-256 hash of the refresh token

  Scenario: JWT contains correct claims
    Given a user exists with ID "c9d0e1f2-a3b4-5678-def0-789012345678" and email "jwttest@example.com"
    When I send a POST request to "/auth/login" with body:
      """
      {
        "email": "jwttest@example.com",
        "password": "TestPassword123"
      }
      """
    Then the response status should be 200
    And I decode the "accessToken" JWT
    And the JWT should have claim "sub" equal to "c9d0e1f2-a3b4-5678-def0-789012345678"
    And the JWT should have claim "exp" set to 1 hour from now
    And the JWT should have claim "iat" set to current timestamp
    And the JWT should be signed with algorithm "HS256"
