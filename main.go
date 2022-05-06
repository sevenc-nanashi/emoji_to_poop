package main

import (
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"github.com/urakozz/go-emoji"
)

func main() {

	godotenv.Load()

	config := oauth1.NewConfig(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	token := oauth1.NewToken(os.Getenv("ACCESS_KEY"), os.Getenv("ACCESS_SECRET"))

	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	userId := strings.Split(os.Getenv("ACCESS_KEY"), "-")[0]

	start(client, userId)
}

func getLastId() int64 {
	if _, err := os.Stat("last_id.txt"); os.IsNotExist(err) {
		return 0
	}
	file, err := ioutil.ReadFile("last_id.txt")
	if err != nil {
		return 0
	}
	data, err := strconv.Atoi(string(file))
	if err != nil {
		return 0
	}
	return int64(data)
}

func setLastId(id int64) {
	ioutil.WriteFile("last_id.txt", []byte(strconv.Itoa(int(id))), 0644)
}

func start(client *twitter.Client, userId string) {
	for {
		println("Starting...")
		lastId := getLastId()
		println("Last id: " + strconv.Itoa(int(lastId)))
		tweets, _, err := client.Timelines.MentionTimeline(&twitter.MentionTimelineParams{
			Count:   200,
			SinceID: lastId,
		})
		if err != nil {
			panic(err)
		}
		if len(tweets) > 0 {
			sort.Slice(tweets, func(i, j int) bool { return tweets[i].ID > tweets[j].ID })
			setLastId(tweets[0].ID)
			for _, tweet := range tweets {
				processTweet(client, tweet, userId)
			}
		} else {
			println("No new tweets")
		}
		time.Sleep(time.Second * 15)
	}
}

func processTweet(client *twitter.Client, tweet twitter.Tweet, userId string) {
	println("Received tweet: " + tweet.Text)
	if tweet.User.IDStr == userId {
		println("Ignoring own tweet")
		return
	}
	println("URL: " + "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IDStr)
	println(tweet.InReplyToStatusID)
	if tweet.InReplyToStatusID == 0 {
		replyTweet, _, err := client.Statuses.Update(
			"@"+tweet.User.ScreenName+"\n"+
				"Hello! Sorry, but this bot only responds to tweets that are replies to other tweets.\n"+
				"Please mention me in a reply!",
			&twitter.StatusUpdateParams{
				InReplyToStatusID: tweet.ID,
			},
		)
		if err != nil {
			println("An error occurred: " + err.Error())
		} else {
			println("Replied help, https://twitter.com/" + replyTweet.User.ScreenName + "/status/" + replyTweet.IDStr)
		}
		return
	}
	replyTweets, _, err := client.Statuses.Lookup([]int64{tweet.InReplyToStatusID}, nil)
	if err != nil || len(replyTweets) == 0 {
		return
	}
	replyTweet := replyTweets[0]
	text := replyTweet.Text
	parser := emoji.NewEmojiParser()

	replaced := parser.ReplaceAllString(text, "💩")
	replaceReplyTweet, _, err := client.Statuses.Update(
		"@"+replyTweet.User.ScreenName+"\n"+replaced,
		&twitter.StatusUpdateParams{
			InReplyToStatusID: tweet.ID,
		},
	)
	if err != nil {
		println("An error occurred: " + err.Error())
	} else {
		println("Replied with replaced text, https://twitter.com/" + replaceReplyTweet.User.ScreenName + "/status/" + replaceReplyTweet.IDStr)
	}
}
