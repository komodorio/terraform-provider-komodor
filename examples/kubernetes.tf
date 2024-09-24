resource  "komodor_kubernetes" "k8s_cluster" {
    cluster_name = "cluster-test-diff"
}

data "komodor_kubernetes" "data_k8s_cluster" {
  cluster_name = komodor_kubernetes.k8s_cluster.cluster_name
}

output "cluster_api_key" {
  value = komodor_kubernetes.k8s_cluster.id
}

output "data_cluster_api_key" {
  value = data.komodor_kubernetes.data_k8s_cluster.id
}