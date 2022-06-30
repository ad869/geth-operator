package shared

// Resources is node compute and storage resources
type Resources struct {
	// CPU is cpu cores the node requires
	CPU string `json:"cpu,omitempty"`
	// CPULimit is cpu cores the node is limited to
	CPULimit string `json:"cpuLimit,omitempty"`
	// Memory is memmory requirements
	Memory string `json:"memory,omitempty"`
	// MemoryLimit is cpu cores the node is limited to
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// Storage is disk space storage requirements
	Storage string `json:"storage,omitempty"`
	// StorageClass is the volume storage class
	StorageClass *string `json:"storageClass,omitempty"`
}
