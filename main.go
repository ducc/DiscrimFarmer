package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"strings"
	"time"
)

type apiResponse struct {
	Responses []struct {
		Username string `json:"username"`
	} `json:"response"`
}

var (
	errorNoUsernames                 = errors.New("No usernames found for current discrim")
	discrims                         []string
	defaultUsername, token, password string
	api                              bool
)

func init() {
	discrims = strings.Split(*flag.String("d", "", "discrims comma seperated"), ",")
	flag.StringVar(&defaultUsername, "u", "", "discord username")
	flag.StringVar(&token, "t", "", "discord token")
	flag.StringVar(&password, "p", "", "discord password")
	flag.BoolVar(&api, "api", false, "use api.pandentia.cf")
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

func findUsernameWithAPI(dg *discordgo.Session) (string, error) {
	me, err := dg.User("@me")
	if err != nil {
		return "", err
	}
	resp, err := http.Get("https://api.pandentia.cf/discord/users/discriminator/" + me.Discriminator)
	if err != nil {
		return "", err
	}
	var ar apiResponse
	err = json.NewDecoder(resp.Body).Decode(&ar)
	if err != nil {
		return "", err
	}
	if len(ar.Responses) == 0 {
		return "", errors.New("aaaa no discrims WTF!!")
	}
	return ar.Responses[0].Username, nil
}

func populateGuildMembers(dg *discordgo.Session) {
	guilds, err := dg.UserGuilds(100, "", "")
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
	if !api {
		populateGuildMembers(dg)
	}
	var first = true
	for {
		if !first {
			if len(defaultUsername) > 0 {
				time.Sleep(time.Minute * 60)
			} else {
				time.Sleep(time.Minute * 30)
			}
		} else {
			first = false
		}
		var username string
		if api {
			username, err = findUsernameWithAPI(dg)
			if err != nil {
				log.WithError(err).Fatal("Error finding new username")
				return
			}
		} else {
			username, err = findUsername(dg)
			if err != nil {
				log.WithError(err).Fatal("Error finding new username")
				return
			}
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
		if len(defaultUsername) > 0 {
			_, err := dg.UserUpdate("", password, defaultUsername, "", "")
			if err != nil {
				log.WithError(err).Warn("Error updating user")
				continue
			}
		}
	}
	log.Info("Finished discrim farming!")
}
