# gomodel [![wercker status](https://app.wercker.com/status/9c6ef0eec7d6d217bd831bbdc3a3ace2/s "wercker status")](https://app.wercker.com/project/bykey/9c6ef0eec7d6d217bd831bbdc3a3ace2) [![GoDoc](https://godoc.org/github.com/cosiner/gomodel?status.png)](http://godoc.org/github.com/cosiner/gomodel)
gomodel provide another method to interact with database.   
Instead of reflection, use bitset represent fields of CRUD, sql/stmt cache and generate model code for you, high performance.

# Install
```sh
$ go get github.com/cosiner/gomodel
$ cd /path/of/gomodel/cmd/gomodel
$ go install # it will install the gomodel binary file to your $GOPATH/bin
$ gomodel -cp # copy model.tmpl to default path $HOME/.config/go/model.tmpl
              # or just put it to your model package, gomodel will search it first 
```

[SQL convertion for structures](https://github.com/cosiner/gomodel/tree/master/cmd/gomodel).

# Example
```Go
type User struct {
    Id int `column:"user_id"`
    Age int
    Name string
}

$ gomodel -i user.go -m User -o user_gen.go
// You will get blow constants and other functions, if need UserId rather 
// than USER_ID, add -cc option for gomodel to enable CamelCase
const (
    USER_ID uint = 1 << iota
    USER_AGE
    USER_NAME
    userFieldsEnd = iota
    userFieldsAll = 1 << userFieldsEnd - 1
)
```
* __DB__
```Go
db := gomodel.NewDB()
```
* __Insert__
```Go
u := &User{Age:1, Name:"abcde"}
db.Insert(u, USER_AGE|USER_NAME, gomodel.RES_ID) // get last inserted id
```

* __Delete__
```Go
u := &User{Id:1, Age:20}
db.Delete(u, USER_ID|USER_AGE)
```

* __Update__
```Go
u := &User{Id:1, Age:5, Name:"abcde"}
db.Update(u, USER_AGE|USER_NAME, USER_ID) // update age by id
```

* __One__
```Go
u := &User{Id:1}
userFieldsExcpId := userFieldsAll & (^USER_ID)
db.One(u, userFieldsExcpId, USER_ID) // select one by id
```

* __Limit__
```Go
u := &User{Age:10}
users := &users{Fields:userFieldsAll} // users is generated by gomodel
db.Limit(users, u, userFieldsAll, USER_AGE, 0, 10)
return users.Values // []User
```

* __All__
```Go
u := &User{Age:10}
users := &users{Fields:userFieldsAll} // users is generated by gomodel
db.All(users, u, userFieldsAll, USER_AGE)
return users.Values // []User
```

# LICENSE
MIT.
