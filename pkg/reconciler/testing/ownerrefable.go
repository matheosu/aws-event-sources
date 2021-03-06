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

package testing

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"knative.dev/pkg/kmeta"
)

// NewOwnerRefable returns a OwnerRefable with the given attributes.
func NewOwnerRefable(name string, gvk schema.GroupVersionKind, uid types.UID) *FakeOwnerRefable {
	return &FakeOwnerRefable{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  uid,
		},
		GroupVersionKind: gvk,
	}
}

// DummyOwnerRefable returns a OwnerRefable with dummy attributes.
func DummyOwnerRefable() *FakeOwnerRefable {
	return &FakeOwnerRefable{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake",
			UID:  types.UID("00000000-0000-0000-0000-000000000000"),
		},
		GroupVersionKind: schema.GroupVersionKind{
			Group:   "fakegroup.fakeapi",
			Version: "v0",
			Kind:    "FakeKind",
		},
	}
}

var _ kmeta.OwnerRefable = (*FakeOwnerRefable)(nil)

// FakeOwnerRefable implements OwnerRefable.
type FakeOwnerRefable struct {
	metav1.ObjectMeta
	schema.GroupVersionKind
}

// GetGroupVersionKind returns the GroupVersionKind from the object.
func (o *FakeOwnerRefable) GetGroupVersionKind() schema.GroupVersionKind {
	return o.GroupVersionKind
}
