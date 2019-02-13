package router

import (
	"regexp"
	"strings"
	"time"

	"github.com/alex19pov31/croner"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type routeCallback func(update tgbotapi.Update)
type timerCallback func(t time.Time)

type route struct {
	comand       string
	queryData    []byte
	containText  string
	pregTemplate string
	callback     routeCallback
}

type routeTimer struct {
	duration time.Duration
	callback timerCallback
}

func (rt *routeTimer) Run() {
	t := time.NewTicker(rt.duration)
	for {
		rt.callback(<-t.C)
	}
}

type routeCronTimer struct {
	*croner.CronTimer
	callback timerCallback
}

func (rct *routeCronTimer) Run() {
	t := time.NewTicker(time.Minute)
	for {
		tm := <-t.C
		if rct.Check(tm) {
			rct.callback(tm)
		}
	}
}

func (r *route) check(update tgbotapi.Update, callback chan routeCallback) {
	if update.CallbackQuery != nil && len(r.queryData) > 0 &&
		string(r.queryData) == string(update.CallbackQuery.Data) {
		callback <- r.callback
		return
	}

	if update.Message == nil {
		return
	}

	context := update.Message.Text
	if r.comand != "" && r.comand == context {
		callback <- r.callback
		return
	}

	if r.containText != "" && strings.Contains(context, r.containText) {
		callback <- r.callback
		return
	}

	if r.pregTemplate != "" {
		if checked, _ := regexp.MatchString(r.pregTemplate, context); checked {
			callback <- r.callback
			return
		}
	}
}

// ITimerRoute timer interface
type ITimerRoute interface {
	Run()
}

// RouteGroup - группа роутов
type RouteGroup struct {
	routes      []route
	routeTimers []ITimerRoute
	callback    chan routeCallback
}

// AddQueryRoute - роутер для Callback запросов
func (rg *RouteGroup) AddQueryRoute(data []byte, callback routeCallback) {
	rg.routes = append(rg.routes, route{queryData: data, callback: callback})
}

// AddCommandRoute - роутер для простой команды
func (rg *RouteGroup) AddCommandRoute(comand string, callback routeCallback) {
	rg.routes = append(rg.routes, route{comand: comand, callback: callback})
}

// AddContainRoute - роутер с проверкой содержания текста
func (rg *RouteGroup) AddContainRoute(containText string, callback routeCallback) {
	rg.routes = append(rg.routes, route{containText: containText, callback: callback})
}

// AddPregRoute - роутер с проверкой гегулярного выражения
func (rg *RouteGroup) AddPregRoute(pregTemplate string, callback routeCallback) {
	rg.routes = append(rg.routes, route{pregTemplate: pregTemplate, callback: callback})
}

// AddTimer - таймер
func (rg *RouteGroup) AddTimer(duration time.Duration, callback timerCallback) {
	rg.routeTimers = append(rg.routeTimers, &routeTimer{duration: duration, callback: callback})
}

// AddCronTimer - таймер с правилом запуска
func (rg *RouteGroup) AddCronTimer(cronRecord string, callback timerCallback) {
	rg.routeTimers = append(rg.routeTimers, &routeCronTimer{CronTimer: croner.NewCronTimer(cronRecord), callback: callback})
}

// Run - запуск роутера
func (rg *RouteGroup) Run(update tgbotapi.Update) {
	rg.callback = make(chan routeCallback)
	for _, rt := range rg.routes {
		go func(update tgbotapi.Update, rt route, callback chan routeCallback) {
			rt.check(update, callback)
		}(update, rt, rg.callback)
	}

	go func(rg *RouteGroup) {
		select {
		case fn := <-rg.callback:
			fn(update)
		case <-time.NewTicker(time.Second).C:
		}
	}(rg)
}

// RunTimer - запуск таймеров
func (rg *RouteGroup) RunTimer() {
	for _, rTimer := range rg.routeTimers {
		go rTimer.Run()
	}
}

// NewRouteGroup - создает новую группу роутов
func NewRouteGroup() *RouteGroup {
	return &RouteGroup{}
}
