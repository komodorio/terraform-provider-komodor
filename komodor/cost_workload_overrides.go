package komodor

const (
	workloadOverrideActionInclude = "include"
	workloadOverrideActionExclude = "exclude"
)

var workloadOverrideActions = []string{workloadOverrideActionInclude, workloadOverrideActionExclude}

type WorkloadOverrideBase struct {
	Action      string  `json:"action"`
	ClusterName string  `json:"clusterName"`
	Namespace   string  `json:"namespace"`
	Kind        string  `json:"kind"`
	Name        string  `json:"name"`
	PolicyId    *string `json:"policyId,omitempty"`
}

type WorkloadOverride struct {
	Id             string  `json:"id"`
	Action         string  `json:"action"`
	ClusterName    string  `json:"clusterName"`
	Namespace      string  `json:"namespace"`
	Kind           string  `json:"kind"`
	Name           string  `json:"name"`
	PolicyId       *string `json:"policyId,omitempty"`
	CreatedAt      *string `json:"createdAt,omitempty"`
	UpdatedAt      *string `json:"updatedAt,omitempty"`
	CreatedByEmail *string `json:"createdByEmail,omitempty"`
	LastUpdatedBy  *string `json:"lastUpdatedBy,omitempty"`
}

type WorkloadOverrideWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type WorkloadOverrideItemMeta struct {
	Warnings []WorkloadOverrideWarning `json:"warnings,omitempty"`
}

type WorkloadOverrideItemResponse struct {
	Meta                         *WorkloadOverrideItemMeta `json:"meta,omitempty"`
	PolicyScopeExceptionWorkload WorkloadOverride          `json:"policyScopeExceptionWorkload"`
}

type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type WorkloadOverrideListResponse struct {
	Meta                          PaginationMeta     `json:"meta"`
	PolicyScopeExceptionWorkloads []WorkloadOverride `json:"policyScopeExceptionWorkloads"`
}

type workloadOverrideTFData struct {
	Action      string
	ClusterName string
	Namespace   string
	Kind        string
	Name        string
	PolicyId    string

	Id        string
	CreatedBy string
	UpdatedBy string
	CreatedAt string
	UpdatedAt string
}
