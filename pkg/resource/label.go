package resource

type NameGetter interface {
	GetName() string
}

func GetLabelsForCR(cr NameGetter) map[string]string {
	labels := map[string]string{
		"app": cr.GetName(),
	}
	return labels
}
