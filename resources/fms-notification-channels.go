package resources

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

const FMSNotificationChannelResource = "FMSNotificationChannel"

func init() {
	registry.Register(&registry.Registration{
		Name:   FMSNotificationChannelResource,
		Scope:  nuke.Account,
		Lister: &FMSNotificationChannelLister{},
	})
}

type FMSNotificationChannelLister struct{}

func (l *FMSNotificationChannelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := fms.New(opts.Session)
	resources := make([]resource.Resource, 0)

	if _, err := svc.GetNotificationChannel(&fms.GetNotificationChannelInput{}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if strings.Contains(aerr.Message(), "No default admin could be found") {
				logrus.Infof("FMSNotificationChannel: %s. Ignore if you haven't set it up.", aerr.Message())
				return nil, nil
			}
		}
	} else {
		resources = append(resources, &FMSNotificationChannel{
			svc: svc,
		})
	}

	return resources, nil
}

type FMSNotificationChannel struct {
	svc *fms.FMS
}

func (f *FMSNotificationChannel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteNotificationChannel(&fms.DeleteNotificationChannelInput{})

	return err
}

func (f *FMSNotificationChannel) String() string {
	return "fms-notification-channel"
}

func (f *FMSNotificationChannel) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("NotificationChannelEnabled", "true")
	return properties
}
