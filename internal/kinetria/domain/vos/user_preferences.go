package vos

import (
	"encoding/json"
	"fmt"
)

// Theme constants
type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
)

// Language constants
type Language string

const (
	LanguagePtBR Language = "pt-BR"
	LanguageEnUS Language = "en-US"
)

// UserPreferences holds user UI preferences.
type UserPreferences struct {
	Theme    Theme    `json:"theme"`
	Language Language `json:"language"`
}

// DefaultUserPreferences returns the default preferences.
func DefaultUserPreferences() UserPreferences {
	return UserPreferences{
		Theme:    ThemeLight,
		Language: LanguagePtBR,
	}
}

// Validate returns an error if any field has an invalid value.
func (p UserPreferences) Validate() error {
	switch p.Theme {
	case ThemeDark, ThemeLight:
	default:
		return fmt.Errorf("invalid theme %q: must be \"dark\" or \"light\"", p.Theme)
	}
	switch p.Language {
	case LanguagePtBR, LanguageEnUS:
	default:
		return fmt.Errorf("invalid language %q: must be \"pt-BR\" or \"en-US\"", p.Language)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p UserPreferences) MarshalJSON() ([]byte, error) {
	type alias UserPreferences
	return json.Marshal(alias(p))
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *UserPreferences) UnmarshalJSON(data []byte) error {
	type alias UserPreferences
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*p = UserPreferences(a)
	return nil
}
