# Secret Server
1. **Submitted by: KHOA BUI**
2. **Time spent: 10 hours**

## Introduction
Your task is to implement a secret server. The secret server can be used to store and share secrets using the random generated URL. But the secret can be read only a limited number of times after that it will expire and won’t be available. The secret may have TTL. After the expiration time the secret won’t be available anymore. You can find the detailed API documentation in the swagger.yaml file. We recommend to use [Swagger](https://editor.swagger.io/) or any other OpenAPI implementation to read the documentation.

Here is the [swagger.yaml](swagger.yaml), what describes the Secret Server API

## Task
**Implementation:** You have to implement the whole Secret Server API. You can choose the database you want to use, however it would be wise to store the data using encryption. The response can be XML or JSON too. Use a configuration file to switch between the two response type.

## Requirements
The list of redirection should be maintained in a command line tool, what can:
- [x] Implement the API what listen and server on the endpoints what swagger.yaml describes.

## Bonus
- [ ] Use data encryption for stored data
- [ ] Deploy your server. There are many of free solutions to do this.
- [ ] Monitor number of requests and their response time.
