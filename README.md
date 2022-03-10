# K8s Swarm Operator

Swarm Operator solves scenarios were Collaborative Workload consumption is required, scenarios  where parallel job processing does not increase overall performance.
They are special cases as they are not the typical deployment scenarios, but there are many scenarios like that, from stream consuming where packets cannot be marked as processing (in-flight), so that, multiple consumers will just reprocess the same source.
As example, dedicated streams as real time video, long run jobs (database backups) and many others.

## Project Goals
THe Operator starts reading from a list of stream source that must be assigned in a deterministic way over a pool of workers. The overall idea here is that workload assignation over the workers is deterministic, computed many times will return the same association.
An easy aprox to get this scenario is using StatefulSets, as they are ordered, scaling Up/Down will be addresses always at the end of the statefulset.
The Operator listens for labeled statefulsets an its associated pods, that way is able to create a pool of members that will assume its part in the worload, to do so the operator balances all the jobs to match the total number of workers.
Operator computes job association to all workers in the pool giving as a result an updated confimap that is used by the workers to get is workload tasks.

First iteration was achieving exactly that, workload assignation to any worker in a shared worker config file based on configmps, on that stage workers are expected to watch config changes and perform hot reload.
Config Map changes need time to be propagated to all mounted volumes, it's just an eventual consistent mechanism, so we know that it should happen but not foresee exactly when.

A second iteration has been done, the operator calculates what are the workers that must be refreshed due to config changes, as a result the operator schedules Pod deletion to speed up workers refresh state.

Operator is able to scale up/down workers pool, write assignations and control pool state.

## Project State
Still on POC consolidation, tons of things need to be improved, starting from testing coverage... project is evolving iteration by iteration, one of the near goals is to use CRDs to reflect the Operator state in a direct way.

## Project base
Right now implementation is quite lazy, as config convergence between assignations and worker state is not direct queried, so that, worker pods do not know nothing about operator existence. In fact they just read its own config part.
At the end the Operator tries to ensure that pods are using the last version of the configmap but it is not querying pods to check their versions.

## Further Iterations
- Add CRD to reflect operator state
- Add QoS policy on the workload streams. This opens the door to dedicate workload jobs to unique pods, behaving as some kind of resource reservation.
- Allow deterministic assignations in top of Deployments, statefulset as an option
- Kustomize deploy

## Development workflow
The whole system can be tested using minikube, a few notes to make it work:
-Start minikube cluster
```
minikube start --addons=ingress --cpus=2 --cni=flannel --install-addons=true --kubernetes-version=stable --memory=6g --driver=docker
```
- Apply required manifests (in order), namespace, rbac, configmaps, operator and statefulset.
```
kubectl get pods -w

NAME                                READY   STATUS    RESTARTS   AGE
swarm-controller-7bcc789689-sj628   1/1     Running   0          10h
swarm-worker-0                      1/1     Running   0          10h
swarm-worker-1                      1/1     Running   0          10h

```
Check worker assignations from worker config:
```
kubectl describe cm swarm-worker-config 
Namespace:    swarm
Labels:       <none>
Annotations:  <none>

Data
====
config.yml:
----
workloads:
  swarm-worker-0:
    jobs:
      - stream:xxrtve1
      - stream:xxrtve2
      - stream:zrtve2:new
      - stream:zrtve1:new
      - stream:zrtve0:new
      - stream:cctv0:updated
      - stream:history:new
      - stream:foo:new
      - stream:xxctv3:new
      - stream:cctv3:updated
      - stream:xxctv0:updated
      - stream:xxctv10:updated
      - stream:xxctv11:updated
      - stream:xxctv12:updated
    version: 0
  swarm-worker-1:
    jobs:
      - stream:xxctv13:updated
      - stream:xxctv14:updated
      - stream:yxctv1:updated
      - stream:yxctv2:updated
      - stream:yxctv3:updated
      - stream:xabcn0:updated
      - stream:xacb01:updated
      - stream:xacb02:updated
      - stream:xacb03:updated
      - stream:xacb04:updated
      - stream:sportnews0:updated
      - stream:cars:new
      - stream:cars:updated 
```

You can even introspect what are the configs mounted as volumes in the worker pods:
```
kubectl exec -ti swarm-worker-0 -- cat /app/config/config.yml
```
Scaling Up/Down
```
`kuectl scale statefulset swarm-worker --replicas=3 -n swarm 
```
On Scaling Up Operator will pick up pool changes, balance workloads over the new updated workers pool, updated workers shared config (swarm-worker-config configmap), it will schedule to reboot worker0 and 1, so that, after refresh cycle all workers will belong to the same config version.

## Build docker image
Bare in mind that minikube images are expected to be local, if you check k8s manifest you will see that imagePullPolicy is Never in all the cases. Before building images ensure minikube is pointing docker env:
```
eval $(minikube docker-env)
```
### Controller image
```
docker build -t swarm-controller . --build-arg SERVICE=controller --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```
### Worker image
```
docker build -t swarm-worker . --build-arg SERVICE=worker --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```

## Minikube rollout
```
docker tag swarm-controller:latest swarm-controller:a4c0d90019d8
```
```
kubectl set image deployment/swarm-controller swarm-controller=swarm-controller:01a68000eb00 -n swarm
kubectl set image statefulset/swarm-worker swarm-worker=swarm-worker:2ddc673264ed -n swarm
```
```
kubectl rollout restart deployment/swarm-controller
kubectl rollout restart statefulset/swarm-worker
```

## Pending development things
- Clean configs
  - local (Single node)
  - worker configMap as placeHolder (empty until controller assignation)
- Track controller config version
  - Using configMap
- controller 
  - http interface [Declare on receipts]
    - useful to introspect doing port-forwarding
- Config Issues ¿?
  - Separated builds
    - worker with config
    - controller does not need it
  - but:
    - we can consider controller config as the keys to share (example)
- Kustomize entry point
  - kustomize build swarm-operator/envs/dev &> dev.yaml
  - It can dump config.yaml to configmap
