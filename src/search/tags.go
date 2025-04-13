package search

import (
    "context"
    "encoding/json"
    "bytes"
)

func SuggestTags(query string) ([]string, error) {
    searchQuery := map[string]interface{}{
        "size": 0,
        "aggs": map[string]interface{}{
            "tag_suggestions": map[string]interface{}{
                "terms": map[string]interface{}{
                    "field": "tags",
                    "include": query + ".*",
                    "size": 10,
                },
            },
        },
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

    aggregations := result["aggregations"].(map[string]interface{})
    tagSuggestions := aggregations["tag_suggestions"].(map[string]interface{})
    buckets := tagSuggestions["buckets"].([]interface{})

    var tags []string
    for _, bucket := range buckets {
        bucketMap := bucket.(map[string]interface{})
        tags = append(tags, bucketMap["key"].(string))
    }

    return tags, nil
}