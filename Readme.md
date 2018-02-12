# Building the operator

Kubernetes version: 1.9

This operator use the kubernetes code-generator for
  * clientset: used to manipulate objects defined in the CRD (cassandraCluster)
  * informers: cache for registering to events on objects defined in the CRD
  * listers
  * deep copy
  
1. Get dependencies packages in your $GOPATH with `go get k8s.io/kubernetes k8s.io/apimachinery`
1. Initialize the dependancies with `dep init` in the project root directory
1. Clone the [code-generator](https://github.com/kubernetes/code-generator) repo in the `vendor/k8s.io` directory.
Be careful of cloning the branch matching the Kubernetes version
`git clone -b <branch> https://github.com/kubernetes/code-generator`
2. Run the script `vendor/k8s.io/code-generator/generate-groups.sh all github.com/vgkowski/cassandra-operator/pkg/client github.com/vgkowski/cassandra-operator/pkg/apis cassandra:v1` or `hack/update-codegen.sh`


# Improvements

* Currently the relationship between native Kubernetes objects and CassandraClusters is done with the name which is equal. 
For multiple resources of the same type (services for instance) a postfix is added to the name ("<name>-internode" for internode service and "<name>-access" for access service)
A better option would be to use the `metadata.ownerReference` but it requires to search and get objects by this reference instead of by name.
For indexed objects like Pods or PVCs the search and get is done through the label `CassandraCluster` which contains the name of the cluster
* Implement proper Cassandra admin logic (current code isn't working)

 