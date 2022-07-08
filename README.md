# news-feeder

## Description

This service has two components running in separate containers:
1. A REST API based application that allows a mobile client to list articles with filters such as provider and category. Client can also use it to share news articles via Twitter.
2. A worker task that periodically parses the RSS feeds provided and saves it in the DB.


## How It Works
### API:
#### ListArticles
- GET /articles
- retrieves a list of articles
- sample query params:
  ```
  ?categories=uk,technology&providers=bbc
  ```

#### ShareArticle
- POST /article/share
- shares an article via social media e.g. Twitter
- sample JSON request body:
  ```json
  {
    "link": "www.bbc.co.uk",
    "medium": "twitter"
  }
  ```

### Worker:
- Worker has no exposed endpoint but it is doing  work periodically with interval that can be changed in the config file.
- Article links are saved to the DB after deduplicating with GUID link as the unique constraint.

## Local Development
- Dockerfile has been provided to containerize the application and PostgreSQL DB
```shell
go mod tidy
make run-docker
```