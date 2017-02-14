# kapprover

_kapprover_ is a tool meant to be deployed in Kubernetes clusters that uses the
[TLS client certificate bootstrapping] flow for kubelets. It will then monitor
and automatically approves Certificate Signing Requests submitted by kubelets,
based on the the policy selected at startup.

As of today, a single approval policy, called `always` exists, and approves any
pending CSRs without making any kind of validation besides checking that the
requester's user/group are respectively `kubelet-bootstrap` /
`system:kubelet-bootstrap`. Long term, we hope to support advanced policies,
such as validating that the requester is part of a given AWS's AutoScalingGroup.

The easiest way to deploy _kapprover_ is to use the provided `deployment.yaml`
resource.

[TLS client certificate bootstrapping]: https://kubernetes.io/docs/admin/kubelet-tls-bootstrapping/
