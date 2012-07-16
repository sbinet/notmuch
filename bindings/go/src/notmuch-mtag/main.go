// notmuch-mtag tags a bunch of messages, a la 'notmuch tag'
package main

import (
	// stdlib imports
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	// 3rd-party imports
	"notmuch"
	"github.com/kless/goconfig/config"
	)

type Tag struct {
	Cmd string
	Query string
}

func main() {
	var cfg *config.Config
	var err error

	// honor NOTMUCH_CONFIG
	home := os.Getenv("NOTMUCH_CONFIG")
	if home == "" {
		home = os.Getenv("HOME")
	}

	if cfg, err = config.ReadDefault(path.Join(home, ".notmuch-config")); err != nil {
		log.Fatalf("error loading config file:", err)
	}

	db_path, err := cfg.String("database", "path")
	if err != nil {
		log.Fatalf("no field 'path' in section 'database'")
	}
	db_path = "/home/binet/mail-notmuch-tst"

	tag_fname, err := cfg.String("tag", "script")
	if err != nil {
		log.Printf("no field 'script' in section 'tag'")
		tag_fname = "/home/binet/emacs/notmuch-lib/tags.json" 
	}

	fmt.Printf("db_path:   [%s]\n", db_path)
	fmt.Printf("tag_fname: [%s]\n", tag_fname)

	// open the database
	db, status := notmuch.OpenDatabase(db_path,
		notmuch.DATABASE_MODE_READ_WRITE)
	if status != notmuch.STATUS_SUCCESS {
		log.Fatalf("Failed to open the database: %v\n", status)
	}

	// open the tag commands
	tagfile, err := os.Open(tag_fname)
	if err != nil {
		log.Fatalf("Failed to open the tag-commands file: %v\n", err)
	}

	tagcmds := []Tag{}
	{
		dec := json.NewDecoder(tagfile)
		if dec == nil {
			log.Fatalf("Failed to create a new json-decoder\n")
		}
		err = dec.Decode(&tagcmds)
		if dec == nil {
			log.Fatalf("Failed to decode the tag-commands file: %v\n", err)
		}
	}
	for i,tagcmd := range tagcmds {
		fmt.Printf("tagcmd[%d] = %v\n", i, tagcmd)
	}

	new_msg_tag_fmt := "(tag:new AND tag:inbox AND NOT tag:%s) AND (%s)"
	for _,tag := range tagcmds {
		tag_cmd := tag.Cmd
		if len(tag_cmd) > 1 && tag_cmd[0] == '+' {
			tag_cmd = tag_cmd[1:]
		} else {
			log.Printf("invalid tag command: [%s]\n", tag_cmd)
			continue
		}
		query_str := fmt.Sprintf(
			new_msg_tag_fmt,
			tag_cmd,
			tag.Query,
			)
		fmt.Printf(">> [%s]\n", query_str)
		// pass 1: look at all new messages in the inbox
		query := db.CreateQuery(query_str)
		msgs := query.SearchMessages()
		fmt.Printf("got: [%v] new messages...\n", len(msgs))
		for _,msg := range msgs {
			fmt.Printf("==> %s\n", msg.GetMessageId())
			if msg.Freeze() != notmuch.STATUS_SUCCESS {
				log.Printf("could not freeze message [%v]\n", msg.GetMessageId())
			}
			tags := msg.GetTags()
			if msg.AddTag(tag_cmd) != notmuch.STATUS_SUCCESS {
				fmt.Printf("**errorr**\n")
				continue
			}
			tags = msg.GetTags()
			fmt.Printf("==> tags: %v\n", tags)
			fmt.Printf("==> subject: '%v'\n", msg.GetHeader("subject"))
			fmt.Printf("==> to:   '%v'\n", msg.GetHeader("to"))
			fmt.Printf("==> from: '%v'\n", msg.GetHeader("from"))
			if msg.Thaw() != notmuch.STATUS_SUCCESS {
				log.Printf("could not freeze message [%v]\n", msg.GetMessageId())
			}
		}
	} //> tag cmds
	return
}
