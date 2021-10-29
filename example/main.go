// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

// This is an example of usage

/*
import (
	"io/ioutil"
	"log"

	"accelbyte.net/justice-lobby-server/pkg/api"
	"accelbyte.net/justice-lobby-server/pkg/config"
	envvariables "accelbyte.net/justice-lobby-server/pkg/utils/env-variables"
	"github.com/caarlos0/env"
	"gopkg.in/yaml.v3"
)

func main() {
	projectMainFileDir := "accelbyte.net/justice-lobby-server/cmd/justice-lobby-server"
	errorCodeDir := "accelbyte.net/justice-lobby-server/pkg/log"

	// create an instance of the service
	cfg := &config.Config{}

	// parse ENV variables into the config
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("unable to parse environment variables: ", err)
	}

	s := api.PrepareLobby(cfg, "", "", "")
	if err != nil {
		log.Fatalf("unable to prepare the service, Err: %s", err)
	}
	if s == nil {
		log.Fatal("unable to process an empty service")
	}

	// parse error codes configuration
	msgCfg, err := LoadAppCodesConfiguration("api/errors.yml")
	if err != nil {
		log.Fatal(err)
	}

	handlersWithCodes, err := Process(
		s.GetHandlers(),
		projectMainFileDir,
		errorCodeDir,
		msgCfg,
	)

	result, err := yaml.Marshal(handlersWithCodes)
	if err != nil {
		log.Fatalf("unable to marshal dst data %s", err.Error())
	}

	// write dst file
	err = ioutil.WriteFile("route-app-codes.yaml", result, 0600)
	if err != nil {
		log.Fatalf("unable to write dst file %s", err.Error())
	}
}
*/
