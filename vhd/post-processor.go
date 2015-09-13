// Package vhd implements the packer.PostProcessor interface and adds a
// post-processor that produces a standalone VHD file.
package vhd

import (
	"fmt"

	"github.com/mitchellh/packer/builder/qemu"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Map Builders to Providers: these are the types of artifacts we know how to
// convert.
var providers = map[string]Provider{
	vboxcommon.BuilderId: new(VirtualBoxProvider),
	qemu.BuilderId:       new(QEMUProvider),
}

// Config contains the post-processor configuration.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Where the VHD will be output to.
	OutputPath string `mapstructure:"output"`

	// Whether to keep the Provider artifact (e.g., VirtualBox VMDK).
	KeepInputArtifact bool `mapstructure:"keep_input_artifict"`

	ctx interpolate.Context
}

// PostProcessor satisfies the packer.PostProcessor interface.
type PostProcessor struct {
	config Config
}

// Configure the PostProcessor, rendering templated values if necessary.
func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}
	return nil
}

// PostProcess is the main entry point. It calls a Provider's Convert() method
// to delegate conversion to that Provider's command-line tool.
func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	provider, err := providerForBuilderId(artifact.BuilderId())
	if err != nil {
		return nil, false, err
	}

	ui.Say(fmt.Sprintf("Converting %s image to VHD file...", provider))

	err = provider.Convert(ui, artifact, p.config.OutputPath)
	if err != nil {
		return nil, false, err
	}

	ui.Say(fmt.Sprintf("Converted VHD: %s", p.config.OutputPath))
	artifact = NewArtifact(provider.String(), p.config.OutputPath)
	keep := p.config.KeepInputArtifact

	return artifact, keep, nil
}

// Pick a provider to use from known builder sources.
func providerForBuilderId(builderId string) (Provider, error) {
	if provider, ok := providers[builderId]; ok {
		return provider, nil
	}
	return nil, fmt.Errorf("Unknown artifact type: %s", builderId)
}
