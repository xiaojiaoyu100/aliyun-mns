package alimns

import "regexp"

const (
	base64Regex = `^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`
)

func IsBase54(s string) bool {
	rp, _ := regexp.Compile(base64Regex)
	return rp.MatchString(s)
}
