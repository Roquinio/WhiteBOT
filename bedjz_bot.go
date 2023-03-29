package main

import (
	"encoding/json"
	"fmt"
	mcrcon "github.com/Kelwing/mc-rcon"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	Token     string
	BotPrefix string

	config *configStruct
)

type configStruct struct {
	Token     string `json : "Token"`
	BotPrefix string `json : "BotPrefix"`
}

func ReadConfig() error {
	fmt.Println("Reading config file...")
	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	Token = config.Token
	BotPrefix = config.BotPrefix

	return nil

}

var BotId string
var goBot *discordgo.Session

func start() {
	// create bot session
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// make the bot a user
	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	BotId = u.ID
	goBot.AddHandler(messageHandler)
	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	goBot.Close()

}
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == BotId {
		return
	}
	// Ignore all messages that don't start with the prefix
	if !strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}
	switch {
	case strings.Contains(m.Content, "whitelist"):
		if m.ChannelID != "" {
			s.ChannelMessageSend(m.ChannelID, "Veuillez faire votre demande dans le channel #whitelist")
		} else {
			pseudo := pseudoTreatement(m.Content)
			valid := pseudoVerify(pseudo)
			if valid {
				s.ChannelMessageSendComplex(m.ChannelID, whitelistPlayer(pseudo))
			} else {
				s.ChannelMessageSendComplex(m.ChannelID, rejectPlayer(pseudo))
			}
		}

		// s.ChannelMessageSend(m.ChannelID, pseudo)
	case strings.Contains(m.Content, "admin"):
		if m.ChannelID != "secret" {
			s.ChannelMessageSend(m.ChannelID, "Vous n'êtes pas autorisé à utiliser cette commande")

		} else {
			s.ChannelMessageSend(m.ChannelID, "Vous êtes autorisé à utiliser cette commande")
			cmd := pseudoTreatement(m.Content)
			minecraftServer(cmd)

		}
	}
}

func pseudoTreatement(msg string) string {

	r, _ := regexp.Compile(`\s+([^\s]+)`)
	pseudo1 := r.FindString(msg)

	pseudo := strings.TrimSpace(pseudo1)
	return pseudo
}

func pseudoVerify(name string) bool {

	uri := strings.TrimSpace(name)
	url := "https://api.mojang.com/users/profiles/minecraft/" + uri

	client := http.Client{Timeout: 5 * time.Second}

	response, err := client.Get(url)
	if err != nil {
		return false
	}

	if response.StatusCode == 200 {
		return true
	}

	defer response.Body.Close()

	return false
}

func whitelistPlayer(pseudo string) *discordgo.MessageSend {
	conn := new(mcrcon.MCConn)
	err := conn.Open("localhost:25575", "secret")

	if err != nil {
		fmt.Println(err.Error())
	}
	defer conn.Close()

	err = conn.Authenticate()

	if err != nil {
		fmt.Println(err.Error())
	}
	conn.SendCommand("/whitelist add " + pseudo)
	conn.SendCommand("/say Le joueur " + pseudo + " a été ajouté à la whitelist")

	embed := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Whitelist",
			Description: "Le joueur " + pseudo + " a été ajouté à la whitelist",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Pseudo", Value: pseudo, Inline: true},
				{Name: "Date", Value: time.Now().Format("02/01/2006 15:04:05"), Inline: true},
			},
		},
		},
	}

	return embed
}

func rejectPlayer(pseudo string) *discordgo.MessageSend {

	embed := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Whitelist",
			Description: "Le pseudo " + pseudo + " est incorrect",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Pseudo", Value: pseudo, Inline: true},
				{Name: "Date", Value: time.Now().Format("02/01/2006 15:04:05"), Inline: true},
			},
		},
		},
	}
	return embed

}

func minecraftServer(arg string) error {

	switch {
	case strings.Contains(arg, "start"):
		exec.Command("systemctl", "start", "minecraft.service").Start()
		fmt.Println("Minecraft server started")
		return nil

	case strings.Contains(arg, "stop"):
		exec.Command("systemctl", "stop", "minecraft.service").Start()
		fmt.Println("Minecraft server stopped")
		return nil
	}
	return nil
}

func main() {
	err := ReadConfig()
	if err != nil {
		fmt.Println(err.Error())
	}

	go minecraftServer("start")
	if err != nil {
		fmt.Println(err.Error())
	}
	start()

	return

}
