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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Kubernetes cluster",
			},
		},
		CreateContext: resourceKomodorKubernetesCreate,
		ReadContext:   resourceKomodorKubernetesRead,
		DeleteContext: resourceKomodorKubernetesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKomodorKubernetesImport,
		},
		Description: "Creates a new Komodor integration with a Kubernetes cluster.\n\n" +
			"This integration allows Komodor to monitor and analyze the cluster's activity.",
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

	if err := d.Set("cluster_name", ""); err != nil {
		return diag.Errorf("error clearing cluster_name: %s", err)
	}

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
			return nil
		}

		return diag.Errorf("Error reading Kubernetes cluster: %s", err)
	}

	d.SetId(kubernetes.Id)

	return nil
}

// resourceKomodorKubernetesImport imports an existing cluster integration by
// cluster name, not by resource ID. The resource ID is the integration's apiKey,
// which resourceKomodorKubernetesRead resolves from cluster_name — so cluster_name
// must be seeded here before Read runs.
func resourceKomodorKubernetesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("cluster_name", d.Id()); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
