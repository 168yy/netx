package bot

import (
	"encoding/json"
	"fmt"
	telebot "github.com/168yy/gfbot"
	"github.com/168yy/gfbot/handlers"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/x/app"
	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/hop"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	HopAdd         = "hopAdd"
	HopUpdate      = "hopUpdate"
	HopExampleJson = `
{
  "name": "hop-0",
  "interface": "192.168.1.2",
  "selector": {
    "strategy": "rand",
    "maxFails": 3,
    "failTimeout": 60
  },
  "bypass": "bypass-0",
  "nodes": [
    {
      "name": "node-0",
      "addr": ":1080",
      "interface": "eth1",
      "bypass": "bypass-0",
      "connector": {
        "type": "socks5",
        "auth": {
          "username": "user",
          "password": "pass"
        },
        "metadata": {
          "foo": "bar"
        }
      },
      "dialer": {
        "type": "tcp",
        "auth": {
          "username": "user",
          "password": "pass"
        },
        "tls": {
          "caFile": "ca.pem",
          "secure": true,
          "serverName": "example.com"
        },
        "metadata": {
          "bar": "baz"
        }
      }
    }
  ]
}
`
)

func (h *hEvent) OnClickHops(c telebot.Context) error {
	var (
		msg string
	)

	cfg := config.Global()
	rowList := make([]telebot.Row, 0)
	btnList := make([]telebot.Btn, 0)
	selector := &telebot.ReplyMarkup{}
	for _, service := range cfg.Hops {
		btnList = append(btnList, selector.Data(fmt.Sprintf("@%s", service.Name), "detailHop", service.Name))
	}
	rowList = append(rowList, selector.Split(MaxCol, btnList)...)
	rowList = append(rowList, selector.Row(
		selector.Data("@添加跳跃点", "addHop", "addHop"),
		selector.Data("« 返回 服务列表", "backServices", "backServices"),
	))

	selector.Inline(
		rowList...,
	)
	msg = "Hops List:\n"
	if c.Callback() != nil {
		return c.Edit(msg, &telebot.SendOptions{ReplyMarkup: selector, ParseMode: telebot.ModeMarkdownV2})
	}

	return c.Reply(msg, &telebot.SendOptions{ReplyMarkup: selector, ParseMode: telebot.ModeMarkdownV2})
}

func (h *hEvent) OnClickDetailHop(c telebot.Context) error {
	var (
		msg string
		str string
		err error
	)

	serviceName := c.Callback().Data
	cfg := config.Global()
	var srv *config.HopConfig
	for _, service := range cfg.Hops {
		if service.Name == serviceName {
			srv = service
		}
	}
	if srv != nil {
		str, err = ConvertJsonMsg(srv)
		if err != nil {
			return c.Reply("OnClickDetailHop ConvertJsonMsg err:", err.Error())
		}
		msg = fmt.Sprintf(CodeMsgTpl, msg, CodeStart, str, CodeEnd)
	}
	selector := &telebot.ReplyMarkup{}
	selector.Inline(
		selector.Row(
			selector.Data("@更新跳跃点", "updateHop", serviceName),
			selector.Data("@删除跳跃点", "delHop", serviceName),
		),
		selector.Row(
			selector.Data("« 返回 服务列表", "backServices", "backServices"),
		),
	)
	return c.Edit(msg, &telebot.SendOptions{ReplyMarkup: selector, ParseMode: telebot.ModeMarkdownV2})
}

func (h *hEvent) OnClickDelHop(c telebot.Context) error {
	serviceName := c.Callback().Data
	svc := app.Runtime.HopRegistry().Get(serviceName)
	if svc == nil {
		return c.Send(ErrNotFound)
	}

	app.Runtime.HopRegistry().Unregister(serviceName)

	_ = config.OnUpdate(func(c *config.Config) error {
		hops := c.Hops
		c.Hops = nil
		for _, s := range hops {
			if s.Name == serviceName {
				continue
			}
			c.Hops = append(c.Hops, s)
		}
		return nil
	})
	_ = c.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("%s 删除成功!", serviceName)})
	return h.OnClickHops(c)
}

func AddHopConversation(entry, cancel string) handlers.Conversation {
	return handlers.NewConversation(
		entry,
		telebot.HandlerFunc(startAddHopHandler), // 入口
		map[string][]telebot.IHandler{
			HopAdd: {telebot.HandlerFunc(addHopHandler)},
		}, // states状态
		&handlers.ConversationOpts{
			ExitName:     cancel,
			ExitHandler:  telebot.HandlerFunc(cancelHopHandler),
			AllowReEntry: true,
		}, // options
	)
}

func UpdateHopConversation(entry, cancel string) handlers.Conversation {
	return handlers.NewConversation(
		entry,
		telebot.HandlerFunc(startUpdateHopHandler), // 入口
		map[string][]telebot.IHandler{
			HopUpdate: {telebot.HandlerFunc(updateHopHandler)},
		}, // states状态
		&handlers.ConversationOpts{
			ExitName:     cancel,
			ExitHandler:  telebot.HandlerFunc(cancelHopHandler),
			AllowReEntry: true,
		}, // options
	)
}

func startAddHopHandler(ctx telebot.Context) error {
	example := fmt.Sprintf(CodeTpl, CodeStart, HopExampleJson, CodeEnd)
	err := ctx.Send(fmt.Sprintf("请输入 跳跃点 配置?\nExample：%s\n您可以随时键入 /cancel 来取消该操作。", example), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	// 设置当前用户下一个入口
	return handlers.NextConversationState(HopAdd)
}

func addHopHandler(ctx telebot.Context) error {
	var (
		data config.HopConfig
	)

	str := ctx.Message().Text
	err := json.Unmarshal([]byte(str), &data)
	if err != nil {
		return ctx.Reply("addHopHandler json.Unmarshal error:", err.Error())
	}
	v, err := parser.ParseHop(&data, logger.Default())
	if err != nil {
		return ctx.Reply(ErrCreate)
	}

	if err = app.Runtime.HopRegistry().Register(data.Name, v); err != nil {
		return ctx.Reply(ErrDup)
	}

	_ = config.OnUpdate(func(c *config.Config) error {
		c.Hops = append(c.Hops, &data)
		return nil
	})

	_ = ctx.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("%s 添加成功!", data.Name)})
	_ = Event.OnClickHops(ctx)

	return handlers.EndConversation()
}

func startUpdateHopHandler(ctx telebot.Context) error {
	srvName := ctx.Callback().Data
	err := ctx.Bot().Store().UpdateData(ctx, HopUpdate, srvName)
	if err != nil {
		return fmt.Errorf("failed UpdateData message: %w", err)
	}
	example := fmt.Sprintf(CodeTpl, CodeStart, HopExampleJson, CodeEnd)
	err = ctx.Send(fmt.Sprintf("请输入 跳跃点 配置?\nExample：%s\n您可以随时键入 /cancel 来取消该操作。", example), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	// 设置当前用户下一个入口
	return handlers.NextConversationState(HopUpdate)
}

func updateHopHandler(ctx telebot.Context) error {
	var (
		data config.HopConfig
	)
	state, err := ctx.Bot().Store().Get(ctx)
	if err != nil {
		return ctx.Reply("updateHopHandler Store().Get error:", err.Error())
	}
	srvName := ""
	if sn, ok := state.Data[HopUpdate]; ok {
		srvName = gconv.String(sn)
	}
	fmt.Println(fmt.Sprintf("srvName :%s", srvName))
	if srvName == "" {
		return ctx.Reply(ErrInvalid)
	}
	str := ctx.Message().Text
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		return ctx.Reply("updateHopHandler json.Unmarshal error:", err.Error())
	}

	if !app.Runtime.HopRegistry().IsRegistered(srvName) {
		return ctx.Reply(ErrNotFound)
	}

	data.Name = srvName

	v, err := parser.ParseHop(&data, logger.Default())
	if err != nil {
		return ctx.Reply(ErrCreate)
	}
	app.Runtime.HopRegistry().Unregister(srvName)

	if err = app.Runtime.HopRegistry().Register(srvName, v); err != nil {
		return ctx.Reply(ErrDup)
	}

	_ = config.OnUpdate(func(c *config.Config) error {
		for i := range c.Hops {
			if c.Hops[i].Name == srvName {
				c.Hops[i] = &data
				break
			}
		}
		return nil
	})

	_ = ctx.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("%s 更新成功!", data.Name)})
	_ = Event.OnClickHops(ctx)

	return handlers.EndConversation()
}

func cancelHopHandler(ctx telebot.Context) error {
	err := ctx.Reply("添加 跳跃点 已被取消。 还有什么我可以为你做的吗？", &telebot.SendOptions{})
	if err != nil {
		return fmt.Errorf("failed to send cancelHandler message: %w", err)
	}
	//return handlers.EndConversation()
	return nil
}
