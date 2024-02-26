my-es
# What's this
for personal study of elastic search

# RUN
```
docker compose up -d
```

# REST API

## Basic

### Index
Create
```
curl -X PUT "http://localhost:9200/my_index" -H 'Content-Type: application/json' -d '
{
  "settings": {
    "number_of_shards": "3",
    "number_of_replicas": "2"
  }
}'
```

Read
```
curl -X GET "http://localhost:9200/my_index/_settings?pretty"
```

Update
```
curl -X PUT "http://localhost:9200/my_index/_settings" -H 'Content-Type: application/json' -d '
{
  "index": {
    "number_of_replicas": "4"
  }
}'
```
Delete
```
curl -X DELETE "http://localhost:9200/my_index"
```

### Mapping
Create
```
curl -X PUT "http://localhost:9200/my_index/_mapping" -H 'Content-Type: application/json' -d '
{
  "properties": {
    "user_name": { "type": "text" },
    "date": { "type": "date" },
    "message": { "type": "text" }
  }
}'
```


Read
```
curl -X GET "http://localhost:9200/my_index/_mapping?pretty"
```

Update (add field)

```
curl -X PUT "http://localhost:9200/my_index/_mapping" -H 'Content-Type: application/json' -d '
{
  "properties": {
    "additional_comment": { "type": "text" }
  }
}'
```

### Document
Create
```
curl -X POST "http://localhost:9200/my_index/_doc" -H 'Content-Type: application/json' -d '
{
  "user_name": "John Smith",
  "date": "2024-01-01T00:00:00",
  "message": "ElasticSearch",
  "additional_comment": "Hello"
}'
```

Read
```
curl -X GET "http://localhost:9200/my_index/_doc/QmE4z40BC8fj59Bmi7OS?pretty"
```

UPDATE
```
curl -X POST "http://localhost:9200/my_index/_update/QmE4z40BC8fj59Bmi7OS" -H 'Content-Type: application/json' -d '
{
  "doc": {
    "user_name": "Pat Smith"
  }
}'
```

Delete
```
curl -X DELETE "http://localhost:9200/my_index/_doc/QmE4z40BC8fj59Bmi7OS"
```

Search
```
curl -X POST "http://localhost:9200/my_index/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "message": "ElasticSearch"
    }
  }
}'
```
## Advanced

### Search in Japanese

Adding a field containing kuromoji analyzer specifications
```
curl -X PUT "http://localhost:9200/my_index/_mapping" -H 'Content-Type: application/json' -d '
{
  "properties": {
    "japanese_comment": {
      "type": "text",
      "analyzer": "kuromoji"
    }
  }
}'
```

Create document
```
curl -X POST "http://localhost:9200/my_index/_doc" -H 'Content-Type: application/json' -d '
{
  "user_name": "John woo",
  "date": "2024-01-01T00:00:00",
  "message": "message",
  "additional_comment": "comment",
  "japanese_comment": "こんにちは世界"
}'
```

Get Analyze information
```
curl -X POST "http://localhost:9200/my_index/_analyze?pretty" -H 'Content-Type: application/json' -d '
{
  "analyzer": "kuromoji",
  "text": "こんにちは世界"
}'
```

Search in Japanese
```
curl -X POST "http://localhost:9200/my_index/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "japanese_comment": "こんにちは"
    }
  }
}'
```

### Synonym Search

Create Index
```
curl -X PUT "http://localhost:9200/my_synonym_index" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "index": {
      "analysis": {
        "analyzer": {
          "my_synonyms": {
            "tokenizer": "whitespace",
            "filter": [ "synonym" ]
          }
        },
        "filter": {
          "synonym": {
            "type": "synonym_graph",
            "synonyms_path": "/usr/share/elasticsearch/config/analysis/synonym.txt",
            "updateable": true
          }
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "text": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "my_synonyms"
      }
    }
  }
}'
```

Create Document
```
curl -X POST "http://localhost:9200/my_synonym_index/_doc" -H 'Content-Type: application/json' -d '
{
  "text": "プロダクトマネージャー"
}'
```

Search by Synonym
```
curl -X POST "http://localhost:9200/my_synonym_index/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "text": "PdM"
    }
  }
}'
```

# App
## Crawler
```
go run main.go --crawl
```

## Searcher
```
go run main.go --search --q=japanese_comment:猫
```
