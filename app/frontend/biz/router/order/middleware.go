// Code generated by hertz generator.

package order

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/qitian118/gomall/app/frontend/middleware"
)

func rootMw() []app.HandlerFunc {
	// your code...
	return []app.HandlerFunc{middleware.AuthMiddleware()}
}

func _orderlistMw() []app.HandlerFunc {
	// your code...
	return nil
}
