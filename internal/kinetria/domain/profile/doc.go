// Package profile provides use cases for managing user profile data.
//
// It exposes two use cases:
//
//   - [GetProfileUC]: retrieves the authenticated user's current profile.
//   - [UpdateProfileUC]: performs a partial update on the authenticated user's profile,
//     applying only the fields explicitly supplied by the caller.
//
// # Validation rules
//
//   - Name: 2–100 characters (whitespace-trimmed); required when provided.
//   - ProfileImageURL: arbitrary non-empty string; no format validation is enforced at
//     the domain layer.
//   - Preferences.Theme: must be one of "dark" or "light".
//   - Preferences.Language: must be one of "pt-BR" or "en-US".
//
// # Errors
//
// Both use cases propagate errors from [ports.UserRepository]. [UpdateProfileUC.Execute]
// additionally returns [domainerrors.ErrMalformedParameters] when validation fails or no
// fields are provided.
package profile
