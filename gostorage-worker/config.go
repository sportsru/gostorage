package main

type StorageCfg struct {
	Address string
	Port    string
}
type MongoCfg struct {
	Address string
	Db      string
}
type MemcacheCfg struct {
	Address   string
	Port      string
	NameSpace string
}
type AppConfig struct {
	Storage  StorageCfg
	Mongo    MongoCfg
	Memcache MemcacheCfg
}

/*
var cfg = AppConfig{
	Storage:  StorageCfg{"0.0.0.0", "9002"},
	Mongo:    MongoCfg{"192.168.1.240", "default"},
	Memcache: MemcacheCfg{"192.168.1.240", "11211", ""},
}
*/
