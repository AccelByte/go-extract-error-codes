// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package goextracterrorcodes

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type AppCodesConfiguration struct {
	Services          map[int]string        `yaml:"services" json:"-"`
	Sections          map[int]string        `yaml:"sections" json:"-"`
	Messages          map[string]AppMessage `yaml:"messages" json:"-"`
	Types             map[string]string     `yaml:"types" json:"-"`
	DefaultType       string                `yaml:"defaultType" json:"-"`
	AllowedDuplicates []int                 `yaml:"allowedDuplicates" json:"-"`
	PackageName       string                `yaml:"packageName" json:"-"`
}

type AppMessage struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func LoadAppCodesConfiguration(yamlFilePath string) (*AppCodesConfiguration, error) {
	fileDataRaw, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read yaml file. Err: %s", err)
	}

	// parse file
	data := &AppCodesConfiguration{}
	err = yaml.Unmarshal(fileDataRaw, &data)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal yaml file. Err: %s", err)
	}

	return data, err
}
