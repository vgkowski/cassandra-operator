# Building the operator

Kubernetes version: 1.9

This operator use the kubernetes code-generator for
  * clientset
  * informers
  * listers
  * deep copy
  
1. Clone the [code-generator](https://github.com/kubernetes/code-generator) repo in the `vendor/k8s.io` directory.
Be careful of cloning the branch matching the Kubernetes version
`git clone -b <branch> https://github.com/kubernetes/code-generator`
2. Run the script `vendor/k8s.io/code-generator/generate-groups.sh all github.com/vgkowski/cassandra-operator/pkg/client github.com/vgkowski/cassandra-operator/pkg/apis cassandra:v1`

 