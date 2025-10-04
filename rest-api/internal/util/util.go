package util

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nyaruka/phonenumbers"
	"github.com/rs/zerolog/log"
)

func NilString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func NilBool(b bool) *bool {
	return &b
}

func NilToEmptyString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func RemoveDuplicateStrings(s []string) []string {
	keys := make(map[string]bool)
	res := []string{}
	for _, v := range s {
		if _, ok := keys[v]; !ok {
			keys[v] = true
			res = append(res, v)
		}
	}
	return res
}

func GetUpdatedJSONFields(b []byte) []string {
	var f []string
	m := make(map[string]any)

	if err := json.Unmarshal(b, &m); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal updated json fields")
		return nil
	}

	for s := range m {
		f = append(f, s)
	}

	return f
}

func PermutateStrings(w []string, fn func([]string)) {
	permutateStrings(w, fn, 0)
}

func permutateStrings(w []string, fn func([]string), i int) {
	if i > len(w) {
		fn(w)
		return
	}
	permutateStrings(w, fn, i+1)
	for j := i + 1; j < len(w); j++ {
		w[i], w[j] = w[j], w[i]
		permutateStrings(w, fn, i+1)
		w[i], w[j] = w[j], w[i]
	}
}

func DecodePgtypeDateText(s string, dst *pgtype.Date) error {
	if strings.HasSuffix(s, " BC") {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimRight(s, " BC"), time.UTC)
		t2 := time.Date(1-t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		if err != nil {
			return err
		}
		*dst = pgtype.Date{Time: t2, Valid: true}
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return err
	}

	*dst = pgtype.Date{Time: t, Valid: true}
	return nil
}

func StringSliceDiff(a, b []string) []string {
	diff := []string{}
	vals := map[string]struct{}{}

	for _, x := range a {
		vals[strings.ToLower(x)] = struct{}{}
	}

	for _, x := range b {
		if _, ok := vals[strings.ToLower(x)]; !ok {
			diff = append(diff, x)
		}
	}

	return diff
}

func IsUrl(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func TrimSpaceSlice(s []string) {
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
}

func TimeAfter(a time.Time, b time.Time) bool {
	return a.Year() >= b.Year() && a.YearDay() > b.YearDay()
}

func TimeBefore(a time.Time, b time.Time) bool {
	return a.Year() <= b.Year() && a.YearDay() < b.YearDay()
}

func TimeEqual(a time.Time, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func MapTags(tags []string) map[string][]string {
	res := make(map[string][]string)
	for _, s := range tags {
		s = strings.TrimSpace(s)
		v := strings.Split(s, ":")
		if len(v) == 2 {
			if _, ok := res[v[0]]; !ok {
				res[v[0]] = []string{v[1]}
			} else {
				res[v[0]] = append(res[v[0]], v[1])
			}
		}
	}

	return res
}

func StringIsEqual(a *string, b *string) bool {
	if a == b || a != nil && b != nil && strings.EqualFold(*a, *b) {
		return true
	}

	return false
}

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func IpToint(ip net.IP) int {
	if ip.IsUnspecified() {
		return 0
	}
	if len(ip) == 16 {
		return int(binary.BigEndian.Uint32(ip[12:16]))
	}
	return int(binary.BigEndian.Uint32(ip))
}

func IsValidEmail(email string) error {
	emailRegex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(emailRegex)

	if email == "" {
		return errors.New("email cannot be empty")
	}

	if !re.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func FormatPhoneNumber(localNumber string, countryCode string) (string, error) {
	parsedNumber, err := phonenumbers.Parse(localNumber, countryCode)
	if err != nil {
		return "", errors.New("error parsing phone number")
	}

	if !phonenumbers.IsValidNumber(parsedNumber) {
		return "", errors.New("invalid phone number")
	}

	formattedNumber := phonenumbers.Format(parsedNumber, phonenumbers.E164)

	return formattedNumber, nil
}
