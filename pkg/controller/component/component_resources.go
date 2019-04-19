package component

import (
	"fmt"

	v1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"

	"github.com/redhat-developer/devconsole-operator/pkg/resource"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newImageStreamFromDocker(cp *devconsoleapi.Component) *imagev1.ImageStream {
	labels := resource.GetLabelsForCR(cp)

	if _, ok := buildTypeImages[cp.Spec.BuildType]; !ok {
		return nil
	}
	return &imagev1.ImageStream{ObjectMeta: metav1.ObjectMeta{
		Name:      cp.Spec.BuildType,
		Namespace: cp.Namespace,
		Labels:    labels,
	}, Spec: imagev1.ImageStreamSpec{
		LookupPolicy: imagev1.ImageLookupPolicy{
			Local: false,
		},
		Tags: []imagev1.TagReference{
			{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: buildTypeImages[cp.Spec.BuildType],
				},
			},
		},
	}}
}

func newOutputImageStream(cp *devconsoleapi.Component) *imagev1.ImageStream {
	labels := resource.GetLabelsForCR(cp)
	return &imagev1.ImageStream{ObjectMeta: metav1.ObjectMeta{
		Name:      cp.Name,
		Namespace: cp.Namespace,
		Labels:    labels,
	}}
}

func newBuildConfig(cp *devconsoleapi.Component, builder *imagev1.ImageStream, gitSource *devconsoleapi.GitSource, secret *corev1.Secret) *buildv1.BuildConfig {
	labels := resource.GetLabelsForCR(cp)
	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: gitSource.Spec.URL,
			Ref: gitSource.Spec.Ref,
		},
		Type: buildv1.BuildSourceGit,
	}
	if secret != nil {
		buildSource.SourceSecret = &corev1.LocalObjectReference{
			Name: secret.Name,
		}
	}
	incremental := true
	return &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{Name: cp.Name, Namespace: cp.Namespace, Labels: labels},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: cp.Name + ":latest",
					},
				},
				Source: buildSource,
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      builder.Name + ":latest",
							Namespace: builder.Namespace,
						},
						Incremental: &incremental,
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				{
					Type: "ConfigChange",
				}, {
					Type:        "ImageChange",
					ImageChange: &buildv1.ImageChangeTrigger{},
				},
			},
		},
	}
}

func newDeploymentConfig(cp *devconsoleapi.Component, output *imagev1.ImageStream, containerPorts []corev1.ContainerPort) *v1.DeploymentConfig {
	labels := resource.GetLabelsForCR(cp)
	if containerPorts == nil {
		containerPorts = []corev1.ContainerPort{{
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		}}
	}
	return &v1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cp.Name,
			Namespace: cp.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentConfigSpec{
			Strategy: v1.DeploymentStrategy{
				Type: v1.DeploymentStrategyTypeRecreate,
			},
			Replicas: 1,
			Selector: labels,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cp.Name,
					Namespace: cp.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  output.Name,
						Image: output.Name + ":latest",
						Ports: containerPorts,
					},
					},
				},
			},
			Triggers: []v1.DeploymentTriggerPolicy{{
				Type: v1.DeploymentTriggerOnConfigChange,
			}, {
				Type: v1.DeploymentTriggerOnImageChange,
				ImageChangeParams: &v1.DeploymentTriggerImageChangeParams{
					Automatic: true,
					ContainerNames: []string{
						output.Name,
					},
					From: corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: output.Name + ":latest",
					},
				},
			},
			},
		},
	}
}

func newService(cp *devconsoleapi.Component, port int32) (*corev1.Service, error) {
	labels := resource.GetLabelsForCR(cp)
	if port > 65536 || port < 1024 {
		return nil, fmt.Errorf("port %d is out of range [1024-65535]", port)
	}
	var svcPorts []corev1.ServicePort
	svcPort := corev1.ServicePort{
		Name:       cp.Name + "-tcp",
		Port:       port,
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(int(port)),
	}
	svcPorts = append(svcPorts, svcPort)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cp.Name,
			Namespace: cp.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": cp.Name,
			},
		},
	}
	return svc, nil
}

func newRoute(cp *devconsoleapi.Component) *routev1.Route {
	labels := resource.GetLabelsForCR(cp)

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cp.Name,
			Namespace: cp.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cp.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{IntVal: cp.Spec.Port, StrVal: fmt.Sprintf("%d-tcp", cp.Spec.Port)},
			},
		},
	}
	return route
}

func newSecret(cr *devconsoleapi.Component, gitSource *devconsoleapi.GitSource) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gitSource.Spec.SecretRef.Name,
			Namespace: cr.Namespace,
		},
	}
	return secret
}
