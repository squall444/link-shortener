package stat

import (
	"context"
	"goadv/pkg/event"
	"log"
)

type StatServiceDeps struct {
	StatRepository *StatRepository
	EventBus       *event.EventBus
}

type StatService struct {
	StatRepository *StatRepository
	EventBus       *event.EventBus
}

func NewStatService(deps *StatServiceDeps) *StatService {
	return &StatService{
		StatRepository: deps.StatRepository,
		EventBus:       deps.EventBus,
	}
}

func (s *StatService) AddClick(ctx context.Context) {
	for {
		select {
		case msg := <-s.EventBus.Subscribe():
			if msg.Type == event.EventLinkVisited {
				id, ok := msg.Data.(uint)
				if !ok {
					log.Fatalln("Bad EventLinkVisited Data:", msg.Data)
					continue
				}
				s.StatRepository.AddClick(id)
			}
		case <-ctx.Done():
			log.Println("StatService stopped")
			return
		}
	}
}
