package application

import "github.com/flyluman/scratch/internal/ports"

type ItemModule struct {
	Create *CreateItemUseCase
	Get    *GetItemUseCase
	List   *ListItemUseCase
	Update *UpdateItemUseCase
	Delete *DeleteItemUseCase
}

func NewItemModule(repo ports.ItemRepository, idGen ports.IDGenerator, clock ports.Clock, bus ports.EventBus, flags ports.FeatureFlags) *ItemModule {
	return &ItemModule{
		Create: NewCreateItemUseCase(repo, idGen, clock, bus),
		Get:    NewGetItemUseCase(repo),
		List:   NewListItemUseCase(repo),
		Update: NewUpdateItemUseCase(repo, clock, bus),
		Delete: NewDeleteItemUseCase(repo, flags, bus),
	}
}
