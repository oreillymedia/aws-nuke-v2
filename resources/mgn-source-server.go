package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/mgn"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

const MGNSourceServerResource = "MGNSourceServer"

func init() {
	registry.Register(&registry.Registration{
		Name:   MGNSourceServerResource,
		Scope:  nuke.Account,
		Lister: &MGNSourceServerLister{},
	})
}

type MGNSourceServerLister struct{}

func (l *MGNSourceServerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mgn.DescribeSourceServersInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeSourceServers(params)
		if err != nil {
			if IsAWSError(err, mgn.ErrCodeUninitializedAccountException) {
				logrus.Info("MGNSourceServer: Account not initialized for Application Migration Service. Ignore if you haven't set it up.")
				return nil, nil
			}
			return nil, err
		}

		for _, sourceServer := range output.Items {
			resources = append(resources, &MGNSourceServer{
				svc:            svc,
				sourceServerID: sourceServer.SourceServerID,
				arn:            sourceServer.Arn,
				tags:           sourceServer.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNSourceServer struct {
	svc            *mgn.Mgn
	sourceServerID *string
	arn            *string
	tags           map[string]*string
}

func (f *MGNSourceServer) Remove(_ context.Context) error {
	// Disconnect source server from service first before delete
	if _, err := f.svc.DisconnectFromService(&mgn.DisconnectFromServiceInput{
		SourceServerID: f.sourceServerID,
	}); err != nil {
		return err
	}

	_, err := f.svc.DeleteSourceServer(&mgn.DeleteSourceServerInput{
		SourceServerID: f.sourceServerID,
	})

	return err
}

func (f *MGNSourceServer) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("SourceServerID", f.sourceServerID)
	properties.Set("ARN", f.arn)

	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}

	return properties
}

func (f *MGNSourceServer) String() string {
	return *f.sourceServerID
}
