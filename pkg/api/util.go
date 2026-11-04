package api

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/joohoi/acme-dns/pkg/acmedns"
)

func jsonError(message string) []byte {
	return []byte(fmt.Sprintf("{\"error\": \"%s\"}", message))
}

func getValidUsername(u string) (uuid.UUID, error) {
	uname, err := uuid.Parse(u)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uname, nil
}

func validKey(k string) bool {
	kn := acmedns.SanitizeString(k)
	if utf8.RuneCountInString(k) == 40 && utf8.RuneCountInString(kn) == 40 {
		// Correct length and all chars valid
		return true
	}
	return false
}

func getIPListFromHeader(header string) []string {
	iplist := []string{}
	for _, v := range strings.Split(header, ",") {
		if len(v) > 0 {
			// Ignore empty values
			iplist = append(iplist, strings.TrimSpace(v))
		}
	}
	return iplist
}

func validSubdomain(s string) bool {
	// URL safe base64 alphabet without padding as defined in ACME
	RegExp := regexp.MustCompile("^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$")
	return RegExp.MatchString(s)
}

// validTXT accepts a value of 1 to 255 characters, all from the URL-safe
// base64 alphabet that SanitizeString keeps.
//
// Upstream accepts exactly 43 characters, the length of an ACME DNS-01 key
// authorization digest. That length is not defined by ACME, and some
// certificate authorities issue domain-validation values of other lengths.
// 255 is the most a single TXT character-string can hold (RFC 1035 §3.3).
func validTXT(s string) bool {
	sn := acmedns.SanitizeString(s)
	n := utf8.RuneCountInString(s)
	if n >= 1 && n <= 255 && utf8.RuneCountInString(sn) == n {
		return true
	}
	return false
}
