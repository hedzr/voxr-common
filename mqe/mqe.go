/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mqe

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

const (
	IM_EVENT_BUS      = "im_event_bus"      // IM 标准事件总线，未单列的全局事件一律走此交换机; 单列的则独立配置交换机;
	IM_EVENT_CAST     = "im_event_cast"     // 广播事件
	IM_HOOK_EVENT_BUS = "im_hook_event_bus" // webhooks' eventBus
)

type (
	MqHub struct {
		MqBase
		mqAll
		demoSend    bool
		demoBusName string
	}
)

//
// starts the daemon for producer/publisher
//
func StartPublisherDaemon(chExitGlobal chan struct{}) *MqHub {
	var hub = &MqHub{}

	hub.chExitExternal = chExitGlobal
	hub.chExit = make(chan struct{}, 3)
	hub.chReconnect = make(chan bool, 2)
	hub.conf = loadConfigSections()

	hub.reconnecFunc = hub.defaultReconnectLoop

	if err := hub.Open(hub.run); err != nil {
		logrus.Warnf("[MQ][HUB] CANNOT connect to MQ.")
	}

	hub.EnableReconnectLoop()

	return hub
}

func (s *MqHub) WithExitSignal(chExit chan struct{}) *MqHub {
	s.chExitExternal = chExit
	return s
}

func (s *MqHub) WithDebug(debug bool, busName string) *MqHub {
	s.demoSend = debug
	s.demoBusName = busName
	return s
}

// func Publish(busName, routingKey, contentType string, msg []byte) {
// 	hub.Publish(busName, routingKey, contentType, msg)
// }

func (s *MqHub) Publish(msg []byte, busName, routingKey, contentType string) {
	s.PublishX(msg, busName, routingKey, contentType, false, false)
}

func (s *MqHub) PublishX(msg []byte, busName, routingKey, contentType string, mandatory, immediate bool) {
	if s.channel == nil {
		s.reconnect()
		return
	}

	if bus, ok := s.conf.Publish[busName]; ok {
		err := s.channel.Publish(bus.Exchange.Exchange, routingKey,
			mandatory, immediate,
			amqp.Publishing{ContentType: contentType, Body: msg})
		if err != nil {
			logrus.Errorf("PublishX Err: %v", err)
			// failOnError(err, "Publish")
			s.reconnect()
		}
	}
}

func (s *MqHub) sendDemo(tm time.Time) {
	msg := fmt.Sprintf("Hello World! %v", tm)
	s.Publish([]byte(msg), s.demoBusName, "ev.fx.im.test", "text/plain")
}

func (s *MqHub) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		logrus.Debug("[MQ][HUB] MqHub.run() stopped.")
		ticker.Stop()
	}()

	logrus.Debugf("[MQ][HUB] MqHub.run() started.")

	for {
		select {
		case tm := <-ticker.C:
			if s.demoSend {
				s.sendDemo(tm)
			}
		case <-s.chExit:
			return
		case <-s.chExitExternal:
			return
		}
	}
}
