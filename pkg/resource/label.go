package resource

// CRLabelGetter is an interface which contains getter functions
// to retrieve data from the custom resource.
type CRLabelGetter interface {
	GetLabelName() string
	GetLabelComponent() string
	GetLabelInstance() string
	GetLabelPartOf() string
	GetLabelVersion() string
}

// GetLabelsForCR retrieves labels for the custom resource
func GetLabelsForCR(cr CRLabelGetter) map[string]string {
	labels := make(map[string]string)

	name := cr.GetLabelName()
	if name != "" {
		labels["app.kubernetes.io/name"] = name
		labels["app"] = name
	}

	component := cr.GetLabelComponent()
	if component != "" {
		labels["app.kubernetes.io/component"] = component
	}

	partOf := cr.GetLabelPartOf()
	if partOf != "" {
		labels["app.kubernetes.io/part-of"] = partOf
	}

	instance := cr.GetLabelInstance()
	if instance != "" {
		labels["app.kubernetes.io/instance"] = instance
	}

	version := cr.GetLabelVersion()
	if version != "" {
		labels["app.kubernetes.io/version"] = version
	}

	return labels
}
