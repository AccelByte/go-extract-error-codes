// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package goextracterrorcodes

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
)

func Process(
	restfulConfiguration []*restful.WebService,
	projectMainFileDir string,
	errorCodeDir string,
	msgConfig *AppCodesConfiguration,
) ([]*HandlerDetailsWithAppCodes, error) {
	// locate handlers
	handlerDetails := locateHandlers(restfulConfiguration)

	callGraphWalker := NewCallGraphWalker(
		projectMainFileDir,
		errorCodeDir,
		handlerDetails,
		msgConfig,
	)

	handlersWithCodes, err := callGraphWalker.LocateHandlersWithAppCodes()
	if err != nil {
		return nil, fmt.Errorf("unable to locate handlers, Err: %s", err)
	}

	return handlersWithCodes, nil
}
