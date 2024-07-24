package resources

import (
	"context"

	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserAccessKeyResource = "IAMUserAccessKey"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMUserAccessKeyResource,
		Scope:  nuke.Account,
		Lister: &IAMUserAccessKeyLister{},
		DeprecatedAliases: []string{
			"IamUserAccessKeys",
		},
	})
}

type IAMUserAccessKey struct {
	svc         iamiface.IAMAPI
	accessKeyId string
	createDate  *time.Time
	userName    string
	status      string
	createDate  *time.Time
	userTags    []*iam.Tag
}

func init() {
	register("IAMUserAccessKey", ListIAMUserAccessKeys)
}

func ListIAMUserAccessKeys(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)

	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0)
	for _, role := range resp.Users {
		resp, err := svc.ListAccessKeys(
			&iam.ListAccessKeysInput{
				UserName: role.UserName,
			})
		if err != nil {
			return nil, err
		}

		userTags, err := svc.ListUserTags(&iam.ListUserTagsInput{UserName: role.UserName})
		if err != nil {
			return nil, err
		}

		for _, meta := range resp.AccessKeyMetadata {
			resources = append(resources, &IAMUserAccessKey{
				svc:         svc,
				accessKeyId: *meta.AccessKeyId,
				createDate:  meta.CreateDate,
				userName:    *meta.UserName,
				status:      *meta.Status,
				userTags:    userTags.Tags,
			})
		}
	}

	return resources, nil
}

func (e *IAMUserAccessKey) Remove() error {
	_, err := e.svc.DeleteAccessKey(
		&iam.DeleteAccessKeyInput{
			AccessKeyId: &e.accessKeyID,
			UserName:    &e.userName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserAccessKey) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("UserName", e.userName)
	properties.Set("AccessKeyID", e.accessKeyID)
	properties.Set("CreateDate", e.createDate.Format(time.RFC3339))

	if e.createDate != nil {
		properties.Set("CreateDate", e.createDate.Format(time.RFC3339))
	}

	for _, tag := range e.userTags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (e *IAMUserAccessKey) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.accessKeyID)
}

// --------------

type IAMUserAccessKeyLister struct{}

func (l *IAMUserAccessKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, role := range resp.Users {
		resp, err := svc.ListAccessKeys(
			&iam.ListAccessKeysInput{
				UserName: role.UserName,
			})
		if err != nil {
			return nil, err
		}

		userTags, err := svc.ListUserTags(&iam.ListUserTagsInput{UserName: role.UserName})
		if err != nil {
			return nil, err
		}

		for _, meta := range resp.AccessKeyMetadata {
			resources = append(resources, &IAMUserAccessKey{
				svc:         svc,
				accessKeyID: *meta.AccessKeyId,
				userName:    *meta.UserName,
				status:      *meta.Status,
				createDate:  meta.CreateDate,
				userTags:    userTags.Tags,
			})
		}
	}

	return resources, nil
}
