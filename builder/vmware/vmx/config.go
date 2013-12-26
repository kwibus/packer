package vmx

import (
	"fmt"
	"os"

	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	vmwcommon.OutputConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig    `mapstructure:",squash"`

	SourcePath string `mapstructure:"source_path"`
	VMName     string `mapstructure:"vm_name"`

	tpl *packer.ConfigTemplate
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	md, err := common.DecodeConfig(c, raws...)
	if err != nil {
		return nil, nil, err
	}

	c.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, nil, err
	}
	c.tpl.UserVars = c.PackerUserVars

	// Defaults
	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", c.PackerBuildName)
	}

	// Prepare the errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(c.tpl, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.VMXConfig.Prepare(c.tpl)...)

	templates := map[string]*string{
		"source_path": &c.SourcePath,
		"vm_name":     &c.VMName,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.SourcePath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	} else {
		if _, err := os.Stat(c.SourcePath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("source_path is invalid: %s", err))
		}
	}

	// Warnings
	var warnings []string

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
