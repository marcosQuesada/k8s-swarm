# K8s Swarm

Collaborative Workload consumption
- where parallel job processing does not increase overall performance
- at least once semantics
- Long running jobs
  - Database snapshot/backup

Example:
- Unbounded stream consumption
  - Video
  - Iot...

## Evolutions
- Jobs with QoS, enabling resource reservation in the pod space
  - moving keys from:
    - cnbc -> cnbc:100 (cnbc consumer will occupy 100% of pod resources)


## Build docker image
### Controller image
```
docker build -t swarm-controller . --build-arg SERVICE=controller --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```
### Worker image
```
docker build -t swarm-worker . --build-arg SERVICE=worker --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```


## Pending
- Clean configs
  - local (Single node)
  - worker configMap as placeHolder (empty until controller assignation)
- Track controller config version
  - Using configMap
- 
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



	// @TODO:
	// - conciliator Scaling Up /Down
	// - create expectations that can be asserted
```go

//podEventsQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
//podEventHandler := operator.NewEventHandler(podFilter, podEventsQueue)


// @TODO: REMOVE IT!
func (c *controller) backoffRetry(ev Event, err error) {
	if c.queue.NumRequeues(ev) >= 5 {
		log.Errorf("controller.processNextItem: key %s with error %v, no more retries", ev.GetKey(), err)
		c.queue.Forget(ev)
		utilruntime.HandleError(err)
		return
	}

	c.queue.AddRateLimited(ev)
}

```
