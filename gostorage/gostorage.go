package gostorage

import (
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/davecgh/go-spew/spew"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"time"
)

type Client struct {
	Cfg            Config
	MgoSession     *mgo.Session
	MemClient      *memcache.Client
	Verbose, Debug bool
}

type MongoCfg struct {
	Url string
	Db  string
}

type MemcacheCfg struct {
	Servers   []string
	NameSpace string
}

type Config struct {
	Mongo          MongoCfg
	Memcache       MemcacheCfg
	Verbose, Debug bool
}

// store data format for MongoDB
type Storage struct {
	//Id         bson.ObjectId   `bson:"_id,omitempty" json:"-"`
	Uid string
	// Data Mixed
	// Data interface{}
	Data map[string]interface{}
	Tags map[string]int32
	// Tags Mixed
	Version    int64
	Last_visit int64 // timestamp
}

/*
func init() {
	_ = mgo.ErrNotFound
	_ = bson.MaxKey
	_ = spew.Config
}
*/

func New(cfg Config) *Client {
	c := &Client{Cfg: cfg}

	// test Mongo connection
	session, err := mgo.Dial(cfg.Mongo.Url)
	if err != nil {
		panic(err)
	}
	c.MgoSession = session

	c.MemClient = memcache.New(cfg.Memcache.Servers...)
	if cfg.Debug {
		fmt.Print("memcache servers: ")
		spew.Dump(cfg.Memcache.Servers)
	}

	c.Verbose = cfg.Verbose
	c.Debug = cfg.Debug
	return c
}

func (c *Client) SetTags(uid string, fields map[string]interface{}) error {
	session := c.MgoSession.New()
	defer session.Close()

	store := session.DB(c.Cfg.Mongo.Db).C("Storage")

	fields["version"] = 0
	timestamp := int64(time.Now().Unix() / 1000)
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{"last_visit": timestamp},
			"$inc": bson.M(fields),
		},
		ReturnNew: true,
		Upsert:    true,
	}
	resultUps := Storage{}
	info, err := store.Find(bson.M{"uid": uid}).Apply(change, &resultUps)

	if c.Debug {
		fmt.Print("set tags info: ")
		spew.Dump(info)
		fmt.Print("resultUps: ")
		spew.Dump(resultUps)
	}

	if err != nil {
		panic(err)
	}

	return nil
}

func (c *Client) SetData(uid string, fields map[string]interface{}) error {
	session := c.MgoSession.New()
	defer session.Close()

	store := session.DB(c.Cfg.Mongo.Db).C("Storage")

	if c.Debug {
		fmt.Println("SetData")
		fmt.Print("fields: ")
		spew.Dump(fields)
	}

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M(fields),
			"$inc": bson.M{"version": 1},
		},
		ReturnNew: true,
		Upsert:    true,
	}

	resultUps := Storage{}
	info, err := store.Find(bson.M{"uid": uid}).Apply(change, &resultUps)

	if c.Debug {
		fmt.Print("info: ")
		spew.Dump(info)
		fmt.Print("resultUps: ")
		spew.Dump(resultUps)
	}

	if err != nil {
		panic(err)
	}

	// XXX: not sure here 
	c.Uncache(uid, resultUps.Version)
	//c.Uncache(uid, resultUps.Version+1) // <-- race condition here ?

	return nil
}

func (c *Client) GetTagsJSON(uid string) string {
	doc := c.GetDoc(uid)
	// TODO: improve error processing
	if doc == nil {
		return ""
	}
	if len(doc.Tags) == 0 {
		return "{}"
	}

	res, err := json.Marshal(doc.Tags)
	if err != nil {
		panic("json marshal error: " + string(err.Error()))
	}

	return string(res)
}

func (c *Client) GetDataJSON(uid string) string {
	doc := c.GetDoc(uid)
	// TODO: improve error processing
	if doc == nil {
		return ""
	}

	if len(doc.Data) == 0 {
		return "{}"
	}

	res, err := json.Marshal(doc.Data)
	if err != nil {
		panic("json marshal error: " + string(err.Error()))
	}
	return string(res)
}

func (c *Client) GetDoc(uid string) *Storage {
	session := c.MgoSession.New()
	defer session.Close()

	store := session.DB(c.Cfg.Mongo.Db).C("Storage")
	resultDoc := Storage{}
	err := store.Find(bson.M{"uid": uid}).One(&resultDoc)

	if c.Debug {
		fmt.Print("resultDoc: ")
		spew.Dump(resultDoc)
	}

	// TODO: add better error processing
	if err == mgo.ErrNotFound {
		return nil
	}
	if err != nil {
		panic("mgo err: " + string(err.Error()))
	}

	return &resultDoc
}

func (c *Client) GetVersion(uid string) string {
	doc := c.GetDoc(uid)

	if c.Debug {
		fmt.Print("resultDoc: ")
		spew.Dump(doc)
	}

	version := int64(-1)
	if doc != nil {
		version = doc.Version
		// XXX: nit sure here
		c.Uncache(uid, version)
	}

	return strconv.FormatInt(version, 10)
}

func (c *Client) Uncache(uid string, version int64) {
	key1 := c.Cfg.Memcache.NameSpace + uid
	//value := []byte(string(version))
	value := []byte("{version: " + strconv.FormatInt(version, 10) + "}")
	// TODO: use const for Expiration value
	setItem := &memcache.Item{
		Key:        key1,
		Value:      value,
		Expiration: 2 * 24 * 60 * 60,
	}
	if c.Debug {
		spew.Dump(setItem)
	}
	setErr := c.MemClient.Set(setItem)

	// XXX: gentle Set error processing
	if setErr != nil {
		panic("set cache failed: " + string(setErr.Error()))
	}

	// XXX: gentle Cas error processing
	// ErrCacheMiss
	item, getErr := c.MemClient.Get(key1)
	// check: Is cas_id here?
	if c.Debug {
		spew.Dump(item)
	}

	if getErr != nil {
		panic("get cache failed: " + string(getErr.Error()))
	}
	casErr := c.MemClient.CompareAndSwap(item)
	// TODO: check ErrCASConflict
	_ = casErr
	//spew.Dump(casErr)

	//it, err := c.Mem.Get(key1)
}

// http://denis.papathanasiou.org/2012/10/14/go-golang-and-mongodb-using-mgo/
