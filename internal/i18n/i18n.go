package i18n

import (
	"errors"
	"net/http"
	"strings"

	"easy-clock/internal/domain"
	domainuser "easy-clock/internal/domain/user"
)

type Lang string

const (
	PL Lang = "pl"
	EN Lang = "en"
)

// DetectLang picks PL or EN from an Accept-Language header value.
// Defaults to EN when neither is present or Polish isn't preferred.
func DetectLang(acceptLang string) Lang {
	// Walk through comma-separated tags, return PL on first "pl" match.
	for _, part := range strings.Split(acceptLang, ",") {
		tag := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if strings.HasPrefix(tag, "pl") {
			return PL
		}
		if strings.HasPrefix(tag, "en") {
			return EN
		}
	}
	return EN
}

// DomainError returns a user-facing message and HTTP status for a domain error.
func DomainError(err error, lang Lang) (string, int) {
	type entry struct {
		sentinel error
		status   int
		pl, en   string
	}
	table := []entry{
		{domain.ErrNotFound, http.StatusNotFound,
			"Nie znaleziono zasobu.", "Resource not found."},
		{domain.ErrEmptyName, http.StatusBadRequest,
			"Nazwa nie może być pusta.", "Name cannot be empty."},
		{domain.ErrInvalidTimezone, http.StatusBadRequest,
			"Nieprawidłowa strefa czasowa (wymagany format IANA, np. Europe/Warsaw).",
			"Invalid timezone (expected IANA format, e.g. Europe/Warsaw)."},
		{domain.ErrInvalidHourRange, http.StatusBadRequest,
			"Godzina początkowa musi być mniejsza od końcowej.", "Start hour must be less than end hour."},
		{domain.ErrActivityOverlap, http.StatusBadRequest,
			"Aktywności w tym samym pierścieniu nie mogą się nakładać.", "Activities in the same ring must not overlap."},
		{domain.ErrImageRequired, http.StatusBadRequest,
			"Ścieżka do obrazka jest wymagana.", "Image path is required."},
		{domain.ErrInvalidTimeRange, http.StatusBadRequest,
			"Czas 'od' musi być wcześniejszy niż 'do'.", "'From' time must be before 'to' time."},
		{domain.ErrEventProfileXorActivities, http.StatusBadRequest,
			"Wydarzenie może zawierać profil albo aktywności — nie oba jednocześnie.",
			"An event may contain either a profile or activities, not both."},
		{domainuser.ErrNotFound, http.StatusNotFound,
			"Nie znaleziono użytkownika.", "User not found."},
		{domainuser.ErrEmailTaken, http.StatusConflict,
			"Ten adres e-mail jest już zajęty.", "This email is already taken."},
		{domainuser.ErrInvalidCredentials, http.StatusUnauthorized,
			"Nieprawidłowy e-mail lub hasło.", "Invalid email or password."},
		{domainuser.ErrNotActive, http.StatusForbidden,
			"Konto nie zostało jeszcze aktywowane. Sprawdź skrzynkę e-mail.",
			"Account not activated yet. Please check your email."},
		{domainuser.ErrInvalidToken, http.StatusBadRequest,
			"Link weryfikacyjny jest nieprawidłowy lub wygasł.", "Verification link is invalid or has expired."},
		{domainuser.ErrEmptyName, http.StatusBadRequest,
			"Imię nie może być puste.", "Name cannot be empty."},
		{domainuser.ErrEmptyEmail, http.StatusBadRequest,
			"Adres e-mail nie może być pusty.", "Email cannot be empty."},
	}

	for _, e := range table {
		if errors.Is(err, e.sentinel) {
			if lang == PL {
				return e.pl, e.status
			}
			return e.en, e.status
		}
	}

	if lang == PL {
		return "Wystąpił błąd serwera. Spróbuj ponownie.", http.StatusInternalServerError
	}
	return "Internal server error. Please try again.", http.StatusInternalServerError
}
