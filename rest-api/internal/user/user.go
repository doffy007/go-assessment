package user

import (
	"errors"
	"rest-api/internal/util"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID               uint64             `json:"id,string"`
	Name             *string            `json:"name"`
	Username         string             `json:"username"`
	Email            string             `json:"email"`
	Password         *string            `json:"password"`
	EmailVerified    *bool              `json:"email_verified"`
	Bio              *string            `json:"bio"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	DeactivatedAt    pgtype.Timestamptz `json:"deactivated_at"`
	DeletedAt        pgtype.Timestamptz `json:"deleted_at"`
	LastSeen         pgtype.Timestamptz `json:"last_seen"`
	IPAddress        *string            `json:"ip_address"`
	UserAgent        *string            `json:"user_agent"`
	Role             *string            `json:"role,omitempty"`
	Gender           *string            `json:"gender,omitempty"`
	CountryCode      *string            `json:"country_code"`
	PhoneNumber      *string            `json:"phone_number"`
	Location         *string            `json:"location"`
	Nationality      *string            `json:"nationality"`
	NationalIDNumber *string            `json:"national_id_number"`
	Language         *string            `json:"language"`
	Birthdate        pgtype.Date        `json:"birthdate"`
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrDeleteToken  = errors.New("error delete token")
	ErrUserNotFound = errors.New("user not found")
)

func (o *User) RemoveSensitivePII() {
	o.Email = ""
	o.Gender = nil
	o.Birthdate.Valid = false
	o.Password = nil
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasSpace     bool
		hasUpperCase bool
		hasDigit     bool
	)

	for _, r := range password {
		switch {
		case unicode.IsSpace(r):
			hasSpace = true
		case unicode.IsUpper(r):
			hasUpperCase = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	return !hasSpace && hasUpperCase && hasDigit
}

func (o *User) ValidateEmail() error {
	if o.Email == "" {
		return nil
	}

	if cfg.SuspiciousEmailDetectionEnable {
		parts := strings.Split(o.Email, "@")
		if len(parts) < 2 {
			return errors.New("invalid email format")
		}

		emailLocalPart := parts[0]
		if cfg.SuspiciousEmailRegex.MatchString(emailLocalPart) {
			log.Warn().Str("email", o.Email).Msg("suspicious email detected!")
			return errors.New("suspicious email address")
		}

		if o.Name != nil {
			nameWords := strings.Fields(strings.ToLower(*o.Name))
			if len(nameWords) > 0 {
				matched := false

				util.PermutateStrings(nameWords, func(w []string) {
					joined := strings.Join(w, "")
					if strings.HasPrefix(emailLocalPart, joined) {
						matched = true
					}
				})

				if matched {
					log.Warn().Str("email", o.Email).Msg("suspicious email detected! (name match)")
					return errors.New("suspicious email address")
				}
			}
		}

		for _, r := range emailLocalPart {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
				log.Warn().Str("email", o.Email).Msg("suspicious email detected! (weird characters)")
				return errors.New("suspicious email address")
			}
		}
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CompareHashAndPassword(hashedPassword string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.New("invalid password")
	}
	return nil
}

var (
	ErrUsernameOrEmailExists = errors.New("username or email already exists")
)
