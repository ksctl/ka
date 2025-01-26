package wasm

import (
	"context"

	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kwasmPlus "github.com/ksctl/ka/internal/stacks/wasm/kwasm"
	spinkubeStandard "github.com/ksctl/ka/internal/stacks/wasm/spinkube"
)

func ShouldPerformAdditionalProcessing(stackID stack.ID) bool {
	return stackID == kwasmPlus.SKU || stackID == spinkubeStandard.SKU
}

func AfterInstall(ctx context.Context, c client.Client) error {
	l := log.FromContext(ctx)
	nodes := &corev1.NodeList{}
	if err := c.List(ctx, nodes, &client.ListOptions{}); err != nil {
		return err
	}

	for _, node := range nodes.Items {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			l.Info("Annotating node", "targetNodeName", node.Name)

			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
			} else {
				if _, ok := node.Annotations["kwasm.sh/kwasm-node"]; ok {
					l.Info("Skipped, Node already annotated", "targetNodeName", node.Name)
					return nil
				}
			}
			node.Annotations["kwasm.sh/kwasm-node"] = "true"

			if err := c.Update(ctx, &node, &client.UpdateOptions{}); err != nil {
				l.Error(err, "Failed to annotate node, retrying", "targetNodeName", node.Name)
				return err
			}
			l.Info("Annotated node", "targetNodeName", node.Name)

			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func AfterRemoval(ctx context.Context, c client.Client) error {
	l := log.FromContext(ctx)
	nodes := &corev1.NodeList{}
	if err := c.List(ctx, nodes, &client.ListOptions{}); err != nil {
		return err
	}

	for _, node := range nodes.Items {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if node.Annotations != nil {
				l.Info("Removing annotatation from node", "targetNodeName", node.Name)
				if _, ok := node.Annotations["kwasm.sh/kwasm-node"]; ok {
					delete(node.Annotations, "kwasm.sh/kwasm-node")

					if err := c.Update(ctx, &node, &client.UpdateOptions{}); err != nil {
						l.Error(err, "Failed to remove annotatation from node, retrying", "targetNodeName", node.Name)
						return err
					}
					l.Info("Removed Annotation from node", "targetNodeName", node.Name)
				} else {
					l.Info("Skipped, Node doesn't have the annotation", "targetNodeName", node.Name)
				}
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}
