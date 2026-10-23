// Package config хранит зафиксированные параметры сервиса.
// Значения меняются только здесь — агент не должен придумывать свои.
package config

// Profile — стратегия чанкинга и скоринга.
type Profile string

const (
	// Naive — фиксированные чанки без заголовков и чистый косинус. Это «до».
	Naive Profile = "naive"
	// Fixed — чанки по разделам + гибридный скор. Это «после».
	Fixed Profile = "fixed"
)

// Config — единая точка правды для путей, модели и весов.
type Config struct {
	DocsDir       string
	DBPath        string
	OllamaURL     string
	EmbedModel    string
	Dims          int
	K             int     // сколько соседей берём из индекса
	Top           int     // сколько результатов отдаём
	LexicalWeight float64 // вес лексической части в гибридном скоре
	MaxChunkRunes int
	Profile       Profile
}

// Default возвращает конфигурацию для демо.
func Default() Config {
	return Config{
		DocsDir:       "data/docs",
		DBPath:        "kb.sqlite",
		OllamaURL:     "http://localhost:11434",
		EmbedModel:    "bge-m3",
		Dims:          1024,
		K:             12,
		Top:           5,
		LexicalWeight: 0.2,
		MaxChunkRunes: 700,
		Profile:       Fixed,
	}
}

// HeadingAware сообщает, резать ли документ по заголовкам.
func (c Config) HeadingAware() bool { return c.Profile == Fixed }

// Hybrid сообщает, применять ли гибридный скор.
func (c Config) Hybrid() bool { return c.Profile == Fixed }

// Dedupe сообщает, схлопывать ли выдачу по документу.
// Первая версия агента этого не делает — в топ-5 попадают пять кусков одного файла.
func (c Config) Dedupe() bool { return c.Profile == Fixed }
