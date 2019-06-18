/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mqe

import (
	"bytes"
	"fmt"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

const (
	defaultHeartbeat         = 10 * time.Second
	defaultConnectionTimeout = 30 * time.Second
	defaultLocale            = "en_US"
)

type (
	mqAll struct {
		chExit         chan struct{}
		conf           *Publishes
		conn           *amqp.Connection
		channel        *amqp.Channel
		reconnecFunc   func()
		runLooper      func()
		chReconnect    chan bool
		chExitExternal chan struct{}
	}

	MqBase struct {
	}

	Publishes struct {
		Publish   map[string]*Bus
		mqUrl     string
		vhost     string
		exchange  string
		queueName string
		keyPath   string
		Backend   string
		Debug     bool
	}

	Bus struct {
		Exchange ExchangeDeclare
		Queue    QueueDeclare
		Bind     QueueBind
	}

	ExchangeDeclare struct {
		reserved1  uint16
		Exchange   string
		Type       string
		Passive    bool
		Durable    bool
		AutoDelete bool
		Internal   bool
		NoWait     bool
		Arguments  amqp.Table
	}

	QueueDeclare struct {
		reserved1  uint16
		Queue      string
		Passive    bool
		Durable    bool
		Exclusive  bool
		AutoDelete bool
		NoWait     bool
		Arguments  amqp.Table
	}

	QueueBind struct {
		reserved1  uint16
		Queue      string
		Exchange   string
		RoutingKey string
		NoWait     bool
		Arguments  amqp.Table
	}

	Table map[string]interface{}
)

func bytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}

func failOnError(err error, msg string) {
	if err != nil {
		logrus.Errorf("ERR: %v, e: %v", msg, err)
		// panic("error")
	}
}

//
func loadConfigSections() (conf *Publishes) {
	var err error
	// var b []byte
	// o := vxconf.GetMapR("server.pub.deps.mq", nil)
	// b, err = yaml.Marshal(o)
	// if err != nil {
	// 	return
	// }
	// err = yaml.Unmarshal(b, &conf)
	// if err != nil {
	// 	logrus.Warnf("[MQ][A] There are some errors in loading config section 'server.pub.deps.mq':\n%v", err)
	// }
	if err = vxconf.GetSectionR("server.pub.deps.mq", &conf); err != nil {
		logrus.Warnf("[MQ][A] There are some errors in loading config section 'server.pub.deps.mq':\n%v", err)
	}
	if conf == nil {
		logrus.Fatal("[MQ][A] There are some errors in loading config section 'server.pub.deps.mq':\nsection not defined.")
	}
	logrus.Debugf("[MQ][A] conf: %v", conf)

	runmode := vxconf.GetStringR("runmode", "devel")
	// backend := vxconf.GetStringR("server.mq.backend")
	conf.keyPath = fmt.Sprintf("server.mq.backends.%v.%v", conf.Backend, runmode)
	conf.mqUrl = vxconf.GetStringR(fmt.Sprintf("%v.url", conf.keyPath), "")
	conf.vhost = vxconf.GetStringR(fmt.Sprintf("%v.vhost", conf.keyPath), "")

	return
}

func (s *mqAll) Open(runLooper func()) (err error) {
	// logrus.Debugf("[MQ][A]   connecting to %v ...", s.conf.mqUrl)
	if len(s.conf.vhost) == 0 {
		s.conn, err = amqp.Dial(s.conf.mqUrl)
	} else {
		s.conn, err = amqp.DialConfig(s.conf.mqUrl, amqp.Config{
			Vhost:     s.conf.vhost,
			Heartbeat: defaultHeartbeat,
			Locale:    defaultLocale})
	}
	if err != nil {
		logrus.Errorf("ERR: [MQ][A] %v to %v, e: %v", "connect", s.conf.mqUrl, err)
		return
	}

	// logrus.Debugf("[MQ][HUB]   channel ...")
	s.channel, err = s.conn.Channel()
	if err != nil {
		logrus.Errorf("ERR: [MQ][A] %v, e: %v", "", err)
		return
	}

	logrus.Debug("[MQ][A]   ensure MQ excahnges and queues ...")
	for k, v := range s.conf.Publish {
		logrus.Debugf("[MQ][HUB]     ensure '%v'...", k)
		if v != nil && len(v.Queue.Queue) > 0 && len(v.Exchange.Exchange) > 0 {
			s.ensureBus(v)
		}
	}
	logrus.Debug("[MQ][A]   ensure completed. run looper...")

	s.runLooper = runLooper
	go s.runLooper()

	return
}

func (s *mqAll) Close() {
	if s.channel != nil || s.conn != nil {
		s.chExit <- struct{}{}

		if s.channel != nil {
			if err := s.channel.Close(); err != nil {
				logrus.Errorf("[MQ][A] close channel failed: %v", err)
			}
			s.channel = nil
		}
		if s.conn != nil {
			if err := s.conn.Close(); err != nil {
				logrus.Errorf("[MQ][A] close conn failed: %v", err)
			}
			s.conn = nil
		}
	}
}

// close all resources and reconnect looper by `mqe.StartPublisherDaemon()`, `mqe.NewClient()`, `MqHub.Open()`, `MqClient.Open()`
func (s *mqAll) CloseAll() {
	s.Close()
	s.chExitExternal <- struct{}{}
}

func (s *mqAll) DropAll() {
	logrus.Debug("  [MQ][A] dropping MQ excahnges and queues ...")
	for k, v := range s.conf.Publish {
		logrus.Debugf("    [MQ][A] dropping for '%v'...", k)
		if v != nil && len(v.Queue.Queue) > 0 && len(v.Exchange.Exchange) > 0 {
			s.dropBus(v)
		}
	}
	logrus.Debug("  [MQ][A] dropped")
}

func (s *mqAll) dropBus(bus *Bus) {
	if s.channel == nil {
		s.reconnect()
		return
	}

	var err error

	err = s.channel.QueueUnbind(bus.Queue.Queue, bus.Bind.RoutingKey, bus.Exchange.Exchange, bus.Bind.Arguments)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "QueueUnbind")
		return
	}

	_, err = s.channel.QueueDelete(bus.Queue.Queue, false, false, false)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "QueueDelete")
		return
	}

	err = s.channel.ExchangeDelete(bus.Exchange.Exchange, false, false)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "ExchangeDelete")
		return
	}
}

func (s *mqAll) ensureBus(bus *Bus) {
	if s.channel == nil {
		s.reconnect()
		return
	}

	var err error

	err = s.channel.ExchangeDeclare(bus.Exchange.Exchange, bus.Exchange.Type, bus.Exchange.Durable,
		bus.Exchange.AutoDelete, bus.Exchange.Internal, bus.Exchange.NoWait, bus.Exchange.Arguments)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "ExchangeDeclare")
		return
	}

	_, err = s.channel.QueueDeclare(bus.Queue.Queue, bus.Queue.Durable, bus.Queue.AutoDelete,
		bus.Queue.Exclusive, bus.Queue.NoWait, bus.Queue.Arguments)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "QueueDeclare")
		return
	}

	err = s.channel.QueueBind(bus.Queue.Queue, bus.Bind.RoutingKey, bus.Exchange.Exchange, bus.Bind.NoWait, bus.Bind.Arguments)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "QueueBind")
		return
	}

	// err = s.channel.Publish(s.conf.exchange, "info", false, false, amqp.Publishing{
	// 	ContentType: "text/plain", Body: []byte(mgsConnect),
	// })
	// failOnError(err, "Publish")
}

func (s *mqAll) push() {
	if s.channel == nil {
		s.reconnect()
	}

	mgsConnect := "hello world"
	err := s.channel.ExchangeDeclare(s.conf.exchange, "direct", false, false, false, false, nil)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "ExchangeDeclare")
		return
	}

	_, err = s.channel.QueueDeclare(s.conf.queueName, false, false,
		false, false, nil)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "QueueDeclare")
		return
	}

	err = s.channel.QueueBind(s.conf.queueName, "info", s.conf.exchange, false, nil)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "QueueBind")
		return
	}

	err = s.channel.Publish(s.conf.exchange, "info", false, false, amqp.Publishing{
		ContentType: "text/plain", Body: []byte(mgsConnect),
	})
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "Publish")
		return
	}

	fmt.Println("[MQ][A] push ok")
}

func (s *mqAll) receive() {
	if s.channel == nil {
		s.reconnect()
	}

	msg, ok, err := s.channel.Get(s.conf.queueName, false)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "")
		return
	}
	if !ok {
		fmt.Println("[MQ][A] do not get msg")
		return
	}

	err = s.channel.Ack(msg.DeliveryTag, false)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
		// failOnError(err, "")
		return
	}

	str := string(msg.Body)
	// s := bytesToString(&(msg.Body))
	logrus.Printf("[MQ][A] received msg is :%s", str)
}

func (s *mqAll) AckFor(msg amqp.Delivery, multiple bool) (err error) {
	err = s.channel.Ack(msg.DeliveryTag, multiple)
	if err != nil {
		logrus.Errorf("[MQ][A] Err: %v", err)
	}
	return
}

func (s *mqAll) EnableReconnectLoop() {
	go s.reconnecFunc()
}

func (s *mqAll) SetAndEnableReconnectLoop(fn func()) {
	s.reconnecFunc = fn
	s.EnableReconnectLoop()
}

func (s *mqAll) reconnect() {
	s.chReconnect <- true
}

func (s *mqAll) defaultReconnectLoop() {
	bReconnect := false
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		logrus.Debugf("[MQ][A] maAll.defaultReconnectLoop() stopped.")
		ticker.Stop()
	}()

	for {
		if bReconnect {
			// logrus.Debug("[MQ][A] reconnecting...")
			s.Close()
			if err := s.Open(s.runLooper); err == nil {
				logrus.Debug("[MQ][A] reconnected.")
				bReconnect = false
			}
		}

		select {
		case b := <-s.chReconnect:
			bReconnect = b
		case <-ticker.C:
			if bReconnect {
				// logrus.Debugf("[MQ][A] reconnect loop: bReconnect=%v", bReconnect)
			}
		case <-s.chExitExternal:
			return
		}
	}
}
