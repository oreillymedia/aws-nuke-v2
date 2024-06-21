package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserResource = "IAMUser"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMUserResource,
		Scope:  nuke.Account,
		Lister: &IAMUsersLister{},
		DependsOn: []string{
			IAMUserAccessKeyResource,
			IAMUserHTTPSGitCredentialResource,
			IAMUserGroupAttachmentResource,
			IAMUserPolicyAttachmentResource,
			IAMVirtualMFADeviceResource,
		},
		DeprecatedAliases: []string{
			"IamUser",
		},
	})
}

type IAMUser struct {
	svc              *iam.IAM
	name             string
	tags             []*iam.Tag
	createDate       *time.Time
	passwordLastUsed *time.Time
}

func init() {
	register("IAMUser", ListIAMUsers)
}

func GetIAMUser(svc *iam.IAM, userName *string) (*iam.User, error) {
	params := &iam.GetUserInput{
		UserName: userName,
	}
	resp, err := svc.GetUser(params)
	return resp.User, err
}

func ListIAMUsers(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)
	var resources []Resource

	err := svc.ListUsersPages(nil, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, out := range page.Users {
			user, err := GetIAMUser(svc, out.UserName)
			if err != nil {
				logrus.Errorf("Failed to get user %s: %v", *out.UserName, err)
				continue
			}
			resources = append(resources, &IAMUser{
				svc:              svc,
				name:             *user.UserName,
				tags:             user.Tags,
				createDate:       user.CreateDate,
				passwordLastUsed: user.PasswordLastUsed,
			})
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (e *IAMUser) Remove() error {
	_, err := e.svc.DeleteUser(&iam.DeleteUserInput{
		UserName: &e.name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUser) String() string {
	return e.name
}

func (e *IAMUser) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", e.name)

	if e.createDate != nil {
		properties.Set("CreateDate", e.createDate.Format(time.RFC3339))
	}
	if e.passwordLastUsed != nil {
		properties.Set("PasswordLastUsed", e.passwordLastUsed.Format(time.RFC3339))
	}

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

// --------------

func GetIAMUser(svc *iam.IAM, userName *string) (*iam.User, error) {
	params := &iam.GetUserInput{
		UserName: userName,
	}
	resp, err := svc.GetUser(params)
	return resp.User, err
}

// --------------

type IAMUsersLister struct{}

func (l *IAMUsersLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := iam.New(opts.Session)

	var resources []resource.Resource

	if err := svc.ListUsersPages(nil, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, out := range page.Users {
			user, err := GetIAMUser(svc, out.UserName)
			if err != nil {
				logrus.Errorf("Failed to get user %s: %v", *out.UserName, err)
				continue
			}
			resources = append(resources, &IAMUser{
				svc:  svc,
				name: *out.UserName,
				tags: user.Tags,
			})
		}
		return true
	}); err != nil {
		return nil, err
	}

	return resources, nil
}
