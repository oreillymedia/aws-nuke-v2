package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2ImageResource = "EC2Image"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2ImageResource,
		Scope:  nuke.Account,
		Lister: &EC2ImageLister{},
	})
}

type EC2ImageLister struct{}

func (l *EC2ImageLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	params := &ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String("self"),
		},
	}
	resp, err := svc.DescribeImages(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.Images {
		resources = append(resources, &EC2Image{
			svc:          svc,
			creationDate: *out.CreationDate,
			id:           *out.ImageId,
			name:         *out.Name,
			tags:         out.Tags,
		})
	}

	return resources, nil
}

type EC2Image struct {
	svc          *ec2.EC2
	creationDate string
	id           string
	name         string
	tags         []*ec2.Tag
}

func (e *EC2Image) Remove(_ context.Context) error {
	_, err := e.svc.DeregisterImage(&ec2.DeregisterImageInput{
		ImageId: &e.id,
	})
	return err
}

func (e *EC2Image) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("CreationDate", e.creationDate)
	properties.Set("Name", e.name)

	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *EC2Image) String() string {
	return e.id
}
