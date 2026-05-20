package komodor

type RightSizingPolicyPercentile int32

const (
	N70 RightSizingPolicyPercentile = 70
	N80 RightSizingPolicyPercentile = 80
	N90 RightSizingPolicyPercentile = 90
	N95 RightSizingPolicyPercentile = 95
	N99 RightSizingPolicyPercentile = 99
)

var validPercentiles = []int{int(N70), int(N80), int(N90), int(N95), int(N99)}

func (p RightSizingPolicyPercentile) IsValid() bool {
	switch p {
	case N70, N80, N90, N95, N99:
		return true
	}
	return false
}

type RightSizingMultiScopePolicy struct {
	Name                string                       `json:"name"`
	Description         *string                      `json:"description,omitempty"`
	Priority            int32                        `json:"priority"`
	OptimizationPreset  string                       `json:"optimizationPreset"`
	Percentile          *RightSizingPolicyPercentile `json:"percentile,omitempty"`
	ApplyProtocol       string                       `json:"applyProtocol"`
	AllowQoSUpgrade     *string                      `json:"allowQoSUpgrade,omitempty"`
	AllowQoSUpgradeV2   *bool                        `json:"allowQoSUpgradeV2,omitempty"`
	AllowQoSDowngrade   *bool                        `json:"allowQoSDowngrade,omitempty"`
	AllowHpaRightSizing *bool                        `json:"allowHpaRightSizing,omitempty"`
	AllowRestart        *bool                        `json:"allowRestart,omitempty"`
	GuardRails          *PolicyGuardRails            `json:"guardRails,omitempty"`
	Scopes              []PolicyResourceScope        `json:"scopes"`
	Tags                *[]string                    `json:"tags,omitempty"`

	CreatedBy      *string `json:"createdBy,omitempty"`
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`
	CreatedAt      *string `json:"createdAt,omitempty"`
	UpdatedAt      *string `json:"updatedAt,omitempty"`
}

type GetMultiScopePolicyResponse struct {
	Id             string                      `json:"id"`
	Policy         RightSizingMultiScopePolicy `json:"policy"`
	CreatedBy      *string                     `json:"createdBy,omitempty"`
	LastModifiedBy *string                     `json:"lastModifiedBy,omitempty"`
	CreatedAt      *string                     `json:"createdAt,omitempty"`
	UpdatedAt      *string                     `json:"updatedAt,omitempty"`
}

type GetAllRightSizingPoliciesResponse struct {
	Policies []GetAllRightSizingPoliciesRow `json:"policies"`
}

type GetAllRightSizingPoliciesRow struct {
	Id                 string  `json:"id"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	ClusterName        string  `json:"clusterName"`
	OptimizationPreset *string `json:"optimizationPreset,omitempty"`
	Priority           *int32  `json:"priority,omitempty"`
	CreatedBy          string  `json:"createdBy"`
	LastModified       string  `json:"lastModified"`
	LastModifiedBy     *string `json:"lastModifiedBy,omitempty"`
}

type PolicyResourceScope struct {
	Clusters              *[]string      `json:"clusters,omitempty"`
	ClustersPatterns      *PolicyPattern `json:"clustersPatterns,omitempty"`
	Namespaces            *[]string      `json:"namespaces,omitempty"`
	NamespacesPatterns    *PolicyPattern `json:"namespacesPatterns,omitempty"`
	ResourceTypes         *[]string      `json:"resourceTypes,omitempty"`
	ResourceTypesPatterns *PolicyPattern `json:"resourceTypesPatterns,omitempty"`
	Workloads             *[]string      `json:"workloads,omitempty"`
	WorkloadsPatterns     *PolicyPattern `json:"workloadsPatterns,omitempty"`
}

type PolicyPattern struct {
	Include *string `json:"include,omitempty"`
	Exclude *string `json:"exclude,omitempty"`
}

type PolicyGuardRails struct {
	ManagedResources               PolicyGuardRailsManagedResources     `json:"managedResources"`
	AllowRightSizingUp             bool                                 `json:"allowRightSizingUp"`
	RightSizingConstraints         PolicyGuardRailsConstraints          `json:"rightSizingConstraints"`
	RightSizingAbsoluteConstraints *PolicyGuardRailsAbsoluteConstraints `json:"rightSizingAbsoluteConstraints,omitempty"`
	Buffer                         PolicyGuardRailsBuffer               `json:"buffer"`
}

type PolicyGuardRailsManagedResources struct {
	CpuRequests    bool `json:"cpuRequests"`
	CpuLimits      bool `json:"cpuLimits"`
	MemoryRequests bool `json:"memoryRequests"`
	MemoryLimits   bool `json:"memoryLimits"`
}

type PolicyGuardRailsConstraints struct {
	IncreaseCpuBy    ToggleableValue `json:"increaseCpuBy"`
	DecreaseCpuBy    ToggleableValue `json:"decreaseCpuBy"`
	IncreaseMemoryBy ToggleableValue `json:"increaseMemoryBy"`
	DecreaseMemoryBy ToggleableValue `json:"decreaseMemoryBy"`
}

type PolicyGuardRailsAbsoluteConstraints struct {
	CpuRequestMillicoresMin ToggleableValue `json:"cpuRequestMillicoresMin"`
	CpuRequestMillicoresMax ToggleableValue `json:"cpuRequestMillicoresMax"`
	CpuLimitsMillicoresMin  ToggleableValue `json:"cpuLimitsMillicoresMin"`
	CpuLimitsMillicoresMax  ToggleableValue `json:"cpuLimitsMillicoresMax"`
	MemoryRequestBytesMin   ToggleableValue `json:"memoryRequestBytesMin"`
	MemoryRequestBytesMax   ToggleableValue `json:"memoryRequestBytesMax"`
	MemoryLimitsBytesMin    ToggleableValue `json:"memoryLimitsBytesMin"`
	MemoryLimitsBytesMax    ToggleableValue `json:"memoryLimitsBytesMax"`
}

type PolicyGuardRailsBuffer struct {
	Cpu    ToggleableValue `json:"cpu"`
	Memory ToggleableValue `json:"memory"`
}

type ToggleableValue struct {
	Enabled bool  `json:"enabled"`
	Value   int64 `json:"value"`
}

type rightSizingPolicyTFData struct {
	Name                string
	Description         string
	Priority            int
	Scopes              []scopeTFData
	ApplyProtocol       string
	AllowRestart        bool
	AllowHpaRightSizing bool
	OptimizationPreset  string
	GuardRails          *guardRailsTFData
	Tags                []string
	ForceDelete         bool

	Id             string
	CreatedBy      string
	LastModifiedBy string
	CreatedAt      string
	UpdatedAt      string
}

type scopeTFData struct {
	Clusters              []string
	ClustersPatterns      *patternTFData
	Namespaces            []string
	NamespacesPatterns    *patternTFData
	ResourceTypes         []string
	ResourceTypesPatterns *patternTFData
	WorkloadNames         []string
	WorkloadNamesPatterns *patternTFData
}

type patternTFData struct {
	Include string
	Exclude string
}

type guardRailsTFData struct {
	Percentile          int
	ManagedResources    managedResourcesTFData
	AllowRightSizingUp  bool
	AllowQoSUpgrade     bool
	AllowQoSDowngrade   bool
	Constraints         constraintsTFData
	AbsoluteConstraints *absoluteConstraintsTFData
	Buffer              bufferTFData
}

type managedResourcesTFData struct {
	CpuRequests    bool
	CpuLimits      bool
	MemoryRequests bool
	MemoryLimits   bool
}

type constraintsTFData struct {
	IncreaseCpuBy    toggleableValueTFData
	DecreaseCpuBy    toggleableValueTFData
	IncreaseMemoryBy toggleableValueTFData
	DecreaseMemoryBy toggleableValueTFData
}

type toggleableValueTFData struct {
	Enabled bool
	Value   int
}

type absoluteConstraintsTFData struct {
	CpuRequestMillicoresMin toggleableValueTFData
	CpuRequestMillicoresMax toggleableValueTFData
	CpuLimitsMillicoresMin  toggleableValueTFData
	CpuLimitsMillicoresMax  toggleableValueTFData
	MemoryRequestBytesMin   toggleableValueTFData
	MemoryRequestBytesMax   toggleableValueTFData
	MemoryLimitsBytesMin    toggleableValueTFData
	MemoryLimitsBytesMax    toggleableValueTFData
}

type bufferTFData struct {
	Cpu    toggleableValueTFData
	Memory toggleableValueTFData
}
