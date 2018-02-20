### Overview
A simple container based rabbit consumer written in go for copying, moving and deleting local directories.

### Usage (1 line)
``` bash
#/bin/bash
docker run -e RABBIT_USER=user \
-e RABBIT_PASSWORD=pwd \
-e RABBIT_HOST=rabbit.mydomain.com \
-e RABBIT_PORT=5672 \
-e RABBIT_QUEUE_FILE_OPS=my-queue \
-e WAIT_TIME=1 \
--name local-file-worker jackytck/rabbit-go-file-docker:v0.0.1
```

### Usage (with .env)
``` text
# .env
RABBIT_USER=user
RABBIT_PASSWORD=pwd
RABBIT_HOST=rabbit.mydomain.com
RABBIT_PORT=5672
RABBIT_QUEUE_FILE_OPS=my-queue
WAIT_TIME=1
```

``` bash
#!/bin/bash
docker run --env-file .env --name local-file-worker jackytck/rabbit-go-file-docker:v0.0.1
```
