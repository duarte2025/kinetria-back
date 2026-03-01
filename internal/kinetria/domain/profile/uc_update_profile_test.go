package profile_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	domainprofile "github.com/kinetria/kinetria-back/internal/kinetria/domain/profile"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
	"go.opentelemetry.io/otel/trace/noop"
)

func ptr[T any](v T) *T { return &v }

func TestUpdateProfileUC_Execute(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	userID := uuid.New()

	baseUser := func() *entities.User {
		return &entities.User{
			ID:          userID,
			Email:       "alice@example.com",
			Name:        "Alice",
			Preferences: vos.DefaultUserPreferences(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	tests := []struct {
		name        string
		setupRepo   func(*mockProfileUserRepo)
		input       domainprofile.UpdateProfileInput
		wantErr     error
		checkResult func(t *testing.T, user *entities.User)
	}{
		{
			name: "update name successfully",
			setupRepo: func(r *mockProfileUserRepo) {
				r.byID[userID] = baseUser()
			},
			input:   domainprofile.UpdateProfileInput{Name: ptr("Bob")},
			wantErr: nil,
			checkResult: func(t *testing.T, user *entities.User) {
				if user.Name != "Bob" {
					t.Errorf("Name = %q, want %q", user.Name, "Bob")
				}
			},
		},
		{
			name: "update profileImageURL successfully",
			setupRepo: func(r *mockProfileUserRepo) {
				r.byID[userID] = baseUser()
			},
			input:   domainprofile.UpdateProfileInput{ProfileImageURL: ptr("https://example.com/avatar.png")},
			wantErr: nil,
			checkResult: func(t *testing.T, user *entities.User) {
				if user.ProfileImageURL != "https://example.com/avatar.png" {
					t.Errorf("ProfileImageURL = %q, want %q", user.ProfileImageURL, "https://example.com/avatar.png")
				}
			},
		},
		{
			name: "update preferences successfully",
			setupRepo: func(r *mockProfileUserRepo) {
				r.byID[userID] = baseUser()
			},
			input: domainprofile.UpdateProfileInput{
				Preferences: &vos.UserPreferences{
					Theme:    vos.ThemeDark,
					Language: vos.LanguageEnUS,
				},
			},
			wantErr: nil,
			checkResult: func(t *testing.T, user *entities.User) {
				if user.Preferences.Theme != vos.ThemeDark {
					t.Errorf("Theme = %q, want %q", user.Preferences.Theme, vos.ThemeDark)
				}
				if user.Preferences.Language != vos.LanguageEnUS {
					t.Errorf("Language = %q, want %q", user.Preferences.Language, vos.LanguageEnUS)
				}
			},
		},
		{
			name: "update multiple fields",
			setupRepo: func(r *mockProfileUserRepo) {
				r.byID[userID] = baseUser()
			},
			input: domainprofile.UpdateProfileInput{
				Name:            ptr("Charlie"),
				ProfileImageURL: ptr("https://example.com/charlie.png"),
				Preferences: &vos.UserPreferences{
					Theme:    vos.ThemeDark,
					Language: vos.LanguagePtBR,
				},
			},
			wantErr: nil,
			checkResult: func(t *testing.T, user *entities.User) {
				if user.Name != "Charlie" {
					t.Errorf("Name = %q, want %q", user.Name, "Charlie")
				}
				if user.ProfileImageURL != "https://example.com/charlie.png" {
					t.Errorf("ProfileImageURL = %q, want %q", user.ProfileImageURL, "https://example.com/charlie.png")
				}
				if user.Preferences.Theme != vos.ThemeDark {
					t.Errorf("Theme = %q, want %q", user.Preferences.Theme, vos.ThemeDark)
				}
			},
		},
		{
			name:      "name too short (< 2 chars after trim)",
			setupRepo: func(r *mockProfileUserRepo) {},
			input:     domainprofile.UpdateProfileInput{Name: ptr("A")},
			wantErr:   domainerrors.ErrMalformedParameters,
		},
		{
			name:      "name too long (> 100 chars)",
			setupRepo: func(r *mockProfileUserRepo) {},
			input:     domainprofile.UpdateProfileInput{Name: ptr(strings.Repeat("x", 101))},
			wantErr:   domainerrors.ErrMalformedParameters,
		},
		{
			name:      "name with only spaces - trimmed to empty string - too short",
			setupRepo: func(r *mockProfileUserRepo) {},
			input:     domainprofile.UpdateProfileInput{Name: ptr("   ")},
			wantErr:   domainerrors.ErrMalformedParameters,
		},
		{
			name:      "invalid preferences theme",
			setupRepo: func(r *mockProfileUserRepo) {},
			input: domainprofile.UpdateProfileInput{
				Preferences: &vos.UserPreferences{
					Theme:    "neon",
					Language: vos.LanguagePtBR,
				},
			},
			wantErr: domainerrors.ErrMalformedParameters,
		},
		{
			name:      "invalid preferences language",
			setupRepo: func(r *mockProfileUserRepo) {},
			input: domainprofile.UpdateProfileInput{
				Preferences: &vos.UserPreferences{
					Theme:    vos.ThemeLight,
					Language: "fr-FR",
				},
			},
			wantErr: domainerrors.ErrMalformedParameters,
		},
		{
			name:      "no fields provided - ErrMalformedParameters",
			setupRepo: func(r *mockProfileUserRepo) {},
			input:     domainprofile.UpdateProfileInput{},
			wantErr:   domainerrors.ErrMalformedParameters,
		},
		{
			name: "user not found",
			setupRepo: func(r *mockProfileUserRepo) {
				r.getByIDErr = domainerrors.ErrNotFound
			},
			input:   domainprofile.UpdateProfileInput{Name: ptr("Valid Name")},
			wantErr: domainerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProfileUserRepo{
				byID: make(map[uuid.UUID]*entities.User),
			}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			uc := domainprofile.NewUpdateProfileUC(tracer, repo)
			user, err := uc.Execute(context.Background(), userID, tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Execute() expected error wrapping %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Execute() error = %v, wantErr wrapping %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			if user == nil {
				t.Fatal("Execute() returned nil user on success")
			}

			if tt.checkResult != nil {
				tt.checkResult(t, user)
			}
		})
	}
}
