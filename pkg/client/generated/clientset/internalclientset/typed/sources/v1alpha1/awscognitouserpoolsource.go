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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	scheme "github.com/triggermesh/aws-event-sources/pkg/client/generated/clientset/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AWSCognitoUserPoolSourcesGetter has a method to return a AWSCognitoUserPoolSourceInterface.
// A group's client should implement this interface.
type AWSCognitoUserPoolSourcesGetter interface {
	AWSCognitoUserPoolSources(namespace string) AWSCognitoUserPoolSourceInterface
}

// AWSCognitoUserPoolSourceInterface has methods to work with AWSCognitoUserPoolSource resources.
type AWSCognitoUserPoolSourceInterface interface {
	Create(*v1alpha1.AWSCognitoUserPoolSource) (*v1alpha1.AWSCognitoUserPoolSource, error)
	Update(*v1alpha1.AWSCognitoUserPoolSource) (*v1alpha1.AWSCognitoUserPoolSource, error)
	UpdateStatus(*v1alpha1.AWSCognitoUserPoolSource) (*v1alpha1.AWSCognitoUserPoolSource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.AWSCognitoUserPoolSource, error)
	List(opts v1.ListOptions) (*v1alpha1.AWSCognitoUserPoolSourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSCognitoUserPoolSource, err error)
	AWSCognitoUserPoolSourceExpansion
}

// aWSCognitoUserPoolSources implements AWSCognitoUserPoolSourceInterface
type aWSCognitoUserPoolSources struct {
	client rest.Interface
	ns     string
}

// newAWSCognitoUserPoolSources returns a AWSCognitoUserPoolSources
func newAWSCognitoUserPoolSources(c *SourcesV1alpha1Client, namespace string) *aWSCognitoUserPoolSources {
	return &aWSCognitoUserPoolSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the aWSCognitoUserPoolSource, and returns the corresponding aWSCognitoUserPoolSource object, and an error if there is any.
func (c *aWSCognitoUserPoolSources) Get(name string, options v1.GetOptions) (result *v1alpha1.AWSCognitoUserPoolSource, err error) {
	result = &v1alpha1.AWSCognitoUserPoolSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AWSCognitoUserPoolSources that match those selectors.
func (c *aWSCognitoUserPoolSources) List(opts v1.ListOptions) (result *v1alpha1.AWSCognitoUserPoolSourceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.AWSCognitoUserPoolSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested aWSCognitoUserPoolSources.
func (c *aWSCognitoUserPoolSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a aWSCognitoUserPoolSource and creates it.  Returns the server's representation of the aWSCognitoUserPoolSource, and an error, if there is any.
func (c *aWSCognitoUserPoolSources) Create(aWSCognitoUserPoolSource *v1alpha1.AWSCognitoUserPoolSource) (result *v1alpha1.AWSCognitoUserPoolSource, err error) {
	result = &v1alpha1.AWSCognitoUserPoolSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		Body(aWSCognitoUserPoolSource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a aWSCognitoUserPoolSource and updates it. Returns the server's representation of the aWSCognitoUserPoolSource, and an error, if there is any.
func (c *aWSCognitoUserPoolSources) Update(aWSCognitoUserPoolSource *v1alpha1.AWSCognitoUserPoolSource) (result *v1alpha1.AWSCognitoUserPoolSource, err error) {
	result = &v1alpha1.AWSCognitoUserPoolSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		Name(aWSCognitoUserPoolSource.Name).
		Body(aWSCognitoUserPoolSource).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *aWSCognitoUserPoolSources) UpdateStatus(aWSCognitoUserPoolSource *v1alpha1.AWSCognitoUserPoolSource) (result *v1alpha1.AWSCognitoUserPoolSource, err error) {
	result = &v1alpha1.AWSCognitoUserPoolSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		Name(aWSCognitoUserPoolSource.Name).
		SubResource("status").
		Body(aWSCognitoUserPoolSource).
		Do().
		Into(result)
	return
}

// Delete takes name of the aWSCognitoUserPoolSource and deletes it. Returns an error if one occurs.
func (c *aWSCognitoUserPoolSources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *aWSCognitoUserPoolSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched aWSCognitoUserPoolSource.
func (c *aWSCognitoUserPoolSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSCognitoUserPoolSource, err error) {
	result = &v1alpha1.AWSCognitoUserPoolSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("awscognitouserpoolsources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
