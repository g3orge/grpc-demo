package cache

import (
	"errors"
	"log"
	"sync"

	"github.com/g3orge/grpc-demo/inv"
)

type mCache struct {
	sync.RWMutex
	// expAt   time.Duration
	// cleanAt time.Duration
	users map[string]inv.User
}

// type mUser struct {
// 	inv.User
// 	Created    time.Time
// 	Expiration int64
// }

func New() *mCache {
	users := make(map[string]inv.User)

	cache := mCache{
		users: users,
	}

	return &cache
}

func (c *mCache) GetAll() map[string]inv.User {
	c.RLock()
	defer c.RUnlock()

	return c.users
}

func (c *mCache) Set(key string, req inv.User) {
	// var exp int64

	// if dur == 0 {
	// 	dur = c.expAt
	// }

	// if dur > 0 {
	// 	exp = time.Now().Add(dur).UnixNano()
	// }

	c.Lock()
	defer c.Unlock()

	c.users[key] = inv.User{
		Id:       req.Id,
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Admin:    req.Admin,
	}
	log.Println("new user set")
}

func (c *mCache) Get(key string) (*inv.User, bool) {
	c.RLock()
	defer c.RUnlock()

	users, err := c.users[key]
	if !err {
		return &inv.User{}, false
	}

	// if users.Expiration > 0 {
	// 	if time.Now().UnixNano() > users.Expiration {
	// 		return nil, false
	// 	}
	// }

	return &users, true
}

func (c *mCache) GetById(key string) (*inv.User, bool) {
	c.RLock()
	defer c.RUnlock()

	// users, err := c.users[key]
	// if !err {
	// 	return &inv.User{}, false

	// var keys []string
	var users inv.User
	for k, v := range c.users {
		if v.Id == key {
			users, _ = c.users[k]
		}
	}

	return &users, true
}

func (c *mCache) GetByName(key string) (*inv.User, bool) {
	c.RLock()
	defer c.RUnlock()

	// users, err := c.users[key]
	// if !err {
	// 	return &inv.User{}, false

	// var keys []string
	var users inv.User
	for k, v := range c.users {
		if v.Username == key {
			users, _ = c.users[k]
		}
	}

	return &users, true
}

func (c *mCache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	if _, v := c.users[key]; !v {
		return errors.New("key not found")
	}

	delete(c.users, key)

	return nil
}

func (c *mCache) StartGC() {
	go c.GC()
}

func (c *mCache) GC() {
	// for {
	// 	<-time.After(c.cleanAt)

	// 	if c.users == nil {
	// 		return
	// 	}

	// 	if k := c.expKeys(); len(k) != 0 {
	// 		c.cleanKeys(k)
	// 	}
	// }
	if c.users == nil {
		return
	}

	var keys []string
	for k, _ := range c.users {
		keys = append(keys, k)
	}

	c.cleanKeys(keys)
}

// func (c *mCache) expKeys() (keys []string) {
// 	c.RLock()
// 	defer c.RUnlock()

// 	for k, i := range c.users {
// 		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
// 			keys = append(keys, k)
// 		}
// 	}

// 	return
// }

func (c *mCache) cleanKeys(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		delete(c.users, k)
	}
}
