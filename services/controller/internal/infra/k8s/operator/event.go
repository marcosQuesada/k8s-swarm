package operator

import "k8s.io/apimachinery/pkg/runtime"

type Event interface {
	GetKey() string
}

type event struct {
	key    string
	obj    runtime.Object
}

func (e *event) GetKey() string {
	return e.key
}

type updateEvent struct {
	key    string
	oldObj runtime.Object
	newObj runtime.Object
}

func (e *updateEvent) GetKey() string {
	return e.key
}
