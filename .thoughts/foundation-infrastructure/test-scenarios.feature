Feature: Foundation & Infrastructure

  As a developer
  I want to have a solid foundation infrastructure
  So that I can build features on top of a reliable and well-structured codebase

  # =========================================================================
  # Docker & Database Infrastructure
  # =========================================================================

  Scenario: Starting the development environment with Docker Compose
    Given the repository has docker-compose.yml configured
    When I run "docker-compose up -d"
    Then the PostgreSQL container should be running
    And the app container should be running
    And the PostgreSQL should accept connections on port 5432
    And the app should be listening on port 8080

  Scenario: PostgreSQL container is healthy
    Given Docker Compose is running
    When I check the PostgreSQL health check
    Then the health check should pass
    And the database should be ready to accept queries

  Scenario: App container waits for PostgreSQL to be ready
    Given Docker Compose is starting
    When the PostgreSQL container is not yet healthy
    Then the app container should wait
    And the app container should only start after PostgreSQL is healthy

  # =========================================================================
  # Database Migrations
  # =========================================================================

  Scenario: Applying all migrations on fresh database
    Given a fresh PostgreSQL database
    And all 7 migration files exist (001 to 007)
    When the migrations are applied
    Then the database should have 7 tables created
    And the tables should be: users, workouts, exercises, sessions, set_records, refresh_tokens, audit_log
    And all indexes should be created
    And no migration errors should occur

  Scenario: Migrations create users table with correct schema
    Given migrations have been applied
    When I query the users table schema
    Then it should have columns: id (UUID), email (VARCHAR UNIQUE), name (VARCHAR), password_hash (VARCHAR), profile_image_url (VARCHAR nullable), created_at (TIMESTAMPTZ), updated_at (TIMESTAMPTZ)
    And it should have an index on email
    And id should be the primary key

  Scenario: Migrations create workouts table with type and intensity fields
    Given migrations have been applied
    When I query the workouts table schema
    Then it should have a foreign key to users(id) with ON DELETE CASCADE
    And it should have columns: type (VARCHAR), intensity (VARCHAR), duration (INT), image_url (VARCHAR)
    And it should have indexes on user_id, type, and (user_id, type)
    And it should NOT have a status ENUM

  Scenario: Migrations create exercises table with workout relationship and JSONB muscles
    Given migrations have been applied
    When I query the exercises table schema
    Then it should have a foreign key to workouts(id) with ON DELETE CASCADE
    And muscles should be JSONB type (array of strings)
    And it should have columns: sets (INT), reps (VARCHAR), rest_time (INT), weight (INT, grams), order_index (INT)
    And it should have a GIN index on muscles
    And it should NOT have exercise_category or muscle_group ENUMs

  Scenario: Migrations create sessions table with correct status values
    Given migrations have been applied
    When I query the sessions table schema
    Then it should have a foreign key to users(id) with ON DELETE CASCADE
    And it should have a foreign key to workouts(id) with ON DELETE RESTRICT
    And status should use CHECK constraint with values: active, completed, abandoned
    And finished_at should be nullable
    And it should have a UNIQUE partial index on (user_id) WHERE status = 'active'

  Scenario: Migrations create set_records table with weight in grams and UNIQUE constraint
    Given migrations have been applied
    When I query the set_records table schema
    Then weight should be INT (grams, not DECIMAL)
    And reps and weight should be NOT NULL with default 0
    And status should use CHECK constraint with values: completed, skipped
    And recorded_at should exist (not created_at)
    And UNIQUE constraint on (session_id, exercise_id, set_number) should exist
    And set_number should have CHECK constraint (>= 1)

  Scenario: Migrations create refresh_tokens table with revoked_at pointer
    Given migrations have been applied
    When I query the refresh_tokens table schema
    Then token should be UNIQUE (column name is "token", not "token_hash")
    And it should have indexes on user_id, token, expires_at
    And revoked_at should be TIMESTAMPTZ nullable (not a BOOLEAN)
    And it should have a partial index on (user_id, revoked_at) WHERE revoked_at IS NULL

  Scenario: Migrations create audit_log table with free-form action and occurred_at
    Given migrations have been applied
    When I query the audit_log table schema
    Then action_data should be JSONB type (not "metadata")
    And it should have a GIN index on action_data
    And user_id should be NOT NULL with ON DELETE RESTRICT
    And action should be VARCHAR (not an ENUM)
    And occurred_at should exist (not created_at)
    And it should have a composite index on (user_id, occurred_at DESC)

  Scenario: Re-running migrations does not cause errors
    Given migrations have already been applied once
    When I restart Docker Compose with the same database volume
    Then the migrations should not be re-applied
    And no duplicate table errors should occur

  Scenario: Recreating database applies migrations from scratch
    Given Docker Compose is running
    When I run "docker-compose down -v" to remove volumes
    And I run "docker-compose up -d" again
    Then all migrations should be applied successfully
    And all 7 tables should be created fresh

  # =========================================================================
  # Database Connection Pool
  # =========================================================================

  Scenario: Application connects to PostgreSQL successfully
    Given PostgreSQL is running and healthy
    And the config has correct DB credentials
    When the application starts
    Then the database connection pool should be created
    And a ping to the database should succeed
    And no connection errors should be logged

  Scenario: Application fails to start with wrong DB credentials
    Given PostgreSQL is running
    And the config has incorrect DB password
    When the application tries to start
    Then the database pool creation should fail
    And an error should be logged: "failed to create database pool"
    And the application should exit with non-zero code

  Scenario: Application fails to start when PostgreSQL is not running
    Given PostgreSQL is not running
    When the application tries to start
    Then the database ping should fail
    And an error should be logged: "failed to ping database"
    And the application should exit with non-zero code

  # =========================================================================
  # Health Check Endpoint
  # =========================================================================

  Scenario: Health check endpoint returns healthy status
    Given the application is running
    When I send a GET request to "/health"
    Then the response status should be 200 OK
    And the response Content-Type should be "application/json"
    And the response body should contain "status": "healthy"
    And the response body should contain "service": "kinetria"
    And the response body should contain "version": "<version>"

  Scenario: Health check endpoint is publicly accessible
    Given the application is running
    And I am not authenticated
    When I send a GET request to "/health"
    Then the response status should be 200 OK
    And no authentication error should occur

  Scenario: Health check endpoint responds quickly
    Given the application is running
    When I send a GET request to "/health"
    Then the response should be received in less than 100ms

  # =========================================================================
  # Domain Entities
  # =========================================================================

  Scenario: User entity is defined with correct fields
    Given the domain entities package
    When I inspect the User entity
    Then it should have fields: ID (UserID), Email (string), Name (string), PasswordHash (string), ProfileImageURL (string), CreatedAt (time.Time), UpdatedAt (time.Time)
    And UserID should be an alias of uuid.UUID

  Scenario: Workout entity has type, intensity, duration and image fields
    Given the domain entities package
    When I inspect the Workout entity
    Then it should have a UserID field
    And it should have fields: Type (string), Intensity (string), Duration (int), ImageURL (string)
    And it should NOT have a Status field

  Scenario: Exercise entity belongs to a Workout and has JSONB muscles
    Given the domain entities package
    When I inspect the Exercise entity
    Then it should have a WorkoutID field
    And it should have Muscles ([]string), Sets (int), Reps (string), RestTime (int), Weight (int, grams), OrderIndex (int)
    And it should NOT have Category or PrimaryMuscleGroup fields

  Scenario: Session entity tracks workout progress with correct status values
    Given the domain entities package
    When I inspect the Session entity
    Then it should have UserID, WorkoutID, Status (string)
    And it should have StartedAt (time.Time) and FinishedAt (*time.Time)
    And FinishedAt should be a pointer (nullable — null = still active)
    And status values should be: active, completed, abandoned

  Scenario: SetRecord entity uses grams for weight and has status
    Given the domain entities package
    When I inspect the SetRecord entity
    Then Weight should be int (grams, not float)
    And SetNumber, Weight, Reps should be required (not pointers)
    And Status should exist with values: completed, skipped
    And RecordedAt should exist (not CreatedAt)

  Scenario: RefreshToken entity uses pointer for revocation tracking
    Given the domain entities package
    When I inspect the RefreshToken entity
    Then it should have Token (string — hash), ExpiresAt (time.Time), RevokedAt (*time.Time)
    And RevokedAt should be a pointer (null = valid token)

  Scenario: AuditLog entity captures events with required user context
    Given the domain entities package
    When I inspect the AuditLog entity
    Then UserID should be required (not pointer)
    And it should have EntityType (string), EntityID (uuid.UUID), Action (string)
    And it should have ActionData (json.RawMessage) and OccurredAt (time.Time)
    And it should have IPAddress and UserAgent fields

  # =========================================================================
  # Value Objects (VOs)
  # =========================================================================

  Scenario: WorkoutType VO has valid enum values
    Given the WorkoutType value object
    When I check the available constants
    Then it should define: FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO
    And all constants should be of type WorkoutType

  Scenario: WorkoutType validates correct values
    Given a WorkoutType with value "FORÇA"
    When I call Validate()
    Then it should return nil (no error)

  Scenario: WorkoutType rejects invalid values
    Given a WorkoutType with value "invalid_type"
    When I call Validate()
    Then it should return an error
    And the error should wrap ErrMalformedParameters

  Scenario: WorkoutIntensity VO has valid enum values
    Given the WorkoutIntensity value object
    When I check the available constants
    Then it should define: BAIXA, MODERADA, ALTA

  Scenario: SessionStatus VO has correct values
    Given the SessionStatus value object
    When I check the available constants
    Then it should define: active, completed, abandoned
    And it should NOT define: in_progress or cancelled

  Scenario: SetRecordStatus VO has valid values
    Given the SetRecordStatus value object
    When I check the available constants
    Then it should define: completed, skipped

  # =========================================================================
  # Constants
  # =========================================================================

  Scenario: Default asset constants are defined
    Given the constants package
    When I check the defaults
    Then DefaultUserAvatarURL should be "/assets/avatars/default.png"
    And DefaultExerciseThumbnailURL should be "/assets/exercises/generic.png"
    And DefaultWorkoutImage* constants should exist for each type (Forca, Hipertrofia, Mobilidade, Condicionamento)
    And DefaultExerciseRestTime should be 60 (seconds)

  Scenario: Validation constants are defined
    Given the constants package
    When I check the validation rules
    Then MinNameLength, MaxNameLength should be defined
    And MaxDescriptionLength should be 500 (for Workout.Description)
    And MaxNotesLength should be 1000 (for Session.Notes)
    And MinSetNumber (1), MaxSetNumber (20) should be defined
    And MaxWeight should be 500_000 (grams)

  # =========================================================================
  # Configuration
  # =========================================================================

  Scenario: Config parses database environment variables
    Given environment variables are set with DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
    When I call ParseConfigFromEnv()
    Then the Config should have DBHost, DBPort, DBUser, DBPassword, DBName populated
    And no error should occur

  Scenario: Config fails when required DB variables are missing
    Given DB_HOST environment variable is not set
    When I call ParseConfigFromEnv()
    Then an error should be returned
    And the error should mention "failed to parse config"

  Scenario: Config uses default values for optional fields
    Given DB_PORT is not set
    When I call ParseConfigFromEnv()
    Then Config.DBPort should default to 5432
    And Config.DBSSLMode should default to "require"
    And Config.HTTPPort should default to 8080

  # =========================================================================
  # Integration Tests
  # =========================================================================

  Scenario: End-to-end test - Start Docker, check health, verify DB
    Given I have a clean environment
    When I run "docker-compose up -d"
    And I wait for services to be healthy
    And I send a GET request to "http://localhost:8080/health"
    Then the response status should be 200
    And I can connect to PostgreSQL at localhost:5432
    And all 7 tables exist in the database

  Scenario: Verify database cascade delete behavior (User → Workouts)
    Given migrations are applied
    And I have a user with ID "user-123" in the database
    And the user has 2 workouts
    When I delete the user from the database
    Then the workouts should also be deleted (CASCADE)

  Scenario: Verify database restrict behavior (Workout → Sessions)
    Given migrations are applied
    And I have a workout with ID "workout-456"
    And the workout has an active session
    When I try to delete the workout
    Then the delete should fail with a foreign key constraint error (RESTRICT)

  Scenario: Verify only one active session per user is allowed
    Given migrations are applied
    And I have a user with ID "user-123" with an active session
    When I try to insert another active session for the same user
    Then the insert should fail with a unique constraint violation
    And the UNIQUE partial index on (user_id) WHERE status = 'active' should be the cause

  Scenario: Verify set_record UNIQUE constraint prevents duplicate sets
    Given migrations are applied
    And I have a session-exercise pair
    When I try to insert two set_records with the same (session_id, exercise_id, set_number)
    Then the second insert should fail with a unique constraint violation

  Scenario: Verify audit log is preserved even with user restriction
    Given migrations are applied
    And I have a user with ID "user-789"
    And the user has audit log entries
    When I try to delete the user
    Then the delete should fail with a foreign key constraint error (RESTRICT)
    And the audit log entries should remain untouched

  # =========================================================================
  # Error Handling
  # =========================================================================

  Scenario: Invalid SQL in migrations causes rollback
    Given a migration file has syntax errors
    When migrations are applied
    Then the migration should fail
    And the transaction should be rolled back
    And the database should remain in the previous consistent state

  Scenario: Duplicate index creation is handled gracefully
    Given migrations have been applied
    And an index already exists
    When I accidentally try to create the same index again
    Then the operation should either succeed (IF NOT EXISTS) or fail with a clear error

  # =========================================================================
  # Performance & Observability
  # =========================================================================

  Scenario: Database indexes improve query performance
    Given 10,000 users exist in the database
    When I query users by email
    Then the query should use the idx_users_email index
    And the query should complete in less than 10ms

  Scenario: Application logs database connection pool stats
    Given the application is running
    When I check the application logs
    Then I should see logs indicating successful database connection
    And connection pool statistics should be available (if implemented)

  # =========================================================================
  # Docker Cleanup
  # =========================================================================

  Scenario: Stopping Docker Compose cleans up containers
    Given Docker Compose is running
    When I run "docker-compose down"
    Then all containers should be stopped
    And all containers should be removed
    But the database volume should persist (data preserved)

  Scenario: Removing volumes clears all database data
    Given Docker Compose is running
    When I run "docker-compose down -v"
    Then all containers should be removed
    And the postgres_data volume should be deleted
    And the next "docker-compose up" should start with a fresh database
