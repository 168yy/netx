package bot

import (
	"flag"
	"fmt"
	telebot "github.com/168yy/gfbot"
	"github.com/168yy/netx/x/config"
	"github.com/gogf/gf/v2/util/gconv"
	"strings"
)

func (h *hEvent) OnParsingCommand(c telebot.Context) error {
	var (
		services stringList
		nodes    stringList
		msg      string
		err      error
	)

	payload := c.Message().Text
	str := strings.Split(payload, " ")
	cmd := flag.NewFlagSet(gconv.String(str[:1]), flag.ContinueOnError)
	cmd.Var(&services, "L", "service list")
	cmd.Var(&nodes, "F", "chain node list")
	err = cmd.Parse(str[1:])
	if err != nil {
		return c.Reply("OnParsingCommand err:", err.Error())
	}
	cfg, err := buildConfigFromCmd(services, nodes)
	if err != nil {
		return c.Reply("OnParsingCommand err:", err.Error())
	}
	if cfg != nil {
		msg, err = ConvertJsonMsg(cfg)
		if err != nil {
			return c.Reply("OnParsingCommand ConvertJsonMsg err:", err.Error())
		}
		fmt.Println("OnParsingCommand msg:", msg)
		msg = fmt.Sprintf(CodeTpl, CodeStart, msg, CodeEnd)
	}

	return c.Reply(msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
}

func (h *hEvent) OnStartCommand(c telebot.Context) error {
	payload := c.Message().Text
	user := c.Message().Sender
	str := strings.Split(payload, " ")
	token := ""
	if len(str) >= 2 {
		token = strings.Join(str[1:], "")
	}
	return c.Send(fmt.Sprintf("欢迎 %s 加入 参数:%s", user.Username, token))
}

func (h *hEvent) OnGostCommand(c telebot.Context) error {
	var (
		services stringList
		nodes    stringList
		err      error
	)

	payload := c.Message().Text
	str := strings.Split(payload, " ")
	cmd := flag.NewFlagSet(gconv.String(str[:1]), flag.ContinueOnError)
	cmd.Var(&services, "L", "service list")
	cmd.Var(&nodes, "F", "chain node list")
	err = cmd.Parse(str[1:])
	if err != nil {
		return c.Reply("OnGostCommand err:", err.Error())
	}
	cfg, err := buildConfigFromCmd(services, nodes)
	if err != nil {
		return c.Reply("OnGostCommand err:", err.Error())
	}
	config.Set(cfg)
	for _, svc := range buildService(cfg) {
		svc := svc
		go func() {
			svc.Serve()
		}()
	}
	return h.OnClickServices(c)
}
