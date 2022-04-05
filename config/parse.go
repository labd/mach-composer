package config

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/labd/mach-composer-go/utils"
	"gopkg.in/yaml.v3"
)

func Load(filename string, varFilename string) (*MachConfig, error) {

	var vars *Variables
	if varFilename != "" {
		var err error
		vars, err = loadVariables(varFilename)
		if err != nil {
			panic(err)
		}
	}

	body, err := utils.AFS.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	schemaVersion, err := GetSchemaVersion(body)
	if err != nil {
		return nil, err
	}

	if !ValidateConfig(body, schemaVersion) {
		return nil, fmt.Errorf("failed to load config %s due to errors", filename)
	}

	cfg, err := Parse(body, vars)
	if err != nil {
		panic(err)
	}

	if err := defaults.Set(cfg); err != nil {
		panic(err)
	}

	cfg.Variables = vars
	cfg.Filename = filepath.Base(filename)
	Process(cfg)

	return cfg, nil
}

func GetSchemaVersion(data []byte) (int, error) {

	type PartialMachConfig struct {
		MachComposer MachComposer `yaml:"mach_composer"`
	}

	// Decode the yaml in an intermediate config file
	intermediate := &PartialMachConfig{}
	err := yaml.Unmarshal(data, intermediate)
	if err != nil {
		return 0, err
	}

	v := intermediate.MachComposer.Version
	if val, err := strconv.Atoi(v); err == nil {
		return val, err
	}

	parts := strings.SplitN(v, ".", 2)
	if val, err := strconv.Atoi(parts[0]); err == nil {
		return val, err
	}

	return 0, errors.New("No valid version identifier found")
}

func Parse(data []byte, vars *Variables) (*MachConfig, error) {

	// Decode the yaml in an intermediate config file
	intermediate := &_RawMachConfig{}
	err := yaml.Unmarshal(data, intermediate)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if vars != nil {
		InterpolateVars(intermediate, vars)
	}

	cfg := &MachConfig{
		Filename:     intermediate.Filename,
		MachComposer: intermediate.MachComposer,
		Global:       intermediate.Global,
	}
	err = intermediate.Sites.Decode(&cfg.Sites)
	if err != nil {
		log.Fatalf("Decode: %v", err)
	}

	intermediate.Components.Decode(&cfg.Components)
	if err != nil {
		log.Fatalf("Decode: %v", err)
	}

	return cfg, nil
}
