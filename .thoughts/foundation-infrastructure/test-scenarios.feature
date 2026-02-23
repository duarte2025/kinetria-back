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
    And all ENUMs should be created
    And no migration errors should occur

  Scenario: Migrations create users table with correct schema
    Given migrations have been applied
    When I query the users table schema
    Then it should have columns: id (UUID), email (VARCHAR UNIQUE), name (VARCHAR), password_hash (VARCHAR), created_at (TIMESTAMPTZ), updated_at (TIMESTAMPTZ)
    And it should have an index on email
    And id should be the primary key

  Scenario: Migrations create workouts table with correct relationships
    Given migrations have been applied
    When I query the workouts table schema
    Then it should have a foreign key to users(id) with ON DELETE CASCADE
    And it should have the workout_status ENUM type
    And it should have indexes on user_id, status, and (user_id, status)

  Scenario: Migrations create exercises table with ENUMs
    Given migrations have been applied
    When I query the exercises table schema
    Then it should have exercise_category ENUM with values: strength, cardio, flexibility, balance
    And it should have muscle_group ENUM with values: chest, back, legs, shoulders, arms, core, full_body
    And difficulty_level should have a CHECK constraint (1-5)

  Scenario: Migrations create sessions table with status tracking
    Given migrations have been applied
    When I query the sessions table schema
    Then it should have a foreign key to users(id) with ON DELETE CASCADE
    And it should have a foreign key to workouts(id) with ON DELETE RESTRICT
    And completed_at should be nullable
    And status should use session_status ENUM

  Scenario: Migrations create set_records table with validations
    Given migrations have been applied
    When I query the set_records table schema
    Then reps, weight_kg, and duration_seconds should be nullable
    And weight_kg should be DECIMAL(6,2)
    And all numeric fields should have CHECK constraints (>= 0)
    And set_number should have CHECK constraint (> 0)

  Scenario: Migrations create refresh_tokens table with security features
    Given migrations have been applied
    When I query the refresh_tokens table schema
    Then token_hash should be UNIQUE
    And it should have indexes on user_id, token_hash, expires_at
    And revoked should default to FALSE

  Scenario: Migrations create audit_log table with JSONB support
    Given migrations have been applied
    When I query the audit_log table schema
    Then metadata should be JSONB type
    And it should have a GIN index on metadata
    And user_id foreign key should have ON DELETE SET NULL
    And action should use audit_action ENUM

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
    Then it should have fields: ID (UserID), Email (string), Name (string), PasswordHash (string), CreatedAt (time.Time), UpdatedAt (time.Time)
    And UserID should be an alias of uuid.UUID

  Scenario: Workout entity references User correctly
    Given the domain entities package
    When I inspect the Workout entity
    Then it should have a UserID field
    And it should have a Status field of type WorkoutStatus

  Scenario: Exercise entity has all catalog fields
    Given the domain entities package
    When I inspect the Exercise entity
    Then it should have Category (ExerciseCategory), PrimaryMuscleGroup (MuscleGroup), DifficultyLevel (int)
    And it should have optional fields: VideoURL, ThumbnailURL

  Scenario: Session entity tracks workout progress
    Given the domain entities package
    When I inspect the Session entity
    Then it should have UserID, WorkoutID, Status (SessionStatus)
    And it should have StartedAt (time.Time) and CompletedAt (*time.Time)
    And CompletedAt should be a pointer (nullable)

  Scenario: SetRecord entity has flexible metrics
    Given the domain entities package
    When I inspect the SetRecord entity
    Then Reps, WeightKg, and DurationSeconds should be pointers (nullable)
    And SetNumber should be required (int, not pointer)

  Scenario: RefreshToken entity supports authentication
    Given the domain entities package
    When I inspect the RefreshToken entity
    Then it should have TokenHash (string), ExpiresAt (time.Time), Revoked (bool)

  Scenario: AuditLog entity captures events
    Given the domain entities package
    When I inspect the AuditLog entity
    Then it should have UserID (*UserID - pointer/nullable), Action (AuditAction)
    And it should have Metadata (map or interface for JSONB)
    And it should have IPAddress and UserAgent fields

  # =========================================================================
  # Value Objects (VOs)
  # =========================================================================

  Scenario: WorkoutStatus VO has valid enum values
    Given the WorkoutStatus value object
    When I check the available constants
    Then it should define: WorkoutStatusDraft, WorkoutStatusPublished, WorkoutStatusArchived
    And all constants should be of type WorkoutStatus

  Scenario: WorkoutStatus validates correct values
    Given a WorkoutStatus with value "published"
    When I call Validate()
    Then it should return nil (no error)

  Scenario: WorkoutStatus rejects invalid values
    Given a WorkoutStatus with value "invalid_status"
    When I call Validate()
    Then it should return an error
    And the error should wrap ErrMalformedParameters

  Scenario: SessionStatus VO has valid enum values
    Given the SessionStatus value object
    When I check the available constants
    Then it should define: SessionStatusInProgress, SessionStatusCompleted, SessionStatusCancelled

  Scenario: ExerciseCategory VO has valid enum values
    Given the ExerciseCategory value object
    When I check the available constants
    Then it should define: strength, cardio, flexibility, balance

  Scenario: MuscleGroup VO has valid enum values
    Given the MuscleGroup value object
    When I check the available constants
    Then it should define: chest, back, legs, shoulders, arms, core, full_body

  Scenario: AuditAction VO has comprehensive action list
    Given the AuditAction value object
    When I check the available constants
    Then it should define actions for: user (created/updated/deleted), workout (created/updated/deleted), session (started/completed/cancelled), set (recorded/updated/deleted), auth (login/logout/token_refreshed)

  # =========================================================================
  # Constants
  # =========================================================================

  Scenario: Default constants are defined
    Given the constants package
    When I check the defaults
    Then DefaultExerciseThumbnail should be defined
    And DefaultExerciseVideo should be defined
    And DefaultDifficultyLevel should be 3

  Scenario: Validation constants are defined
    Given the constants package
    When I check the validation rules
    Then MinNameLength, MaxNameLength should be defined
    And MinDifficultyLevel (1), MaxDifficultyLevel (5) should be defined
    And MinSetNumber, MaxSetNumber should be defined

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

  Scenario: Verify audit log preserves data after user deletion
    Given migrations are applied
    And I have a user with ID "user-789"
    And the user has audit log entries
    When I delete the user
    Then the audit log entries should remain
    And the user_id in audit log should be NULL (SET NULL)

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
