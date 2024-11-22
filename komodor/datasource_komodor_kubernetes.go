package komodor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorKubernetes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKomodorKubernetesRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id and api key of the cluster integration",
			},
			"cluster_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster; must be unique to a Komodor account",
			},
		},
		Description: "Retrieves an existing Komodor Kubernetes cluster integration by name",
	}
}

func dataSourceKomodorKubernetesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	clusterName := d.Get("cluster_name").(string)

	kubernetes, _, err := c.GetKubernetesCluster(clusterName)
	if err != nil {
		return diag.Errorf("Could not get kubernetes cluster integration by name %s", err)
	}

	d.SetId(kubernetes.Id)

	return nil
}
