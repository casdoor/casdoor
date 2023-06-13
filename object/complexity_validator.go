package object

import "regexp"

type ValidatorFunc func(password string) string

var (
	regexLowerCase = regexp.MustCompile(`[a-z]`)
	regexUpperCase = regexp.MustCompile(`[A-Z]`)
	regexDigit     = regexp.MustCompile(`\d`)
	regexSpecial   = regexp.MustCompile(`[!@#$%^&*]`)
)

func isValidOptionAtLeast8(password string) string {
	if len(password) < 8 {
		return "AtLeast8"
	}
	return ""
}

func isValidOptionAa123(password string) string {
	hasLowerCase := regexLowerCase.MatchString(password)
	hasUpperCase := regexUpperCase.MatchString(password)
	hasDigit := regexDigit.MatchString(password)

	if !hasLowerCase || !hasUpperCase || !hasDigit {
		return "Aa123"
	}
	return ""
}

func isValidOptionSpecialChar(password string) string {
	if !regexSpecial.MatchString(password) {
		return "SpecialChar"
	}
	return ""
}

func isValidOptionNoRepeat(password string) string {
	for i := 0; i < len(password)-1; i++ {
		if password[i] == password[i+1] {
			return "NoRepeat"
		}
	}
	return ""
}
