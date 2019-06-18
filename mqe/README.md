# mqe - MQ extensions

帮助连接到 MQ，管理 MQ 连接，与其交互。

帮助实现 Topic Publish/Subscribe 模型。

具有自动重连机制。




### 实现发布者

注意要提供一个全局的退出信号，在主程序退出时发出该信号能够安全地停止发布者的内部循环。

```go
var mqEngine *mqe.MqHub

func main(){
	mqEngine = mqe.Open(common.AppExitCh)
	
	// 使能内部的发件定时器以发送调试用的消息包
	// mqe.WithDebug(true)
	
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer func() {
			ticker.Stop()
		}()
	
		for {
			select {
			case tm := <-ticker.C:
				msg := fmt.Sprintf("Hello World! %v", tm)
				mqEngine.Publish(IM_EVENT_BUS, "ev.fx.im.test", "text/plain", []byte(msg))
			}
		}
	}()
	
	mqEngine.CloseAll()
}
```


### 实现消费者

注意要提供一个全局的退出信号，在主程序退出时发出该信号能够安全地停止消费者的内部循环。

```go
mqe.NewClient(mqe.IM_EVENT_BUS, common.AppExitCh).
	// WithExitSignal(common.AppExitCh).
	NewConsumer("abc", mqe.IM_EVENT_BUS, func(d amqp.Delivery) {
		logrus.Debugf(" [x] %v | %v", string(d.Body), d.Body)
	})

// optional: mqe.CloseClient(client)
```




## `fanout`

实现 fanout 消息的订阅，在发布端，需要调整bus配置参数，此处略过；在客户端代码上有一点不同。

### 实现多个消费者以便同时消费fanout消息

```go
mqe.NewClient(DEFAULT_BUS, common.AppExitCh).
	NewConsumerWithQueueName("abc", "fx.q.recv1", DEFAULT_BUS, func(d amqp.Delivery) {
		logrus.Debugf(" [x] %v | %v", string(d.Body), d.Body)
	}).
	NewConsumerWithQueueName("def", "fx.q.recv3", DEFAULT_BUS, func(d amqp.Delivery) {
		logrus.Debugf(" [-] %v | %v", string(d.Body), d.Body)
	}).
	NewConsumerWithQueueName("ghi", "fx.q.recv2", DEFAULT_BUS, func(d amqp.Delivery) {
		logrus.Debugf(" [+] %v | %v", string(d.Body), d.Body)
	})
```

由于 `fanout` 消息在被发布到 exchange 时，自动复制给 exchange 的所有绑定队列，所以客户端应该使用不同的队列名才能达到接收的效果。

按照示例代码的写法，每当发布者向 DEFAULT_BUS 发布一条消息时，所有消费者将会分别同时收到该消息。

也可以使用 `NewConsumerWith(consumerName, queueName, busName, autoAck, exclusive, noLocal, noWait, args, onRecv)`:

```go
mqe.NewClient(DEFAULT_BUS, common.AppExitCh).
	NewConsumerWith("abc", "fx.q.recv1", DEFAULT_BUS, true, false, false, false, nil, func(d amqp.Delivery) {
		logrus.Debugf(" [x] %v | %v", string(d.Body), d.Body)
	})
```



## 配置文件章节

根据配置文件的 `server.mq.publish` 章节定义的 bus 清单，`mqe.Open(...)` 会
建立和管理所有的 exchanges, queues。

你无需为每个 exchange, queue, 或者 bus 与 bind 建立一个 mqe.Open(...) 实例
对象，所有的 总线(bus) 使用同一个 MqHub 对象完成管理。

预定义的配置章节结构如下：

```yaml
server:
  mq:
    backend: rabbitmq    # current backend
    env: devel           # current mode: devel/staging/prod, ...
    debug: true          # uses debug mode
    backends:
      rabbitmq:
        devel:
          url: "amqp://guest:guest@localhost:5672/"
          connectionTimeout: 30000
          maxOpenConns: 100
          maxIdleConns: 10
        prod:
          url: "amqp://guest:guest@localhost:5672/"
          connectionTimeout: 30000
          maxOpenConns: 100
          maxIdleConns: 10

    clients:
      - im_event_bus

    publish:
      logger_bus:
      monitor_bus:
      config_cast:

      # im-platform event cast
      im_event_bus:
        exchange:
          exchange:   fx.ex.event_bus
          type:       topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.event_bus
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.im.#
          noWait:     false
          arguments:  {}

      im_hook_event_bus:
        exchange:
          exchange:   fx.ex.event_bus
          type:       topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.event_bus.hooks
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.im.hooks.#
          noWait:     false
          arguments:  {}

      im_app_event_bus:
        exchange:
          exchange:   fx.ex.event_bus
          type:       topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.event_bus.apps
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.im.apps.#
          noWait:     false
          arguments:  {}

      # im-platform event bus
      im_event_cast:
        exchange:
          exchange:   fx.ex.event_cast
          type:       fanout # direct, fanout, topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.event_cast
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.im.# # as a sample: info,warning,error
          noWait:     false
          arguments:  {}

      sms_req:
        exchange:
          exchange:   fx.ex.sms_req
          type:       topic # direct, fanout, topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.sms_req
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.sms.#
          noWait:     false
          arguments:  {}
      mail_req:
        exchange:
          exchange:   fx.ex.mail_req
          type:       topic       # direct, fanout, topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.email_req
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.mail.#      # fx.mail.{user.{login,register,find-pwd},org.{sns.{like,fav,mentioned,...},ann.{publish,revoke}}}
          noWait:     false
          arguments:  {}
      cmdlet:
        exchange:
          exchange:   fx.ex.cmdlet
          type:       topic       # direct, fanout, topic
          passive:    false
          durable:    false
          autoDelete: false
          internal:   false
          noWait:     false
          arguments:  {}
        queue:
          queue:      fx.q.cmdlet
          passive:    false
          durable:    false
          exclusive:  false
          autoDelete: false
          noWait:     false
          arguments:  {}
        bind:
          queue:
          exchange:
          routingKey: fx.exec.#
          noWait:     false
          arguments:  {}

```


