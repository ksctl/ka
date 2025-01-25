package controller

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StackState struct {
	Stacks map[string]AppState `json:"stacks"`
}
type AppState struct {
	Components map[string]ComponentState `json:"components"`
}
type ComponentState struct {
	Ver string `json:"version"`
}

func getConfigmap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ka-state",
			Namespace: "ka-system",
		},
		BinaryData: map[string][]byte{},
	}
}

func (r *StackReconciler) Save(ctx context.Context) error {
	cf := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: "ka-system", Name: "ka-state"}, cf); err != nil {
		if errors.IsNotFound(err) {
			_cf := getConfigmap()
			_cf.BinaryData["data"], err = json.Marshal(r.state)
			if err != nil {
				return err
			}
			if err := r.Create(ctx, _cf); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	var err error
	cf.BinaryData["data"], err = json.Marshal(r.state)
	if err != nil {
		return err
	}
	if err := r.Update(ctx, cf); err != nil {
		return err
	}
	return nil
}

func (r *StackReconciler) Load(ctx context.Context) error {
	cf := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: "ka-system", Name: "ka-state"}, cf); err != nil {
		if errors.IsNotFound(err) {
			r.state = &StackState{Stacks: map[string]AppState{}}
			return nil
		}
		return err
	}

	if cf.BinaryData == nil {
		r.state = &StackState{Stacks: map[string]AppState{}}
	} else {
		if _, ok := cf.BinaryData["data"]; !ok {
			r.state = &StackState{Stacks: map[string]AppState{}}
			return nil
		}
		r.state = &StackState{Stacks: map[string]AppState{}}
		if err := json.Unmarshal(cf.BinaryData["data"], r.state); err != nil {
			return err
		}
	}
	return nil
}

func (r *StackReconciler) WasStackInstalled(stackName string) bool {
	if _, ok := r.state.Stacks[stackName]; !ok {
		return false
	}
	return true
}

func (r *StackReconciler) WasComponentInstalled(stackName, componentName string) bool {
	if !r.WasStackInstalled(stackName) {
		return false
	}

	if _, okC := r.state.Stacks[stackName].Components[componentName]; !okC {
		return false
	}
	return true
}
