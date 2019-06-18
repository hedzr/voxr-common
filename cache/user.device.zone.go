/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/hedzr/voxr-common/xs/mjwt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

// 返回值 err == redis.Nil 表示 uid+did 不存在, 此时并非关键性错误
func GetUserZone(uid uint64, did string) (zoneId string, err error) {
	token, e := GetUserToken(uid, did)
	if e != nil {
		err = e
		return // logrus.Fatalf("GetUserZone() 坏了，redis读不了了: %v", e)
	}
	if e == redis.Nil {
		err = e
		return // not exists
	}

	sc := hub.rd.Get(fmt.Sprintf("%s:token-to-zone:%v", conf.AppName, token))
	// if sc.Err() != nil && err != redis.Nil {
	// 	logrus.Fatalf("GetUserZone() 坏了，redis读不了了: %v", sc)
	// }
	zoneId = sc.Val()
	err = sc.Err()
	return
}

// 返回值 err == redis.Nil 表示 uid+did 不存在, 此时并非关键性错误
func GetUserToken(uid uint64, did string) (token string, err error) {
	zid := vxconf.GetStringR("server.id", "") // see also: id.GenerateInstanceId()
	sc := hub.rd.Get(fmt.Sprintf("%s:zones:t:%v:usr:%v", conf.AppName, zid, did))
	token = sc.Val()
	err = sc.Err()
	return
}

func IsUserInZone(uid uint64, did string, zoneId string) (yes bool) {
	token, err := GetUserToken(uid, did)
	if err == redis.Nil {
		return // not exists
	}
	if err != nil {
		return // logrus.Fatalf("GetUserZone() 坏了，redis读不了了: %v", err)
	}

	ic := hub.rd.Exists(fmt.Sprintf("%s:token-to-zone:%s", conf.AppName, token))
	// if ic.Err() != nil {
	// 	if ic.Err() != redis.Nil {
	// 		logrus.Fatalf("GetUserZone() 坏了，redis读不了了: %v", ic)
	// 	} else {
	// 		return // not exists
	// 	}
	// }
	yes = ic.Val() > 0
	return
}

func TokenToUserId(token string) (uid uint64, did string) {
	tk := jwtDecode(token)
	if tk != nil {
		if c, ok := tk.Claims.(*mjwt.ImClaims); ok {
			uid, _ = strconv.ParseUint(c.Id, 10, 64)
			did = c.DeviceId
		}
	}
	return
}

func IsUserTokenValid(token string) (valid bool) {
	valid, _ = JwtVerifyToken(token)
	// st := hub.rd.HGet(fmt.Sprintf("%s:tokens:%v", conf.AppName, token), "device")
	// if st.Err() != nil {
	// 	// logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", st)
	// 	return
	// }
	// if len(st.String()) > 0 {
	// 	logrus.Debugf("      [cache] token found: %v -> %v", token, st.String())
	// 	valid = true
	// }
	return
}

func AddZone() {
	var zid = vxconf.GetStringR("server.id", "") // see also: id.GenerateInstanceId()
	key := fmt.Sprintf("%s:zones:z", conf.AppName)
	// it := hub.rw.ZAdd(key, redis.Z{0, zid,})
	it := hub.rw.SAdd(key, zid)
	if it.Err() != nil {
		logrus.Warnf("      [cache] AddZone() 坏了，redis写不了了: %v", it)
		return
	}
	logrus.Debugf("      [cache] %v: append zid(%v) if not exists", key, zid)
}

func DelZone() {
	if hub.exited {
		return
	}

	var zid = vxconf.GetStringR("server.id", "") // see also: id.GenerateInstanceId()
	key := fmt.Sprintf("%s:zones:z", conf.AppName)
	// it := hub.rw.ZRem(key, redis.Z{0, zid,})
	it := hub.rw.SRem(key, zid)
	if it.Err() != nil {
		logrus.Warnf("      [cache] DelZone() 坏了，redis写不了了: %v", it)
		return
	}
	logrus.Debugf("      [cache] %v: append zid(%v) if not exists", key, zid)
}

func FindLivingUids(uid uint64) (ret map[string][]string) {
	key := fmt.Sprintf("%s:zones:z", conf.AppName)
	ssc := hub.rd.SMembers(key)
	if ssc.Err() != nil {
		logrus.Warnf("      [cache] FindLivingUids() 坏了，redis读不出来了: %v", ssc)
		return
	}

	// key: zoneId, value: redis key string list which contains uid
	ret = make(map[string][]string)

	// loop all zones
	for _, zid := range ssc.Val() {
		// logrus.Debugf("      [cache] loop for #%v - zid(%v)...", ix, zid)
		ret[zid] = findUidInZone(uid, zid)
	}
	return
}

func findUidInZone(uid uint64, zid string) (keys []string) {
	uidString := strconv.FormatUint(uid, 10)

	keyPattern := fmt.Sprintf("%s:zones:t:%v:usr:*", conf.AppName, zid)
	ssc := hub.rd.Keys(keyPattern)
	if ssc.Err() != nil {
		logrus.Warnf("      [cache] findUidInZone() 坏了，redis读不出来了: %v", ssc)
		return
	}

	// loop all deviceIds
	for _, key := range ssc.Val() {
		// logrus.Debugf("      [cache] loop for #%v - key(%v)...", ix, key)
		ss := strings.Split(key, ":")
		if len(ss) >= 7 {
			// uid := ss[5]
			// did := ss[6]
			if uidString == ss[5] {
				keys = append(keys, key)
			}
		}
	}

	return
}

/**
<root>
	"tokens":token = uid(hex) -> did
	"token-to-zone":token = zoneId
	"users":uid:did = token
	"zones":zoneId = token

<root>:
  im-core:
	zones:
	  z: []zid						// instance id
	  t:							// tokens...
		<zid>:
		  tks: []token				// token string
		  usr:						// users and devices...
			<uid>:<did>: = token	// token string
	  u:
		<uid>:<did>: = token
*/
func PutUserHash(uid uint64, did string, token string, expire time.Duration) {

	var (
		key string
		it  *redis.IntCmd
		// st *redis.StringCmd
		stc *redis.StatusCmd
		// bt *redis.BoolCmd
		zid = vxconf.GetStringR("server.id", "") // see also: id.GenerateInstanceId()
		// zidHMAC := vxconf.GetStringR("service.id.hmac")
	)

	// key = fmt.Sprintf("%s:zones:z", conf.AppName)
	// it = hub.rw.ZAdd(key, redis.Z{0,zid, })
	// // if it.Err() != nil {
	// // 	logrus.Warnf("    [cache] LPushX() 坏了，redis写不了了: %v", it)
	// // 	return
	// // }
	// logrus.Debugf("    [cache] %v: append zid(%v) if not exists", key, zid)

	// add `token` to `<app>:zones:t:<zid>:tks` ordered-list.
	key = fmt.Sprintf("%s:zones:t:%v:tks", conf.AppName, zid)
	it = hub.rw.ZAdd(key, redis.Z{0, token})
	if it.Err() != nil {
		logrus.Warnf("    [cache] LPushX() 坏了，redis写不了了: %v", it)
		return
	}
	// logrus.Debugf("    [cache] %v: append token(%v) if not exists", key, token)

	key = fmt.Sprintf("%s:zones:t:%v:usr:%v:%v", conf.AppName, zid, uid, did)
	stc = hub.rw.Set(key, token, expire)
	if stc.Err() != nil {
		logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", stc)
	}
	// logrus.Debugf("    [cache] %v: set token(%v)", key, token)

	if vxconf.GetBoolR("server.cache.verbose-log", false) {
		logrus.Debugf("    [cache] %v: append token(%v) if not exists", key, token)
	}

	// key = fmt.Sprintf("%s:tokens:%v", conf.AppName, token)
	// bt = hub.rw.HSet(key, did, uid)
	// if bt.Err() != nil {
	// 	logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", bt)
	// }
	// bt = hub.rw.HSet(key, "device", did)
	// if bt.Err() != nil {
	// 	logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", bt)
	// }
	// logrus.Debugf("    [cache] %v: device = did(%v), did(%v) -> uid(%v)", key, did, did, uid)
	//
	// key = fmt.Sprintf("%s:token-to-zone:%v", conf.AppName, token)
	// stc = hub.rw.Set(key, zid, 0)
	// if stc.Err() != nil {
	// 	logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", stc)
	// }
	// logrus.Debugf("    [cache] %v: zid = %v", key, zid)
	//
	// key = fmt.Sprintf("%s:users:%v:%v", conf.AppName, uid, did)
	// stc = hub.rw.Set(key, token, 0)
	// if stc.Err() != nil {
	// 	logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", stc)
	// }
	// logrus.Debugf("    [cache] %v: token = %v", key, token)
	//
	// key = fmt.Sprintf("%s:zones:%v", conf.AppName, zid)
	// stc = hub.rw.Set(key, token, 0)
	// if stc.Err() != nil {
	// 	logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", stc)
	// }
	// logrus.Debugf("    [cache] %v: token = %v", key, token)
}

func DelUserHash(uid uint64, did string, token string) {
	var (
		key string
		it  *redis.IntCmd
		// st *redis.StringCmd
		stc *redis.StatusCmd
		// bt *redis.BoolCmd
		zid = vxconf.GetStringR("server.id", "") // see also: id.GenerateInstanceId()
		// zidHMAC := vxconf.GetStringR("service.id.hmac")
	)

	key = fmt.Sprintf("%s:zones:t:%v:tks", conf.AppName, zid)
	it = hub.rw.ZRem(key, redis.Z{0, token})
	if it.Err() != nil {
		logrus.Warnf("    [cache] LPushX() 坏了，redis写不了了: %v", it)
		return
	}
	// logrus.Debugf("    [cache] %v: erase token(%v)", key, token)

	key = fmt.Sprintf("%s:zones:t:%v:usr:%v:%v", conf.AppName, zid, uid, did)
	stc = hub.rw.Set(key, nil, 0)
	if stc.Err() != nil {
		logrus.Fatalf("PutUserHash() 坏了，redis写不了了: %v", stc)
	}
	// logrus.Debugf("    [cache] %v: erase token to nil", key, token)

	if vxconf.GetBoolR("server.cache.verbose-log", false) {
		logrus.Debugf("    [cache] %v: erase token(%v)", key, token)
	}
}

func inthex(i uint64) string {
	return strconv.FormatUint(i, 16)
}
