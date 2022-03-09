package operator

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type filter struct {
	label  string
	object runtime.Object
}

func NewFilter(watchedLabel string, objectType runtime.Object) Filter {
	return &filter{
		label:  watchedLabel,
		object: objectType,
	}
}

func (f *filter) Object() runtime.Object {
	return f.object
}

func (f *filter) Validate(obj runtime.Object) error {
	v, err := f.hasWatchedLabel(obj)
	if err != nil {
		log.Debugf("unable to get object %T app label, error %v", obj, err)
		return err
	}

	if !v {
		l, err := f.getAppLabel(obj)
		if err != nil {
			return fmt.Errorf("filter validate error, not watched label, app label not found, skip object, error %v", err)
		}
		return fmt.Errorf("label %s not handled", l)
	}

	return nil
}

func (f *filter) hasWatchedLabel(obj runtime.Object) (bool, error) {
	label, err := f.getAppLabel(obj)
	if err != nil {
		return false, fmt.Errorf("meta app label error: %v", err)
	}

	return label == f.label, nil
}

func (f *filter) getAppLabel(obj runtime.Object) (string, error) {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("meta accessor error: %v", err)
	}

	labels := acc.GetLabels()
	v, ok := labels["app"]
	if !ok {
		return "", ErrNoAppLabelFound
	}

	return v, nil
}
