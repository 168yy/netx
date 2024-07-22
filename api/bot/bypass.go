package bot

import (
	"encoding/json"
	"fmt"
	telebot "github.com/168yy/gfbot"
	"github.com/168yy/gfbot/handlers"
	"github.com/168yy/netx/x/app"
	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/bypass"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	BypassAdd         = "bypassAdd"
	BypassUpdate      = "bypassUpdate"
	BypassExampleJson = `
{
  "name": "bypass-0",
  "whitelist": false,
  "matchers": [
    "*.example.com",
    ".example.org",
    "0.0.0.0/8"
  ]
}
`
)

func (h *hEvent) OnClickBypasses(c telebot.Context) error {
	var (
		msg string
	)

	cfg := config.Global()
	rowList := make([]telebot.Row, 0)
	btnList := make([]telebot.Btn, 0)
	selector := &telebot.ReplyMarkup{}
	for _, service := range cfg.Bypasses {
		btnList = append(btnList, selector.Data(fmt.Sprintf("@%s", service.Name), "detailBypass", service.Name))
	}
	rowList = append(rowList, selector.Split(MaxCol, btnList)...)
	rowList = append(rowList, selector.Row(
		selector.Data("@添加分流器", "addBypass", "addBypass"),
		selector.Data("« 返回 服务列表", "backServices", "backServices"),
	))

	selector.Inline(
		rowList...,
	)
	msg = "Bypasss List:\n"
	if c.Callback() != nil {
		return c.Edit(msg, &telebot.SendOptions{ReplyMarkup: selector, ParseMode: telebot.ModeMarkdownV2})
	}

	return c.Reply(msg, &telebot.SendOptions{ReplyMarkup: selector, ParseMode: telebot.ModeMarkdownV2})
}

func (h *hEvent) OnClickDetailBypass(c telebot.Context) error {
	var (
		msg string
		str string
		err error
	)

	serviceName := c.Callback().Data
	cfg := config.Global()
	var srv *config.BypassConfig
	for _, service := range cfg.Bypasses {
		if service.Name == serviceName {
			srv = service
		}
	}
	if srv != nil {
		str, err = ConvertJsonMsg(srv)
		if err != nil {
			return c.Reply("OnClickDetailBypass ConvertJsonMsg err:", err.Error())
		}
		msg = fmt.Sprintf(CodeMsgTpl, msg, CodeStart, str, CodeEnd)
	}
	selector := &telebot.ReplyMarkup{}
	selector.Inline(
		selector.Row(
			selector.Data("@更新分流器", "updateBypass", serviceName),
			selector.Data("@删除分流器", "delBypass", serviceName),
		),
		selector.Row(
			selector.Data("« 返回 服务列表", "backServices", "backServices"),
		),
	)
	return c.Edit(msg, &telebot.SendOptions{ReplyMarkup: selector, ParseMode: telebot.ModeMarkdownV2})
}

func (h *hEvent) OnClickDelBypass(c telebot.Context) error {
	serviceName := c.Callback().Data
	svc := app.Runtime.BypassRegistry().Get(serviceName)
	if svc == nil {
		return c.Send(ErrNotFound)
	}

	app.Runtime.BypassRegistry().Unregister(serviceName)

	_ = config.OnUpdate(func(c *config.Config) error {
		bypasses := c.Bypasses
		c.Bypasses = nil
		for _, s := range bypasses {
			if s.Name == serviceName {
				continue
			}
			c.Bypasses = append(c.Bypasses, s)
		}
		return nil
	})
	_ = c.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("%s 删除成功!", serviceName)})
	return h.OnClickBypasses(c)
}

func AddBypassConversation(entry, cancel string) handlers.Conversation {
	return handlers.NewConversation(
		entry,
		telebot.HandlerFunc(startAddBypassHandler), // 入口
		map[string][]telebot.IHandler{
			BypassAdd: {telebot.HandlerFunc(addBypassHandler)},
		}, // states状态
		&handlers.ConversationOpts{
			ExitName:     cancel,
			ExitHandler:  telebot.HandlerFunc(cancelBypassHandler),
			AllowReEntry: true,
		}, // options
	)
}

func UpdateBypassConversation(entry, cancel string) handlers.Conversation {
	return handlers.NewConversation(
		entry,
		telebot.HandlerFunc(startUpdateBypassHandler), // 入口
		map[string][]telebot.IHandler{
			BypassUpdate: {telebot.HandlerFunc(updateBypassHandler)},
		}, // states状态
		&handlers.ConversationOpts{
			ExitName:     cancel,
			ExitHandler:  telebot.HandlerFunc(cancelBypassHandler),
			AllowReEntry: true,
		}, // options
	)
}

func startAddBypassHandler(ctx telebot.Context) error {
	example := fmt.Sprintf(CodeTpl, CodeStart, BypassExampleJson, CodeEnd)
	err := ctx.Send(fmt.Sprintf("请输入 分流器 配置?\nExample：%s\n您可以随时键入 /cancel 来取消该操作。", example), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	// 设置当前用户下一个入口
	return handlers.NextConversationState(BypassAdd)
}

func addBypassHandler(ctx telebot.Context) error {
	var (
		data config.BypassConfig
	)

	str := ctx.Message().Text
	err := json.Unmarshal([]byte(str), &data)
	if err != nil {
		return ctx.Reply("addBypassHandler json.Unmarshal error:", err.Error())
	}
	v := parser.ParseBypass(&data)
	if err = app.Runtime.BypassRegistry().Register(data.Name, v); err != nil {
		return ctx.Reply(ErrDup)
	}

	_ = config.OnUpdate(func(c *config.Config) error {
		c.Bypasses = append(c.Bypasses, &data)
		return nil
	})

	_ = ctx.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("%s 添加成功!", data.Name)})
	_ = Event.OnClickBypasses(ctx)

	return handlers.EndConversation()
}

func startUpdateBypassHandler(ctx telebot.Context) error {
	srvName := ctx.Callback().Data
	err := ctx.Bot().Store().UpdateData(ctx, BypassUpdate, srvName)
	if err != nil {
		return fmt.Errorf("failed UpdateData message: %w", err)
	}
	example := fmt.Sprintf(CodeTpl, CodeStart, BypassExampleJson, CodeEnd)
	err = ctx.Send(fmt.Sprintf("请输入 分流器 配置?\nExample：%s\n您可以随时键入 /cancel 来取消该操作。", example), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	// 设置当前用户下一个入口
	return handlers.NextConversationState(BypassUpdate)
}

func updateBypassHandler(ctx telebot.Context) error {
	var (
		data config.BypassConfig
	)
	state, err := ctx.Bot().Store().Get(ctx)
	if err != nil {
		return ctx.Reply("updateBypassHandler Store().Get error:", err.Error())
	}
	srvName := ""
	if sn, ok := state.Data[BypassUpdate]; ok {
		srvName = gconv.String(sn)
	}
	fmt.Println(fmt.Sprintf("srvName :%s", srvName))
	if srvName == "" {
		return ctx.Reply(ErrInvalid)
	}
	str := ctx.Message().Text
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		return ctx.Reply("updateBypassHandler json.Unmarshal error:", err.Error())
	}

	if !app.Runtime.BypassRegistry().IsRegistered(srvName) {
		return ctx.Reply(ErrNotFound)
	}

	data.Name = srvName

	v := parser.ParseBypass(&data)

	app.Runtime.BypassRegistry().Unregister(srvName)

	if err = app.Runtime.BypassRegistry().Register(srvName, v); err != nil {
		return ctx.Reply(ErrDup)
	}

	_ = config.OnUpdate(func(c *config.Config) error {
		for i := range c.Bypasses {
			if c.Bypasses[i].Name == srvName {
				c.Bypasses[i] = &data
				break
			}
		}
		return nil
	})

	_ = ctx.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("%s 更新成功!", data.Name)})
	_ = Event.OnClickBypasses(ctx)

	return handlers.EndConversation()
}

func cancelBypassHandler(ctx telebot.Context) error {
	err := ctx.Reply("添加 分流器 已被取消。 还有什么我可以为你做的吗？", &telebot.SendOptions{})
	if err != nil {
		return fmt.Errorf("failed to send cancelHandler message: %w", err)
	}
	//return handlers.EndConversation()
	return nil
}
