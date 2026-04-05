package komodor

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// jsonDiffSuppress suppresses diffs for JSON string fields that are semantically
// equal but differ in key ordering or whitespace (e.g. jsonencode vs json.Marshal).
func jsonDiffSuppress(_, old, new string, _ *schema.ResourceData) bool {
	var oldVal, newVal interface{}
	if json.Unmarshal([]byte(old), &oldVal) != nil || json.Unmarshal([]byte(new), &newVal) != nil {
		return false
	}
	oldNorm, err1 := json.Marshal(oldVal)
	newNorm, err2 := json.Marshal(newVal)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(oldNorm) == string(newNorm)
}

func ExpandStringList(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, &val)
		}
	}
	return vs
}

func ExpandStringSet(configured *schema.Set) []*string {
	return ExpandStringList(configured.List())
}
