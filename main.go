package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	esapi "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "application for personal study of elastic search",
	RunE: func(cmd *cobra.Command, args []string) error {
		modeCrawl, _ := cmd.Flags().GetBool("crawl")
		modeSearch, _ := cmd.Flags().GetBool("search")
		if modeCrawl == modeSearch {
			return errors.New("please specify either --search or --crawl")
		}
		ctx := context.Background()
		client, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		})
		if err != nil {
			return err
		}
		if err := prepare(ctx, client); err != nil {
			return err
		}
		if modeCrawl {
			return crawl(ctx, client)
		}
		q, _ := cmd.Flags().GetString("q")
		keyVal := strings.Split(q, ":")
		if len(keyVal) != 2 {
			return errors.New("Invalid query format. Expected key:value")
		}
		return search(ctx, client, keyVal[0], keyVal[1])
	},
}

func init() {
	rootCmd.Flags().Bool("search", false, "Search mode")
	rootCmd.Flags().Bool("crawl", false, "Crawl mode")
	rootCmd.Flags().String("q", "", "Query for search mode in 'key:value' format")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

const indexName = "posts"

const mapping = `{
        "mappings": {
            "properties": {
                "id": {
                    "type": "integer"
                },
                "title": {
                    "type": "text"
                },
                "body": {
                    "type": "text"
                },
                "japanese_body": {
                    "type": "text",
                    "analyzer": "kuromoji"
                }
            }
        }
    }`

func prepare(ctx context.Context, client *elasticsearch.Client) (err error) {
	res, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, res.Body)
	defer func() {
		if e := res.Body.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()
	if res.StatusCode != http.StatusNotFound {
		log.Println("Index already exists")
		return nil
	}
	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  bytes.NewReader([]byte(mapping)),
	}
	rez, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	if rez.IsError() {
		return fmt.Errorf("Error: %s", rez.String())
	}
	_, _ = io.Copy(io.Discard, res.Body)
	defer func() {
		if e := res.Body.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()
	return nil
}

type Post struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	JapanseBody string `json:"japanese_body"`
}

var japaneseBodies = []string{
	"吾輩は猫である。名前はまだ無い。",
	"どこで生れたかとんと見当がつかぬ。",
	"雨ニモマケズ風ニモマケズ雪ニモ夏ノ暑サニモマケヌ丈夫ナカラダヲモチ",
	"人間は皆、死ぬべき運命にある。 ",
	"おれは人間である、猿ではない。 ",
	"彼の掌に載せられてしまった私は、もう逃げるべくもなくなってしまった。",
	"旅をするということは、ただ単に移動することだけではない。",
	"月がきれいですね。",
	"山路を登りながら、こう考えた。",
	"人間は皆、自分自身の幸せを求める生き物だ。",
}

const (
	crawlURL = "https://jsonplaceholder.typicode.com/posts"
	firstID  = "1"
)

func crawl(ctx context.Context, client *elasticsearch.Client) (err error) {
	req := &esapi.GetRequest{
		Index:      indexName,
		DocumentID: firstID,
	}
	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, res.Body)
	defer func() {
		if e := res.Body.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()
	if res.StatusCode != http.StatusNotFound {
		log.Println("Documents already exists")
		return nil
	}

	var posts []Post
	rez, err := http.Get(crawlURL)
	if err != nil {
		return err
	}
	defer func() {
		if e := rez.Body.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()
	if err := json.NewDecoder(rez.Body).Decode(&posts); err != nil {
		return err
	}

	for i, p := range posts {
		i := i // for < v1.22
		err := func() (err error) {
			ji := i % 10
			p.JapanseBody = japaneseBodies[ji]
			b, err := json.Marshal(p)
			if err != nil {
				return err
			}
			req := esapi.IndexRequest{
				Index:      indexName,
				DocumentID: fmt.Sprintf("%d", p.ID),
				Body:       bytes.NewReader(b),
				Refresh:    "true",
			}
			res, err := req.Do(ctx, client)
			if err != nil {
				return err
			}
			_, _ = io.Copy(io.Discard, res.Body)
			defer func() {
				if e := res.Body.Close(); e != nil {
					err = errors.Join(err, e)
				}
			}()
			if res.IsError() {
				return fmt.Errorf("Error: %s", res.String())
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

type SearchResponse struct {
	Hits struct {
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func search(ctx context.Context, client *elasticsearch.Client, key, val string) (err error) {
	query := fmt.Sprintf(`{"query": {"match": {"%s": "%s"}}}`, key, val)
	res, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(indexName),
		client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return err
	}
	defer func() {
		if e := res.Body.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()
	if res.IsError() {
		return fmt.Errorf("Error: %s", res.String())
	}
	sr := &SearchResponse{}
	err = json.NewDecoder(res.Body).Decode(&sr)
	if err != nil {
		return err
	}
	posts := make([]*Post, len(sr.Hits.Hits))
	for i, hit := range sr.Hits.Hits {
		var p Post
		if err := json.Unmarshal(hit.Source, &p); err != nil {
			return err
		}
		posts[i] = &p
	}
	if err != nil {
		return err
	}
	fmt.Printf("matched %d posts\n", len(posts))
	for _, p := range posts {
		fmt.Printf("%v\n", p.ID)
	}
	return nil
}
