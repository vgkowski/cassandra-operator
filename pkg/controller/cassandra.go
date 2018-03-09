package controller

import "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"

func (c *Controller) deleteCassandraCluster(name string) error {
	// deleted the statefulset
	err := c.DeleteStatefulSet(name)
	if err != nil {
		return err
	}
	// delete the pvc
	err = c.DeletePVC(name)
	if err != nil {
		return err
	}
	// delete the service
	err = c.DeleteService(name)
	if err != nil {
		return err
	}
	return err
}

func (c *Controller) createOrUpdateCassandraCluster(cc *v1.CassandraCluster) error {
	// reconciliates the statefulset
	repair,err := c.CreateOrUpdateStatefulSet(cc)
	if err != nil {
		return err
	}

	// if required, launch a repair
	if repair == true {

	}
	// reconciliates the service
	err = c.CreateOrUpdateService(cc)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) fullRepair(cc *v1.CassandraCluster) error {
	// get the statefulset
	sts, err := c.statefulsetsLister.StatefulSets(c.namespace).Get(cc.Name)
	if err != nil {
		return err
	}
	// get the pods owned by the statefulset
	pods := c.podLister.Pods(c.namespace).List()

	// iterate over the range and execucte "nodetool repair -pr"
	c.ExecCmd(pod,"nodetool repair -pr")
}

