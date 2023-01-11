package komodor

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func String(v string) *string { // strange function, looks redundant
	return &v
}

func ExpandStringList(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, String(v.(string))) // why not just cast to *string?
		}
	}
	return vs
}

func ExpandStringSet(configured *schema.Set) []*string {
	return ExpandStringList(configured.List())
}

func GetStatusCodeFromErrorMessage(err error) string { // is there a way to check the type of it instead? Then use type assertion and get statusCode from there. Alternatively, return third field in response with status code and ignore it in some places. Current approach with generating the string and then parsing it back looks unelegant
	parts := strings.Split(err.Error(), ",")
	status := parts[0]
	statusParts := strings.Split(status, " ")
	statusCode := statusParts[len(statusParts)-1]

	return statusCode
}
