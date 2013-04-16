package gostorage

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"time"
)

type Client struct {
	Cfg            Config
	MgoSession     *mgo.Session
	Verbose, Debug bool
}

type MongoCfg struct {
	Url string
	Db  string
}

type Config struct {
	Mongo          MongoCfg
	Verbose, Debug bool
}

// store data format for MongoDB
type Storage struct {
	//Id         bson.ObjectId   `bson:"_id,omitempty" json:"-"`
	Uid string
	// Data Mixed
	// Data interface{}
	Data map[string]string
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

	c.Verbose = cfg.Verbose
	c.Debug = cfg.Debug
	return c
}

func (c *Client) SetTags(uid string, fields map[string]interface{}) error {
	session := c.MgoSession.New()
	defer session.Close()

	store := session.DB(c.Cfg.Mongo.Db).C("Storage")

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
		fmt.Println("fields: ")
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

	return nil
}

func (c *Client) GetTagsJSON(uid string) string {
	doc := c.GetDoc(uid)
	// TODO: improve error processing
	if doc == nil {
		return ""
	}

	res, err := json.Marshal(doc.Tags)
	if err != nil {
		panic("json marshal error: " + string(err.Error()))
	}
	if len(res) == 0 {
		return "{}"
	}

	return string(res)
}

func (c *Client) GetDataJSON(uid string) string {
	doc := c.GetDoc(uid)
	// TODO: improve error processing
	if doc == nil {
		return ""
	}

	res, err := json.Marshal(doc.Data)
	if err != nil {
		panic("json marshal error: " + string(err.Error()))
	}
	if len(res) == 0 {
		return "{}"
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
	}

	return strconv.FormatInt(version, 10)
}

// http://denis.papathanasiou.org/2012/10/14/go-golang-and-mongodb-using-mgo/
