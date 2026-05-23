package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

var cfg = elasticsearch.Config{
	Addresses: []string{
		"https://4a216f9feb56400fbc5c280865be1062.psc.asia-southeast2.gcp.elastic-cloud.com",
	},
	Username: "elastic",
	Password: "KiMFCYvuOMCcmdEzpU2G7xZ7",
}

func main() {

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	res, err := es.Ping(es.Ping.WithContext(context.Background()))
	if err != nil {
		log.Fatalf("failed to ping elasticsearch: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("elasticsearch ping returned error: %s", res.Status())
	}

	fmt.Printf("elasticsearch ping successful: %s\n", res.Status())

	info, err := es.Info()
	if err != nil {
		log.Fatalf("failed to get elasticsearch info: %v", err)
	}
	defer info.Body.Close()

	fmt.Printf("elasticsearch info: %s\n", info.String())

	// if err := createIndex(es); err != nil {
	// 	log.Fatalf("failed to create index: %v", err)
	// }

	// if err := insertDocument(es); err != nil {
	// 	log.Fatalf("failed to insert document: %v", err)
	// }

	// if err := listDocuments(es); err != nil {
	// 	log.Fatalf("failed to list documents: %v", err)
	// }

	fmt.Println(deleteIndex(es))
	// listIndices(es)
}

type Ad struct {
	AdID  string `json:"ad_id"`
	Title string `json:"title"`
}

var dummyAds = []Ad{
	{AdID: "ad-001", Title: "Buy cheap shoes online"},
	{AdID: "ad-002", Title: "Best laptop deals this week"},
	{AdID: "ad-003", Title: "Fresh groceries delivered to your door"},
}

func insertDocument(es *elasticsearch.Client) error {
	const indexName = "ads_testing"

	for _, ad := range dummyAds {
		body, err := json.Marshal(ad)
		if err != nil {
			return fmt.Errorf("failed to marshal ad %s: %w", ad.AdID, err)
		}

		res, err := es.Index(
			indexName,
			strings.NewReader(string(body)),
			es.Index.WithDocumentID(ad.AdID),
			es.Index.WithContext(context.Background()),
		)
		if err != nil {
			return fmt.Errorf("request error for ad %s: %w", ad.AdID, err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("insert failed for ad %s: %s", ad.AdID, res.String())
		}

		fmt.Printf("inserted document %q: %s\n", ad.AdID, res.Status())
	}

	return nil
}

func listDocuments(es *elasticsearch.Client) error {
	const indexName = "astro_ads_index_v1"
	const query = `{ "query": { "match_all": {} } }`

	res, err := es.Search(
		es.Search.WithIndex(indexName),
		es.Search.WithBody(strings.NewReader(query)),
		es.Search.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("search failed: %s", res.String())
	}

	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Println(string(pretty))

	return nil
}

func listIndices(es *elasticsearch.Client) error {
	res, err := es.Cat.Indices(
		es.Cat.Indices.WithContext(context.Background()),
		es.Cat.Indices.WithH("index", "docs.count", "store.size", "health"),
		es.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("list indices failed: %s", res.String())
	}

	var indices []struct {
		Index     string `json:"index"`
		DocsCount string `json:"docs.count"`
		StoreSize string `json:"store.size"`
		Health    string `json:"health"`
	}

	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("%-40s %10s %10s %s\n", "index", "docs", "size", "health")
	fmt.Printf("%-40s %10s %10s %s\n", "-----", "----", "----", "------")
	for _, idx := range indices {
		fmt.Printf("%-40s %10s %10s %s\n", idx.Index, idx.DocsCount, idx.StoreSize, idx.Health)
	}

	return nil
}

func deleteIndex(es *elasticsearch.Client) error {
	const indexName = "astro_ads_index_v1"

	res, err := es.Indices.Delete(
		[]string{indexName},
		es.Indices.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index deletion failed: %s", res.String())
	}

	fmt.Printf("index %q deleted: %s\n", indexName, res.Status())
	return nil
}

func createIndex(es *elasticsearch.Client) error {
	const indexName = "ads_testing"
	const mapping = `{
		"mappings": {
			"properties": {
				"ad_id": { "type": "keyword" },
				"title": { "type": "text" }
			}
		}
	}`

	res, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithBody(strings.NewReader(mapping)),
		es.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index creation failed: %s", res.String())
	}

	fmt.Printf("index %q created: %s\n", indexName, res.Status())
	return nil
}
