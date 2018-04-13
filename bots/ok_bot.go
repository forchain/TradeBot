package bots

import (
	"github.com/forchain/cryptotrader/okex"
	"context"
	"time"
	"github.com/forchain/cryptotrader/model"
)

type OkBot struct {
	Bot

	client_ *okex.Client
}

func (_b *OkBot) Init() {
	_b.Bot.Init("okex", 0.9985)
	_b.client_ = okex.New("868f9fbd-5857-4019-9ce9-282c207e0cc2", "DB0EF2FC1C09D56FD190CACC879975A6")
}

func (_b *OkBot) GetRecords(_base, _quote string) (records []model.Record, err error) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	records, err = _b.client_.GetRecords(context, _base, _quote, "1hour", 0, 0)
	defer cancel()
	return
}
