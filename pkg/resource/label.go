package resource

// NameGetter comment
type NameGetter interface {
	GetName() string
	GetComponent() string
	GetInstance() string
	GetPartOf() string
	GetVersion() string
}

// GetLabelsForCR comment
func GetLabelsForCR(cr NameGetter) map[string]string {
	labels := make(map[string]string)

	name := cr.GetName()
	if name != "" {
		labels["app.kubernetes.io/name"] = name
		labels["app"] = name
	}

	component := cr.GetComponent()
	if component != "" {
		labels["app.kubernetes.io/component"] = component
	}

	partOf := cr.GetPartOf()
	if partOf != "" {
		labels["app.kubernetes.io/part-of"] = partOf
	}

	instance := cr.GetInstance()
	if instance != "" {
		labels["app.kubernetes.io/instance"] = instance
	}

	version := cr.GetVersion()
	if version != "" {
		labels["app.kubernetes.io/version"] = version
	}

	return labels
}
