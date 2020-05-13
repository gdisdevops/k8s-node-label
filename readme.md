# K8S Node Label

K8S Node Label is a small tool to label Nodes based on their role in a kubernetes cluster.

## Master Node

Starting from 1.16 the --node-labels flag in kubelet is deprecated for security reasons. This tool runs inside the cluster and marks a node when it starts with the
"node-role.kubernetes.io/master" label.

Two more options are available:

* Exclude node from loadbalancing. When this feature is enabled a master node also gets excluded from loadbalancers created via "LoadBalancer"-Type Service.
* Exclude node from eviction. When this feature is enabled a master node also gets marked to exclude from eviction. "Master workloads will not be evicted if the master is NotReady for longer than the grace period" [1]

For more details please see: https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md

[1] https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md

## Worker Node

To have a similar solution available for worker nodes it also marks worker nodes with the "node-role.kubernetes.io/worker" label.

## Spot instances

Additionally this tool supports spot instance role to mark nodes in case they are based on spot instances.
Therefore it assigns the "node-role.kubernetes.io/spot-worker" label to nodes, that are part of a spot request.
Currently only aws is supported, but it can be extended. Pull requests for further providers are welcome :-)