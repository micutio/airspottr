# Target Architecture using DDD

```
internal/
├── domain/           # entities, value objects, events, repository INTERFACES
├── application/      # commands, queries, services (orchestration)
├── infrastructure/   # Postgres/sqlc repositories, outbox relay, config
└── interface/        # REST controllers, DTOs
```

| Layer           | 	May import                       | 	Must never import                                               |
|:----------------|:---------------------------------|:----------------------------------------------------------------|
| domain          | 	stdlib, google/uuid              | 	application, infrastructure, interface, any framework or driver |
| application 	    | domain 	                          | infrastructure, interface, Echo, pgx                            |
| infrastructure 	 | domain, application              | 	interface                                                       |
| interface       | 	application, domain (types only) | infrastructure                                                  |

## Main function

- usually lives in `cmd/youAppName/main.go`
- construct concrete repositories
- injects them into services
- hands services to controllers
- NO other file knows ALL the layers

## TODOs for DDD Conversion

- take `application.Dashboard` apart bit by bit and create entities/interfaces/repos for its individual functionalities
- separate `infrastructure.Domain` into aircraft and flightroute repos