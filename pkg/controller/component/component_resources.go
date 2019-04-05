package component

import (
	"context"
	"fmt"

	v1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	componentsv1alpha1 "github.com/redhat-developer/devconsole-operator/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-operator/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newImageStreamFromDocker(cr *componentsv1alpha1.Component) *imagev1.ImageStream {
	labels := resource.GetLabelsForCR(cr)

	if _, ok := buildTypeImages[cr.Spec.BuildType]; !ok {
		return nil
	}
	return &imagev1.ImageStream{ObjectMeta: metav1.ObjectMeta{
		Name:      cr.Spec.BuildType,
		Namespace: cr.Namespace,
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
					Name: buildTypeImages[cr.Spec.BuildType],
				},
			},
		},
	}}
}

func newOutputImageStream(cr *componentsv1alpha1.Component) *imagev1.ImageStream {
	labels := resource.GetLabelsForCR(cr)
	return &imagev1.ImageStream{ObjectMeta: metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Labels:    labels,
	}}
}

func (r *ReconcileComponent) newBuildConfig(cr *componentsv1alpha1.Component, builder *imagev1.ImageStream, gitSource *devconsoleapi.GitSource) *buildv1.BuildConfig {
	labels := resource.GetLabelsForCR(cr)
	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: gitSource.Spec.URL,
			Ref: gitSource.Spec.Ref,
		},
		Type: buildv1.BuildSourceGit,
	}
	incremental := true

	// Check if secrets provided exist or not
	if gitSource.Spec.SecretRef != nil && gitSource.Spec.SecretRef.Name != "" {
		u := corev1.Secret{}

		err := r.client.Get(context.TODO(), client.ObjectKey{
			Namespace: cr.Namespace,
			Name:      gitSource.Spec.SecretRef.Name,
		}, &u)
		if err != nil {
			log.Error(err, fmt.Sprintf("Failed to get provided secret please check if secrets %s present or not", gitSource.Spec.SecretRef.Name))
		} else {
			// Since error is nil, moving ahead to add secret reference in buildConfig
			buildSource.SourceSecret = &corev1.LocalObjectReference{
				Name: gitSource.Spec.SecretRef.Name,
			}
		}
	}

	return &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{Name: cr.Name, Namespace: cr.Namespace, Labels: labels},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: cr.Name + ":latest",
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

func newDeploymentConfig(cr *componentsv1alpha1.Component, output *imagev1.ImageStream) *v1.DeploymentConfig {
	labels := resource.GetLabelsForCR(cr)
	return &v1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
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
					Name:      cr.Name,
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  output.Name,
						Image: output.Name + ":latest",
						Ports: []corev1.ContainerPort{{ // do we plan to have several ports exposed?
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
						},
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

func newService(cr *componentsv1alpha1.Component, port int32) (*corev1.Service, error) {
	labels := resource.GetLabelsForCR(cr)
	if port > 65536 || port < 1024 {
		return nil, fmt.Errorf("port %d is out of range [1024-65535]", port)
	}
	var svcPorts []corev1.ServicePort
	svcPort := corev1.ServicePort{
		Name:       cr.Name + "-tcp",
		Port:       port,
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(int(port)),
	}
	svcPorts = append(svcPorts, svcPort)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": cr.Name,
			},
		},
	}
	return svc, nil
}

func newRoute(cr *componentsv1alpha1.Component) *routev1.Route {
	labels := resource.GetLabelsForCR(cr)

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{IntVal: cr.Spec.Port, StrVal: fmt.Sprintf("%d-tcp", cr.Spec.Port)},
			},
		},
	}
	return route
}
