package router

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type routeCallback func(update tgbotapi.Update)

type route struct {
	comand       string
	containText  string
	pregTemplate string
	callback     routeCallback
}

func (r *route) check(context string, callback chan routeCallback) {
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

// RouteGroup - группа роутов
type RouteGroup struct {
	routes   []route
	callback chan routeCallback
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

// Run - запуск роутера
func (rg *RouteGroup) Run(update tgbotapi.Update) {
	go func(context string, callback chan routeCallback) {
		for _, route := range rg.routes {
			go route.check(context, callback)
		}
	}(update.Message.Text, rg.callback)

	go func(rg *RouteGroup) {
		select {
		case fn := <-rg.callback:
			fn(update)
		case <-time.NewTicker(time.Second).C:
		}
	}(rg)
}

// NewRouteGroup - создает новую группу роутов
func NewRouteGroup() *RouteGroup {
	return &RouteGroup{}
}
