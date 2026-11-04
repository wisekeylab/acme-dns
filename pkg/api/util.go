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

// validTXTValue is what validTXT accepts: 1 to 255 characters from the
// URL-safe base64 alphabet plus '='.
var validTXTValue = regexp.MustCompile(`^[A-Za-z0-9_=-]{1,255}$`)

// validTXT accepts a value of 1 to 255 characters from the URL-safe base64
// alphabet plus '='.
//
// Upstream accepts exactly 43 characters from the URL-safe base64 alphabet,
// the shape of an ACME DNS-01 key authorization digest. That length is not
// defined by ACME, and some certificate authorities issue domain-validation
// values of other lengths, in the form label=hex. 255 is the most a single TXT
// character-string can hold (RFC 1035 §3.3).
func validTXT(s string) bool {
	return validTXTValue.MatchString(s)
}
