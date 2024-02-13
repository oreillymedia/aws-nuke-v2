package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type ElasticacheReplicationGroup struct {
	svc        *elasticache.ElastiCache
	groupID    *string
	createTime *time.Time
}

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticacheReplicationGroupResource,
		Scope:  nuke.Account,
		Lister: &ElasticacheReplicationGroupLister{},
	})
}

type ElasticacheReplicationGroupLister struct{}

func (l *ElasticacheReplicationGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticache.New(opts.Session)
	var resources []resource.Resource

	params := &elasticache.DescribeReplicationGroupsInput{MaxRecords: aws.Int64(100)}

	for {
		resp, err := svc.DescribeReplicationGroups(params)
		if err != nil {
			return nil, err
		}

		for _, replicationGroup := range resp.ReplicationGroups {
			resources = append(resources, &ElasticacheReplicationGroup{
				svc:        svc,
				groupID:    replicationGroup.ReplicationGroupId,
				createTime: replicationGroup.ReplicationGroupCreateTime,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

func (i *ElasticacheReplicationGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", i.groupID)

	if i.createTime != nil {
		properties.Set("CreateTime", i.createTime.Format(time.RFC3339))
	}

	return properties
}

func (i *ElasticacheReplicationGroup) Remove() error {
	params := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: i.groupID,
	}

	_, err := i.svc.DeleteReplicationGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *ElasticacheReplicationGroup) String() string {
	return *i.groupID
}
