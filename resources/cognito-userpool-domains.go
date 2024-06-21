package resources

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CognitoUserPoolDomainResource = "CognitoUserPoolDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:   CognitoUserPoolDomainResource,
		Scope:  nuke.Account,
		Lister: &CognitoUserPoolDomainLister{},
	})
}

type CognitoUserPoolDomainLister struct{}

func (l *CognitoUserPoolDomainLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cognitoidentityprovider.New(opts.Session)

	userPoolsLister := &CognitoUserPoolLister{}
	userPools, poolErr := userPoolsLister.List(ctx, o)
	if poolErr != nil {
		return nil, poolErr
	}

	resources := make([]resource.Resource, 0)
	for _, userPoolResource := range userPools {
		userPool, ok := userPoolResource.(*CognitoUserPool)
		if !ok {
			logrus.Errorf("Unable to case CognitoUserPool")
			continue
		}

		describeParams := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: userPool.id,
		}
		userPoolDetails, err := svc.DescribeUserPool(describeParams)
		if err != nil {
			return nil, err
		}
		if userPoolDetails.UserPool.Domain == nil {
			// No domain on this user pool so skip
			continue
		}

		resources = append(resources, &CognitoUserPoolDomain{
			svc:          svc,
			name:         userPoolDetails.UserPool.Domain,
			userPoolName: userPool.name,
			userPoolID:   userPool.id,
		})
	}

	return resources, nil
}

type CognitoUserPoolDomain struct {
	svc          *cognitoidentityprovider.CognitoIdentityProvider
	name         *string
	userPoolName *string
	userPoolID   *string
}

func (f *CognitoUserPoolDomain) Remove(_ context.Context) error {
	params := &cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     f.name,
		UserPoolId: f.userPoolID,
	}
	_, err := f.svc.DeleteUserPoolDomain(params)

	return err
}

func (f *CognitoUserPoolDomain) String() string {
	return fmt.Sprintf("%s -> %s", *f.userPoolName, *f.name)
}
