package randomizer

type Randomizer interface {
	String() (string, error)
}

func NewStringRandomizer(cfg StringRandomizerConfig) Randomizer {
	return &randomizer{
		size:          cfg.Size,
		characterPool: cfg.CharacterPool,
	}
}
