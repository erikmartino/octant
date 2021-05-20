/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/octant/internal/link"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// deploymentAppsV1 creates status for an v1/apps deployment. This is
// not the final implementation. It is included to generate output.
func deploymentAppsV1(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("deployment is nil")
	}

	deployment := &appsv1.Deployment{}

	if err := scheme.Scheme.Convert(object, deployment, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to apps/v1 deployment")
	}

	status := deployment.Status
	properties := []component.Property{{Label: "Deployment Strategy", Value: component.NewText(string(deployment.Spec.Strategy.Type))},
		{Label: "Selectors", Value: getSelector(deployment)}}

	switch {
	case status.Replicas == status.UnavailableReplicas:
		return ObjectStatus{
			nodeStatus: component.NodeStatusError,
			Details:    []component.Component{component.NewText("No replicas exist for this deployment")},
			Properties: properties,
		}, nil
	case status.Replicas == status.AvailableReplicas:
		return ObjectStatus{
			nodeStatus: component.NodeStatusOK,
			Details:    []component.Component{component.NewText("Deployment is OK")},
			Properties: properties,
		}, nil
	default:
		return ObjectStatus{
			nodeStatus: component.NodeStatusWarning,
			Details: []component.Component{
				component.NewText(
					fmt.Sprintf("Expected %d replicas, but %d are available",
						status.Replicas, status.AvailableReplicas))},
			Properties: properties,
		}, nil
	}
}

func getSelector(deployment *appsv1.Deployment) component.Component {
	if selector := deployment.Spec.Selector; selector != nil {
		var selectors []component.Selector

		for _, lsr := range selector.MatchExpressions {
			o, _ := component.MatchOperator(string(lsr.Operator))

			es := component.NewExpressionSelector(lsr.Key, o, lsr.Values)
			selectors = append(selectors, es)
		}

		for k, v := range selector.MatchLabels {
			ls := component.NewLabelSelector(k, v)
			selectors = append(selectors, ls)
		}

		return component.NewSelectors(selectors)
	}
	return component.NewText("None")
}
