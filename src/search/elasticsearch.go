package search

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "github.com/elastic/go-elasticsearch/v8"
)

var esClient *elasticsearch.Client

func InitElasticsearch() {
    cfg := elasticsearch.Config{
        Addresses: []string{os.Getenv("ES_URL")},
        Username:  os.Getenv("ES_USER"),
        Password:  os.Getenv("ES_PASSWORD"),
    }

    var err error
    esClient, err = elasticsearch.NewClient(cfg)
    if err != nil {
        log.Fatal("Error creating Elasticsearch client:", err)
    }
}

type SearchResult struct {
    ID          int64    `json:"id"`
    Description string   `json:"description"`
    Tags        []string `json:"tags"`
}

func SearchImages(query string) ([]SearchResult, error) {
    searchQuery := map[string]interface{}{
        "query": map[string]interface{}{
            "multi_match": map[string]interface{}{
                "query":  query,
                "fields": []string{"description", "tags"},
                "type":   "best_fields",
            },
        },
    }

    var buf bytes.Buffer
    if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
        return nil, err
    }

    res, err := esClient.Search(
        esClient.Search.WithContext(context.Background()),
        esClient.Search.WithIndex("images"),
        esClient.Search.WithBody(&buf),
    )
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return nil, err
    }

    var searchResults []SearchResult
    hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
    
    for _, hit := range hits {
        source := hit.(map[string]interface{})["_source"].(map[string]interface{})
        searchResults = append(searchResults, SearchResult{
            ID:          int64(source["id"].(float64)),
            Description: source["description"].(string),
            Tags:        interfaceArrayToStringArray(source["tags"].([]interface{})),
        })
    }

    return searchResults, nil
}

func interfaceArrayToStringArray(arr []interface{}) []string {
    result := make([]string, len(arr))
    for i, v := range arr {
        result[i] = v.(string)
    }
    return result
}

func IndexImage(id int64, description string, tags []string) error {
    document := map[string]interface{}{
        "id":          id,
        "description": description,
        "tags":        tags,
    }

    var buf bytes.Buffer
    if err := json.NewEncoder(&buf).Encode(document); err != nil {
        return err
    }

    res, err := esClient.Index(
        "images",
        &buf,
        esClient.Index.WithDocumentID(fmt.Sprintf("%d", id)),
        esClient.Index.WithContext(context.Background()),
    )
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error indexing document: %s", res.String())
    }

    return nil
}