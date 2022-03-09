package pod

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net"
	"strconv"
	"strings"
)

var ErrBadStatefulSetPodName = errors.New("malformed pod name, expected statefulset pattern")

type Pool interface {
	AddWorkerIfNotExists(idx int, name string, IP net.IP) bool
	RemoveWorkerByName(name string)
}

type Handler struct {
	state Pool
}

func NewHandler(st Pool) *Handler {
	return &Handler{
		state: st,
	}
}

func (h *Handler) Created(_ context.Context, obj runtime.Object) {
	pod := obj.(*api.Pod)

	if !isReadyPod(pod) {
		return
	}

	idx, err := podIndex(pod)
	if err != nil {
		log.Errorf("unable to get pod index %v", err)
		return
	}

	if !h.state.AddWorkerIfNotExists(idx, pod.Name, net.ParseIP(pod.Status.PodIP)) {
		return
	}

	log.Infof("Created Pod %s IP %s", pod.Name, pod.Status.PodIP)
}

func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	pod := new.(*api.Pod)

	// Quick deletion detection
	if hasDeletionTimestamp(pod) || isTerminated(pod) {
		h.Deleted(ctx, pod)
		return
	}

	if !isReadyPod(pod) {
		return
	}

	for _, cond := range pod.Status.Conditions {
		log.Debugf("Pod %s Condition type %s Status %s", pod.Name, cond.Type, cond.Status)
	}

	//diff := cmp.Diff(old, new)
	//cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
	//	return !unicode.IsGraphic(r)
	//})
	//fmt.Println("UPDATE POD diff: ", cleanDiff)

	if !isReady(pod) {
		log.Debugf("Pod is not ready yet %v", pod.Status.Conditions)
		return
	}

	idx, err := podIndex(pod)
	if err != nil {
		log.Errorf("unable to get pod index %v", err)
		return
	}

	if !h.state.AddWorkerIfNotExists(idx, pod.Name, net.ParseIP(pod.Status.PodIP)) {
		return
	}

	log.Infof("Updated Pod %s READY", pod.Name)

}

func (h *Handler) Deleted(_ context.Context, obj runtime.Object) {
	pod := obj.(*api.Pod)
	log.Infof("Deleted POD %s", pod.Name)

	h.state.RemoveWorkerByName(pod.Name)
}

func isReadyPod(pod *api.Pod) bool {
	if isHostNetworked(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is host networked.")
		return false
	} else if !hasIPAddress(pod) {
		log.WithField("pod", pod.Name).Debug("Pod does not have an IP address.")
		return false
	} else if !isScheduled(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not scheduled.")
		return false
	} else if !isRunning(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not running.")
		return false
	}
	return true
}

func isScheduled(pod *api.Pod) bool {
	return pod.Spec.NodeName != ""
}

func isHostNetworked(pod *api.Pod) bool {
	return pod.Spec.HostNetwork
}

func hasIPAddress(pod *api.Pod) bool {
	return pod.Status.PodIP != ""
}

func isRunning(pod *api.Pod) bool {
	return pod.Status.Phase == "Running"
}

func isReady(pod *api.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

func hasDeletionTimestamp(pod *api.Pod) bool {
	return pod.ObjectMeta.DeletionTimestamp != nil
}

func isTerminated(pod *api.Pod) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil {
			return true
		}
	}
	return false
}

func podIndex(pod *api.Pod) (int, error) {
	parts := strings.Split(pod.Name, "-")
	if len(parts) < 2 {
		return 0, ErrBadStatefulSetPodName
	}

	idx := parts[len(parts)-1]
	i, err := strconv.ParseInt(idx, 10, 64)

	return int(i), err
}
