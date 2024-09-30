package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKomodorKubernetes() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id and api key of the cluster integration",
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		CreateContext: resourceKomodorKubernetesCreate,
		ReadContext:   resourceKomodorKubernetesRead,
		DeleteContext: resourceKomodorKubernetesDelete,
	}
}

func resourceKomodorKubernetesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	clusterName := d.Get("cluster_name").(string)

	kubernetes, err := c.CreateKubernetesCluster(clusterName)
	if err != nil {
		return diag.Errorf("Error onboarding Kubernetes cluster: %s", err)
	}

	d.SetId(kubernetes.Id)

	log.Printf("[INFO] Kubernetes cluster created successfully: %s", clusterName)

	return resourceKomodorKubernetesRead(ctx, d, meta)
}

func resourceKomodorKubernetesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	clusterName := d.Id()

	log.Printf("[INFO] Deleting Kubernetes cluster: %s", clusterName)
	if err := c.DeleteKubernetesCluster(clusterName); err != nil {
		return diag.Errorf("Error deleting Kubernetes cluster: %s", err)
	}

	d.Set("cluster_name", "")

	return nil
}

func resourceKomodorKubernetesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	clusterName := d.Get("cluster_name").(string)

	kubernetes, statusCode, err := c.GetKubernetesCluster(clusterName)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Kubernetes cluster %s not found - removing from state", clusterName)
			d.SetId("")
		}

		return diag.Errorf("Error reading Kubernetes cluster: %s", err)
	}

	d.SetId(kubernetes.Id)

	return nil
}
