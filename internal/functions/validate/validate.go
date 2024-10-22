package validate

import (
	"net/http"
	"regexp"
)

func IsMatchesTemplate(addr string, pattern string) (bool, error) {
	res, err := MatchString(pattern, addr)
	if err != nil {
		return false, err
	}

	return res, err
}

func MatchString(pattern string, s string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err == nil {
		return re.MatchString(s), nil
	}

	return false, err
}

func IsMethodPost(method string) bool {
	return method == http.MethodPost
}

func IsMethodGet(method string) bool {
	return method == http.MethodGet
}
