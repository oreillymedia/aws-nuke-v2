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

const MGNJobResource = "MGNJob"

func init() {
	resource.Register(resource.Registration{
		Name:   MGNJobResource,
		Scope:  nuke.Account,
		Lister: &MGNJobLister{},
	})
}

type MGNJobLister struct{}

func (l *MGNJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mgn.DescribeJobsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeJobs(params)
		if err != nil {
			if IsAWSError(err, mgn.ErrCodeUninitializedAccountException) {
				logrus.Info("MGNJob: Account not initialized for Application Migration Service. Ignore if you haven't set it up.")
				return nil, nil
			}
			return nil, err
		}

		for _, job := range output.Items {
			resources = append(resources, &MGNJob{
				svc:   svc,
				jobID: job.JobID,
				arn:   job.Arn,
				tags:  job.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNJob struct {
	svc   *mgn.Mgn
	jobID *string
	arn   *string
	tags  map[string]*string
}

func (f *MGNJob) Remove(_ context.Context) error {
	_, err := f.svc.DeleteJob(&mgn.DeleteJobInput{
		JobID: f.jobID,
	})

	return err
}

func (f *MGNJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobID", f.jobID)
	properties.Set("ARN", f.arn)

	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}

	return properties
}

func (f *MGNJob) String() string {
	return *f.jobID
}
