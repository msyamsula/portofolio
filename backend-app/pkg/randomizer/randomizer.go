package randomizer

type StringRandomizer struct {
	size          int
	characterPool string
}

func NewStringRandomizer(cfg StringRandomizerConfig) *StringRandomizer {
	return &StringRandomizer{
		size:          cfg.Size,
		characterPool: cfg.CharacterPool,
	}
}
