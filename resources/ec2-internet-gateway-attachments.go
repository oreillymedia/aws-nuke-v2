package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type EC2InternetGatewayAttachment struct {
	svc        *ec2.EC2
	vpcId      *string
	vpcOwnerID *string
	vpcTags    []*ec2.Tag
	igwId      *string
	igwOwnerID *string
	igwTags    []*ec2.Tag
	defaultVPC bool
}

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2InternetGatewayAttachmentResource,
		Scope:  nuke.Account,
		Lister: &EC2InternetGatewayAttachmentLister{},
		DeprecatedAliases: []string{
			"EC2InternetGatewayAttachement",
		},
	})
}

type EC2InternetGatewayAttachmentLister struct{}

func (l *EC2InternetGatewayAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, vpc := range resp.Vpcs {
		params := &ec2.DescribeInternetGatewaysInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("attachment.vpc-id"),
					Values: []*string{vpc.VpcId},
				},
			},
		}

		resp, err := svc.DescribeInternetGateways(params)
		if err != nil {
			return nil, err
		}

		for _, igw := range resp.InternetGateways {
			resources = append(resources, &EC2InternetGatewayAttachment{
				svc:        svc,
				vpcId:      vpc.VpcId,
				vpcOwnerID: vpc.OwnerId,
				vpcTags:    vpc.Tags,
				igwId:      igw.InternetGatewayId,
				igwOwnerID: igw.OwnerId,
				igwTags:    igw.Tags,
				defaultVPC: *vpc.IsDefault,
			})
		}
	}

	return resources, nil
}

type EC2InternetGatewayAttachment struct {
	svc        *ec2.EC2
	vpcID      *string
	vpcOwnerID *string
	vpcTags    []*ec2.Tag
	igwID      *string
	igwOwnerID *string
	igwTags    []*ec2.Tag
	defaultVPC bool
}

func (e *EC2InternetGatewayAttachment) Remove(_ context.Context) error {
	params := &ec2.DetachInternetGatewayInput{
		VpcId:             e.vpcID,
		InternetGatewayId: e.igwID,
	}

	_, err := e.svc.DetachInternetGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2InternetGatewayAttachment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DefaultVPC", e.defaultVPC)
	properties.SetWithPrefix("vpc", "OwnerID", e.vpcOwnerID)
	properties.SetWithPrefix("igw", "OwnerID", e.igwOwnerID)

	for _, tagValue := range e.igwTags {
		properties.SetTagWithPrefix("igw", tagValue.Key, tagValue.Value)
	}
	for _, tagValue := range e.vpcTags {
		properties.SetTagWithPrefix("vpc", tagValue.Key, tagValue.Value)
	}
	properties.Set("DefaultVPC", e.defaultVPC)
	properties.SetPropertyWithPrefix("vpc", "OwnerID", e.vpcOwnerID)
	properties.SetPropertyWithPrefix("igw", "OwnerID", e.igwOwnerID)
	return properties
}

func (e *EC2InternetGatewayAttachment) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(e.igwID), ptr.ToString(e.vpcID))
}
