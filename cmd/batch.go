package cmd

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/lhecker/tumblr-scraper/account"
	"github.com/lhecker/tumblr-scraper/database"
	"github.com/lhecker/tumblr-scraper/scraper"
)

var (
	batchCmd = &cobra.Command{
		Use:   "batch",
		Short: "Scrape multiple blogs at once",
		Run:   batchRun,
	}
)

func init() {
	rootCmd.AddCommand(batchCmd)
}

func batchRun(cmd *cobra.Command, args []string) {
	defer func() {
		err := account.Logout()
		if err != nil {
			log.Printf("failed to logout: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer cancel()

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		defer signal.Stop(ch)
		<-ch
	}()

	u, _ := url.Parse("https://www.tumblr.com")
	cookies := singletons.HTTPClient.Jar.Cookies(u)
	data, err := json.Marshal(cookies)
	_, _ = data, err

	if len(singletons.Config.Username) != 0 {
		account.Setup(singletons.HTTPClient, singletons.Config, singletons.Database)
	}

	s := scraper.NewScraper(singletons.HTTPClient, singletons.Config, singletons.Database)

	for _, blog := range singletons.Config.Blogs {
		highestPostID, err := s.Scrape(ctx, blog)
		if err != nil {
			if !isContextCanceledError(err) {
				log.Println(err)
			}
			return
		}

		err = singletons.Database.SetHighestID(blog.Name, highestPostID)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func recoverCookies(httpClient *http.Client, db *database.Database) {
	u, err := url.Parse("https://www.tumblr.com")
	if err != nil {
		panic(err)
	}

	httpClient.Jar.SetCookies(u, db.GetCookies(database.WwwTumblrComCookiesKey))
}
