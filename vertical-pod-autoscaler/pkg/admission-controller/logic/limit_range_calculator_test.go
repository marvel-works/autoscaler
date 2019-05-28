/*
Copyright 2019 The Kubernetes Authors.

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

package logic

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
)

const defaultNamespace = "default"

func TestNewNoopLimitsChecker(t *testing.T) {
	nlc := NewNoopLimitsCalculator()
	limitRange, err := nlc.GetContainerLimitRangeItem(defaultNamespace)
	assert.NoError(t, err)
	assert.Nil(t, limitRange)
}

func TestNoLimitRange(t *testing.T) {
	cs := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(cs, 0)
	lc, err := NewLimitsRangeCalculator(factory)

	if assert.NoError(t, err) {
		limitRange, err := lc.GetContainerLimitRangeItem(defaultNamespace)
		assert.NoError(t, err)
		assert.Nil(t, limitRange)
	}
}

func TestGetContainerLimitRangeItem(t *testing.T) {
	containerLimitRangeWithMax := test.LimitRange().WithName("default").WithNamespace(defaultNamespace).WithType(apiv1.LimitTypeContainer).WithMax(test.Resources("2", "2")).Get()
	containerLimitRangeWithDefault := test.LimitRange().WithName("default").WithNamespace(defaultNamespace).WithType(apiv1.LimitTypeContainer).WithDefault(test.Resources("2", "2")).Get()
	testCases := []struct {
		name           string
		limitRanges    []runtime.Object
		expectedErr    error
		expectedLimits *apiv1.LimitRangeItem
	}{
		{
			name: "no matching limit ranges",
			limitRanges: []runtime.Object{
				test.LimitRange().WithName("different-namespace").WithNamespace("different").WithType(apiv1.LimitTypeContainer).WithMax(test.Resources("2", "2")).Get(),
				test.LimitRange().WithName("differen-type").WithNamespace(defaultNamespace).WithType(apiv1.LimitTypePersistentVolumeClaim).WithMax(test.Resources("2", "2")).Get(),
			},
			expectedErr:    nil,
			expectedLimits: nil,
		},
		{
			name: "matching container limit range",
			limitRanges: []runtime.Object{
				containerLimitRangeWithMax,
			},
			expectedErr:    nil,
			expectedLimits: &containerLimitRangeWithMax.Spec.Limits[0],
		},
		{
			name: "with default value",
			limitRanges: []runtime.Object{
				containerLimitRangeWithDefault,
			},
			expectedErr:    nil,
			expectedLimits: &containerLimitRangeWithDefault.Spec.Limits[0],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset(tc.limitRanges...)
			factory := informers.NewSharedInformerFactory(cs, 0)
			lc, err := NewLimitsRangeCalculator(factory)
			if assert.NoError(t, err) {
				limitRange, err := lc.GetContainerLimitRangeItem(defaultNamespace)
				if tc.expectedErr == nil {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
				assert.Equal(t, tc.expectedLimits, limitRange)
			}
		})

	}
}
