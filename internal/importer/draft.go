package importer

type Card map[string]string

type Draft struct {
	Profile string `yaml:"profile"`
	Cards   []Card `yaml:"cards"`
}
