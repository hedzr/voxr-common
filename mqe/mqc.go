/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mqe

import (
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

type (
	MqClient struct {
		MqBase
		mqAll
		busName        string
		chReconnect    chan *consumerPkg
		chExitExternal chan struct{}
	}

	consumerPkg struct {
		consumerName string
		queueName    string
		autoAck      bool
		exclusive    bool
		noLocal      bool
		noWait       bool
		args         amqp.Table
		bus          *Bus
		onRecv       func(d amqp.Delivery)
	}
)

//
// create an new consumer
//
func NewClient(defaultBusName string, chGlobalExit chan struct{}) *MqClient {
	var c = &MqClient{}
	c.chExit = make(chan struct{}, 3)
	c.chReconnect = make(chan *consumerPkg, 3)
	c.chExitExternal = chGlobalExit // make(chan bool, 3)
	c.conf = loadConfigSections()
	c.busName = defaultBusName
	if err := c.Open(c.dummy); err != nil {
		logrus.Warnf("[MQ][C] cannot connect to MQ")
	}

	c.SetAndEnableReconnectLoop(c.reconnectLoop)
	// go c.reconnectLoop()
	// //c.EnableReconnectLoop()

	return c
}

//
// stops the daemon by `NewClient()`
//
func CloseClient(c *MqClient) {
	c.Close()
}

//
//
//

func (s *MqClient) WithExitSignal(chExit chan struct{}) *MqClient {
	s.chExitExternal = chExit
	return s
}

func (s *MqClient) dummy() {
}

func (s *MqClient) reconnect(pkg *consumerPkg) {
	s.chReconnect <- pkg
}

func (s *MqClient) reconnectLoop() {
	bReconnect := false
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		logrus.Debugf("[MQ][C] MqHub.reconnectLoop() stopped.")
		ticker.Stop()
	}()

	var pkg *consumerPkg

	for {
		if bReconnect && pkg != nil {
			// logrus.Debug("[MQ][C] reconnecting...")
			s.Close()
			if err := s.Open(s.runLooper); err == nil {
				logrus.Debug("[MQ][C] reconnected.")
				bReconnect = false
				ClearChanBuffer(s.chExit, time.Second)
				logrus.Debug("[MQ][C] rerun run().")
				go s.runForConsumer(pkg)
			}
		}

		select {
		case pkg = <-s.chReconnect:
			bReconnect = true
		case <-ticker.C:
			if bReconnect {
				// logrus.Debugf("[MQ][C] reconnect loop: bReconnect=%v", bReconnect)
			}
		case <-s.chExitExternal:
			return
		}
	}
}

func (s *MqClient) NewConsumerWith(consumerName, queueName, busName string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table, onRecv func(d amqp.Delivery)) *MqClient {
	if s.channel != nil {
		if bus, ok := s.conf.Publish[busName]; ok {
			if len(queueName) != 0 {
				_, err := s.channel.QueueDeclare(queueName, bus.Queue.Durable, bus.Queue.AutoDelete,
					bus.Queue.Exclusive, bus.Queue.NoWait, bus.Queue.Arguments)
				if err != nil {
					logrus.Errorf("[MQ][C] Err: %v", err)
					// failOnError(err, "QueueDeclare")
					return s
				}

				err = s.channel.QueueBind(queueName, bus.Bind.RoutingKey, bus.Exchange.Exchange,
					bus.Bind.NoWait, bus.Bind.Arguments)
				if err != nil {
					logrus.Errorf("[MQ][C] Err: %v", err)
					// failOnError(err, "QueueBind")
					return s
				}
			}

			go s.runForConsumer(&consumerPkg{consumerName, queueName, autoAck, exclusive, noLocal, noWait, args, bus, onRecv})
		}
	}
	return s
}

func (s *MqClient) NewConsumerWithQueueName(consumerName, queueName, busName string, onRecv func(d amqp.Delivery)) *MqClient {
	return s.NewConsumerWith(consumerName, queueName, busName, true, false, false, false, nil, onRecv)
}

func (s *MqClient) NewConsumer(consumerName, busName string, onRecv func(d amqp.Delivery)) *MqClient {
	return s.NewConsumerWithQueueName(consumerName, "", busName, onRecv)
}

func (s *MqClient) runForConsumer(pkg *consumerPkg) {
	defer func() {
		logrus.Debug("[MQ][C] mq client run() loop stopped.")
	}()

	logrus.Debug("[MQ][C] mq client run() loop started.")
	time.Sleep(time.Second) // 在重连成功后，给予一个稳定延迟，rmq的可用性需要此延迟

	qName := pkg.queueName
	if len(qName) == 0 {
		qName = pkg.bus.Queue.Queue
	}

	msgs, err := s.channel.Consume(
		qName,
		pkg.consumerName, // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		logrus.Errorf("[MQ][C] Err: %v", err)
		// failOnError(err, "[MQ][C] Failed to register a consumer")
	}

	// go func() {
	// 	for d := range msgs {
	// 		logrus.Debugf(" [x] %s", d.Body)
	// 	}
	// }()

	for {
		if s.conn.IsClosed() {
			logrus.Debugf("[MQ][C] conn is closed, reconnecting...")
			s.reconnect(pkg)
			return
		}

		select {
		case d := <-msgs:
			if len(d.Body) > 0 {
				pkg.onRecv(d)
			}

		case <-s.chExit:
			logrus.Debugf("[MQ][C] chexit return...")
			return
		case <-s.chExitExternal:
			logrus.Debugf("[MQ][C] global chexit return...")
			return
		}
	}
}
