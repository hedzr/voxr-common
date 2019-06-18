/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache

import (
	"github.com/sirupsen/logrus"
	"time"
)

const insertPeriod = 1 * time.Second

func (h *Hub) run() {
	for {
		h.exited = false
		ticker := time.NewTicker(insertPeriod)
		select {
		case tm := <-ticker.C:
			logrus.Debugf("refreshing clients at %v", tm)
			// uid := rand.Intn(math.MaxInt32)
			// did := strconv.Itoa(rand.Intn(32))
			// zid := rand.Intn(32)
			// PutUserHash(uint64(uid), did, int2str(zid))

		case exit := <-h.exitCh:
			if exit {
				logrus.Infof("chat hub exiting. (WebSocket message processing service)")
				h.exited = true
				return
			}

			// case client := <-h.register:
			// 	h.clients[client] = true
			// 	size := int(unsafe.Sizeof(*client))
			// 	logrus.Debugf("=== new client in. %v clients x %v bytes. %s", hub.clients, size, client.userAgent) // and log it for debugging
			//
			// case client := <-h.unregister:
			// 	if _, ok := h.clients[client]; ok {
			// 		logrus.Println("=== the client leaved.", hub.clients) // and log it for debugging
			// 		delete(h.clients, client)
			// 		close(client.textSend)
			// 	}

		}
	}
}
