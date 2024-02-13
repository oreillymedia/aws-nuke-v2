package resources

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/machinelearning"
	"github.com/sirupsen/logrus"
)

const MachineLearningEvaluationResource = "MachineLearningEvaluation"

func init() {
	registry.Register(&registry.Registration{
		Name:   MachineLearningEvaluationResource,
		Scope:  nuke.Account,
		Lister: &MachineLearningEvaluationLister{},
	})
}

type MachineLearningEvaluationLister struct{}

func (l *MachineLearningEvaluationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := machinelearning.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &machinelearning.DescribeEvaluationsInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeEvaluations(params)
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
			resources = append(resources, &MachineLearningEvaluation{
				svc: svc,
				ID:  result.EvaluationId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MachineLearningEvaluation struct {
	svc *machinelearning.MachineLearning
	ID  *string
}

func (f *MachineLearningEvaluation) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEvaluation(&machinelearning.DeleteEvaluationInput{
		EvaluationId: f.ID,
	})

	return err
}

func (f *MachineLearningEvaluation) String() string {
	return *f.ID
}
