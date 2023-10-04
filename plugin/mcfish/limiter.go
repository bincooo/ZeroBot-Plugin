package mcfish

import (
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"sync"
	"time"
)

var (
	// 买卖限流，3秒1次
	saleLimiterManager = rate.NewManager[int64](time.Second*3, 1)
	// 买卖渔获，用户隔离锁
	sMu SaleMutex
)

type SaleMutex struct {
	sync.Mutex
	muM map[int64]*_Mutex
}
type _Mutex struct {
	sync.Mutex     // uid锁
	c          int // 锁次数，归零清除
}

// 玩家上锁
func (s *SaleMutex) SLock(uid int64) {
	s.Lock()
	if s.muM == nil {
		s.muM = make(map[int64]*_Mutex)
	}

	if _, ok := s.muM[uid]; !ok {
		var m _Mutex
		s.muM[uid] = &m
	}
	s.muM[uid].c++
	s.Unlock()

	s.muM[uid].Lock()
}

// 玩家解锁（锁计数归零时删除锁）
func (s *SaleMutex) SUnlock(uid int64) {
	s.Lock()

	mu, ok := s.muM[uid]
	if !ok {
		logrus.Error("买卖渔获的锁丢失了: ", uid)
		s.Unlock()
		return
	}

	mu.c--
	if mu.c <= 0 {
		delete(s.muM, uid)
	}
	s.Unlock()

	mu.Unlock()
}

// 买卖限流，3秒1次
func CustomLimitByUser(ctx *zero.Ctx) *rate.Limiter {
	return saleLimiterManager.Load(ctx.Event.UserID)
}
