package komodor

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func String(v string) *string {
	return &v
}

func ExpandStringList(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, String(v.(string)))
		}
	}
	return vs
}

func ExpandStringSet(configured *schema.Set) []*string {
	return ExpandStringList(configured.List())
}
