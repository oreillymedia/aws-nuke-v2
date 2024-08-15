package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/elastictranscoder"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ElasticTranscoderPresetResource = "ElasticTranscoderPreset"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticTranscoderPresetResource,
		Scope:  nuke.Account,
		Lister: &ElasticTranscoderPresetLister{},
	})
}

type ElasticTranscoderPresetLister struct{}

func (l *ElasticTranscoderPresetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elastictranscoder.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &elastictranscoder.ListPresetsInput{}

	for {
		resp, err := svc.ListPresets(params)
		if err != nil {
			return nil, err
		}

		for _, preset := range resp.Presets {
			resources = append(resources, &ElasticTranscoderPreset{
				svc:      svc,
				PresetID: preset.Id,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ElasticTranscoderPreset struct {
	svc      *elastictranscoder.ElasticTranscoder
	PresetID *string
}

func (f *ElasticTranscoderPreset) Filter() error {
	if strings.HasPrefix(*f.PresetID, "1351620000001") {
		return fmt.Errorf("cannot delete elastic transcoder system presets")
	}
	return nil
}

func (f *ElasticTranscoderPreset) Remove(_ context.Context) error {
	_, err := f.svc.DeletePreset(&elastictranscoder.DeletePresetInput{
		Id: f.PresetID,
	})

	return err
}

func (f *ElasticTranscoderPreset) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *ElasticTranscoderPreset) String() string {
	return *f.PresetID
}
