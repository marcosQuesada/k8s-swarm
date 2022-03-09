package statefulset

import (
	"context"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Pool interface {
	UpdateExpectedSize(size int)
}

type Handler struct{
	state Pool
}

func NewHandler(st Pool) *Handler {
	return &Handler{
		state: st,
	}
}

func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	ss := obj.(*api.StatefulSet)

	log.Infof("Created StatefulSet %s replicas %d", ss.Name, uint64(*ss.Spec.Replicas))

	h.state.UpdateExpectedSize(int(*ss.Spec.Replicas))
}

func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	ss := new.(*api.StatefulSet)

	// @TODO Replicas Spec
	// CurrentReplicas
	// AvailableReplicas:
	// ReadyReplicas
	h.state.UpdateExpectedSize(int(*ss.Spec.Replicas))

/*	diff := cmp.Diff(old, new)
	cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
	fmt.Println("UPDATE STATEFULSET diff: ", cleanDiff)*/
}

func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	ss := obj.(*api.StatefulSet)

	log.Infof("Deleted StatefulSet %s", ss.Name)
	h.state.UpdateExpectedSize(0)
}
