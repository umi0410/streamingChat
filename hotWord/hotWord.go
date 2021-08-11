package hotWord

import (
	"context"
	"fmt"
	"github.com/umi0410/streamingChat/adapter"
	"sort"
	"time"
)

var (
	ttl int = 10
)

type HotWordCalculator struct {
	Ctx            context.Context
	messageAdapter adapter.MessageAdapter
}

func NewHotWordCalculator(ctx context.Context, messageAdapter adapter.MessageAdapter) *HotWordCalculator {
	return &HotWordCalculator{
		Ctx:            ctx,
		messageAdapter: messageAdapter,
	}
}

func (c *HotWordCalculator) Run() error {
	for {
		select {
		case <-c.Ctx.Done():
			return nil
		default:
			time.Sleep(time.Second * 3)
			ranking := c.messageAdapter.GetWordFrequency(c.Ctx)
			sort.Slice(ranking, func(i, j int) bool {
				return ranking[i].Frequency > ranking[j].Frequency
			})
			fmt.Println("====================== HOT WORDS RANKING ======================")
			bound := 10
			if len(ranking) < 10 {
				bound = len(ranking)
			}
			for i, hotWord := range ranking[:bound] {
				fmt.Printf("%-4dth | %-10s | %d\n", i, hotWord.Word, hotWord.Frequency)
			}
		}
	}
}

//func GetRanking() {
//    adapter.MessageAdapter
//}
