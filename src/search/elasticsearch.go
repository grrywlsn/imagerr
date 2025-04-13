package search

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"
    "github.com/elastic/go-elasticsearch/v8"
    "github.com/grrywlsn/imagerr/src/db"
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

func getIndexName() string {
    prefix := os.Getenv("ES_INDEX_PREFIX")
    if prefix != "" {
        return prefix + "_images"
    }
    return "images"
}

type SearchResult struct {
    ID               int64     `json:"id"`
    OriginalFilename string    `json:"original_filename"`
    UUIDFilename     string    `json:"uuid_filename"`
    Description      string    `json:"description"`
    URL              string    `json:"url"`
    Tags             []string  `json:"tags"`
    StoragePath      string    `json:"storage_path"`
    CreatedAt        time.Time `json:"created_at"`
    ViewCount        int       `json:"view_count"`
}

func SearchImages(q string, tags string) ([]SearchResult, error) {
    searchQuery := map[string]interface{}{
        "query": map[string]interface{}{
            "bool": map[string]interface{}{},
        },
    }

    // Build query based on provided parameters
    boolQuery := searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})
    
    if q != "" || tags != "" {
        var shouldClauses []map[string]interface{}
        
        if tags != "" {
            shouldClauses = append(shouldClauses, map[string]interface{}{
                "terms": map[string]interface{}{
                    "tags": strings.Split(tags, ","),
                    "boost": 2.0,
                },
            })
        }
        
        if q != "" {
            shouldClauses = append(shouldClauses, map[string]interface{}{
                "match": map[string]interface{}{
                    "description": map[string]interface{}{
                        "query": q,
                        "fuzziness": "AUTO",
                        "boost": 1.0,
                    },
                },
            })
        }
        
        boolQuery["should"] = shouldClauses
        boolQuery["minimum_should_match"] = 1
        
        searchQuery["sort"] = []map[string]interface{}{
            {
                "_score": map[string]interface{}{
                    "order": "desc",
                },
            },
        }
    } else {
        // If no search parameters, return newest images
        searchQuery["sort"] = []map[string]interface{}{
            {
                "id": map[string]interface{}{
                    "order": "desc",
                },
            },
        }
        searchQuery["size"] = 9
    }

    var buf bytes.Buffer
    if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
        return nil, err
    }

    res, err := esClient.Search(
        esClient.Search.WithContext(context.Background()),
        esClient.Search.WithIndex(getIndexName()),
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
            ID:               getInt64Value(source["id"]),
            OriginalFilename: getString(source["original_filename"]),
            UUIDFilename:     getString(source["uuid_filename"]),
            Description:      getString(source["description"]),
            Tags:             interfaceArrayToStringArray(source["tags"].([]interface{})),
            StoragePath:      getString(source["storage_path"]),
            CreatedAt:        time.Unix(0, getInt64Value(source["created_at"])),
            ViewCount:        int(getInt64Value(source["view_count"])),
        })
    }

    return searchResults, nil
}

func interfaceArrayToStringArray(arr []interface{}) []string {
    result := make([]string, len(arr))
    for i, v := range arr {
        result[i] = getString(v)
    }
    return result
}

func getString(v interface{}) string {
    if v == nil {
        return ""
    }
    switch v := v.(type) {
    case string:
        return v
    case float64:
        return fmt.Sprintf("%v", v)
    default:
        return fmt.Sprintf("%v", v)
    }
}

func getInt64Value(v interface{}) int64 {
    if v == nil {
        return 0
    }
    switch v := v.(type) {
    case float64:
        return int64(v)
    case string:
        if val, err := strconv.ParseInt(v, 10, 64); err == nil {
            return val
        }
    case int64:
        return v
    case int:
        return int64(v)
    }
    return 0
}

func IndexImage(image *db.Image) error {
    document := map[string]interface{}{
        "id":               image.ID,
        "original_filename": image.OriginalFilename,
        "uuid_filename":     image.UUIDFilename,
        "description":      image.Description,
        "tags":             image.Tags,
        "storage_path":      image.StoragePath,
        "created_at":       image.CreatedAt,
        "view_count":       image.ViewCount,
    }

    var buf bytes.Buffer
    if err := json.NewEncoder(&buf).Encode(document); err != nil {
        return err
    }

    res, err := esClient.Index(
        getIndexName(),
        &buf,
        esClient.Index.WithDocumentID(fmt.Sprintf("%d", image.ID)),
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

func DeleteIndex() error {
    res, err := esClient.Indices.Delete([]string{getIndexName()})
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error deleting index: %s", res.String())
    }
    return nil
}

func createIndexMapping() error {
    mapping := `{
        "mappings": {
            "properties": {
                "id": { "type": "long" },
                "original_filename": { "type": "keyword" },
                "uuid_filename": { "type": "keyword" },
                "description": { "type": "text", "analyzer": "standard" },
                "url": { "type": "keyword", "index": false },
                "tags": { "type": "keyword" },
                "storage_path": { "type": "keyword" },
                "created_at": { "type": "date" },
                "view_count": { "type": "integer" }
            }
        }
    }`

    res, err := esClient.Indices.Create(
        getIndexName(),
        esClient.Indices.Create.WithBody(strings.NewReader(mapping)),
    )
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error creating index mapping: %s", res.String())
    }
    return nil
}

func ReindexAll(images []db.Image) error {
    // Delete existing index
    if err := DeleteIndex(); err != nil {
        return fmt.Errorf("failed to delete index: %v", err)
    }

    // Create new index with mapping
    if err := createIndexMapping(); err != nil {
        return fmt.Errorf("failed to create index mapping: %v", err)
    }

    // Reindex all images
    for _, image := range images {
        if err := IndexImage(&image); err != nil {
            return fmt.Errorf("failed to index image %d: %v", image.ID, err)
        }
    }
    return nil
}