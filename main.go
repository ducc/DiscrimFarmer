package main

import (
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

var (
	errorNoUsernames = errors.New("No usernames found for current discrim")
	discrims         []string
	token, password  string
)

func init() {
	discrims = strings.Split(*flag.String("d", "", "discrims comma seperated"), ",")
	flag.StringVar(&token, "t", "", "discord token")
	flag.StringVar(&password, "p", "", "discord password")
	flag.Parse()
}

func isGoodDiscrim(discrim string) bool {
	if discrim[0] == discrim[1] && discrim[1] == discrim[2] && discrim[2] == discrim[3] {
		return true
	}
	if discrim[0] == discrim[1] && discrim[1] == discrim[2] && discrim[2] == 0 {
		return true
	}
	for _, d := range discrims {
		if discrim == d {
			return true
		}
	}
	return false
}

func findUsername(dg *discordgo.Session) (string, error) {
	me, err := dg.User("@me")
	if err != nil {
		return "", err
	}
	for _, guild := range dg.State.Guilds {
		for _, member := range guild.Members {
			if member.User.Discriminator == me.Discriminator && member.User.Username != me.Username {
				return member.User.Username, nil
			}
		}
	}
	return "", errorNoUsernames
}

func populateGuildMembers(dg *discordgo.Session) {
	guilds, err := dg.UserGuilds()
	if err != nil {
		log.WithError(err).Fatal("Error getting user guilds")
		return
	}
	for _, guild := range guilds {
		log.WithFields(log.Fields{
			"name": guild.Name,
			"id":   guild.ID,
		}).Debug("Populating guild members")
		var after string
		for {
			members, err := dg.GuildMembers(guild.ID, after, 1000)
			if err != nil {
				log.WithError(err).Fatal("Error getting guild members")
				return
			}
			if len(members) < 1000 {
				break
			}
			after = members[len(members)-1].User.ID
		}
	}
	log.Info("Loaded guild members")
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("Starting...")
	dg, err := discordgo.New(token)
	if err != nil {
		log.WithError(err).Fatal("Error creating discord session")
		return
	}
	u, err := dg.User("@me")
	if err != nil {
		log.WithError(err).Fatal("Error getting user details")
		return
	}
	err = dg.Open()
	if err != nil {
		log.WithError(err).Fatal("Error opening connection")
		return
	}
	log.WithField("user", fmt.Sprintf("%s#%s (%s)", u.Username, u.Discriminator, u.ID)).Info("Started!")
	populateGuildMembers(dg)
	var first = true
	for {
		if !first {
			time.Sleep(time.Minute * 30)
		} else {
			first = false
		}
		username, err := findUsername(dg)
		if err != nil {
			log.WithError(err).Fatal("Error finding new username")
			return
		}
		u, err := dg.UserUpdate("", password, username, "", "")
		if err != nil {
			log.WithError(err).Warn("Error updating user")
			continue
		}
		log.WithFields(log.Fields{
			"username":      u.Username,
			"discriminator": u.Discriminator,
		}).Info("Updated user")
		if isGoodDiscrim(u.Discriminator) {
			log.Info("Found sweet mf discrim")
			break
		}
	}
	log.Info("Finished discrim farming!")
}
