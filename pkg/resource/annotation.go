package resource

// CRAnnotationGetter is an interface which contains getter functions
// to retrieve data from the custom resource.
type CRAnnotationGetter interface {
	GetAnnotationVcsUri() string
	GetAnnotationVcsRef() string
}

// GetAnnotationsForCR retrieves annotations for the custom resource.
func GetAnnotationsForCR(cr CRAnnotationGetter) map[string]string {
	annotations := make(map[string]string)

	uri := cr.GetAnnotationVcsUri()
	if uri != "" {
		annotations["app.openshift.io/vcs-uri"] = uri
	}

	ref := cr.GetAnnotationVcsRef()
	if ref != "" {
		annotations["app.openshift.io/vcs-ref"] = ref
	}

	return annotations
}
