package pqtgogen

type Component uint64

const (
	ComponentInsert Component = 1 << (64 - 1 - iota)
	ComponentFind
	ComponentUpdate
	ComponentUpsert
	ComponentCount
	ComponentDelete
	ComponentHelpers

	ComponentRepository = ComponentInsert | ComponentFind | ComponentUpdate | ComponentUpsert | ComponentCount | ComponentDelete
	ComponentAll        = ComponentRepository | ComponentHelpers
)
