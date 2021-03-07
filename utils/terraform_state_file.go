package utils

// TerraformSate ..
type TerraformSate struct {
	Resources []Resources `json:"resources"`
	Modules   []Modules   `json:"modules"`
}

// Resources ..
type Resources struct {
	Instances    []Instances `json:"instances"`
	Mode         string      `json:"mode"`
	ResourceType string      `json:"type"`
	ResourceName string      `json:"name"`
}

// Instances ..
type Instances struct {
	Mode       string        `json:"schema_version"`
	Attributes AttributeBody `json:"attributes"`
}

// AttributeBody ..
type AttributeBody struct {
	ID string `json:"id"`
}

// ResourceList ..
type ResourceList []Resource

// Resource ..
type Resource struct {
	ID           string
	ResourceType string
	ResourceName string
}

// Modules ..
type Modules struct {
	Resources map[string]ResourceBody `json:"resources"`
}

// ResourceBody .
type ResourceBody struct {
	Primary      map[string]PrimaryBody `json:"primary"`
	Provider     string                 `json:"provider"`
	ResourceType string                 `json:"type"`
}

// PrimaryBody .
type PrimaryBody struct {
	ID            string   `json:"id"`
	Location      string   `json:"location"`
	AttributeData []string `json:"attributes"`
}
