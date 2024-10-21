package validate

import (
	"net/http"
	"regexp"
)

func IsMatchesTemplate(addr string, pattern string) (bool, error) {
	res, err := MatchString(pattern, addr)
	if err != nil {
		return false, err
	} else {
		return res, err
	}
}

func MatchString(pattern string, s string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err == nil {
		return re.MatchString(s), nil
	} else {
		return false, err
	}
}

func IsMethodPost(method string) bool {
	if method == http.MethodPost {
		return true
	} else {
		return false
	}
}

func IsMethodGet(method string) bool {
	if method == http.MethodGet {
		return true
	} else {
		return false
	}
}
