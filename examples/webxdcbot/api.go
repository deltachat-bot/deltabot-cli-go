/*
RPC API definitions.

Here you can define functions that can be called on the mini-app side.

```
webxdc.sendUpdate({payload: {id: "1", method: "Noop", params: []}}, "");
```
*/
package main

import (
	"github.com/deltachat-bot/deltabot-cli-go/v2/xdcrpc"
)

// You must put your available RPC API in a custom type
type API struct{}

// Function without arguments or return value
func (api *API) Noop() {
	// do nothing
}

// Function with return value but no *xdcrpc.Error
func (api *API) Echo(text string) string {
	return text
}

// Function that might return an xdcrpc.Error.
// Functions must return `*xdcrpc.Error` instead of `error`
func (api *API) Divide(a int, b int) (int, *xdcrpc.Error) {
	if b == 0 {
		return 0, &xdcrpc.Error{Code: 1, Message: "Division by zero"}
	}
	return a / b, nil
}
