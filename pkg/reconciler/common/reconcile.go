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

package common

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/reconciler"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/event"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/semantic"
)

// List of annotations set on Knative Serving objects by the Knative Serving
// admission webhook.
var knativeServingAnnotations = []string{
	serving.CreatorAnnotation,
	serving.UpdaterAnnotation,
}

// AdapterDeploymentBuilderFunc builds a Deployment object for a source's adapter.
type AdapterDeploymentBuilderFunc func(sinkURI *apis.URL) *appsv1.Deployment

// ReconcileSource reconciles an event source type.
func (r *GenericDeploymentReconciler) ReconcileSource(ctx context.Context, adb AdapterDeploymentBuilderFunc) reconciler.Event {
	src := v1alpha1.SourceFromContext(ctx)

	src.GetStatusManager().CloudEventAttributes = CreateCloudEventAttributes(
		src.AsEventSource(), src.GetEventTypes())

	sinkURI, err := r.resolveSinkURL(ctx)
	if err != nil {
		src.GetStatusManager().MarkNoSink()
		return controller.NewPermanentError(reconciler.NewEvent(corev1.EventTypeWarning,
			ReasonBadSinkURI, "Could not resolve sink URI: %s", err))
	}
	src.GetStatusManager().MarkSink(sinkURI)

	if err := r.reconcileAdapter(ctx, adb(sinkURI)); err != nil {
		return fmt.Errorf("failed to reconcile adapter: %w", err)
	}
	return nil
}

// resolveSinkURL resolves the URL of a sink reference.
func (r *GenericDeploymentReconciler) resolveSinkURL(ctx context.Context) (*apis.URL, error) {
	src := v1alpha1.SourceFromContext(ctx)
	sink := *src.GetSink()

	if sinkRef := &src.GetSink().Ref; *sinkRef != nil && (*sinkRef).Namespace == "" {
		(*sinkRef).Namespace = src.GetNamespace()
	}

	return r.SinkResolver.URIFromDestinationV1(ctx, sink, src)
}

// reconcileAdapter reconciles the state of the source's adapter.
func (r *GenericDeploymentReconciler) reconcileAdapter(ctx context.Context, desiredAdapter *appsv1.Deployment) error {
	src := v1alpha1.SourceFromContext(ctx)

	currentAdapter, err := r.getOrCreateAdapter(ctx, desiredAdapter)
	if err != nil {
		src.GetStatusManager().PropagateDeploymentAvailability(ctx, currentAdapter, r.PodClient(src.GetNamespace()))
		return err
	}

	currentAdapter, err = r.syncAdapterDeployment(ctx, currentAdapter, desiredAdapter)
	if err != nil {
		return fmt.Errorf("failed to synchronize adapter Deployment: %w", err)
	}
	src.GetStatusManager().PropagateDeploymentAvailability(ctx, currentAdapter, r.PodClient(src.GetNamespace()))

	return nil
}

// getOrCreateAdapter returns the existing adapter Deployment for a given
// source, or creates it if it is missing.
func (r *GenericDeploymentReconciler) getOrCreateAdapter(ctx context.Context, desiredAdapter *appsv1.Deployment) (*appsv1.Deployment, error) {
	src := v1alpha1.SourceFromContext(ctx)

	adapter, err := r.FindAdapter(src)
	switch {
	case apierrors.IsNotFound(err):
		adapter, err = r.Client(src.GetNamespace()).Create(ctx, desiredAdapter, metav1.CreateOptions{})
		if err != nil {
			return nil, reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedAdapterCreate,
				"Failed to create adapter Deployment %q: %s", desiredAdapter.Name, err)
		}
		event.Normal(ctx, ReasonAdapterCreate, "Created adapter Deployment %q", adapter.Name)

	case err != nil:
		return nil, fmt.Errorf("failed to get adapter Deployment from cache: %w", err)
	}

	return adapter, nil
}

// FindAdapter returns the adapter Deployment for a given source if it exists.
func (r *GenericDeploymentReconciler) FindAdapter(src v1alpha1.EventSource) (*appsv1.Deployment, error) {
	o, err := findAdapter(r, src)
	if err != nil {
		return nil, err
	}
	return o.(*appsv1.Deployment), nil
}

// syncAdapterDeployment synchronizes the desired state of an adapter Deployment
// against its current state in the running cluster.
func (r *GenericDeploymentReconciler) syncAdapterDeployment(ctx context.Context,
	currentAdapter, desiredAdapter *appsv1.Deployment) (*appsv1.Deployment, error) {

	if semantic.Semantic.DeepEqual(desiredAdapter, currentAdapter) {
		return currentAdapter, nil
	}

	// resourceVersion must be returned to the API server unmodified for
	// optimistic concurrency, as per Kubernetes API conventions
	desiredAdapter.ResourceVersion = currentAdapter.ResourceVersion

	// (fake Clientset) preserve status to avoid resetting conditions
	desiredAdapter.Status = currentAdapter.Status

	adapter, err := r.Client(currentAdapter.Namespace).Update(ctx, desiredAdapter, metav1.UpdateOptions{})
	if err != nil {
		return nil, reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedAdapterUpdate,
			"Failed to update adapter Deployment %q: %s", desiredAdapter.Name, err)
	}
	event.Normal(ctx, ReasonAdapterUpdate, "Updated adapter Deployment %q", adapter.Name)

	return adapter, nil
}

// AdapterServiceBuilderFunc builds a Service object for a source's adapter.
type AdapterServiceBuilderFunc func(sinkURI *apis.URL) *servingv1.Service

// ReconcileSource reconciles an event source type.
func (r *GenericServiceReconciler) ReconcileSource(ctx context.Context, adb AdapterServiceBuilderFunc) reconciler.Event {
	src := v1alpha1.SourceFromContext(ctx)

	src.GetStatusManager().CloudEventAttributes = CreateCloudEventAttributes(
		src.AsEventSource(), src.GetEventTypes())

	sinkURI, err := r.resolveSinkURL(ctx)
	if err != nil {
		src.GetStatusManager().MarkNoSink()
		return controller.NewPermanentError(reconciler.NewEvent(corev1.EventTypeWarning,
			ReasonBadSinkURI, "Could not resolve sink URI: %s", err))
	}
	src.GetStatusManager().MarkSink(sinkURI)

	if err := r.reconcileAdapter(ctx, adb(sinkURI)); err != nil {
		return fmt.Errorf("failed to reconcile adapter: %w", err)
	}
	return nil
}

// resolveSinkURL resolves the URL of a sink reference.
func (r *GenericServiceReconciler) resolveSinkURL(ctx context.Context) (*apis.URL, error) {
	src := v1alpha1.SourceFromContext(ctx)
	sink := *src.GetSink()

	if sinkRef := &src.GetSink().Ref; *sinkRef != nil && (*sinkRef).Namespace == "" {
		(*sinkRef).Namespace = src.GetNamespace()
	}

	return r.SinkResolver.URIFromDestinationV1(ctx, sink, src)
}

// reconcileAdapter reconciles the state of the source's adapter.
func (r *GenericServiceReconciler) reconcileAdapter(ctx context.Context, desiredAdapter *servingv1.Service) error {
	src := v1alpha1.SourceFromContext(ctx)

	currentAdapter, err := r.getOrCreateAdapter(ctx, desiredAdapter)
	if err != nil {
		src.GetStatusManager().PropagateServiceAvailability(currentAdapter)
		return err
	}

	currentAdapter, err = r.syncAdapterService(ctx, currentAdapter, desiredAdapter)
	if err != nil {
		return fmt.Errorf("failed to synchronize adapter Service: %w", err)
	}
	src.GetStatusManager().PropagateServiceAvailability(currentAdapter)

	return nil
}

// getOrCreateAdapter returns the existing adapter Service for a given
// source, or creates it if it is missing.
func (r *GenericServiceReconciler) getOrCreateAdapter(ctx context.Context, desiredAdapter *servingv1.Service) (*servingv1.Service, error) {
	src := v1alpha1.SourceFromContext(ctx)

	adapter, err := r.FindAdapter(src)
	switch {
	case apierrors.IsNotFound(err):
		adapter, err = r.Client(src.GetNamespace()).Create(ctx, desiredAdapter, metav1.CreateOptions{})
		if err != nil {
			return nil, reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedAdapterCreate,
				"Failed to create adapter Service %q: %s", desiredAdapter.Name, err)
		}
		event.Normal(ctx, ReasonAdapterCreate, "Created adapter Service %q", adapter.Name)

	case err != nil:
		return nil, fmt.Errorf("failed to get adapter Service from cache: %w", err)
	}

	return adapter, nil
}

// FindAdapter returns the adapter Service for a given source if it exists.
func (r *GenericServiceReconciler) FindAdapter(src v1alpha1.EventSource) (*servingv1.Service, error) {
	o, err := findAdapter(r, src)
	if err != nil {
		return nil, err
	}
	return o.(*servingv1.Service), nil
}

// syncAdapterService synchronizes the desired state of an adapter Service
// against its current state in the running cluster.
func (r *GenericServiceReconciler) syncAdapterService(ctx context.Context,
	currentAdapter, desiredAdapter *servingv1.Service) (*servingv1.Service, error) {

	if semantic.Semantic.DeepEqual(desiredAdapter, currentAdapter) {
		return currentAdapter, nil
	}

	// resourceVersion must be returned to the API server unmodified for
	// optimistic concurrency, as per Kubernetes API conventions
	desiredAdapter.ResourceVersion = currentAdapter.ResourceVersion

	// immutable Knative annotations must be preserved
	for _, ann := range knativeServingAnnotations {
		if val, ok := currentAdapter.Annotations[ann]; ok {
			metav1.SetMetaDataAnnotation(&desiredAdapter.ObjectMeta, ann, val)
		}
	}

	// (fake Clientset) preserve status to avoid resetting conditions
	desiredAdapter.Status = currentAdapter.Status

	adapter, err := r.Client(currentAdapter.Namespace).Update(ctx, desiredAdapter, metav1.UpdateOptions{})
	if err != nil {
		return nil, reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedAdapterUpdate,
			"Failed to update adapter Service %q: %s", desiredAdapter.Name, err)
	}
	event.Normal(ctx, ReasonAdapterUpdate, "Updated adapter Service %q", adapter.Name)

	return adapter, nil
}

// findAdapter returns the adapter object for a given source if it exists.
func findAdapter(genericReconciler interface{}, src v1alpha1.EventSource) (metav1.Object, error) {
	// the combination of standard labels {name,instance} is unique and
	// immutable
	sel := labels.SelectorFromSet(labels.Set{
		AppNameLabel:     AdapterName(src),
		AppInstanceLabel: src.GetName(),
	})

	var objs []metav1.Object
	var gr schema.GroupResource

	switch r := genericReconciler.(type) {
	case *GenericDeploymentReconciler:
		depls, err := r.Lister(src.GetNamespace()).List(sel)
		if err != nil {
			return nil, err
		}

		for _, d := range depls {
			objs = append(objs, d)
		}

		gr = appsv1.Resource("deployment")

	case *GenericServiceReconciler:
		svcs, err := r.Lister(src.GetNamespace()).List(sel)
		if err != nil {
			return nil, err
		}

		for _, s := range svcs {
			objs = append(objs, s)
		}

		gr = servingv1.Resource("service")
	}

	for _, obj := range objs {
		if metav1.IsControlledBy(obj, src) {
			return obj, nil
		}
	}

	return nil, newNotFoundForSelector(gr, sel)
}

// newNotFoundForSelector returns an error which indicates that no object of
// the type matching the given GroupResource was found for the given label
// selector.
func newNotFoundForSelector(gr schema.GroupResource, sel labels.Selector) *apierrors.StatusError {
	err := apierrors.NewNotFound(gr, "")
	err.ErrStatus.Message = fmt.Sprint(gr, " not found for selector ", sel)
	return err
}
