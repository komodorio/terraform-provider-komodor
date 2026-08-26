package komodor

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func expandWorkloadOverride(d *schema.ResourceData) workloadOverrideTFData {
	return workloadOverrideTFData{
		Action:      d.Get("action").(string),
		ClusterName: d.Get("cluster_name").(string),
		Namespace:   d.Get("namespace").(string),
		Kind:        d.Get("kind").(string),
		Name:        d.Get("name").(string),
		PolicyId:    d.Get("policy_id").(string),
	}
}

func tfToAPIWorkloadOverride(tf workloadOverrideTFData) WorkloadOverrideBase {
	api := WorkloadOverrideBase{
		Action:      tf.Action,
		ClusterName: tf.ClusterName,
		Namespace:   tf.Namespace,
		Kind:        tf.Kind,
		Name:        tf.Name,
	}
	if tf.PolicyId != "" {
		api.PolicyId = stringPtr(tf.PolicyId)
	}
	return api
}

func apiToTFWorkloadOverride(api WorkloadOverride) workloadOverrideTFData {
	return workloadOverrideTFData{
		Action:      api.Action,
		ClusterName: api.ClusterName,
		Namespace:   api.Namespace,
		Kind:        api.Kind,
		Name:        api.Name,
		PolicyId:    stringValue(api.PolicyId),
		Id:          api.Id,
		CreatedBy:   stringValue(api.CreatedByEmail),
		UpdatedBy:   stringValue(api.LastUpdatedBy),
		CreatedAt:   stringValue(api.CreatedAt),
		UpdatedAt:   stringValue(api.UpdatedAt),
	}
}

func flattenWorkloadOverride(d *schema.ResourceData, tf workloadOverrideTFData) error {
	for k, v := range map[string]interface{}{
		"action":       tf.Action,
		"cluster_name": tf.ClusterName,
		"namespace":    tf.Namespace,
		"kind":         tf.Kind,
		"name":         tf.Name,
		"policy_id":    tf.PolicyId,
		"created_by":   tf.CreatedBy,
		"updated_by":   tf.UpdatedBy,
		"created_at":   tf.CreatedAt,
		"updated_at":   tf.UpdatedAt,
	} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}
