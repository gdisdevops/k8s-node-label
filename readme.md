# K8S Node Label

K8S Node Label is a small tool to label Nodes based on their role in a kubernetes cluster.

## Control Plane Node

Starting from 1.16 the --node-labels flag in kubelet is deprecated for security reasons. This tool runs inside the cluster and marks a node when it starts with the
"node-role.kubernetes.io/control-plane" label.

Two more options are available:

* Exclude node from loadbalancing. When this feature is enabled a master node also gets excluded from loadbalancers created via "LoadBalancer"-Type Service.
* Exclude node from eviction. When this feature is enabled a master node also gets marked to exclude from eviction. "Master workloads will not be evicted if the master is NotReady for longer than the grace period" [1]

For more details please see: https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md

[1] https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md

### Legacy master node support
Since kubernetes 1.20 taint and label `node-role.kubernetes.io/master` are deprecated. Cluster administrators should migrate both taints and labels to `node-role.kubernetes.io/control-plane`. 
During transition period you can control behaviour of K8s Node Label with 2 flags:
* `-control-plane-legacy-label` - if set to true will add `node-role.kubernetes.io/master` or `node-role.kubernetes.io/spot-master` label next to `node-role.kubernetes.io/control-plane`
* `-control-plane-taint` - by default K8s Node Label detects control-plane nodes by checking `node-role.kubernetes.io/control-plane` taint. Using this flag it can be switched to look for different flag (for example legacy: `node-role.kubernetes.io/master`)

## Worker Node

To have a similar solution available for worker nodes it also marks worker nodes with the "node-role.kubernetes.io/worker" label.

## Spot instances

Additionally this tool supports spot instance role to mark nodes in case they are based on spot instances.
Therefore it assigns the "node-role.kubernetes.io/spot-worker" label to nodes, that are part of a spot request.
Currently only aws is supported, but it can be extended. Pull requests for further providers are welcome :-)

## Custom node-role labels

It is possible to label your nodes with role taken from custom label (for example `custom-label`). To enable this node use this tool with parameter `custom-role-label` equal to the name of that custom label. Then nodes with this `custom-label` will be also labelled with corresponding `node-role.kubernetes.io/*` label.

For example, node with `custom-label=special-node` will be also labelled with `node-role.kubernetes.io/special-node`.

## Karpenter nodes

Nodes labeled with `karpenter.sh/nodepool` will be also labelled with `node-role.kubernetes.io/karpenter`. This behaviour can be turned off with `-karpenter=false` flag.
