# Go Simple Rest Sample
Simple todo list api

# How to use
## clone and run

```
$ git clone https://github.com/dai65527/go_simpleREST.git
$ cd go_simpleREST/srcs
$ go mod tidy
$ go run main.go
```

Or use docker.

```
$ git clone https://github.com/dai65527/go_simpleREST.git
$ cd go_simpleREST
$ docker build -t go_simpleREST .
$ docker run --rm -d -p 4000:4000 go_simpleREST
```

## End Points

|End Point|Request Method|Effect|
| ---- | ---- | ---- |
|`/items`|GET|Get all items as json|
|`/items`|POST|Add new item|
|`/items`|DELETE|Delete items|
|`/items/:id`|DELETE|Delete one item|
|`/items/:id/done`|PUT|Mark a task as done|
|`/items/:id/done`|DELETE|Mark a task as undone|

## Example
By curl.

```sh
# add 5 new items
$ curl -d '{"name":"item1"}' localhost:4000/items
$ curl -d '{"name":"item2"}' localhost:4000/items
$ curl -d '{"name":"item3"}' localhost:4000/items
$ curl -d '{"name":"item4"}' localhost:4000/items  
$ curl -d '{"name":"item5"}' localhost:4000/items

# get all items
$ curl localhost:4000/items
[{"id":1,"name":"item1","done":false},{"id":2,"name":"item2","done":false},{"id":3,"name":"item3","done":false},{"id":4,"name":"item4","done":false},{"id":5,"name":"item5","done":false}]

# done items2, 3 and 4
$ curl -X PUT localhost:4000/items/2/done
$ curl -X PUT localhost:4000/items/3/done
$ curl -X PUT localhost:4000/items/4/done
$ curl localhost:4000/items              
[{"id":1,"name":"item1","done":false},{"id":2,"name":"item2","done":true},{"id":3,"name":"item3","done":true},{"id":4,"name":"item4","done":true},{"id":5,"name":"item5","done":false}]

# undone item3
$ curl -X DELETE localhost:4000/items/3/done
$ curl localhost:4000/items                 
[{"id":1,"name":"item1","done":false},{"id":2,"name":"item2","done":true},{"id":3,"name":"item3","done":false},{"id":4,"name":"item4","done":true},{"id":5,"name":"item5","done":false}]

# delete all done item
$ curl -X DELETE localhost:4000/items/ 
$ curl localhost:4000/items/          
[{"id":1,"name":"item1","done":false},{"id":3,"name":"item3","done":false},{"id":5,"name":"item5","done":false}]

# delete a item
$ curl -X DELETE localhost:4000/items/3
$ curl localhost:4000/items         
[{"id":1,"name":"item1","done":false},{"id":5,"name":"item5","done":false}]
```

Or, you can access this via browser using https://github.com/dai65527/tstodo-client.

![browser](/sampleimages/todoAPI.gif)
