# zstate: Type-Safe State Machine Library for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/upamune/zstate)](https://goreportcard.com/report/github.com/upamune/zstate)
[![GoDoc](https://godoc.org/github.com/upamune/zstate?status.svg)](https://godoc.org/github.com/upamune/zstate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/github/upamune/zstate/graph/badge.svg?token=GPJ6L4P8AO)](https://codecov.io/github/upamune/zstate)

zstate is a simple and type-safe state machine library for Go. It allows you to manage complex state transitions and clearly express business logic.

## Features

- 🔒 **Type-Safe**: Leverages generics for compile-time type checking
- 🛠 **Flexible**: Control transitions with custom guard functions
- 🧩 **Simple API**: Easy to use with an intuitive builder pattern
- 📦 **No Dependencies**: Uses only the standard library

## Installation

```bash
go get github.com/upamune/zstate
```

## Usage Example

Here's an example of a state machine managing a door's state, including a guard function:

```go
package main

import (
    "context"
    "fmt"
    "github.com/upamune/zstate"
)

type DoorState int

const (
    Closed DoorState = iota
    Open
    Locked
)

type DoorEvent string

const (
    OpenDoor  DoorEvent = "OpenDoor"
    CloseDoor DoorEvent = "CloseDoor"
    LockDoor  DoorEvent = "LockDoor"
    UnlockDoor DoorEvent = "UnlockDoor"
)

func main() {
    var isLocked bool

    doorBuilder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
    door, _ := doorBuilder.
        AddState(Closed).
        AddState(Open).
        AddState(Locked).
        SetInitialState(Closed).
        AddTransition(Closed, Open, OpenDoor).
        AddTransition(Open, Closed, CloseDoor).
        AddTransition(Closed, Locked, LockDoor, zstate.WithGuard[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) bool {
            return !isLocked
        })).
        AddTransition(Locked, Closed, UnlockDoor).
        Build()

    ctx := context.Background()
    
    fmt.Println(door.GetCurrentState()) // Output: Closed
    
    _ = door.Trigger(ctx, OpenDoor)
    fmt.Println(door.GetCurrentState()) // Output: Open
    
    _ = door.Trigger(ctx, CloseDoor)
    fmt.Println(door.GetCurrentState()) // Output: Closed
    
    isLocked = true
    err := door.Trigger(ctx, LockDoor)
    if err != nil {
        fmt.Println("Cannot lock the door") // This will be printed
    }
    fmt.Println(door.GetCurrentState()) // Output: Closed

    isLocked = false
    _ = door.Trigger(ctx, LockDoor)
    fmt.Println(door.GetCurrentState()) // Output: Locked
}
```

In this example, we've added a guard function to the `LockDoor` transition. The door can only be locked if `isLocked` is false. This demonstrates how you can use external conditions to control state transitions.

## Documentation

For detailed documentation, please visit [GoDoc](https://godoc.org/github.com/upamune/zstate).

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

If you have any questions or feedback, please open an issue on [GitHub Issues](https://github.com/upamune/z-state/issues).

---

Made with ❤️ by [upamune](https://github.com/upamune)