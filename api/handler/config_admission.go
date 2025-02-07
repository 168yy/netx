package handler

import (
	"context"
	"github.com/168yy/netx/x/app"
	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/admission"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Admission = hAdmission{}
)

type hAdmission struct{}

type CreateAdmissionReq struct {
	g.Meta `path:"/admissions" method:"post" tags:"Admissions" summary:"Create a new admission, the name of admission must be unique in admission list."`
	// in: body
	Data config.AdmissionConfig `json:"data"`
}

func (h *hAdmission) CreateAdmission(ctx context.Context, req *CreateAdmissionReq) (res *NullStructRes, err error) {
	res = &NullStructRes{}
	if req.Data.Name == "" {
		return nil, ErrInvalid
	}

	v := parser.ParseAdmission(&req.Data)

	if err := app.Runtime.AdmissionRegistry().Register(req.Data.Name, v); err != nil {
		return nil, ErrDup
	}

	_ = config.OnUpdate(func(c *config.Config) error {
		c.Admissions = append(c.Admissions, &req.Data)
		return nil
	})

	return res, nil
}

type UpdateAdmissionReq struct {
	g.Meta `path:"/admissions/{admission}" method:"put" tags:"Admissions" summary:"Update admission by name, the admission must already exist."`
	// in: path
	// required: true
	Admission string `json:"admission"`
	// in: body
	Data config.AdmissionConfig `json:"data"`
}

func (h *hAdmission) UpdateAdmission(ctx context.Context, req *UpdateAdmissionReq) (res *NullStructRes, err error) {
	res = &NullStructRes{}

	if !app.Runtime.AdmissionRegistry().IsRegistered(req.Admission) {
		return nil, ErrNotFound
	}

	req.Data.Name = req.Admission

	v := parser.ParseAdmission(&req.Data)

	app.Runtime.AdmissionRegistry().Unregister(req.Admission)

	if err := app.Runtime.AdmissionRegistry().Register(req.Admission, v); err != nil {
		return nil, ErrDup
	}

	_ = config.OnUpdate(func(c *config.Config) error {
		for i := range c.Admissions {
			if c.Admissions[i].Name == req.Admission {
				c.Admissions[i] = &req.Data
				break
			}
		}
		return nil
	})

	return res, nil
}

type DeleteAdmissionReq struct {
	g.Meta `path:"/admissions/{admission}" method:"delete" tags:"Admissions" summary:"Delete admission by name."`
	// in: path
	// required: true
	Admission string `json:"admission"`
}

func (h *hAdmission) DeleteAdmission(ctx context.Context, req *DeleteAdmissionReq) (res *NullStructRes, err error) {
	res = &NullStructRes{}

	if !app.Runtime.AdmissionRegistry().IsRegistered(req.Admission) {
		return nil, ErrNotFound
	}
	app.Runtime.AdmissionRegistry().Unregister(req.Admission)

	_ = config.OnUpdate(func(c *config.Config) error {
		admissiones := c.Admissions
		c.Admissions = nil
		for _, s := range admissiones {
			if s.Name == req.Admission {
				continue
			}
			c.Admissions = append(c.Admissions, s)
		}
		return nil
	})

	return res, nil
}
