# go-extract-error-codes

## Overview
This library helps to extract possible error codes from configured go-restful applications
quality level: PoC

## High Level Features
* Extract error codes for HTTP routes and save into a file

## Requirements
* The service should support design-first configuration for error codes in the the form:
```yaml
packageName: log
types:
  numberOfAttempts: int
defaultType: string
allowedDuplicates:
  - 20000 # needs to be resolved
services:
  11: "lobby"
sections:
  2000: "general"
  2001: "general"
messages:
#global error codes
  InternalServerError:
    code: 20000
    text: "{{message}}"
  InternalServerErrorV1:
    code: 20000
    text: "unable to {{action}}: {{reason}}, userID: {{userID}}, details: {{details}}"
  UnauthorizedAccess:
    code: 20001
    text: "unauthorized access"
  LobbyConnectionUnableToQueryConfig:
    code: 112110
    text: "unable to query lobby config"
  LobbyConnectionUnableToQueryCCU:
    code: 112111
    text: "unable to query lobby CCU"
  LobbyConnectionCCULimitReached:
    code: 112112
    text: "lobby CCU limit reached"
  LobbyConnectionUnableToGetBlockedPlayer:
    code: 112109
    text: "unable to get blocked player from storage"
```

## Usage example
stored in the /example directory

## Result example
```yaml
- method: GET
  receiver: ""
  path: /lobby/healthz
  name: xx/pkg/health.(*Health).HealthCheckHandler
  file: xx/pkg/health/health.go
  line: 182
  appcodes: []
- method: GET
  receiver: ""
  path: /lobby/
  name: xx/pkg/lobby/service.(*Lobby).HandleUserConnection
  file: xx/pkg/lobby/service/service.go
  line: 212
  appcodes:
    - code: 112110
      text: unable to query lobby config
      name: LobbyConnectionUnableToQueryConfig
    - code: 112111
      text: unable to query lobby CCU
      name: LobbyConnectionUnableToQueryCCU
    - code: 112112
      text: lobby CCU limit reached
      name: LobbyConnectionCCULimitReached
    - code: 112109
      text: unable to get blocked player from storage
      name: LobbyConnectionUnableToGetBlockedPlayer
    - code: 112106
      text: '{{message}}'
      name: LobbyConnectionUnableCheckConnectedUser
    - code: 112105
      text: '{{message}}'
      name: LobbyConnectionMultipleLoginAttempt
    - code: 112101
      text: '{{message}}'
      name: LobbyConnectionUnableToUpgrade
- method: POST
  receiver: ""
  path: /notification/namespaces/{namespace}/freeform
  name: xx/pkg/notification/service.(*NotificationService).HandleFreeFormNamespaceNotification
  file: xx/pkg/notification/service/service.go
  line: 104
  appcodes:
    - code: 11401
      text: '{{message}}'
      name: FreeNotificationInvalidRequestBody
    - code: 11403
      text: '{{message}}'
      name: FreeNotificationUnablePublishNotification
```
