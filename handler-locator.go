// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package goextracterrorcodes

import (
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"github.com/emicklei/go-restful/v3"
)

type HandlerDetails struct {
	Method string
	Path   string
	Name   string
	File   string
	Line   int
}

func locateHandlers(handlers []*restful.WebService) []*HandlerDetails {
	result := make([]*HandlerDetails, 0)

	for _, h := range handlers {
		routes := getRoutes(h)

		handlers := processRoutes(routes)

		result = append(result, handlers...)
	}

	return result
}

func processRoutes(routes []restful.Route) []*HandlerDetails {
	result := make([]*HandlerDetails, 0)

	for _, r := range routes {
		details := getHandlerDetails(r.Function)

		// add known attributes
		details.Method = r.Method
		details.Path = r.Path

		result = append(result, details)
	}

	return result
}

func getRoutes(s *restful.WebService) []restful.Route {
	rs := reflect.ValueOf(s).Elem()
	rf := rs.FieldByName("routes")
	// rf can't be read or set.
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	// Now rf can be read and set.

	r, ok := rf.Interface().([]restful.Route)
	if !ok {
		return []restful.Route{}
	}

	return r
}

func getHandlerDetails(f restful.RouteFunction) (details *HandlerDetails) {
	details = &HandlerDetails{}

	defer func() {
		// as we can't analyze the route as a function
		// try to call it with invalid parameters and handle panic
		// next we can analyze stack trace of the panic
		if r := recover(); r != nil {
			pc := make([]uintptr, 10)
			n := runtime.Callers(1, pc)
			if n == 0 {
				// No PCs available. This can happen if the first argument to
				// runtime.Callers is large.
				//
				// Return now to avoid processing the zero Frame that would
				// otherwise be returned by frames.Next below.
				return
			}

			pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
			frames := runtime.CallersFrames(pc)

			// Loop to get frames.
			// A fixed number of PCs can expand to an indefinite number of Frames.
			for {
				frame, more := frames.Next()

				// Check whether there are more frames to process after this one.
				if !more {
					break
				}

				// previous item is a required handler
				if strings.HasSuffix(frame.Function, "getHandlerDetails") {
					break
				}

				details.Name = frame.Function
				details.File = frame.File
				details.Line = frame.Line
			}
		}
	}()

	f(nil, nil)

	return details
}
