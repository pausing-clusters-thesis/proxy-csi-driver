package controllerhelpers

import (
	"fmt"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

type objectForFinalizersPatch struct {
	objectMetaForFinalizersPatch `json:"metadata"`
}

type objectMetaForFinalizersPatch struct {
	ResourceVersion string   `json:"resourceVersion"`
	Finalizers      []string `json:"finalizers"`
}

func PrepareAddFinalizerPatch(obj metav1.Object, finalizer string) ([]byte, error) {
	newFinalizers := append(append(make([]string, 0, len(obj.GetFinalizers())+1), obj.GetFinalizers()...), finalizer)
	patch, err := json.Marshal(objectForFinalizersPatch{
		objectMetaForFinalizersPatch: objectMetaForFinalizersPatch{
			ResourceVersion: obj.GetResourceVersion(),
			Finalizers:      newFinalizers,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("can't marshal object for add finalizer patch: %w", err)
	}

	return patch, nil
}

func PrepareRemoveFinalizerPatch(obj metav1.Object, finalizer string) ([]byte, error) {
	skipFinalizerFunc := func(s string) bool {
		return s != finalizer
	}
	newFinalizers := slices.Filter(obj.GetFinalizers(), skipFinalizerFunc)

	patch, err := json.Marshal(objectForFinalizersPatch{
		objectMetaForFinalizersPatch: objectMetaForFinalizersPatch{
			ResourceVersion: obj.GetResourceVersion(),
			Finalizers:      newFinalizers,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("can't marshal object for remove finalizer patch: %w", err)
	}

	return patch, nil
}

func HasFinalizer(obj metav1.Object, finalizer string) bool {
	return slices.Contains(obj.GetFinalizers(), finalizer)
}
