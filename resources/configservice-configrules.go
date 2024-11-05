package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ConfigServiceConfigRuleResource = "ConfigServiceConfigRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   ConfigServiceConfigRuleResource,
		Scope:  nuke.Account,
		Lister: &ConfigServiceConfigRuleLister{},
	})
}

type ConfigServiceConfigRuleLister struct{}

func (l *ConfigServiceConfigRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := configservice.New(opts.Session)
	var resources []resource.Resource

	params := &configservice.DescribeConfigRulesInput{}

	for {
		output, err := svc.DescribeConfigRules(params)
		if err != nil {
			return nil, err
		}

		for _, configRule := range output.ConfigRules {
			resources = append(resources, &ConfigServiceConfigRule{
				svc:            svc,
				configRuleName: configRule.ConfigRuleName,
				createdBy:      configRule.CreatedBy,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ConfigServiceConfigRule struct {
	svc            *configservice.ConfigService
	configRuleName *string
	createdBy      *string
}

func (f *ConfigServiceConfigRule) Filter() error {
	if aws.StringValue(f.createdBy) == "securityhub.amazonaws.com" {
		return fmt.Errorf("cannot remove rule owned by securityhub.amazonaws.com")
	}

	if aws.StringValue(f.createdBy) == "config-conforms.amazonaws.com" {
		return fmt.Errorf("cannot remove rule owned by config-conforms.amazonaws.com")
	}

	return nil
}

func (f *ConfigServiceConfigRule) Remove() error {
	f.svc.DeleteRemediationConfiguration(&configservice.DeleteRemediationConfigurationInput{
		ConfigRuleName: f.configRuleName,
	})

	_, err := f.svc.DeleteConfigRule(&configservice.DeleteConfigRuleInput{
		ConfigRuleName: f.configRuleName,
	})

	return err
}

func (f *ConfigServiceConfigRule) String() string {
	return *f.configRuleName
}

func (f *ConfigServiceConfigRule) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("CreatedBy", f.createdBy)
	return props
}
