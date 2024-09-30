resource  "komodor_kubernetes" "k8s_cluster" {
    cluster_name = "cluster-test-diff"
}

data "komodor_kubernetes" "data_k8s_cluster" {
  cluster_name = komodor_kubernetes.k8s_cluster.cluster_name
}

// the output below represents the API key used to onboard a cluster to app.komodor.com
output "cluster_api_key" {
  value = komodor_kubernetes.k8s_cluster.id
}