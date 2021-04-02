# OAuth2
## Models
- Configuration
- OAuth2Info
- User

## Services
- OAuth2Service

## Repositories
- ConfigurationRepository
- OAuth2UserRepository
- UserRepository

## Installation

Please make sure to initialize a Go module before installing common-go/oauth2:

```shell
go get -u github.com/common-go/oauth2
```

Import:

```go
import "github.com/common-go/oauth2"
```

## Implementations of UserRepository and IntegrationConfigurationRepository
- sql: SqlUserRepository and SqlConfigurationRepository
- [mongo](https://github.com/common-go/oauth2-mongo)
- [dynamodb](https://github.com/common-go/oauth2-dynamodb)
- [firestore](https://github.com/common-go/oauth2-firestore)
- [elasticsearch](https://github.com/common-go/oauth2-elasticsearch)
