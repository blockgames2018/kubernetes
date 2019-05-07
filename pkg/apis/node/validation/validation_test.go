/*
Copyright 2018 The Kubernetes Authors.

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

package validation

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/node"

	"github.com/stretchr/testify/assert"
)

func TestValidateRuntimeClass(t *testing.T) {
	tests := []struct {
		name        string
		rc          node.RuntimeClass
		expectError bool
	}{{
		name:        "invalid name",
		expectError: true,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "&!@#"},
			Handler:    "foo",
		},
	}, {
		name:        "invalid Handler name",
		expectError: true,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Handler:    "&@#$",
		},
	}, {
		name:        "invalid empty RuntimeClass",
		expectError: true,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "empty"},
		},
	}, {
		name:        "valid Handler",
		expectError: false,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Handler:    "bar-baz",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := ValidateRuntimeClass(&test.rc)
			if test.expectError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateRuntimeUpdate(t *testing.T) {
	old := node.RuntimeClass{
		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
		Handler:    "bar",
	}
	tests := []struct {
		name        string
		expectError bool
		old, new    node.RuntimeClass
	}{{
		name: "valid metadata update",
		old:  old,
		new: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "foo",
				Labels: map[string]string{"foo": "bar"},
			},
			Handler: "bar",
		},
	}, {
		name:        "invalid metadata update",
		expectError: true,
		old:         old,
		new: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "empty",
				ClusterName: "somethingelse", // immutable
			},
			Handler: "bar",
		},
	}, {
		name:        "invalid Handler update",
		expectError: true,
		old:         old,
		new: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Handler:    "somethingelse",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// So we don't need to write it in every test case...
			test.old.ObjectMeta.ResourceVersion = "1"
			test.new.ObjectMeta.ResourceVersion = "1"

			errs := ValidateRuntimeClassUpdate(&test.new, &test.old)
			if test.expectError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateTopology(t *testing.T) {
	tests := []struct {
		name       string
		topology   *node.Topology
		expectErrs int
	}{{
		name: "valid topology",
		topology: &node.Topology{
			NodeSelector: &core.NodeSelector{
				NodeSelectorTerms: []core.NodeSelectorTerm{{
					MatchExpressions: []core.NodeSelectorRequirement{{
						Key:      "valid",
						Operator: core.NodeSelectorOpExists,
					}},
				}},
			},
			Tolerations: []core.Toleration{{
				Key:      "valid",
				Operator: core.TolerationOpExists,
				Effect:   core.TaintEffectNoSchedule,
			}},
		},
	}, {
		name:     "empty topology",
		topology: &node.Topology{},
	}, {
		name: "invalid nodeSelector",
		topology: &node.Topology{
			NodeSelector: &core.NodeSelector{
				NodeSelectorTerms: []core.NodeSelectorTerm{{
					MatchExpressions: []core.NodeSelectorRequirement{{
						Key:      "not a valid key!!!",
						Operator: core.NodeSelectorOpExists,
					}},
				}},
			},
		},
		expectErrs: 1,
	}, {
		name: "invalid toleration",
		topology: &node.Topology{
			Tolerations: []core.Toleration{{
				Key:      "valid",
				Operator: core.TolerationOpExists,
				Effect:   core.TaintEffectNoSchedule,
			}, {
				Key:      "not a valid key!!!",
				Operator: core.TolerationOpExists,
				Effect:   core.TaintEffectNoSchedule,
			}},
		},
		expectErrs: 1,
	}, {
		name: "invalid topology",
		topology: &node.Topology{
			NodeSelector: &core.NodeSelector{
				NodeSelectorTerms: []core.NodeSelectorTerm{{
					MatchExpressions: []core.NodeSelectorRequirement{{
						Key:      "not a valid label key!!!",
						Operator: core.NodeSelectorOpExists,
					}},
				}},
			},
			Tolerations: []core.Toleration{{
				Key:      "valid",
				Operator: core.TolerationOpExists,
				Effect:   core.TaintEffectNoSchedule,
			}, {
				Key:      "not a valid toleration key!!!",
				Operator: core.TolerationOpExists,
				Effect:   core.TaintEffectNoSchedule,
			}},
		},
		expectErrs: 2,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rc := &node.RuntimeClass{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Handler:    "bar",
				Topology:   test.topology,
			}
			assert.Len(t, ValidateRuntimeClass(rc), test.expectErrs)
		})
	}
}
