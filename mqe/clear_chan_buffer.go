/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mqe

import "time"

func ClearChanBuffer(ch chan struct{}, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer func() {
		ticker.Stop()
	}()

	for {
		remains := false
		select {
		case <-ch:
			remains = true
		case <-ticker.C:
			remains = false
		}
		if !remains {
			break
		}
	}
}
