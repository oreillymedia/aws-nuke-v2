package resources

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/machinelearning"
	"github.com/sirupsen/logrus"
)

const MachineLearningMLModelResource = "MachineLearningMLModel"

func init() {
	registry.Register(&registry.Registration{
		Name:   MachineLearningMLModelResource,
		Scope:  nuke.Account,
		Lister: &MachineLearningMLModelLister{},
	})
}

type MachineLearningMLModelLister struct{}

func (l *MachineLearningMLModelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := machinelearning.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &machinelearning.DescribeMLModelsInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeMLModels(params)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if strings.Contains(aerr.Message(), "AmazonML is no longer available to new customers") {
					logrus.Info("MachineLearningBranchPrediction: AmazonML is no longer available to new customers. Ignore if you haven't set it up.")
					return nil, nil
				}
			}
			return nil, err
		}

		for _, result := range output.Results {
			resources = append(resources, &MachineLearningMLModel{
				svc: svc,
				ID:  result.MLModelId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MachineLearningMLModel struct {
	svc *machinelearning.MachineLearning
	ID  *string
}

func (f *MachineLearningMLModel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteMLModel(&machinelearning.DeleteMLModelInput{
		MLModelId: f.ID,
	})

	return err
}

func (f *MachineLearningMLModel) String() string {
	return *f.ID
}
