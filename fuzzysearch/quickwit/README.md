# Start Indexer

`docker-compose up`

# Create Index

Named `homewatch`

```
curl -XPOST \
    http://127.0.0.1:7280/api/v1/indexes \
    --header "content-type: application/yaml" \
    --data-binary @./homewatch.index.yml
```


# Add Data

```
curl -XPOST \
    "http://127.0.0.1:7280/api/v1/homewatch/ingest?commit=force" \
    --data-binary @./import.json
```

# ElasticSearch the Data

```
curl -XPOST \
    localhost:7280/api/v1/_elastic/homewatch/_search \
    --data '{"query": { "match": { "MESSAGE": "listening"} } }'
```

# Console

Commands can be run against the console as well:

```
docker run \
    -it \
    --network=$Network \
    --rm \
    -v $(pwd):/data \
    --name quickwit_console \
    quickwit/quickwit:latest $command
```    

An example environment `homewatch` can be used when executing commands against the console
```
Network=host
Index=homewatch
IndexConfig="/input/homewatch.index.yml"
```

## Delete an Index
```
deleteCommand="quickwit/quickwit:latest index delete --index $Index"
```
## Create an Index
```
createCommand="index create --index-config $IndexConfig"
```

## Ingest JSONL immediately to an Index
```
ingestCommand="index ingest --index homewatch --input-path /data:/data/$InputJsonl --force"
```