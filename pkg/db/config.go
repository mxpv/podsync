package db

type Config struct {
	// Dir is a directory to keep database files
	Dir    string        `toml:"dir"`
	Badger *BadgerConfig `toml:"badger"`
}
