# K8S Master Label

K8S Master Label is a small tool to label Master nodes after an upgrade to 1.16 with the master role.

Starting from 1.16 the --node-labels flag disables for security reasons to set the "master" role, this tool runs inside the cluster and marks a node when it starts with the
"node-role.kubernetes.io/master" label.

Two more options are available:

* Exclude node from loadbalancing. When this feature is enabled a master node also gets excluded from loadbalancers created via "LoadBalancer"-Type Service.
* Exclude node from eviction. When this feature is enabled a master node also gets marked to exclude from eviction. "Master workloads will not be evicted if the master is NotReady for longer than the grace period" [1]

For more details please see: https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md

[1] https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md