// notmuch-mtag tags a bunch of messages, a la 'notmuch tag'
package main

import (
	// stdlib imports
	"encoding/json"
	"flag"
	"fmt"
	stdlog "log"
	"os"
	"path"
	"strings"

	// 3rd-party imports
	"github.com/kless/goconfig/config"
	"notmuch"
)

type Tag struct {
	Cmd   string
	Query string
}

func build_tag_cmd(cmd string) (cmds []string, err error) {
	cmds = strings.Split(cmd, " ")
	if len(cmds) <= 0 {
		err = fmt.Errorf("invalid tag-cmd [%s]", cmd)
		return nil, err
	}
	for idx,c := range cmds {
		if len(c) <= 1 {
			err = fmt.Errorf("invalid tag-cmd at index #%d: [%s]", idx, c)
			return nil, err
		}
		switch c[0] {
		case '+', '-':
			// ok.
		default:
			err = fmt.Errorf("invalid tag-cmd at index #%d: [%s]", idx, c)
			return nil, err
		}
	}
	return cmds, err
}
var g_verbose = flag.Bool("verbose", false, "enable verbose output")

func main() {
	var cfg *config.Config
	var err error

	log := stdlog.New(
		os.Stderr,
		"[notmuch-mtag] ",
		stdlog.Flags())

	flag.Parse()
	log.Printf(":: notmuch-mtag...\n")

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

	tag_fname, err := cfg.String("notmuch-mtag", "script")
	if err != nil {
		log.Fatalf("no field 'script' in section 'notmuch-mtag'")
	}

	log.Printf("verbose:   [%v]\n", *g_verbose)
	log.Printf("db_path:   [%s]\n", db_path)
	log.Printf("tag_fname: [%s]\n", tag_fname)

	// open the database
	db, status := notmuch.OpenDatabase(
		db_path,
		notmuch.DATABASE_MODE_READ_WRITE)
	if status != notmuch.STATUS_SUCCESS {
		log.Fatalf("Failed to open the database: %v\n", status)
	}
	defer db.Close()

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

	{
		// remove phony messages (before 1980)
		query := db.CreateQuery("tag:new AND tag:inbox AND ..-0")
		msgs := query.SearchMessages()
		for _, msg := range msgs {
			for _,tag := range []string{"new", "inbox", "unread"} {
				if msg.Freeze() != notmuch.STATUS_SUCCESS {
					log.Printf("could not freeze message [%v]\n", msg.GetMessageId())
				}
				if msg.RemoveTag(tag) != notmuch.STATUS_SUCCESS {
					log.Printf("could not apply tag -%s to message id=%v\n", 
						tag,
						msg.GetMessageId())
				}
				if msg.Thaw() != notmuch.STATUS_SUCCESS {
					log.Printf("could not freeze message [%v]\n", msg.GetMessageId())
				}
			}
		}
	}

	new_msg_tag_fmt := "(tag:new AND tag:inbox AND NOT (%s)) AND (%s)"
	for _, tag := range tagcmds {
		tag_cmds, err := build_tag_cmd(tag.Cmd)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		tag_cmd_str := make([]string, len(tag_cmds))
		for i,tag_cmd_i := range tag_cmds {
			tag_cmd_str[i] = "tag:"+tag_cmd_i[1:]
		}
		query_str := fmt.Sprintf(
			new_msg_tag_fmt,
			strings.Join(tag_cmd_str, " AND "),
			tag.Query,
		)

		// look at all new messages in the inbox
		query := db.CreateQuery(query_str)
		msgs := query.SearchMessages()
		if len(msgs) > 0 {
			log.Printf(">> [%s]\n", query_str)
			log.Printf(">> got: [%v] new messages...\n", len(msgs))
		}
		for _, msg := range msgs {
			if *g_verbose {
				log.Printf("==> %s\n", msg.GetMessageId())
			}
			if msg.Freeze() != notmuch.STATUS_SUCCESS {
				log.Printf("could not freeze message [%v]\n", msg.GetMessageId())
			}
			tags := msg.GetTags()
			for _,tag_cmd := range tag_cmds {
				switch tag_cmd[0] {
					case '+':
					if msg.AddTag(tag_cmd[1:]) != notmuch.STATUS_SUCCESS {
						log.Printf("**error**\n")
						continue
					}
					case '-':
					if msg.RemoveTag(tag_cmd[1:]) != notmuch.STATUS_SUCCESS {
						log.Printf("**error**\n")
						continue
					}
				}
			}
			if *g_verbose {
				tags = msg.GetTags()
				log.Printf("==> tags: %v\n", tags)
			}
			if msg.Thaw() != notmuch.STATUS_SUCCESS {
				log.Printf("could not freeze message [%v]\n", msg.GetMessageId())
			}
		}
	} //> tag cmds

	// remove the 'new' tag
	query := db.CreateQuery("tag:new AND tag:inbox")
	msgs := query.SearchMessages()
	log.Printf(">> applying tag -new...\n")
	log.Printf(">> got: [%v] new messages...\n", len(msgs))
	for _, msg := range msgs {
		if msg.RemoveTag("new") != notmuch.STATUS_SUCCESS {
			log.Printf("could not 'un-new' message id=%v\n", msg.GetMessageId())
		}
	}
	log.Printf(">> applying tag -new... [done]\n")
	log.Printf(":: notmuch-mtag...[done]\n")
	return
}
