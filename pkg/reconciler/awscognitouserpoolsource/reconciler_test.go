/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awscognitouserpoolsource

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"knative.dev/eventing/pkg/reconciler/source"
	fakek8sinjectionclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awscognitouserpoolsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	. "github.com/triggermesh/aws-event-sources/pkg/reconciler/testing"
)

func TestReconcileSource(t *testing.T) {
	adapterCfg := &adapterConfig{
		Image:   "registry/image:tag",
		configs: &source.EmptyVarsGenerator{},
	}

	var (
		ctor      = reconcilerCtor(adapterCfg)
		src       = newEventSource()
		adapterFn = adapterDeploymentBuilder(src, adapterCfg)
	)

	TestReconcile(t, ctor, src, adapterFn)
}

// reconcilerCtor returns a Ctor for a AWSCognitoUserPoolSource Reconciler.
func reconcilerCtor(cfg *adapterConfig) Ctor {
	return func(t *testing.T, ctx context.Context, ls *Listers) controller.Reconciler {
		base := common.GenericDeploymentReconciler{
			SinkResolver: resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
			Lister:       ls.GetDeploymentLister().Deployments,
			Client:       fakek8sinjectionclient.Get(ctx).AppsV1().Deployments,
			PodClient:    fakek8sinjectionclient.Get(ctx).CoreV1().Pods,
		}

		r := &Reconciler{
			base:       base,
			adapterCfg: cfg,
		}

		return reconcilerv1alpha1.NewReconciler(ctx, logging.FromContext(ctx),
			fakeinjectionclient.Get(ctx), ls.GetAWSCognitoUserPoolSourceLister(),
			controller.GetEventRecorder(ctx), r)
	}
}

// newEventSource returns a test source object with a minimal set of pre-filled attributes.
func newEventSource(skipCEAtrributes ...interface{}) *v1alpha1.AWSCognitoUserPoolSource {
	src := &v1alpha1.AWSCognitoUserPoolSource{
		Spec: v1alpha1.AWSCognitoUserPoolSourceSpec{
			ARN: NewARN(cognitoidentityprovider.ServiceName, "userpool/triggermeshtest"),
			Credentials: v1alpha1.AWSSecurityCredentials{
				AccessKeyID: v1alpha1.ValueFromField{
					ValueFromSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "keyId",
					},
				},
				SecretAccessKey: v1alpha1.ValueFromField{
					ValueFromSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "secret",
					},
				},
			},
		},
	}

	Populate(src)

	return src
}
