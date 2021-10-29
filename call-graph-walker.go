// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package goextracterrorcodes

import (
	"fmt"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func NewCallGraphWalker(
	projectMainFileDir string,
	errorCodeDir string,
	handlerDetails []*HandlerDetails,
	appCodesConfig *AppCodesConfiguration,
) *CallGraphWalker {
	return &CallGraphWalker{
		projectMainFileDir: projectMainFileDir,
		errorCodeDir:       errorCodeDir,
		nodes:              make(map[string]Node),
		handlerDetails:     handlerDetails,
		appCodesConfig:     appCodesConfig,
	}
}

type Node struct {
	Caller    string
	Called    []string
	IsHandler bool
	Handlers  []*HandlerDetails

	LocalErrors map[string]struct{}
}

type HandlerDetailsWithAppCodes struct {
	Method   string
	Receiver string
	Path     string
	Name     string
	File     string
	Line     int
	AppCodes FoundAppMessages
}

type FoundAppMessages []*FoundAppMessage

func (i FoundAppMessages) Len() int { return len(i) }

func (i FoundAppMessages) Swap(x, y int) {
	i[x], i[y] = i[y], i[x]
}

func (i FoundAppMessages) Less(x, y int) bool {
	return i[x].Code < i[y].Code
}

type FoundAppMessage struct {
	Code int
	Text string
	Name string
}

type CallGraphWalker struct {
	projectMainFileDir string
	errorCodeDir       string
	handlerDetails     []*HandlerDetails
	appCodesConfig     *AppCodesConfiguration

	nodes map[string]Node
}

// nolint:funlen,cyclop
func (s *CallGraphWalker) LocateHandlersWithAppCodes() ([]*HandlerDetailsWithAppCodes, error) {
	response := make([]*HandlerDetailsWithAppCodes, 0)

	cfg := packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps |
			packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	initial, err := packages.Load(&cfg, s.projectMainFileDir)
	if err != nil {
		return nil, err
	}

	// Create SSA packages for well-typed packages and their dependencies.
	prog, pkgs := ssautil.AllPackages(initial, ssa.GlobalDebug)

	// Build SSA code for the whole program.
	prog.Build()

	result, err := pointer.Analyze(&pointer.Config{
		Mains: pkgs,
		// Mains:           []*ssa.Package{mainPkg},
		Reflection:      false,
		BuildCallGraph:  true,
		Queries:         nil,
		IndirectQueries: nil,
		Log:             nil,
	})
	if err != nil {
		return nil, err
	}

	// Find edges originating from the main package.
	// By converting to strings, we de-duplicate nodes
	// representing the same function due to context sensitivity.

	allHandlers := make([]*HandlerDetails, 0)

	err = callgraph.GraphVisitEdges(result.CallGraph, func(edge *callgraph.Edge) error {
		callerName := ExtractNodeName(edge.Caller.String())
		calleeName := ExtractNodeName(edge.Callee.String())

		callerNode, exists := s.nodes[callerName]
		if !exists {
			callerNode = Node{
				Caller:      callerName,
				Called:      make([]string, 0),
				LocalErrors: make(map[string]struct{}),
			}
		}

		calleeNode, exists := s.nodes[calleeName]
		if !exists {
			calleeNode = Node{
				Caller:      calleeName,
				Called:      make([]string, 0),
				LocalErrors: make(map[string]struct{}),
			}
		}

		isNodeError, name, msg := s.IsNodeError(edge.Callee)
		if isNodeError {
			_ = msg

			// nolint:forbidigo
			fmt.Printf("Error Found: %s\n", name)

			callerNode.LocalErrors[name] = struct{}{}
		}

		isNodeHandler, handlers := s.IsNodeHandler(edge.Callee)
		if isNodeHandler {
			calleeNode.IsHandler = true

			calleeNode.Handlers = append(calleeNode.Handlers, handlers...)

			allHandlers = append(allHandlers, handlers...)
		}

		callerNode.Called = append(callerNode.Called, calleeName)

		s.nodes[callerName] = callerNode
		s.nodes[calleeName] = calleeNode

		return nil
	})
	if err != nil {
		//nolint:forbidigo
		fmt.Printf("CallGraphWalker.LocateHandlersWithAppCodes: GraphVisitEdges, Err: %s", err)
	}

	for _, handler := range allHandlers {
		alreadyVisitedNodes := make(map[string]struct{})

		codes := s.CollectErrorCodesRecursively(handler.Name, alreadyVisitedNodes)

		details := &HandlerDetailsWithAppCodes{
			Method:   handler.Method,
			Receiver: "",
			Path:     handler.Path,
			Name:     handler.Name,
			File:     handler.File,
			Line:     handler.Line,
			AppCodes: make(FoundAppMessages, 0),
		}

		for code := range codes {
			found := false
			for name, codeDeclaration := range s.appCodesConfig.Messages {
				if code == name {
					found = true

					details.AppCodes = append(details.AppCodes, &FoundAppMessage{
						Code: codeDeclaration.Code,
						Text: codeDeclaration.Text,
						Name: code,
					})

					break
				}
			}

			if !found {
				details.AppCodes = append(details.AppCodes, &FoundAppMessage{
					Code: 0,
					Text: code,
					Name: code,
				})
			}
		}

		sort.Sort(details.AppCodes)

		response = append(response, details)
	}

	return response, nil
}

func (s *CallGraphWalker) IsNodeHandler(node *callgraph.Node) (bool, []*HandlerDetails) {
	/*
		(string) (len=92) "n6674:(*accelbyte.net/template-justice-example-service/pkg/example/api.Handlers).HealthCheck"
		(*main.HandlerDetails)(0xc0004b16d0)({
			 Method: (string) (len=3) "GET",
			 Path: (string) (len=8) "/healthz",
			 Name: (string) (len=86) "accelbyte.net/template-justice-example-service/pkg/example/api.(*Handlers).HealthCheck",
			 File: (string) (len=78) "/home/ab/dev/ab/template-justice-example-service/pkg/example/api/health-get.go",
			 Line: (int) 49
		})
	*/

	found := false
	handlers := make([]*HandlerDetails, 0)

	for _, item := range s.handlerDetails {
		if ExtractNodeName(node.String()) == ExtractNodeName(item.Name) {
			handlers = append(handlers, item)
			found = true
		}
	}

	return found, handlers
}

func (s *CallGraphWalker) IsNodeError(node *callgraph.Node) (bool, string, *AppMessage) {
	for name, item := range s.appCodesConfig.Messages {
		pattern := fmt.Sprintf("%s.New", s.errorCodeDir)

		pos := strings.Index(node.String(), pattern)
		if pos <= 0 {
			continue
		}

		code := node.String()[len(pattern)+pos:]

		// check that name is the same as a known error
		if code == name {
			return true, name, &item
		}
	}

	return false, "", nil
}

// CollectErrorCodesRecursively collects error codes from the graph with cycles
// we are passing the variable alreadyVisitedNodes by the pointer to avoid visiting cycles in the graph.
func (s *CallGraphWalker) CollectErrorCodesRecursively(
	nodeName string,
	alreadyVisitedNodes map[string]struct{},
) map[string]struct{} {
	nodeName = ExtractNodeName(nodeName)

	result := make(map[string]struct{})

	if _, alreadyVisited := alreadyVisitedNodes[nodeName]; alreadyVisited {
		return result
	}

	node, isExists := s.nodes[nodeName]
	if !isExists {
		// nolint:forbidigo
		fmt.Println("CallGraphWalker.CollectErrorCodesRecursively: Node not exists")

		return result
	}

	if len(node.LocalErrors) > 0 {
		// nolint:forbidigo
		fmt.Printf("CallGraphWalker.CollectErrorCodesRecursively. %s. Node Errors: %s", nodeName, spew.Sdump(node.LocalErrors))
	}

	for errName := range node.LocalErrors {
		result[errName] = struct{}{}
	}

	// make the node as visited
	alreadyVisitedNodes[nodeName] = struct{}{}

	for _, item := range node.Called {
		subNodeResult := s.CollectErrorCodesRecursively(item, alreadyVisitedNodes)

		if len(subNodeResult) > 0 {
			// nolint:forbidigo
			fmt.Printf("SubFunction Errors: %s", spew.Sdump(subNodeResult))
		}

		for errName := range subNodeResult {
			result[errName] = struct{}{}
		}
	}

	return result
}

func ExtractNodeName(name string) string {
	name = strings.ReplaceAll(name, "*", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	if parts := strings.Split(name, ":"); len(parts) == 2 {
		name = parts[1]
	}

	return name
}
