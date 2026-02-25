Feature: Sessions Completion (RecordSet, FinishSession, AbandonSession)

  Background:
    Given a user is authenticated with a valid JWT token
    And the user has an active workout session

  # ─────────────────────────────────────────────────────────────
  # POST /sessions/:id/sets (RecordSet)
  # ─────────────────────────────────────────────────────────────

  Scenario: Successfully record a completed set
    Given the session has an exercise with id "exercise-123"
    When the user records a set with:
      | exerciseId | setNumber | weight | reps | status    |
      | exercise-123 | 1       | 82.5   | 10   | completed |
    Then the response status should be 201
    And the set record should be created with weight 82500 grams
    And an audit log entry should be created with action "created"

  Scenario: Successfully record a skipped set
    Given the session has an exercise with id "exercise-123"
    When the user records a set with:
      | exerciseId | setNumber | weight | reps | status  |
      | exercise-123 | 2       | 0      | 0    | skipped |
    Then the response status should be 201
    And the set record should be created with status "skipped"

  Scenario: Fail to record set - session not found
    When the user records a set for a non-existent session
    Then the response status should be 404
    And the error code should be "SESSION_NOT_FOUND"

  Scenario: Fail to record set - session not active
    Given the session status is "completed"
    When the user records a set
    Then the response status should be 409
    And the error code should be "SESSION_NOT_ACTIVE"

  Scenario: Fail to record set - exercise not in workout
    When the user records a set with exerciseId "wrong-exercise-id"
    Then the response status should be 404
    And the error code should be "EXERCISE_NOT_FOUND"

  Scenario: Fail to record set - duplicate set number
    Given the user already recorded set number 1 for exercise "exercise-123"
    When the user records set number 1 again for the same exercise
    Then the response status should be 409
    And the error code should be "SET_ALREADY_RECORDED"

  Scenario: Fail to record set - session belongs to another user
    Given the session belongs to another user
    When the user tries to record a set
    Then the response status should be 404
    And the error code should be "SESSION_NOT_FOUND"

  Scenario: Record set with bodyweight exercise (weight = 0)
    Given the session has a bodyweight exercise
    When the user records a set with weight 0
    Then the response status should be 201
    And the set record should be created with weight 0

  Scenario: Record set with failure (reps = 0)
    Given the session has an exercise
    When the user records a set with reps 0 and status "completed"
    Then the response status should be 201
    And the set record should be created with reps 0

  # ─────────────────────────────────────────────────────────────
  # PATCH /sessions/:id/finish (FinishSession)
  # ─────────────────────────────────────────────────────────────

  Scenario: Successfully finish an active session with notes
    Given the session status is "active"
    When the user finishes the session with notes "Great workout!"
    Then the response status should be 200
    And the session status should be "completed"
    And the session finishedAt should be set
    And the session notes should be "Great workout!"
    And an audit log entry should be created with action "completed"

  Scenario: Successfully finish an active session without notes
    Given the session status is "active"
    When the user finishes the session without notes
    Then the response status should be 200
    And the session status should be "completed"
    And the session notes should be empty

  Scenario: Fail to finish - session not found
    When the user tries to finish a non-existent session
    Then the response status should be 404
    And the error code should be "SESSION_NOT_FOUND"

  Scenario: Fail to finish - session already completed
    Given the session status is "completed"
    When the user tries to finish the session
    Then the response status should be 409
    And the error code should be "SESSION_ALREADY_CLOSED"

  Scenario: Fail to finish - session already abandoned
    Given the session status is "abandoned"
    When the user tries to finish the session
    Then the response status should be 409
    And the error code should be "SESSION_ALREADY_CLOSED"

  Scenario: Fail to finish - session belongs to another user
    Given the session belongs to another user
    When the user tries to finish the session
    Then the response status should be 404
    And the error code should be "SESSION_NOT_FOUND"

  Scenario: Finish session without recording any sets
    Given the session status is "active"
    And no sets have been recorded
    When the user finishes the session
    Then the response status should be 200
    And the session status should be "completed"

  # ─────────────────────────────────────────────────────────────
  # PATCH /sessions/:id/abandon (AbandonSession)
  # ─────────────────────────────────────────────────────────────

  Scenario: Successfully abandon an active session
    Given the session status is "active"
    When the user abandons the session
    Then the response status should be 200
    And the session status should be "abandoned"
    And the session finishedAt should be set
    And an audit log entry should be created with action "abandoned"

  Scenario: Fail to abandon - session not found
    When the user tries to abandon a non-existent session
    Then the response status should be 404
    And the error code should be "SESSION_NOT_FOUND"

  Scenario: Fail to abandon - session already completed
    Given the session status is "completed"
    When the user tries to abandon the session
    Then the response status should be 409
    And the error code should be "SESSION_ALREADY_CLOSED"

  Scenario: Fail to abandon - session already abandoned
    Given the session status is "abandoned"
    When the user tries to abandon the session
    Then the response status should be 409
    And the error code should be "SESSION_ALREADY_CLOSED"

  Scenario: Fail to abandon - session belongs to another user
    Given the session belongs to another user
    When the user tries to abandon the session
    Then the response status should be 404
    And the error code should be "SESSION_NOT_FOUND"

  Scenario: Abandon session with recorded sets
    Given the session status is "active"
    And the user has recorded 3 sets
    When the user abandons the session
    Then the response status should be 200
    And the session status should be "abandoned"
    And the recorded sets should remain in the database

  # ─────────────────────────────────────────────────────────────
  # Edge Cases & Concurrency
  # ─────────────────────────────────────────────────────────────

  Scenario: Concurrent set recording (same set number)
    Given the session has an exercise
    When two requests try to record set number 1 simultaneously
    Then one request should succeed with status 201
    And the other request should fail with status 409

  Scenario: Record set after session is finished
    Given the session status is "completed"
    When the user tries to record a set
    Then the response status should be 409
    And the error code should be "SESSION_NOT_ACTIVE"

  Scenario: Finish session twice (idempotency check)
    Given the session status is "active"
    When the user finishes the session
    And the user tries to finish the session again
    Then the second request should fail with status 409
