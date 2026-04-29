# azurenamingtool-client-go

A Go client library for the [Azure Naming Tool](https://github.com/mspnp/AzureNamingTool) API. It is used by [terraform-provider-proactnaming](https://github.com/proact-global/terraform-provider-proactnaming) and can also be imported directly into your own Go programs.

## Installation

```bash
go get github.com/proact-global/azurenamingtool-client-go
```

## Usage

```go
package main

import (
    "fmt"
    "log"

    naming "github.com/proact-global/azurenamingtool-client-go"
)

func main() {
    host   := "https://your-naming-tool.azurewebsites.net"
    apiKey := "your-api-key"
    adminPwd := "your-admin-password" // optional; required for delete & read-by-ID

    client, err := naming.NewClient(&host, &apiKey, &adminPwd)
    if err != nil {
        log.Fatal(err)
    }

    // Generate a name
    resp, err := client.GenerateName(naming.GenerateNameRequest{
        ResourceOrg:         "myorg",
        ResourceType:        "rg",
        ResourceEnvironment: "dev",
        ResourceLocation:    "euw",
        ResourceInstance:    "001",
        CustomComponents:    naming.GenerateNameRequestCustomComponents{
            "application": "webapp",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Generated name:", resp.ResourceName)
    fmt.Println("Naming Tool ID:", resp.ResourceNameDetails.ID)

    // Look up a name by its ID
    details, err := client.GetName(resp.ResourceNameDetails.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Resource type name:", details.ResourceTypeName)

    // Delete a name by ID
    _, err = client.DeleteName(naming.DeleteGeneratedNameRequest{ID: resp.ResourceNameDetails.ID})
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### `NewClient(host, apiKey, adminPassword *string) (*Client, error)`

Creates a new client. Pass `nil` for `adminPassword` if admin operations are not needed.

### `GenerateName(req GenerateNameRequest) (*GenerateNameResponse, error)`

Calls the V2 API to generate and persist a new resource name.

### `GetName(id int64) (*ResourceNameDetails, error)`

Retrieves a previously generated name by its ID via the Admin API. Returns `ErrNotFound` if the ID does not exist.

### `DeleteName(req DeleteGeneratedNameRequest) ([]byte, error)`

Deletes a generated name by its ID via the Admin API. Treats "not found" as success.

## Sentinel Errors

| Error | Meaning |
|-------|---------|
| `azurenamingtool.ErrNotFound` | The requested ID does not exist in the naming tool. Use `errors.Is` to check. |

## Requirements

- Go 1.21+
- Access to an Azure Naming Tool instance (v5+)