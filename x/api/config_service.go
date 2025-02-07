package api

import (
	"fmt"
	"github.com/168yy/netx/x/app"
	"net/http"
	"strings"

	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/service"
	"github.com/gin-gonic/gin"
)

// swagger:parameters createServiceRequest
type createServiceRequest struct {
	// in: body
	Data config.ServiceConfig `json:"data"`
}

// successful operation.
// swagger:response createServiceResponse
type createServiceResponse struct {
	Data Response
}

func createService(ctx *gin.Context) {
	// swagger:route POST /config/services Service createServiceRequest
	//
	// Create a new service, the name of the service must be unique in service list.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: createServiceResponse

	var req createServiceRequest
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Data.Name)
	if name == "" {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeInvalid, "service name is required"))
		return
	}
	req.Data.Name = name

	if app.Runtime.ServiceRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("service %s already exists", name)))
		return
	}

	svc, err := parser.ParseService(&req.Data)
	if err != nil {
		writeError(ctx, NewError(http.StatusInternalServerError, ErrCodeFailed, fmt.Sprintf("create service %s failed: %s", name, err.Error())))
		return
	}

	if err := app.Runtime.ServiceRegistry().Register(name, svc); err != nil {
		svc.Close()
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("service %s already exists", name)))
		return
	}

	go svc.Serve()

	config.OnUpdate(func(c *config.Config) error {
		c.Services = append(c.Services, &req.Data)
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters updateServiceRequest
type updateServiceRequest struct {
	// in: path
	// required: true
	Service string `uri:"service" json:"service"`
	// in: body
	Data config.ServiceConfig `json:"data"`
}

// successful operation.
// swagger:response updateServiceResponse
type updateServiceResponse struct {
	Data Response
}

func updateService(ctx *gin.Context) {
	// swagger:route PUT /config/services/{service} Service updateServiceRequest
	//
	// Update service by name, the service must already exist.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: updateServiceResponse

	var req updateServiceRequest
	ctx.ShouldBindUri(&req)
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Service)

	old := app.Runtime.ServiceRegistry().Get(name)
	if old == nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("service %s not found", name)))
		return
	}
	old.Close()

	req.Data.Name = name

	svc, err := parser.ParseService(&req.Data)
	if err != nil {
		writeError(ctx, NewError(http.StatusInternalServerError, ErrCodeFailed, fmt.Sprintf("create service %s failed: %s", name, err.Error())))
		return
	}

	app.Runtime.ServiceRegistry().Unregister(name)

	if err := app.Runtime.ServiceRegistry().Register(name, svc); err != nil {
		svc.Close()
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("service %s already exists", name)))
		return
	}

	go svc.Serve()

	config.OnUpdate(func(c *config.Config) error {
		for i := range c.Services {
			if c.Services[i].Name == name {
				c.Services[i] = &req.Data
				break
			}
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters deleteServiceRequest
type deleteServiceRequest struct {
	// in: path
	// required: true
	Service string `uri:"service" json:"service"`
}

// successful operation.
// swagger:response deleteServiceResponse
type deleteServiceResponse struct {
	Data Response
}

func deleteService(ctx *gin.Context) {
	// swagger:route DELETE /config/services/{service} Service deleteServiceRequest
	//
	// Delete service by name.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: deleteServiceResponse

	var req deleteServiceRequest
	ctx.ShouldBindUri(&req)

	name := strings.TrimSpace(req.Service)

	svc := app.Runtime.ServiceRegistry().Get(name)
	if svc == nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("service %s not found", name)))
		return
	}

	app.Runtime.ServiceRegistry().Unregister(name)
	svc.Close()

	config.OnUpdate(func(c *config.Config) error {
		services := c.Services
		c.Services = nil
		for _, s := range services {
			if s.Name == name {
				continue
			}
			c.Services = append(c.Services, s)
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}
