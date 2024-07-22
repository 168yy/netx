package api

import (
	telebot "github.com/168yy/gfbot"
	"github.com/168yy/netx/core/service"
	"github.com/gogf/gf/v2/net/ghttp"
	"net"
)

type TGBot struct {
	Bot    *telebot.Bot `json:"bot"`
	Domain string       `json:"domain"`
	Token  string       `json:"token"`
}

type Server struct {
	Srv      *ghttp.Server
	Listener net.Listener
	Bot      *TGBot
}

type IApi interface {
	service.IService
	TGBot() *TGBot
}
